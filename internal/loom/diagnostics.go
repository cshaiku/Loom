package loom

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type LoomErrorInspectionKind string

const (
	LoomErrorInspectionSwift    LoomErrorInspectionKind = "swift"
	LoomErrorInspectionXAML     LoomErrorInspectionKind = "xaml"
	LoomErrorInspectionManifest LoomErrorInspectionKind = "manifest"
	LoomErrorInspectionPatterns LoomErrorInspectionKind = "patterns"
)

type LoomErrorInspectionReport struct {
	SchemaVersion string                       `json:"schema_version"`
	Status        string                       `json:"status"`
	InspectedKind LoomErrorInspectionKind      `json:"inspectedKind"`
	Source        string                       `json:"source"`
	Findings      []LoomErrorInspectionFinding `json:"findings"`
}

type LoomErrorInspectionFinding struct {
	Severity       DiagnosticSeverity `json:"severity"`
	Code           string             `json:"code"`
	Source         string             `json:"source"`
	Message        string             `json:"message"`
	Offset         *int               `json:"offset,omitempty"`
	Line           *int               `json:"line,omitempty"`
	Column         *int               `json:"column,omitempty"`
	SuggestedFixes []SuggestedFix     `json:"suggested_fixes"`
}

type LoomManifest struct {
	SchemaVersion       string   `json:"schema_version"`
	Project             string   `json:"project"`
	Source              string   `json:"source"`
	RootView            string   `json:"rootView"`
	Target              string   `json:"target"`
	ExistingXaml        string   `json:"existingXaml,omitempty"`
	ReferenceLayout     string   `json:"referenceLayout,omitempty"`
	TranslationGuide    string   `json:"translationGuide,omitempty"`
	Components          []string `json:"components"`
	ThemeResourcePrefix string   `json:"themeResourcePrefix,omitempty"`
}

type LoomManifestValidationIssue struct {
	Severity DiagnosticSeverity `json:"severity"`
	Code     string             `json:"code"`
	Path     string             `json:"path"`
	Detail   string             `json:"detail"`
	Fix      string             `json:"fix"`
}

type LoomManifestValidationReport struct {
	SchemaVersion string                        `json:"schema_version"`
	Status        string                        `json:"status"`
	Project       string                        `json:"project,omitempty"`
	Issues        []LoomManifestValidationIssue `json:"issues"`
}

type LoomStatusReport struct {
	SchemaVersion    string       `json:"schema_version"`
	Version          string       `json:"version"`
	WorkingDirectory string       `json:"workingDirectory"`
	Commands         int          `json:"commands"`
	PatternDirectory string       `json:"patternDirectory"`
	PatternStatus    string       `json:"patternStatus"`
	PatternCount     int          `json:"patternCount"`
	Issues           []Diagnostic `json:"issues"`
}

type LoomCommandCatalogCheckReport struct {
	SchemaVersion string       `json:"schema_version"`
	Status        string       `json:"status"`
	Commands      int          `json:"commands"`
	Aliases       int          `json:"aliases"`
	Issues        []Diagnostic `json:"issues"`
}

type LoomVerifyReport struct {
	SchemaVersion  string                        `json:"schema_version"`
	Status         string                        `json:"status"`
	CommandCatalog LoomCommandCatalogCheckReport `json:"commandCatalog"`
	Patterns       LoomPatternValidationReport   `json:"patterns"`
	PatternLint    LoomPatternValidationReport   `json:"patternLint"`
}

type LoomGuardEntry struct {
	Command     string        `json:"command"`
	Access      CommandAccess `json:"access"`
	WriteFlags  []string      `json:"writeFlags"`
	Description string        `json:"description"`
}

type LoomGuardsReport struct {
	SchemaVersion string           `json:"schema_version"`
	Status        string           `json:"status"`
	Entries       []LoomGuardEntry `json:"entries"`
}

type LoomSelfHealEntry struct {
	Command   string `json:"command"`
	Flag      string `json:"flag"`
	Scope     string `json:"scope"`
	Guardrail string `json:"guardrail"`
}

