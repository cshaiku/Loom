package loom

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
)

type runtimeOptions struct {
	quiet      bool
	verbose    bool
	lineEnding string
}

func Run(args []string, stdout io.Writer, stderr io.Writer) error {
	runtime, args, err := parseRuntime(args)
	if err != nil {
		return err
	}
	if len(args) > 1 && args[0] == "help" {
		text := manual(args[1])
		if text == "" {
			return fmt.Errorf("unknown command manual request")
		}
		return writeText(text, stdout, runtime)
	}
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		return writeText("loom: cross-platform interface layout analysis CLI\n\n"+catalogText(""), stdout, runtime)
	}
	if args[0] == "version" || args[0] == "--version" {
		return writeText(fmt.Sprintf("loom %s\n", Version), stdout, runtime)
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
			return writeText(text, stdout, runtime)
		}
		text := catalogText(category)
		if text == "" {
			return fmt.Errorf("unknown or empty command category %s", category)
		}
		return writeText(text, stdout, runtime)
	}
	if args[0] == "man" || args[0] == "explain" {
		if len(args) != 2 {
			return fmt.Errorf("unknown command manual request")
		}
		text := manual(args[1])
		if text == "" {
			return fmt.Errorf("unknown command manual request")
		}
		return writeText(text, stdout, runtime)
	}
	command, ok := resolveCommand(args[0])
	if !ok {
		return fmt.Errorf("unknown command %q; run `loom help` to see available commands or `loom list --json` for agent-readable command metadata", args[0])
	}
	if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
		return writeText(manual(command.Command), stdout, runtime)
	}
	switch command.Command {
	case "status":
		return runStatus(args[1:], stdout, stderr, runtime)
	case "verify":
		return runVerify(args[1:], stdout, runtime)
	case "checks:command-catalog":
		return runChecksCommandCatalog(args[1:], stdout, stderr, runtime)
	case "config:validate":
		return runConfigValidate(args[1:], stdout, stderr, runtime)
	case "config:schema":
		return runConfigSchema(args[1:], stdout, runtime)
	case "guards:summary":
		return runGuardsSummary(args[1:], stdout, runtime)
	case "self-heal:plan":
		return runSelfHealPlan(args[1:], stdout, runtime)
	case "patterns:list", "patterns:show", "patterns:validate", "patterns:lint", "patterns:export":
		return runPattern(command.Command, args[1:], stdout, stderr, runtime)
	case "inspect:xaml":
		return runInspectXAML(args[1:], stdout, stderr, runtime)
	case "inspect:swiftui":
		return runInspectSwiftUI(args[1:], stdout, stderr, runtime)
	case "inspect:qt":
		return runInspectQt(args[1:], stdout, stderr, runtime)
	case "inspect:source":
		return runInspectSource(args[1:], stdout, stderr, runtime)
	case "inspect:ascii":
		return runASCII(args[1:], stdout, stderr, runtime)
	case "inspect:errors":
		return runInspectErrors(args[1:], stdout, stderr, runtime)
	case "inspect:font":
		return runInspectFont(args[1:], stdout, stderr, runtime)
	case "inspect:parity":
		return runInspectParity(args[1:], stdout, stderr, runtime)
	case "inspect:visual-parity":
		return runInspectVisualParity(args[1:], stdout, stderr, runtime)
	case "graph:components":
		return runGraphComponents(args[1:], stdout, stderr, runtime)
	case "project:build":
		return runProjectBuild(args[1:], stdout, stderr, runtime)
	case "generate:xaml", "generate:swiftui", "generate:contracts":
		return runUnavailableCommand(command.Command, stdout, stderr, runtime)
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

func parseRuntime(args []string) (runtimeOptions, []string, error) {
	runtime := runtimeOptions{lineEnding: "lf"}
	var remaining []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--quiet", "-q":
			runtime.quiet = true
		case "--verbose", "-v":
			runtime.verbose = true
		case "--line-ending":
			i++
			if i >= len(args) {
				return runtimeOptions{}, nil, fmt.Errorf("missing value for --line-ending")
			}
			value := strings.ToLower(strings.TrimSpace(args[i]))
			switch value {
			case "lf", "crlf", "native":
				runtime.lineEnding = value
			default:
				return runtimeOptions{}, nil, fmt.Errorf("unsupported line ending %s", args[i])
			}
		default:
			remaining = append(remaining, arg)
		}
	}
	if runtime.quiet {
		runtime.verbose = false
	}
	return runtime, remaining, nil
}

