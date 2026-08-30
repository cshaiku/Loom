package loom

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type runtimeOptions struct {
	quiet   bool
	verbose bool
}

func Run(args []string, stdout io.Writer, stderr io.Writer) error {
	runtime, args := parseRuntime(args)
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(stdout, "Loom: SwiftUI to WinUI Layout Compiler\n\n")
		fmt.Fprint(stdout, catalogText(""))
		return nil
	}
	if args[0] == "version" || args[0] == "--version" {
		fmt.Fprintf(stdout, "loom %s\n", Version)
		return nil
	}
	if args[0] == "list" || args[0] == "commands" {
		category, jsonOut, err := parseList(args[1:])
		if err != nil {
			return err
		}
		if jsonOut {
			out := Commands
			if category != "" {
				out = nil
				for _, c := range Commands {
					if c.Category == category {
						out = append(out, c)
					}
				}
			}
			text, err := prettyJSON(out)
			if err != nil {
				return err
			}
			fmt.Fprint(stdout, text)
			return nil
		}
		text := catalogText(category)
		if text == "" {
			return fmt.Errorf("unknown or empty command category %s", category)
		}
		fmt.Fprint(stdout, text)
		return nil
	}
	if args[0] == "man" || args[0] == "explain" {
		if len(args) != 2 {
			return fmt.Errorf("unknown command manual request")
		}
		text := manual(args[1])
		if text == "" {
			return fmt.Errorf("unknown command manual request")
		}
		fmt.Fprint(stdout, text)
		return nil
	}
	command, ok := resolveCommand(args[0])
	if !ok {
		return fmt.Errorf("unknown command %s", args[0])
	}
	if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
		fmt.Fprint(stdout, manual(command.Command))
		return nil
	}
	switch command.Command {
	case "status":
		return runStatus(args[1:], stdout)
	case "verify":
		return runVerify(args[1:], stdout)
	case "patterns:list", "patterns:show", "patterns:validate", "patterns:lint", "patterns:export":
		return runPattern(command.Command, args[1:], stdout, stderr, runtime)
	case "inspect:xaml":
		return runInspectXAML(args[1:], stdout, stderr, runtime)
	case "inspect:ascii":
		return runASCII(args[1:], stdout, stderr, runtime)
	case "accessibility:audit":
		return runAudit(args[1:], stdout, stderr, runtime)
	case "patterns:transfer":
		return runTransfer(args[1:], stdout, stderr, runtime)
	case "suggestions:os-errors":
		return runSuggestions(args[1:], stdout, stderr, runtime)
	default:
		return fmt.Errorf("command %s is not yet implemented in the Go port", command.Command)
	}
}

func parseRuntime(args []string) (runtimeOptions, []string) {
	var runtime runtimeOptions
	var remaining []string
	for _, arg := range args {
		switch arg {
		case "--quiet", "-q":
			runtime.quiet = true
		case "--verbose", "-v":
			runtime.verbose = true
		default:
			remaining = append(remaining, arg)
		}
	}
	if runtime.quiet {
		runtime.verbose = false
	}
	return runtime, remaining
}