type LoomSelfHealPlan struct {
	SchemaVersion string              `json:"schema_version"`
	Status        string              `json:"status"`
	Entries       []LoomSelfHealEntry `json:"entries"`
}

type LoomPatternValidationReport = PatternValidationReport

func DiagnosticsStatus(patternDirectory string) LoomStatusReport {
	patternReport := ValidatePatterns(patternDirectory)
	issues := make([]Diagnostic, 0, len(patternReport.Issues))
	for _, issue := range patternReport.Issues {
		issues = append(issues, Diagnostic{Severity: issue.Severity, Code: issue.Code, Message: issue.Detail})
	}
	cwd, err := os.Getwd()
	if err != nil {
		issues = append(issues, Diagnostic{Severity: SeverityWarning, Code: "STATUS001", Message: fmt.Sprintf("Working directory could not be resolved: %v.", err)})
		cwd = ""
	}
	return LoomStatusReport{
		SchemaVersion:    "1",
		Version:          Version,
		WorkingDirectory: cwd,
		Commands:         len(Commands),
		PatternDirectory: patternReport.Directory,
		PatternStatus:    patternReport.Status,
		PatternCount:     patternReport.PatternCount,
		Issues:           issues,
	}
}

func DiagnosticsCommandCatalogCheck() LoomCommandCatalogCheckReport {
	commands := Commands
	issues := []Diagnostic{}
	commandNames := map[string]struct{}{}
	symbols := map[string]struct{}{}
	aliasCount := 0
	for _, command := range commands {
		if _, ok := commandNames[command.Command]; ok {
			issues = append(issues, Diagnostic{Severity: SeverityError, Code: "CATALOG001", Message: fmt.Sprintf("Duplicate command %s.", command.Command)})
		}
		commandNames[command.Command] = struct{}{}
		if _, ok := symbols[command.Command]; ok {
			issues = append(issues, Diagnostic{Severity: SeverityError, Code: "CATALOG002", Message: fmt.Sprintf("Command collides with another command or alias: %s.", command.Command)})
		}
		symbols[command.Command] = struct{}{}
		if strings.TrimSpace(command.Name) == "" || strings.TrimSpace(command.Description) == "" || strings.TrimSpace(command.Category) == "" {
			issues = append(issues, Diagnostic{Severity: SeverityError, Code: "CATALOG003", Message: fmt.Sprintf("Command %s has incomplete metadata.", command.Command)})
		}
		if len(command.Synopsis) == 0 {
			issues = append(issues, Diagnostic{Severity: SeverityError, Code: "CATALOG004", Message: fmt.Sprintf("Command %s has no synopsis.", command.Command)})
		}
		for _, synopsis := range command.Synopsis {
			if !strings.HasPrefix(synopsis, "loom "+command.Command) {
				issues = append(issues, Diagnostic{Severity: SeverityError, Code: "CATALOG005", Message: fmt.Sprintf("Synopsis for %s does not start with command.", command.Command)})
			}
		}
		if command.Access == AccessRead && len(command.WriteFlags) > 0 {
			issues = append(issues, Diagnostic{Severity: SeverityError, Code: "CATALOG006", Message: fmt.Sprintf("Read-only command %s declares write flags.", command.Command)})
		}
		for _, alias := range command.Aliases {
			aliasCount++
			if strings.Contains(alias, ":") {
				issues = append(issues, Diagnostic{Severity: SeverityError, Code: "CATALOG007", Message: fmt.Sprintf("Alias %s should be short and non-namespaced.", alias)})
			}
			if _, ok := symbols[alias]; ok {
				issues = append(issues, Diagnostic{Severity: SeverityError, Code: "CATALOG008", Message: fmt.Sprintf("Alias collides with command or alias: %s.", alias)})
			}
			symbols[alias] = struct{}{}
		}
	}
	status := "ok"
	if len(issues) > 0 {
		status = "error"
	}
	return LoomCommandCatalogCheckReport{SchemaVersion: "1", Status: status, Commands: len(commands), Aliases: aliasCount, Issues: issues}
}

