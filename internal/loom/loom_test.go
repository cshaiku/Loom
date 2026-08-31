package loom

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixtureXAML(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "MainWindow.xaml")
	source := `<Grid xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation">` + body + `</Grid>`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadAndValidatePatterns(t *testing.T) {
	patterns, err := LoadPatterns("../../Patterns")
	if err != nil {
		t.Fatal(err)
	}
	if len(patterns) < 20 {
		t.Fatalf("expected at least 20 patterns, got %d", len(patterns))
	}
	report := ValidatePatterns("../../Patterns")
	if report.Status != "ok" {
		t.Fatalf("expected valid patterns, got %#v", report.Issues)
	}
}

func TestAnalyzeXAMLUnsupportedNativeBoundary(t *testing.T) {
	path := fixtureXAML(t, `<NavigationView PaneTitle="Shell"><Button Content="Save" /></NavigationView>`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Diagnostics) != 1 {
		t.Fatalf("expected one diagnostic, got %d", len(analysis.Diagnostics))
	}
	if analysis.Diagnostics[0].Code != "XAML.UNSUPPORTED_COMPONENT_BOUNDARY" {
		t.Fatalf("unexpected diagnostic: %#v", analysis.Diagnostics[0])
	}
	boundary := analysis.Layout.Children[0].Children[0]
	if boundary.Kind != KindComponent {
		t.Fatalf("expected component boundary, got %s", boundary.Kind)
	}
	if boundary.Properties["componentBoundary"] != "native-winui-control" {
		t.Fatalf("missing native boundary metadata: %#v", boundary.Properties)
	}
}