func writeOrPrint(text, output string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	return writeOrPrintChecked(text, output, "", false, stdout, stderr, runtime)
}

func writeOrPrintChecked(text, output, input string, overwrite bool, stdout, stderr io.Writer, runtime runtimeOptions) error {
	text = applyLineEnding(text, runtime.lineEnding)
	if output == "" {
		_, err := fmt.Fprint(stdout, text)
		return err
	}
	outputAbs, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	if input != "" {
		inputAbs, err := filepath.Abs(input)
		if err != nil {
			return err
		}
		if samePath(inputAbs, outputAbs) {
			return fmt.Errorf("refusing to write output over input path %s", output)
		}
	}
	if !overwrite {
		if _, err := os.Stat(outputAbs); err == nil {
			return fmt.Errorf("refusing to overwrite existing output %s without --overwrite", output)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(outputAbs), 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(outputAbs), "."+filepath.Base(outputAbs)+".tmp-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if _, err := temp.WriteString(text); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempName, outputAbs); err != nil {
		return err
	}
	if runtime.verbose {
		if _, err := fmt.Fprint(stderr, applyLineEnding(fmt.Sprintf("[info] wrote %s (%d bytes)\n", output, len(text)), runtime.lineEnding)); err != nil {
			return err
		}
	}
	if !runtime.quiet {
		return writeText(fmt.Sprintf("Wrote %s\n", output), stdout, runtime)
	}
	return nil
}

func samePath(a, b string) bool {
	aEval, aErr := filepath.EvalSymlinks(a)
	bEval, bErr := filepath.EvalSymlinks(b)
	if aErr == nil {
		a = aEval
	}
	if bErr == nil {
		b = bEval
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func writeText(text string, stdout io.Writer, runtime runtimeOptions) error {
	_, err := fmt.Fprint(stdout, applyLineEnding(text, runtime.lineEnding))
	return err
}

func applyLineEnding(text, mode string) string {
	lineEnding := "\n"
	switch mode {
	case "crlf":
		lineEnding = "\r\n"
	case "native":
		if goruntime.GOOS == "windows" {
			lineEnding = "\r\n"
		}
	case "", "lf":
		lineEnding = "\n"
	default:
		lineEnding = "\n"
	}
	if lineEnding == "\n" {
		return strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	}
	normalized := strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	return strings.ReplaceAll(normalized, "\n", lineEnding)
}

func runPattern(command string, args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	directory := DefaultPatternDirectory
	format := "text"
	exportFormat := "loom"
	output := ""
	parsed, err := parseArgs(args, map[string]bool{"--directory": true, "--format": true, "--output": true}, map[string]bool{"--json": true, "--overwrite": true})
	if err != nil {
		return err
	}
	positional := parsed.Positionals
	if parsed.Bools["--json"] {
		format = "json"
	}
	if parsed.Values["--directory"] != "" {
		directory = parsed.Values["--directory"]
	}
	if parsed.Values["--format"] != "" {
		if command == "patterns:export" {
			exportFormat = parsed.Values["--format"]
		} else {
			format = parsed.Values["--format"]
		}
	}
	if format != "text" && format != "json" {
		return fmt.Errorf("--format must be text or json")
	}
	output = parsed.Values["--output"]
	if command != "patterns:show" && len(positional) > 0 {
		if len(positional) > 1 {
			return fmt.Errorf("%s accepts at most one positional directory", command)
		}
		directory = positional[0]
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
			return writeOrPrintChecked(text, output, "", parsed.Bools["--overwrite"], stdout, stderr, runtime)
		}
		return writeOrPrintChecked(PatternListText(patterns), output, "", parsed.Bools["--overwrite"], stdout, stderr, runtime)
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
		return writeOrPrintChecked(text, output, "", parsed.Bools["--overwrite"], stdout, stderr, runtime)
	case "patterns:validate":
		report := ValidatePatterns(directory)
		text := PatternReportText(report)
		if format == "json" {
			text, _ = prettyJSON(report)
		}
		if err := writeOrPrintChecked(text, output, "", parsed.Bools["--overwrite"], stdout, stderr, runtime); err != nil {
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
		if err := writeOrPrintChecked(text, output, "", parsed.Bools["--overwrite"], stdout, stderr, runtime); err != nil {
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
		return writeOrPrintChecked(text, output, "", parsed.Bools["--overwrite"], stdout, stderr, runtime)
	}
	return nil
}

func PatternReportText(report PatternValidationReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "loom pattern validation\nstatus: %s\ndirectory: %s\npatterns: %d\nissues: %d\n", report.Status, report.Directory, report.PatternCount, len(report.Issues))
	for _, issue := range report.Issues {
		fmt.Fprintf(&b, "  [%s] %s %s: %s\n", issue.Severity, issue.Code, issue.Path, issue.Detail)
	}
	return b.String()
}

func runInspectXAML(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	path, format, output, overwrite, err := sourceArgs(args)
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
		return writeOrPrintChecked(text, output, path, overwrite, stdout, stderr, runtime)
	}
	return writeOrPrintChecked(AnalysisText(analysis), output, path, overwrite, stdout, stderr, runtime)
}

func runInspectSwiftUI(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	path, format, output, overwrite, err := sourceArgs(args)
	if err != nil {
		return err
	}
	analysis, err := AnalyzeSwiftUI(path)
	if err != nil {
		return err
	}
	if format == "json" {
		text, err := prettyJSON(analysis)
		if err != nil {
			return err
		}
		return writeOrPrintChecked(text, output, path, overwrite, stdout, stderr, runtime)
	}
	return writeOrPrintChecked(AnalysisText(analysis), output, path, overwrite, stdout, stderr, runtime)
}

func runInspectQt(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	path, format, output, overwrite, err := sourceArgs(args)
	if err != nil {
		return err
	}
	analysis, err := AnalyzeQt(path)
	if err != nil {
		return err
	}
	if format == "json" {
		text, err := prettyJSON(analysis)
		if err != nil {
			return err
		}
		return writeOrPrintChecked(text, output, path, overwrite, stdout, stderr, runtime)
	}
	return writeOrPrintChecked(AnalysisText(analysis), output, path, overwrite, stdout, stderr, runtime)
}

func runInspectSource(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	path, format, output, overwrite, err := sourceArgs(args, "--from")
	if err != nil {
		return err
	}
	analysis, err := AnalyzeByPlatform(path, flagValue(args, "--from"))
	if err != nil {
		return err
	}
	if format == "json" {
		text, err := prettyJSON(analysis)
		if err != nil {
			return err
		}
		return writeOrPrintChecked(text, output, path, overwrite, stdout, stderr, runtime)
	}
	return writeOrPrintChecked(AnalysisText(analysis), output, path, overwrite, stdout, stderr, runtime)
}

func runInspectParity(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	parsed, err := parseArgs(args, map[string]bool{"--target": true, "--xaml": true, "--from": true, "--to": true, "--format": true, "--output": true}, map[string]bool{"--json": true, "--overwrite": true})
	if err != nil {
		return err
	}
	if len(parsed.Positionals) != 1 {
		return fmt.Errorf("inspect:parity requires a source path")
	}
	source := parsed.Positionals[0]
	target := firstNonEmpty(parsed.Values["--target"], parsed.Values["--xaml"])
	if target == "" {
		return fmt.Errorf("inspect:parity requires --target path")
	}
	format := firstNonEmpty(parsed.Values["--format"], "text")
	if parsed.Bools["--json"] {
		format = "json"
	}
	if format != "text" && format != "json" {
		return fmt.Errorf("--format must be text or json")
	}
	output := parsed.Values["--output"]
	report, err := InspectParity(source, target, parsed.Values["--from"], parsed.Values["--to"])
	if err != nil {
		return err
	}
	text := ParityText(report)
	if format == "json" {
		text, err = prettyJSON(report)
		if err != nil {
			return err
		}
	}
	if err := writeOrPrintChecked(text, output, source, parsed.Bools["--overwrite"], stdout, stderr, runtime); err != nil {
		return err
	}
	if report.Status != "ok" {
		return ErrCommandFailed
	}
	return nil
}

func runInspectFont(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	parsed, err := parseArgs(args, map[string]bool{"--family": true, "--format": true, "--output": true}, map[string]bool{"--json": true, "--overwrite": true})
	if err != nil {
		return err
	}
	if len(parsed.Positionals) > 1 {
		return fmt.Errorf("inspect:font accepts at most one font path")
	}
	path := ""
	if len(parsed.Positionals) == 1 {
		path = parsed.Positionals[0]
	}
	format := firstNonEmpty(parsed.Values["--format"], "text")
	if parsed.Bools["--json"] {
		format = "json"
	}
	if format != "text" && format != "json" {
		return fmt.Errorf("--format must be text or json")
	}
	report := InspectFontSource(path, parsed.Values["--family"])
	text := FontInspectionText(report)
	if format == "json" {
		text, err = prettyJSON(report)
		if err != nil {
			return err
		}
	}
	if err := writeOrPrintChecked(text, parsed.Values["--output"], path, parsed.Bools["--overwrite"], stdout, stderr, runtime); err != nil {
		return err
	}
	if report.Status != "ok" {
		return ErrCommandFailed
	}
	return nil
}

func runInspectVisualParity(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	parsed, err := parseArgs(args, map[string]bool{"--target": true, "--xaml": true, "--from": true, "--to": true, "--profile": true, "--source-font": true, "--target-font": true, "--source-font-family": true, "--target-font-family": true, "--format": true, "--output": true}, map[string]bool{"--json": true, "--overwrite": true})
	if err != nil {
		return err
	}
	if len(parsed.Positionals) != 1 {
		return fmt.Errorf("inspect:visual-parity requires a source path")
	}
	source := parsed.Positionals[0]
	target := firstNonEmpty(parsed.Values["--target"], parsed.Values["--xaml"])
	if target == "" {
		return fmt.Errorf("inspect:visual-parity requires --target path")
	}
	format := firstNonEmpty(parsed.Values["--format"], "text")
	if parsed.Bools["--json"] {
		format = "json"
	}
	if format != "text" && format != "json" {
		return fmt.Errorf("--format must be text or json")
	}
	report, err := InspectVisualParity(source, target, parsed.Values["--from"], parsed.Values["--to"], parsed.Values["--profile"], parsed.Values["--source-font"], parsed.Values["--target-font"], parsed.Values["--source-font-family"], parsed.Values["--target-font-family"])
	if err != nil {
		return err
	}
	text := VisualParityText(report)
	if format == "json" {
		text, err = prettyJSON(report)
		if err != nil {
			return err
		}
	}
	if err := writeOrPrintChecked(text, parsed.Values["--output"], source, parsed.Bools["--overwrite"], stdout, stderr, runtime); err != nil {
		return err
	}
	if report.Status != "ok" {
		return ErrCommandFailed
	}
	return nil
}

func runASCII(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	path, _, output, overwrite, err := sourceArgs(args, "--from")
	if err != nil {
		return err
	}
	analysis, err := AnalyzeByPlatform(path, flagValue(args, "--from"))
	if err != nil {
		return err
	}
	return writeOrPrintChecked(ASCIIAnalysis(analysis), output, path, overwrite, stdout, stderr, runtime)
}

func runAudit(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	path, format, output, overwrite, err := sourceArgs(args, "--fail-on")
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
	if err := writeOrPrintChecked(text, output, path, overwrite, stdout, stderr, runtime); err != nil {
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
	path, format, output, overwrite, err := sourceArgs(args, "--patterns-dir", "--from", "--to")
	if err != nil {
		return err
	}
	patternDir := firstNonEmpty(flagValue(args, "--patterns-dir"), DefaultPatternDirectory)
	from := firstNonEmpty(flagValue(args, "--from"), InferSourcePlatform(path))
	to := firstNonEmpty(flagValue(args, "--to"), defaultTransferTarget(from))
	if !validPatternPlatform(from) {
		return fmt.Errorf("--from must be swiftui, winui3, qt, or a supported alias")
	}
	if !validPatternPlatform(to) {
		return fmt.Errorf("--to must be swiftui, winui3, qt, or a supported alias")
	}
	analysis, err := AnalyzeByPlatform(path, from)
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
	return writeOrPrintChecked(text, output, path, overwrite, stdout, stderr, runtime)
}

func runGraphComponents(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	parsed, err := parseArgs(args, map[string]bool{"--root-view": true, "--component": true, "--format": true, "--include": true, "--exclude": true, "--output": true}, map[string]bool{"--json": true, "--overwrite": true})
	if err != nil {
		return err
	}
	if len(parsed.Positionals) != 1 {
		return fmt.Errorf("graph:components requires a source file or directory")
	}
	format := firstNonEmpty(parsed.Values["--format"], "text")
	if parsed.Bools["--json"] {
		format = "json"
	}
	if format != "text" && format != "json" && format != "dot" {
		return fmt.Errorf("--format must be text, json, or dot")
	}
	report, err := GraphComponents(parsed.Positionals[0], parsed.Values["--root-view"], parsed.Values["--component"], parsed.Values["--include"], parsed.Values["--exclude"])
	if err != nil {
		return err
	}
	text := ComponentGraphText(report)
	if format == "json" {
		text, err = prettyJSON(report)
		if err != nil {
			return err
		}
	} else if format == "dot" {
		text = ComponentGraphDOT(report)
	}
	if err := writeOrPrintChecked(text, parsed.Values["--output"], parsed.Positionals[0], parsed.Bools["--overwrite"], stdout, stderr, runtime); err != nil {
		return err
	}
	if report.Status == "error" {
		return ErrCommandFailed
	}
	return nil
}

func runProjectBuild(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	parsed, err := parseArgs(args, map[string]bool{"--project-root": true, "--output-dir": true, "--format": true}, map[string]bool{"--json": true, "--overwrite": true})
	if err != nil {
		return err
	}
	if len(parsed.Positionals) != 1 {
		return fmt.Errorf("project:build requires a manifest path")
	}
	format := firstNonEmpty(parsed.Values["--format"], "text")
	if parsed.Bools["--json"] {
		format = "json"
	}
	if format != "text" && format != "json" {
		return fmt.Errorf("--format must be text or json")
	}
	report, err := ProjectBuild(parsed.Positionals[0], parsed.Values["--project-root"], parsed.Values["--output-dir"], parsed.Bools["--overwrite"])
	text := ProjectBuildText(report)
	if format == "json" {
		text, _ = prettyJSON(report)
	}
	if writeErr := writeOrPrint(text, "", stdout, stderr, runtime); writeErr != nil {
		return writeErr
	}
	if err != nil {
		return err
	}
	if report.Status == "error" {
		return ErrCommandFailed
	}
	return nil
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
	fmt.Fprintf(&b, "loom OS error suggestions\nstatus: %s\nplatform: %s\nquery: %s\nsuggestions: %d\n\n", report.Status, firstNonEmpty(string(report.Platform), "all"), firstNonEmpty(report.Query, "none"), len(report.Suggestions))
	for _, suggestion := range report.Suggestions {
		fmt.Fprintf(&b, "[%s] %s: %s\n  issue: %s\n  reference: %s\n", suggestion.Platform, suggestion.Category, suggestion.Matcher, suggestion.Issue, suggestion.Reference)
		for _, fix := range suggestion.SuggestedFixes {
			fmt.Fprintf(&b, "  - %s: %s - %s\n", fix.Audience, fix.Action, fix.Detail)
		}
	}
	return b.String()
}

func runStatus(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	_ = stderr
	if contains(args, "--help") || contains(args, "-h") {
		return writeText(manual("status"), stdout, runtime)
	}
	dir := firstNonEmpty(flagValue(args, "--patterns-dir"), DefaultPatternDirectory)
	format := firstNonEmpty(flagValue(args, "--format"), "text")
	if contains(args, "--json") {
		format = "json"
	}
	report := DiagnosticsStatus(dir)
	text := LoomStatusText(report)
	if format == "json" {
		text, _ = prettyJSON(report)
	}
	return writeOrPrint(text, "", stdout, stderr, runtime)
}

func runVerify(args []string, stdout io.Writer, runtime runtimeOptions) error {
	dir := firstNonEmpty(flagValue(args, "--patterns-dir"), DefaultPatternDirectory)
	format := firstNonEmpty(flagValue(args, "--format"), "text")
	if contains(args, "--json") {
		format = "json"
	}
	report := DiagnosticsVerify(dir)
	text := LoomVerifyText(report)
	if format == "json" {
		text, _ = prettyJSON(report)
	}
	if err := writeOrPrint(text, "", stdout, os.Stderr, runtime); err != nil {
		return err
	}
	if report.Status != "ok" {
		return ErrCommandFailed
	}
	return nil
}

func runChecksCommandCatalog(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	if contains(args, "--help") || contains(args, "-h") {
		return writeText(manual("checks:command-catalog"), stdout, runtime)
	}
	format := firstNonEmpty(flagValue(args, "--format"), "text")
	if contains(args, "--json") {
		format = "json"
	}
	report := DiagnosticsCommandCatalogCheck()
	text := LoomCommandCatalogCheckText(report)
	if format == "json" {
		text, _ = prettyJSON(report)
	}
	if err := writeOrPrint(text, "", stdout, stderr, runtime); err != nil {
		return err
	}
	if report.Status == "error" {
		return ErrCommandFailed
	}
	return nil
}

func runGuardsSummary(args []string, stdout io.Writer, runtime runtimeOptions) error {
	if contains(args, "--help") || contains(args, "-h") {
		return writeText(manual("guards:summary"), stdout, runtime)
	}
	if contains(args, "--json") {
		text, _ := prettyJSON(DiagnosticsGuardsSummary())
		return writeText(text, stdout, runtime)
	}
	return writeText(LoomGuardsSummaryText(DiagnosticsGuardsSummary()), stdout, runtime)
}

func runSelfHealPlan(args []string, stdout io.Writer, runtime runtimeOptions) error {
	if contains(args, "--help") || contains(args, "-h") {
		return writeText(manual("self-heal:plan"), stdout, runtime)
	}
	if contains(args, "--json") {
		text, _ := prettyJSON(DiagnosticsSelfHealPlan())
		return writeText(text, stdout, runtime)
	}
	return writeText(LoomSelfHealPlanText(DiagnosticsSelfHealPlan()), stdout, runtime)
}

func runConfigSchema(args []string, stdout io.Writer, runtime runtimeOptions) error {
	_ = runtime
	if len(args) != 0 {
		return fmt.Errorf("config:schema accepts no arguments")
	}
	text, err := DiagnosticsProjectConfigSchema()
	if err != nil {
		return err
	}
	return writeText(text, stdout, runtime)
}

func runConfigValidate(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	_ = stderr
	if len(args) == 0 {
		return fmt.Errorf("config:validate requires a manifest path")
	}
	manifestPath := args[0]
	format := firstNonEmpty(flagValue(args, "--format"), "text")
	projectRoot := flagValue(args, "--project-root")
	if contains(args, "--json") {
		format = "json"
	}
	report := DiagnosticsProjectConfigValidate(manifestPath, projectRoot)
	text := LoomManifestValidationText(report)
	if format == "json" {
		text, _ = prettyJSON(report)
	}
	if err := writeOrPrint(text, "", stdout, stderr, runtime); err != nil {
		return err
	}
	if report.Status != "ok" {
		return ErrCommandFailed
	}
	return nil
}

func runInspectErrors(args []string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	parsed, err := parseArgs(args, map[string]bool{"--kind": true, "--root-view": true, "--component": true, "--format": true, "--fail-on": true, "--output": true}, map[string]bool{"--json": true, "--overwrite": true})
	if err != nil {
		return err
	}
	if len(parsed.Positionals) != 1 {
		return fmt.Errorf("inspect:errors requires a source path")
	}
	path := parsed.Positionals[0]
	kind := parsed.Values["--kind"]
	rootView := parsed.Values["--root-view"]
	component := parsed.Values["--component"]
	format := firstNonEmpty(parsed.Values["--format"], "text")
	failOn := firstNonEmpty(parsed.Values["--fail-on"], "none")
	output := parsed.Values["--output"]
	if parsed.Bools["--json"] {
		format = "json"
	}
	if format != "text" && format != "json" {
		return fmt.Errorf("--format must be text or json")
	}
	report := InspectErrors(path, kind, rootView, component, failOn)
	text := LoomErrorInspectionText(report)
	if format == "json" {
		text, _ = prettyJSON(report)
	}
	if err := writeOrPrintChecked(text, output, path, parsed.Bools["--overwrite"], stdout, stderr, runtime); err != nil {
		return err
	}
	if ShouldFailForInspection(report, failOn) {
		return ErrCommandFailed
	}
	return nil
}

func runUnavailableCommand(command string, stdout, stderr io.Writer, runtime runtimeOptions) error {
	_ = stderr
	_ = runtime
	return fmt.Errorf("%s is reserved for catalog parity only and is not yet available in the Go runtime", command)
}

func AnalysisText(analysis Analysis) string {
	return fmt.Sprintf("loom analysis\nsource: %s\nview: %s.%s\nsource nodes: %d\nlayout nodes: %d\n\n%s", analysis.SourcePath, analysis.RootView, analysis.Component, analysis.SyntaxNodeCount, analysis.Layout.RecursiveNodeCount(), ASCIIAnalysis(analysis))
}

type parsedArgs struct {
	Positionals []string
	Values      map[string]string
	Bools       map[string]bool
}

func sourceArgs(args []string, extraValueFlags ...string) (path, format, output string, overwrite bool, err error) {
	valueFlags := map[string]bool{"--format": true, "--output": true}
	for _, flag := range extraValueFlags {
		valueFlags[flag] = true
	}
	parsed, err := parseArgs(args, valueFlags, map[string]bool{"--json": true, "--overwrite": true})
	if err != nil {
		return "", "", "", false, err
	}
	if len(parsed.Positionals) != 1 {
		return "", "", "", false, fmt.Errorf("source path required")
	}
	path = parsed.Positionals[0]
	format = firstNonEmpty(parsed.Values["--format"], "text")
	if parsed.Bools["--json"] {
		format = "json"
	}
	if format != "text" && format != "json" {
		return "", "", "", false, fmt.Errorf("--format must be text or json")
	}
	output = parsed.Values["--output"]
	return path, format, output, parsed.Bools["--overwrite"], nil
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

func parseArgs(args []string, valueFlags map[string]bool, boolFlags map[string]bool) (parsedArgs, error) {
	parsed := parsedArgs{Values: map[string]string{}, Bools: map[string]bool{}}
	seen := map[string]bool{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			parsed.Positionals = append(parsed.Positionals, arg)
			continue
		}
		if seen[arg] {
			return parsedArgs{}, fmt.Errorf("duplicate option %s", arg)
		}
		seen[arg] = true
		if boolFlags[arg] {
			parsed.Bools[arg] = true
			continue
		}
		if valueFlags[arg] {
			i++
			if i >= len(args) || strings.HasPrefix(args[i], "--") {
				return parsedArgs{}, fmt.Errorf("missing value for %s", arg)
			}
			parsed.Values[arg] = args[i]
			continue
		}
		return parsedArgs{}, fmt.Errorf("unknown option %s", arg)
	}
	return parsed, nil
}

func contains(args []string, value string) bool {
	for _, arg := range args {
		if arg == value {
			return true
		}
	}
	return false
}