func DiagnosticsVerify(patternDirectory string) LoomVerifyReport {
	commandCatalog := DiagnosticsCommandCatalogCheck()
	patterns := ValidatePatterns(patternDirectory)
	lint := LintPatterns(patternDirectory)
	status := "ok"
	if commandCatalog.Status != "ok" || patterns.Status != "ok" || lint.Status != "ok" {
		status = "error"
	}
	return LoomVerifyReport{
		SchemaVersion:  "1",
		Status:         status,
		CommandCatalog: commandCatalog,
		Patterns:       LoomPatternValidationReport(patterns),
		PatternLint:    LoomPatternValidationReport(lint),
	}
}

func DiagnosticsGuardsSummary() LoomGuardsReport {
	entries := make([]LoomGuardEntry, 0, len(Commands))
	for _, command := range Commands {
		if command.Access != AccessRead || len(command.WriteFlags) > 0 {
			entries = append(entries, LoomGuardEntry{
				Command:     command.Command,
				Access:      command.Access,
				WriteFlags:  command.WriteFlags,
				Description: command.Description,
			})
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Command < entries[j].Command })
	return LoomGuardsReport{SchemaVersion: "1", Status: "ok", Entries: entries}
}

func DiagnosticsSelfHealPlan() LoomSelfHealPlan {
	return LoomSelfHealPlan{
		SchemaVersion: "1",
		Status:        "ok",
		Entries:       []LoomSelfHealEntry{},
	}
}