func TestAnalyzeXAMLGridDefinitionsBecomeMetadata(t *testing.T) {
	path := fixtureXAML(t, `<Grid.RowDefinitions><RowDefinition Height="Auto" /><RowDefinition Height="*" /></Grid.RowDefinitions><Grid.ColumnDefinitions><ColumnDefinition Width="240" /><ColumnDefinition Width="*" /></Grid.ColumnDefinitions><TextBlock Text="Name" />`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	grid := analysis.Layout.Children[0]
	if got := grid.Properties["xaml.Grid.RowDefinitions"]; got != "Auto,*" {
		t.Fatalf("unexpected row definition metadata: %q", got)
	}
	if got := grid.Properties["xaml.Grid.ColumnDefinitions"]; got != "240,*" {
		t.Fatalf("unexpected column definition metadata: %q", got)
	}
	for _, diagnostic := range analysis.Diagnostics {
		if diagnostic.Code == "XAML.UNSUPPORTED_COMPONENT_BOUNDARY" {
			t.Fatalf("grid definitions should not become unsupported boundaries: %#v", diagnostic)
		}
	}
	if len(grid.Children) != 1 || grid.Children[0].Kind != KindText {
		t.Fatalf("expected only the visible TextBlock child, got %#v", grid.Children)
	}
}

func TestAccessibilityAuditSuggestedFixes(t *testing.T) {
	path := fixtureXAML(t, `<NavigationView /><Button Width="20" Height="20" />`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	report := Audit(analysis)
	if report.Summary.Warnings == 0 && report.Summary.Errors == 0 {
		t.Fatal("expected audit findings")
	}
	foundBoundary := false
	foundFixes := false
	for _, finding := range report.Findings {
		if finding.Code == "AUDIT070" {
			foundBoundary = true
		}
		if len(finding.SuggestedFixes) > 0 {
			foundFixes = true
		}
	}
	if !foundBoundary {
		t.Fatal("expected unsupported native boundary audit finding")
	}
	if !foundFixes {
		t.Fatal("expected structured suggested fixes")
	}
}

func TestTransferFlagsUnsupportedNativeBoundary(t *testing.T) {
	path := fixtureXAML(t, `<NavigationView />`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	patterns, err := LoadPatterns("../../Patterns")
	if err != nil {
		t.Fatal(err)
	}
	report := Transfer(analysis, patterns, "winui3", "macos")
	if report.Summary.Unsupported == 0 {
		t.Fatalf("expected unsupported transfer item, got %#v", report.Summary)
	}
	if !strings.Contains(report.ASCIIPattern, `\-- grid`) {
		t.Fatalf("expected ASCII tree in transfer report, got %q", report.ASCIIPattern)
	}
}

func TestTransferIncludesGridTrackPolicy(t *testing.T) {
	path := fixtureXAML(t, `<Grid.RowDefinitions><RowDefinition Height="Auto" /><RowDefinition Height="*" /></Grid.RowDefinitions><Button Content="Save" />`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	patterns, err := LoadPatterns("../../Patterns")
	if err != nil {
		t.Fatal(err)
	}
	report := Transfer(analysis, patterns, "winui3", "macos")
	found := false
	for _, item := range report.Items {
		if item.Kind == KindGrid && strings.Contains(strings.Join(item.Policies, " "), "row/column tracks") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected grid track transfer policy, got %#v", report.Items)
	}
}

func TestTransferMacOSTargetUsesSwiftUIMappings(t *testing.T) {
	path := fixtureXAML(t, `<Button Content="Save" />`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	patterns, err := LoadPatterns("../../Patterns")
	if err != nil {
		t.Fatal(err)
	}
	report := Transfer(analysis, patterns, "winui3", "macos")
	if report.To != "macos" {
		t.Fatalf("expected public route target to remain macos, got %q", report.To)
	}
	for _, item := range report.Items {
		if item.Kind == KindButton && !contains(item.TargetConstructs, "Button") {
			t.Fatalf("expected macos route to use SwiftUI button mapping, got %#v", item)
		}
	}
}

func TestOSErrorSuggestionsMatchStaticResource(t *testing.T) {
	report := OSErrorSuggestions("winui3", "StaticResource not found")
	if report.Status != "matched" {
		t.Fatalf("expected matched report, got %s", report.Status)
	}
	if len(report.Suggestions) == 0 || len(report.Suggestions[0].SuggestedFixes) == 0 {
		t.Fatalf("expected suggestions with fixes, got %#v", report.Suggestions)
	}
}

func TestOSErrorSuggestionsPlatformAndQueryFiltering(t *testing.T) {
	allWinUI := OSErrorSuggestions("winui3", "")
	if allWinUI.Status != "ok" || len(allWinUI.Suggestions) < 3 {
		t.Fatalf("expected all WinUI suggestions, got %#v", allWinUI)
	}
	xamlParse := OSErrorSuggestions("xaml", "")
	if xamlParse.Status != "ok" || len(xamlParse.Suggestions) != 1 || xamlParse.Suggestions[0].Category != "parse" {
		t.Fatalf("expected xaml parse suggestion, got %#v", xamlParse)
	}
	empty := OSErrorSuggestions("windows", "no matching issue")
	if empty.Status != "empty" || len(empty.Suggestions) != 0 {
		t.Fatalf("expected empty windows no-match report, got %#v", empty)
	}
}

func TestQueryMatchesCaseAndTokenizedInput(t *testing.T) {
	haystack := "winui3 resources staticresource unresolved resource dictionaries"
	for _, query := range []string{"StaticResource", "static-resource failed", "RESOURCE_DICTIONARY"} {
		if !queryMatches(haystack, query) {
			t.Fatalf("expected query %q to match %q", query, haystack)
		}
	}
	if queryMatches(haystack, "xml") {
		t.Fatal("short unrelated tokens should not match")
	}
}

func TestUnknownCommandGuidesHumansAndAgents(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := Run([]string{"not:a-command"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected unknown command to fail")
	}
	message := err.Error()
	for _, needle := range []string{"not:a-command", "loom help", "loom list --json"} {
		if !strings.Contains(message, needle) {
			t.Fatalf("expected unknown command guidance to include %q, got %q", needle, message)
		}
	}
}

func TestCLIJSONAndVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"version"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "loom 0.19.0" {
		t.Fatalf("unexpected version output: %q", got)
	}
	stdout.Reset()
	if err := Run([]string{"list", "--category", "patterns", "--json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var commands []CommandInfo
	if err := json.Unmarshal(stdout.Bytes(), &commands); err != nil {
		t.Fatal(err)
	}
	if len(commands) == 0 {
		t.Fatal("expected commands in JSON output")
	}
	if strings.Contains(stdout.String(), `\u003c`) {
		t.Fatalf("command JSON should keep synopsis placeholders readable, got %q", stdout.String())
	}
}

func TestFunctionJSONUsesStableEmptyArrays(t *testing.T) {
	patternReport := ValidatePatterns("../../Patterns")
	text, err := prettyJSON(patternReport)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text, `"issues": null`) {
		t.Fatalf("expected empty issues array, got %s", text)
	}
	suggestions := OSErrorSuggestions("windows", "no matching issue")
	text, err = prettyJSON(suggestions)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text, `"suggestions": null`) {
		t.Fatalf("expected empty suggestions array, got %s", text)
	}
	transfer := Transfer(Analysis{Layout: Node{Children: []Node{{Kind: KindUnsupported, Expression: "Unknown"}}}}, nil, "winui3", "macos")
	text, err = prettyJSON(transfer)
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{`"sourceConstructs": []`, `"targetConstructs": []`, `"contracts": []`, `"policies": []`} {
		if !strings.Contains(text, needle) {
			t.Fatalf("expected stable empty array %s in %s", needle, text)
		}
	}
}

func TestLineEndingOptionControlsStdoutAndFiles(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"--line-ending", "crlf", "version"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != "loom 0.19.0\r\n" {
		t.Fatalf("expected CRLF stdout, got %q", got)
	}

	path := fixtureXAML(t, `<Button Content="Save" />`)
	out := filepath.Join(t.TempDir(), "tree.txt")
	stdout.Reset()
	if err := Run([]string{"inspect:ascii", path, "--output", out, "--line-ending", "crlf"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(stdout.String(), "\r\n") {
		t.Fatalf("expected CRLF write confirmation, got %q", stdout.String())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("\r\n")) || bytes.Contains(data, []byte("\r\r\n")) {
		t.Fatalf("expected normalized CRLF file output, got %q", string(data))
	}
	if err := Run([]string{"--line-ending", "weird", "version"}, &stdout, &stderr); err == nil {
		t.Fatal("expected invalid line ending to fail")
	}
}

func TestLineEndingPolicyFileExists(t *testing.T) {
	data, err := os.ReadFile("../../.gitattributes")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, rule := range []string{"* text=auto eol=lf", "*.bat text eol=crlf", "*.cmd text eol=crlf", "*.ps1 text eol=crlf"} {
		if !strings.Contains(text, rule) {
			t.Fatalf("missing line-ending policy rule %q in %s", rule, text)
		}
	}
}

func TestCLIQuietOutputWrite(t *testing.T) {
	path := fixtureXAML(t, `<Button Content="Save" />`)
	out := filepath.Join(t.TempDir(), "tree.txt")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"--quiet", "inspect:ascii", path, "--output", out}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("quiet write should not print success chatter, got %q", stdout.String())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "button / Button") {
		t.Fatalf("unexpected ASCII output: %s", string(data))
	}
}

func TestCommandCatalogCoverage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"list", "--json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var commands []CommandInfo
	if err := json.Unmarshal(stdout.Bytes(), &commands); err != nil {
		t.Fatal(err)
	}
	if len(commands) < 16 {
		t.Fatalf("expected expanded command coverage, got %d", len(commands))
	}
	found := 0
	required := map[string]struct{}{"config:validate": {}, "config:schema": {}, "checks:command-catalog": {}, "guards:summary": {}, "self-heal:plan": {}, "inspect:errors": {}, "project:build": {}, "generate:xaml": {}}
	for _, command := range commands {
		if _, ok := required[command.Command]; ok {
			found++
			delete(required, command.Command)
		}
	}
	if len(required) > 0 {
		missing := make([]string, 0, len(required))
		for command := range required {
			missing = append(missing, command)
		}
		t.Fatalf("missing required command coverage: %s", strings.Join(missing, ", "))
	}
	if found != 8 {
		t.Fatalf("unexpected required command count %d", found)
	}
}

func TestDiagnosticCommandJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"checks:command-catalog", "--json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var report LoomCommandCatalogCheckReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "ok" {
		t.Fatalf("expected catalog checks ok, got %s", report.Status)
	}
}

func TestInspectErrorsAndFailMode(t *testing.T) {
	path := fixtureXAML(t, `<Grid><TextBlock></Grid>`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := Run([]string{"inspect:errors", path, "--kind", "xaml", "--format", "json", "--fail-on", "error"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected malformed xaml to fail")
	}
	if !strings.Contains(strings.TrimSpace(stdout.String()), "\"status\": \"error\"") && !strings.Contains(stdout.String(), "LOOM.XAML") {
		t.Fatalf("expected error report, got %q", stdout.String())
	}
	var report LoomErrorInspectionReport
	if err2 := json.Unmarshal(stdout.Bytes(), &report); err2 == nil {
		if report.Status != "error" {
			t.Fatalf("expected report status error, got %s", report.Status)
		}
	}
	if !strings.Contains(err.Error(), "command completed") {
		t.Fatal(err)
	}
	_ = stderr
}

func TestUnavailableSwiftOnlyCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := Run([]string{"generate:xaml", "MainView.swift"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected go-only unsupported error")
	}
}