func writeOrPrint(text, output string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	if output == "" {
		fmt.Fprint(stdout, text)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(output, []byte(text), 0644); err != nil {
		return err
	}
	if runtime.verbose {
		fmt.Fprintf(stderr, "[info] wrote %s (%d bytes)\n", output, len(text))
	}
	if !runtime.quiet {
		fmt.Fprintf(stdout, "Wrote %s\n", output)
	}
	return nil
}

func runPattern(command string, args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	directory := "Patterns"
	format := "text"
	exportFormat := "loom"
	output := ""
	var positional []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			format = "json"
		case "--directory":
			i++
			if i >= len(args) {
				return fmt.Errorf("missing value for --directory")
			}
			directory = args[i]
		case "--format":
			i++
			if i >= len(args) {
				return fmt.Errorf("missing value for --format")
			}
			if command == "patterns:export" {
				exportFormat = args[i]
			} else {
				format = args[i]
			}
		case "--output":
			i++
			if i >= len(args) {
				return fmt.Errorf("missing value for --output")
			}
			output = args[i]
		default:
			positional = append(positional, args[i])
		}
	}
	patterns, err := LoadPatterns(directory)
	if err != nil && command != "patterns:validate" && command != "patterns:lint" {
		return err
	}
	switch command {
	case "patterns:list":
		if format == "json" {
			text, err := prettyJSON(patterns)
			if err != nil {
				return err
			}
			return writeOrPrint(text, output, stdout, stderr, runtime)
		}
		return writeOrPrint(PatternListText(patterns), output, stdout, stderr, runtime)
	case "patterns:show":
		if len(positional) != 1 {
			return fmt.Errorf("patterns:show requires one pattern id")
		}
		pattern, ok := FindPattern(patterns, positional[0])
		if !ok {
			return fmt.Errorf("no pattern named %s in %s", positional[0], directory)
		}
		text, err := prettyJSON(pattern)
		if err != nil {
			return err
		}
		return writeOrPrint(text, output, stdout, stderr, runtime)
	case "patterns:validate":
		report := ValidatePatterns(directory)
		text := PatternReportText(report)
		if format == "json" {
			text, _ = prettyJSON(report)
		}
		if err := writeOrPrint(text, output, stdout, stderr, runtime); err != nil {
			return err
		}
		if report.Status != "ok" {
			return ErrCommandFailed
		}
	case "patterns:lint":
		report := LintPatterns(directory)
		text := PatternReportText(report)
		if format == "json" {
			text, _ = prettyJSON(report)
		}
		if err := writeOrPrint(text, output, stdout, stderr, runtime); err != nil {
			return err
		}
		if report.Status != "ok" {
			return ErrCommandFailed
		}
	case "patterns:export":
		text, err := ExportPatterns(patterns, exportFormat)
		if err != nil {
			return err
		}
		return writeOrPrint(text, output, stdout, stderr, runtime)
	}
	return nil
}

func PatternReportText(report PatternValidationReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Loom pattern validation\nStatus: %s\nDirectory: %s\nPatterns: %d\nIssues: %d\n", report.Status, report.Directory, report.PatternCount, len(report.Issues))
	for _, issue := range report.Issues {
		fmt.Fprintf(&b, "  [%s] %s %s: %s\n", issue.Severity, issue.Code, issue.Path, issue.Detail)
	}
	return b.String()
}

func runInspectXAML(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	path, format, output, err := sourceArgs(args)
	if err != nil {
		return err
	}
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		return err
	}
	if format == "json" {
		text, err := prettyJSON(analysis)
		if err != nil {
			return err
		}
		return writeOrPrint(text, output, stdout, stderr, runtime)
	}
	return writeOrPrint(AnalysisText(analysis), output, stdout, stderr, runtime)
}

func runASCII(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	path, _, output, err := sourceArgs(args)
	if err != nil {
		return err
	}
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		return err
	}
	return writeOrPrint(ASCIIAnalysis(analysis), output, stdout, stderr, runtime)
}

func runAudit(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	path, format, output, err := sourceArgs(args)
	if err != nil {
		return err
	}
	failOn := flagValue(args, "--fail-on")
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		return err
	}
	report := Audit(analysis)
	text := AuditText(report)
	if format == "json" {
		text, _ = prettyJSON(report)
	}
	if err := writeOrPrint(text, output, stdout, stderr, runtime); err != nil {
		return err
	}
	if failOn == "error" && report.Summary.Errors > 0 {
		return ErrCommandFailed
	}
	if failOn == "warning" && (report.Summary.Errors > 0 || report.Summary.Warnings > 0) {
		return ErrCommandFailed
	}
	return nil
}

func runTransfer(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	path, format, output, err := sourceArgs(args)
	if err != nil {
		return err
	}
	patternDir := firstNonEmpty(flagValue(args, "--patterns-dir"), "Patterns")
	from := firstNonEmpty(flagValue(args, "--from"), "winui3")
	to := firstNonEmpty(flagValue(args, "--to"), "swiftui")
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		return err
	}
	patterns, err := LoadPatterns(patternDir)
	if err != nil {
		return err
	}
	report := Transfer(analysis, patterns, from, to)
	text := TransferText(report)
	if format == "json" {
		text, _ = prettyJSON(report)
	}
	return writeOrPrint(text, output, stdout, stderr, runtime)
}