func DiagnosticsProjectConfigValidate(path, projectRoot string) LoomManifestValidationReport {
	path = filepath.Clean(path)
	manifest := LoomManifest{}
	data, readErr := os.ReadFile(path)
	if readErr != nil {
		return LoomManifestValidationReport{SchemaVersion: "1", Status: "error", Project: "", Issues: []LoomManifestValidationIssue{{Severity: SeverityError, Code: "manifest.unreadable", Path: path, Detail: readErr.Error(), Fix: "Correct manifest path or encoding."}}}
	}

	if err := json.Unmarshal(data, &manifest); err != nil {
		return LoomManifestValidationReport{SchemaVersion: "1", Status: "error", Project: "", Issues: []LoomManifestValidationIssue{{Severity: SeverityError, Code: "manifest.decode", Path: path, Detail: err.Error(), Fix: "Correct manifest syntax and required fields."}}}
	}

	known := map[string]struct{}{
		"schema_version":      {},
		"project":             {},
		"source":              {},
		"rootView":            {},
		"target":              {},
		"existingXaml":        {},
		"referenceLayout":     {},
		"translationGuide":    {},
		"components":          {},
		"themeResourcePrefix": {},
	}

	var raw map[string]any
	_ = json.Unmarshal(data, &raw)
	issues := []LoomManifestValidationIssue{}
	for key := range raw {
		if _, ok := known[key]; !ok {
			issues = append(issues, LoomManifestValidationIssue{Severity: SeverityError, Code: "manifest.key.unsupported", Path: key, Detail: fmt.Sprintf("Unsupported manifest key %s.", key), Fix: "Remove the key or check loom config:schema."})
		}
	}

	if manifest.SchemaVersion == "" {
		manifest.SchemaVersion = "1"
	}
	if manifest.SchemaVersion != "1" {
		issues = append(issues, LoomManifestValidationIssue{Severity: SeverityError, Code: "manifest.schema_version", Path: "schema_version", Detail: "Expected schema_version 1.", Fix: "Set schema_version to \"1\"."})
	}
	if strings.TrimSpace(manifest.Project) == "" {
		issues = append(issues, LoomManifestValidationIssue{Severity: SeverityError, Code: "manifest.project.empty", Path: "project", Detail: "Project name is empty.", Fix: "Set project to a stable display name."})
	}
	if strings.TrimSpace(manifest.Target) == "" {
		issues = append(issues, LoomManifestValidationIssue{Severity: SeverityError, Code: "manifest.target.empty", Path: "target", Detail: "Target is empty.", Fix: "Set target to winui3."})
	} else if strings.ToLower(manifest.Target) != "winui3" {
		issues = append(issues, LoomManifestValidationIssue{Severity: SeverityError, Code: "manifest.target.unsupported", Path: "target", Detail: fmt.Sprintf("Unsupported target %s.", manifest.Target), Fix: "Set target to winui3."})
	}
	if strings.TrimSpace(manifest.Source) == "" {
		issues = append(issues, LoomManifestValidationIssue{Severity: SeverityError, Code: "manifest.source.empty", Path: "source", Detail: "Source is empty.", Fix: "Provide a source file path."})
	}
	if strings.TrimSpace(manifest.RootView) == "" {
		issues = append(issues, LoomManifestValidationIssue{Severity: SeverityError, Code: "manifest.rootView.empty", Path: "rootView", Detail: "Root view is empty.", Fix: "Set a root view like ContentView."})
	}

	components := uniqueManifestComponents(manifest.Components)
	if len(components) == 0 {
		issues = append(issues, LoomManifestValidationIssue{Severity: SeverityError, Code: "manifest.components.empty", Path: "components", Detail: "No components are declared.", Fix: "Add one or more layout component names."})
	}

	root := filepath.Dir(path)
	if projectRoot != "" {
		root = projectRoot
	}
	sourcePath := resolveManifestPath(manifest.Source, root)
	if !fileExists(sourcePath) {
		issues = append(issues, LoomManifestValidationIssue{Severity: SeverityError, Code: "source.missing", Path: "source", Detail: fmt.Sprintf("Source does not exist at %s.", sourcePath), Fix: "Correct source or --project-root."})
	} else {
		if _, err := os.ReadFile(sourcePath); err != nil {
			issues = append(issues, LoomManifestValidationIssue{Severity: SeverityError, Code: "source.unreadable", Path: "source", Detail: fmt.Sprintf("Could not read source at %s.", sourcePath), Fix: "Check file permissions and encoding."})
		}
	}

	if manifest.ExistingXaml != "" {
		xamlPath := resolveManifestPath(manifest.ExistingXaml, root)
		if !fileExists(xamlPath) {
			issues = append(issues, LoomManifestValidationIssue{Severity: SeverityError, Code: "xaml.missing", Path: "existingXaml", Detail: fmt.Sprintf("XAML source does not exist at %s.", xamlPath), Fix: "Correct existingXaml or --project-root."})
		}
	}

	for _, field := range []struct {
		field string
		value string
	}{
		{"referenceLayout", manifest.ReferenceLayout},
		{"translationGuide", manifest.TranslationGuide},
	} {
		if field.value != "" {
			path := resolveManifestPath(field.value, root)
			if !fileExists(path) {
				issues = append(issues, LoomManifestValidationIssue{Severity: SeverityWarning, Code: "reference.missing", Path: field.field, Detail: fmt.Sprintf("Reference does not exist at %s.", path), Fix: "Correct or remove the optional reference path."})
			}
		}
	}

	status := "ok"
	for _, issue := range issues {
		if issue.Severity == SeverityError {
			status = "error"
			break
		}
	}
	project := manifest.Project
	if strings.TrimSpace(project) == "" {
		project = path
	}
	return LoomManifestValidationReport{SchemaVersion: "1", Status: status, Project: project, Issues: issues}
}

func DiagnosticsProjectConfigSchema() (string, error) {
	return ManifestSchemaJSON()
}

func DiagnosticsProjectBuild(manifestPath, projectRoot, outputDir string) ([]byte, error) {
	_ = outputDir
	report := DiagnosticsProjectConfigValidate(manifestPath, projectRoot)
	text, _ := json.MarshalIndent(report, "", "  ")
	if report.Status != "ok" {
		return append(text, '\n'), fmt.Errorf("project:build preconditions failed")
	}
	return append(text, '\n'), fmt.Errorf("project:build is reserved for catalog parity only and is not yet available in the Go runtime")
}

func InspectErrors(path string, kind, rootView, component, failOn string) LoomErrorInspectionReport {
	_ = rootView
	_ = component
	path = filepath.Clean(path)
	if kind == "" {
		kind = string(inferErrorInspectionKind(path))
	}
	kind = strings.ToLower(kind)
	findings := []LoomErrorInspectionFinding{}
	switch LoomErrorInspectionKind(kind) {
	case LoomErrorInspectionSwift:
		findings = inspectSwiftErrors(path, rootView, component)
	case LoomErrorInspectionXAML:
		findings = inspectXAMLErrors(path)
	case LoomErrorInspectionManifest:
		findings = inspectManifestErrors(path)
	case LoomErrorInspectionPatterns:
		findings = inspectPatternErrors(path)
	default:
		findings = append(findings, LoomErrorInspectionFinding{Severity: SeverityError, Code: "LOOM.UNKNOWN_KIND", Source: path, Message: fmt.Sprintf("Unknown inspection kind %s.", kind)})
	}
	for i := range findings {
		s := &findings[i]
		if s.SuggestedFixes == nil {
			s.SuggestedFixes = []SuggestedFix{}
		}
		for _, fix := range suggestFixesForInspectionFinding(*s) {
			s.SuggestedFixes = append(s.SuggestedFixes, fix)
		}
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Offset != nil && findings[j].Offset != nil {
			return *findings[i].Offset < *findings[j].Offset
		}
		return findings[i].Code < findings[j].Code
	})
	status := "ok"
	for _, finding := range findings {
		if finding.Severity == SeverityError {
			status = "error"
			break
		}
	}
	if failOn == "warning" {
		for _, finding := range findings {
			if finding.Severity == SeverityWarning {
				status = "error"
				break
			}
		}
	}
	return LoomErrorInspectionReport{SchemaVersion: "1", Status: status, InspectedKind: LoomErrorInspectionKind(kind), Source: path, Findings: findings}
}

func ShouldFailForInspection(report LoomErrorInspectionReport, mode string) bool {
	switch mode {
	case "error":
		return report.Status == "error"
	case "warning":
		if report.Status == "error" {
			return true
		}
		for _, finding := range report.Findings {
			if finding.Severity == SeverityWarning {
				return true
			}
		}
	}
	return false
}