func runSuggestions(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	platform := flagValue(args, "--platform")
	query := firstNonEmpty(flagValue(args, "--message"), flagValue(args, "--query"))
	format := firstNonEmpty(flagValue(args, "--format"), "text")
	output := flagValue(args, "--output")
	report := OSErrorSuggestions(platform, query)
	text := ""
	var err error
	if format == "json" {
		text, err = prettyJSON(report)
	} else {
		text = SuggestionsText(report)
	}
	if err != nil {
		return err
	}
	return writeOrPrint(text, output, stdout, stderr, runtime)
}

func SuggestionsText(report OSErrorSuggestionReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Loom OS error suggestions\nStatus: %s\nPlatform: %s\nQuery: %s\nSuggestions: %d\n\n", report.Status, firstNonEmpty(string(report.Platform), "all"), firstNonEmpty(report.Query, "none"), len(report.Suggestions))
	for _, suggestion := range report.Suggestions {
		fmt.Fprintf(&b, "[%s] %s: %s\n  issue: %s\n  reference: %s\n", suggestion.Platform, suggestion.Category, suggestion.Matcher, suggestion.Issue, suggestion.Reference)
		for _, fix := range suggestion.SuggestedFixes {
			fmt.Fprintf(&b, "  - %s: %s — %s\n", fix.Audience, fix.Action, fix.Detail)
		}
	}
	return b.String()
}

func runStatus(args []string, stdout io.Writer) error {
	dir := firstNonEmpty(flagValue(args, "--patterns-dir"), "Patterns")
	report := ValidatePatterns(dir)
	if contains(args, "--json") {
		text, _ := prettyJSON(map[string]any{"schema_version": "1", "version": Version, "commands": len(Commands), "patternStatus": report.Status, "patternCount": report.PatternCount})
		fmt.Fprint(stdout, text)
		return nil
	}
	fmt.Fprintf(stdout, "Loom status\nVersion: %s\nCommands: %d\nPatterns: %s (%d)\n", Version, len(Commands), report.Status, report.PatternCount)
	return nil
}

func runVerify(args []string, stdout io.Writer) error {
	dir := firstNonEmpty(flagValue(args, "--patterns-dir"), "Patterns")
	patterns := ValidatePatterns(dir)
	lint := LintPatterns(dir)
	status := "ok"
	if patterns.Status != "ok" || lint.Status != "ok" {
		status = "error"
	}
	report := map[string]any{"schema_version": "1", "status": status, "commands": len(Commands), "patterns": patterns, "patternLint": lint}
	if contains(args, "--json") {
		text, _ := prettyJSON(report)
		fmt.Fprint(stdout, text)
	} else {
		fmt.Fprintf(stdout, "Loom verify\nStatus: %s\nPatterns: %s\nPattern lint: %s\n", status, patterns.Status, lint.Status)
	}
	if status != "ok" {
		return ErrCommandFailed
	}
	return nil
}

func AnalysisText(analysis Analysis) string {
	return fmt.Sprintf("Loom analysis\nSource: %s\nView: %s.%s\nSource nodes: %d\nLayout nodes: %d\n\n%s", analysis.SourcePath, analysis.RootView, analysis.Component, analysis.SyntaxNodeCount, analysis.Layout.RecursiveNodeCount(), ASCIIAnalysis(analysis))
}

func sourceArgs(args []string) (path, format, output string, err error) {
	if len(args) == 0 {
		return "", "", "", fmt.Errorf("source path required")
	}
	path = args[0]
	format = firstNonEmpty(flagValue(args, "--format"), "text")
	if contains(args, "--json") {
		format = "json"
	}
	output = flagValue(args, "--output")
	return path, format, output, nil
}

func parseList(args []string) (string, bool, error) {
	category := ""
	jsonOut := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOut = true
		case "--category":
			i++
			if i >= len(args) {
				return "", false, fmt.Errorf("missing value for --category")
			}
			category = args[i]
		default:
			return "", false, fmt.Errorf("unknown list option %s", args[i])
		}
	}
	return category, jsonOut, nil
}

func flagValue(args []string, flag string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			return args[i+1]
		}
	}
	return ""
}

func contains(args []string, value string) bool {
	for _, arg := range args {
		if arg == value {
			return true
		}
	}
	return false
}