func ManifestSchemaJSON() (string, error) {
	const schema = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Loom project manifest",
  "type": "object",
  "required": ["schema_version", "project", "source", "rootView", "target", "components"],
  "properties": {
    "schema_version": { "const": "1" },
    "project": { "type": "string", "minLength": 1 },
    "source": { "type": "string", "minLength": 1 },
    "rootView": { "type": "string", "minLength": 1 },
    "target": { "const": "winui3" },
    "existingXaml": { "type": "string" },
    "referenceLayout": { "type": "string" },
    "translationGuide": { "type": "string" },
    "components": {
      "type": "array",
      "minItems": 1,
      "items": { "type": "string", "minLength": 1 }
    },
    "themeResourcePrefix": { "type": "string" }
  },
  "additionalProperties": false
}`
	var parsed map[string]any
	if err := json.Unmarshal([]byte(schema), &parsed); err != nil {
		return "", err
	}
	output, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return "", err
	}
	return string(output) + "\n", nil
}

func LoomErrorInspectionText(report LoomErrorInspectionReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Loom error inspection\nStatus: %s\nKind: %s\nSource: %s\nFindings: %d\n", report.Status, report.InspectedKind, report.Source, len(report.Findings))
	if len(report.Findings) == 0 {
		b.WriteString("  none\n")
		return b.String()
	}
	for _, finding := range report.Findings {
		location := "unknown"
		if finding.Line != nil && finding.Column != nil {
			location = fmt.Sprintf("%d:%d", *finding.Line, *finding.Column)
		} else if finding.Offset != nil {
			location = fmt.Sprintf("offset %d", *finding.Offset)
		}
		b.WriteString(fmt.Sprintf("[%s] %s %s\n  %s\n", finding.Severity, finding.Code, location, finding.Message))
		if len(finding.SuggestedFixes) > 0 {
			b.WriteString("  suggested fixes:\n")
			for _, fix := range finding.SuggestedFixes {
				b.WriteString(fmt.Sprintf("    - %s: %s - %s\n", fix.Audience, fix.Action, fix.Detail))
				if fix.Command != "" {
					b.WriteString(fmt.Sprintf("      command: %s\n", fix.Command))
				}
			}
		}
	}
	return b.String()
}

func LoomStatusText(report LoomStatusReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Loom status\nVersion: %s\nWorking directory: %s\nCommands: %d\nPattern directory: %s\nPatterns: %s (%d)\n", report.Version, report.WorkingDirectory, report.Commands, report.PatternDirectory, report.PatternStatus, report.PatternCount)
	fmt.Fprintf(&b, "Issues: %d\n", len(report.Issues))
	for _, issue := range report.Issues {
		fmt.Fprintf(&b, "  [%s] %s %s\n", issue.Severity, issue.Code, issue.Message)
	}
	return b.String()
}

func LoomCommandCatalogCheckText(report LoomCommandCatalogCheckReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Loom command catalog check\nStatus: %s\nCommands: %d\nAliases: %d\nIssues: %d\n", report.Status, report.Commands, report.Aliases, len(report.Issues))
	for _, issue := range report.Issues {
		fmt.Fprintf(&b, "  [%s] %s %s\n", issue.Severity, issue.Code, issue.Message)
	}
	return b.String()
}

func LoomVerifyText(report LoomVerifyReport) string {
	var b strings.Builder
	b.WriteString("Loom verify\n")
	fmt.Fprintf(&b, "Status: %s\n", report.Status)
	fmt.Fprintf(&b, "Command catalog: %s\n", report.CommandCatalog.Status)
	fmt.Fprintf(&b, "Patterns: %s\n", report.Patterns.Status)
	fmt.Fprintf(&b, "Pattern lint: %s\n", report.PatternLint.Status)
	fmt.Fprintf(&b, "Issues: %d\n", len(report.CommandCatalog.Issues)+len(report.Patterns.Issues)+len(report.PatternLint.Issues))
	return b.String()
}

func LoomGuardsSummaryText(report LoomGuardsReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Loom guards summary\nStatus: %s\nWriting commands: %d\n\n", report.Status, len(report.Entries))
	for _, entry := range report.Entries {
		flags := "always writes"
		if len(entry.WriteFlags) > 0 {
			flags = strings.Join(entry.WriteFlags, ", ")
		}
		b.WriteString(fmt.Sprintf("%s %s\n", entry.Access, entry.Command))
		b.WriteString(fmt.Sprintf("  %s\n", flags))
		b.WriteString(fmt.Sprintf("  %s\n", entry.Description))
	}
	return b.String()
}

func LoomSelfHealPlanText(report LoomSelfHealPlan) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Loom self-heal plan\nStatus: %s\nHealable actions: %d\n\n", report.Status, len(report.Entries))
	for _, entry := range report.Entries {
		b.WriteString(fmt.Sprintf("%s %s\n", entry.Command, entry.Flag))
		b.WriteString(fmt.Sprintf("  scope: %s\n", entry.Scope))
		b.WriteString(fmt.Sprintf("  guardrail: %s\n", entry.Guardrail))
	}
	return b.String()
}

func LoomManifestValidationText(report LoomManifestValidationReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Loom manifest validation\nStatus: %s\nProject: %s\nIssues: %d\n", report.Status, report.Project, len(report.Issues))
	for _, issue := range report.Issues {
		fmt.Fprintf(&b, "  [%s] %s %s: %s\n  fix: %s\n", issue.Severity, issue.Code, issue.Path, issue.Detail, issue.Fix)
	}
	return b.String()
}

func inferErrorInspectionKind(path string) LoomErrorInspectionKind {
	if path == "" {
		return LoomErrorInspectionSwift
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return LoomErrorInspectionPatterns
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".swift", ".txt":
		return LoomErrorInspectionSwift
	case ".xaml", ".xml":
		return LoomErrorInspectionXAML
	case ".json":
		if strings.HasSuffix(strings.ToLower(filepath.Base(path)), ".pattern.json") {
			return LoomErrorInspectionPatterns
		}
		return LoomErrorInspectionManifest
	default:
		if strings.HasSuffix(strings.ToLower(filepath.Base(path)), ".pattern.json") {
			return LoomErrorInspectionPatterns
		}
		return LoomErrorInspectionSwift
	}
}

func inspectXAMLErrors(path string) []LoomErrorInspectionFinding {
	findings := []LoomErrorInspectionFinding{}
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		return append(findings, LoomErrorInspectionFinding{Severity: SeverityError, Code: "LOOM.XAML", Source: path, Message: err.Error()})
	}
	for _, diagnostic := range analysis.Diagnostics {
		finding := LoomErrorInspectionFinding{Severity: diagnostic.Severity, Code: diagnostic.Code, Source: path, Message: diagnostic.Message}
		if diagnostic.SourceOffset != nil {
			finding.Offset = diagnostic.SourceOffset
		}
		findings = append(findings, finding)
	}
	return findings
}

func inspectManifestErrors(path string) []LoomErrorInspectionFinding {
	report := DiagnosticsProjectConfigValidate(path, "")
	findings := make([]LoomErrorInspectionFinding, 0, len(report.Issues))
	for _, issue := range report.Issues {
		findings = append(findings, LoomErrorInspectionFinding{
			Severity: issue.Severity,
			Code:     issue.Code,
			Source:   path,
			Message:  fmt.Sprintf("%s: %s Fix: %s", issue.Path, issue.Detail, issue.Fix),
		})
	}
	return findings
}

func inspectPatternErrors(directory string) []LoomErrorInspectionFinding {
	validate := ValidatePatterns(directory)
	lint := LintPatterns(directory)
	issues := append(validate.Issues, lint.Issues...)
	findings := make([]LoomErrorInspectionFinding, 0, len(issues))
	for _, issue := range issues {
		findings = append(findings, LoomErrorInspectionFinding{Severity: issue.Severity, Code: issue.Code, Source: issue.Path, Message: fmt.Sprintf("%s: %s", issue.Path, issue.Detail)})
	}
	return findings
}

func inspectSwiftErrors(path, rootView, component string) []LoomErrorInspectionFinding {
	_ = rootView
	_ = component
	_, err := os.ReadFile(path)
	if err != nil {
		return []LoomErrorInspectionFinding{{Severity: SeverityError, Code: "SOURCE001", Source: path, Message: "Could not read source."}}
	}
	return []LoomErrorInspectionFinding{}
}

func suggestFixesForInspectionFinding(finding LoomErrorInspectionFinding) []SuggestedFix {
	query := strings.TrimSpace(strings.Join([]string{finding.Code, finding.Message}, " "))
	if query == "" {
		return nil
	}
	platform := ""
	switch {
	case strings.HasPrefix(finding.Code, "SOURCE") || finding.Code == "SWIFT.PARSE":
		platform = "swiftui"
	case strings.HasPrefix(strings.ToUpper(string(finding.Code)), "LOOM.XAML") || strings.HasPrefix(finding.Code, "XAML") || strings.HasPrefix(finding.Code, "XAML."):
		platform = "xaml"
	default:
		platform = ""
	}
	report := OSErrorSuggestions(platform, query)
	if len(report.Suggestions) == 0 {
		return nil
	}
	out := []SuggestedFix{}
	for _, suggestion := range report.Suggestions {
		out = append(out, suggestion.SuggestedFixes...)
	}
	return out
}

func resolveManifestPath(path string, root string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(root, path))
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func uniqueManifestComponents(components []string) []string {
	seen := map[string]struct{}{}
	result := []string{}
	for _, component := range components {
		trimmed := strings.TrimSpace(component)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
