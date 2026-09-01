package loom

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type GeneratedArtifactReport struct {
	SchemaVersion string       `json:"schema_version"`
	Status        string       `json:"status"`
	SourcePath    string       `json:"sourcePath"`
	From          string       `json:"from"`
	To            string       `json:"to"`
	RootView      string       `json:"rootView"`
	Component     string       `json:"component"`
	OutputKind    string       `json:"outputKind"`
	Text          string       `json:"text"`
	Diagnostics   []Diagnostic `json:"diagnostics"`
}

type ContractReport struct {
	SchemaVersion string         `json:"schema_version"`
	Status        string         `json:"status"`
	SourcePath    string         `json:"sourcePath"`
	Target        string         `json:"target"`
	RootView      string         `json:"rootView"`
	Component     string         `json:"component"`
	Contracts     []ContractItem `json:"contracts"`
	Diagnostics   []Diagnostic   `json:"diagnostics"`
}

type ContractItem struct {
	Path        string   `json:"path"`
	Kind        NodeKind `json:"kind"`
	Expression  string   `json:"expression"`
	Contracts   []string `json:"contracts"`
	Policies    []string `json:"policies"`
	SourceNotes []string `json:"sourceNotes"`
}

func GenerateXAML(analysis Analysis, themePrefix string, patternComments bool) GeneratedArtifactReport {
	var b strings.Builder
	b.WriteString("<Grid xmlns=\"http://schemas.microsoft.com/winfx/2006/xaml/presentation\">\n")
	for _, child := range analysis.Layout.Children {
		writeXAMLNode(&b, child, 1, themePrefix, patternComments)
	}
	b.WriteString("</Grid>\n")
	diagnostics := generatorDiagnostics(analysis)
	return GeneratedArtifactReport{"1", generatedStatus(diagnostics), analysis.SourcePath, sourceDialect(analysis), "winui3", analysis.RootView, analysis.Component, "xaml", b.String(), diagnostics}
}

func GenerateSwiftUI(analysis Analysis, viewName string) GeneratedArtifactReport {
	viewName = sanitizeIdentifier(firstNonEmpty(viewName, analysis.Component, "GeneratedView"))
	var b strings.Builder
	b.WriteString("import SwiftUI\n\n")
	fmt.Fprintf(&b, "struct %s: View {\n", viewName)
	b.WriteString("    var body: some View {\n")
	if len(analysis.Layout.Children) == 0 {
		b.WriteString("        EmptyView()\n")
	} else if len(analysis.Layout.Children) == 1 {
		writeSwiftUINode(&b, analysis.Layout.Children[0], 2)
	} else {
		b.WriteString("        VStack {\n")
		for _, child := range analysis.Layout.Children {
			writeSwiftUINode(&b, child, 3)
		}
		b.WriteString("        }\n")
	}
	b.WriteString("    }\n")
	b.WriteString("}\n")
	diagnostics := generatorDiagnostics(analysis)
	return GeneratedArtifactReport{"1", generatedStatus(diagnostics), analysis.SourcePath, sourceDialect(analysis), "swiftui", analysis.RootView, analysis.Component, "swift", b.String(), diagnostics}
}

func GenerateContracts(analysis Analysis, target string) ContractReport {
	items := []ContractItem{}
	var walk func(Node, string)
	walk = func(node Node, path string) {
		contracts := contractsFor(node)
		policies := policiesFor(node)
		notes := []string{}
		if node.Properties["componentBoundary"] != "" {
			notes = append(notes, "preserved component boundary requires human review")
		}
		if len(contracts) > 0 || len(policies) > 0 || len(notes) > 0 {
			items = append(items, ContractItem{Path: path, Kind: node.Kind, Expression: node.Expression, Contracts: nonNilStrings(contracts), Policies: nonNilStrings(policies), SourceNotes: nonNilStrings(notes)})
		}
		counts := map[NodeKind]int{}
		for _, child := range node.Children {
			index := counts[child.Kind]
			counts[child.Kind]++
			walk(child, path+"/"+indexedPathPart(child.Kind, index))
		}
	}
	counts := map[NodeKind]int{}
	for _, child := range analysis.Layout.Children {
		index := counts[child.Kind]
		counts[child.Kind]++
		walk(child, indexedPathPart(child.Kind, index))
	}
	diagnostics := generatorDiagnostics(analysis)
	status := generatedStatus(diagnostics)
	if status == "ok" && len(items) > 0 {
		status = "review"
	}
	return ContractReport{"1", status, analysis.SourcePath, firstNonEmpty(target, defaultTransferTarget(sourceDialect(analysis))), analysis.RootView, analysis.Component, items, diagnostics}
}

func writeXAMLNode(b *strings.Builder, node Node, indent int, themePrefix string, patternComments bool) {
	pad := strings.Repeat("  ", indent)
	if patternComments {
		fmt.Fprintf(b, "%s<!-- loom %s / %s -->\n", pad, node.Kind, xmlEscape(node.Expression))
	}
	switch node.Kind {
	case KindVerticalStack:
		fmt.Fprintf(b, "%s<StackPanel>\n", pad)
		writeXAMLChildren(b, node, indent+1, themePrefix, patternComments)
		fmt.Fprintf(b, "%s</StackPanel>\n", pad)
	case KindHorizontalStack:
		fmt.Fprintf(b, "%s<StackPanel Orientation=\"Horizontal\">\n", pad)
		writeXAMLChildren(b, node, indent+1, themePrefix, patternComments)
		fmt.Fprintf(b, "%s</StackPanel>\n", pad)
	case KindOverlayStack, KindGrid, KindRoot:
		fmt.Fprintf(b, "%s<Grid>\n", pad)
		writeXAMLChildren(b, node, indent+1, themePrefix, patternComments)
		fmt.Fprintf(b, "%s</Grid>\n", pad)
	case KindScrollView:
		fmt.Fprintf(b, "%s<ScrollViewer>\n", pad)
		writeXAMLChildren(b, node, indent+1, themePrefix, patternComments)
		fmt.Fprintf(b, "%s</ScrollViewer>\n", pad)
	case KindList:
		fmt.Fprintf(b, "%s<ListView>\n", pad)
		writeXAMLChildren(b, node, indent+1, themePrefix, patternComments)
		fmt.Fprintf(b, "%s</ListView>\n", pad)
	case KindText:
		fmt.Fprintf(b, "%s<TextBlock Text=\"%s\" />\n", pad, xmlEscape(firstNonEmpty(node.VisibleLabel, unquote(node.Arguments), node.Expression)))
	case KindButton:
		fmt.Fprintf(b, "%s<Button Content=\"%s\" />\n", pad, xmlEscape(firstNonEmpty(node.VisibleLabel, unquote(node.Arguments), "Button")))
	case KindTextField:
		fmt.Fprintf(b, "%s<TextBox Header=\"%s\" PlaceholderText=\"%s\" />\n", pad, xmlEscape(firstNonEmpty(node.VisibleLabel, node.AccessibleName, "Text")), xmlEscape(node.Placeholder))
	case KindImage:
		fmt.Fprintf(b, "%s<Image Source=\"%s\" AutomationProperties.Name=\"%s\" />\n", pad, xmlEscape(node.Resource), xmlEscape(firstNonEmpty(node.AccessibleName, node.VisibleLabel, "Image")))
	case KindSlider:
		fmt.Fprintf(b, "%s<Slider />\n", pad)
	case KindToggle:
		fmt.Fprintf(b, "%s<CheckBox Content=\"%s\" />\n", pad, xmlEscape(firstNonEmpty(node.VisibleLabel, unquote(node.Arguments), "Option")))
	case KindSpacer:
		fmt.Fprintf(b, "%s<Border />\n", pad)
	case KindDivider:
		fmt.Fprintf(b, "%s<Rectangle Height=\"1\" />\n", pad)
	case KindColor:
		fmt.Fprintf(b, "%s<Border Background=\"%s\" />\n", pad, xmlEscape(firstNonEmpty(node.Resource, node.Arguments, "{ThemeResource "+themePrefix+"SurfaceBrush}")))
	case KindComponent, KindUnsupported:
		fmt.Fprintf(b, "%s<!-- Unsupported component boundary: %s -->\n", pad, xmlEscape(node.Expression))
		fmt.Fprintf(b, "%s<ContentControl />\n", pad)
	default:
		fmt.Fprintf(b, "%s<!-- Unsupported layout node: %s / %s -->\n", pad, node.Kind, xmlEscape(node.Expression))
	}
}

func writeXAMLChildren(b *strings.Builder, node Node, indent int, themePrefix string, patternComments bool) {
	for _, child := range node.Children {
		writeXAMLNode(b, child, indent, themePrefix, patternComments)
	}
}

func writeSwiftUINode(b *strings.Builder, node Node, indent int) {
	pad := strings.Repeat("    ", indent)
	switch node.Kind {
	case KindVerticalStack:
		fmt.Fprintf(b, "%sVStack {\n", pad)
		writeSwiftUIChildren(b, node, indent+1)
		fmt.Fprintf(b, "%s}\n", pad)
	case KindHorizontalStack:
		fmt.Fprintf(b, "%sHStack {\n", pad)
		writeSwiftUIChildren(b, node, indent+1)
		fmt.Fprintf(b, "%s}\n", pad)
	case KindOverlayStack, KindGrid, KindRoot:
		fmt.Fprintf(b, "%sZStack {\n", pad)
		writeSwiftUIChildren(b, node, indent+1)
		fmt.Fprintf(b, "%s}\n", pad)
	case KindScrollView:
		fmt.Fprintf(b, "%sScrollView {\n", pad)
		writeSwiftUIChildren(b, node, indent+1)
		fmt.Fprintf(b, "%s}\n", pad)
	case KindList:
		fmt.Fprintf(b, "%sList {\n", pad)
		writeSwiftUIChildren(b, node, indent+1)
		fmt.Fprintf(b, "%s}\n", pad)
	case KindText:
		fmt.Fprintf(b, "%sText(%q)\n", pad, firstNonEmpty(node.VisibleLabel, unquote(node.Arguments), node.Expression))
	case KindButton:
		fmt.Fprintf(b, "%sButton(%q) {}\n", pad, firstNonEmpty(node.VisibleLabel, unquote(node.Arguments), "Button"))
	case KindTextField:
		fmt.Fprintf(b, "%sTextField(%q, text: .constant(\"\"))\n", pad, firstNonEmpty(node.Placeholder, node.VisibleLabel, node.AccessibleName, "Text"))
	case KindImage:
		fmt.Fprintf(b, "%sImage(%q)\n", pad, firstNonEmpty(node.Resource, node.AccessibleName, "image"))
	case KindSlider:
		fmt.Fprintf(b, "%sSlider(value: .constant(0))\n", pad)
	case KindToggle:
		fmt.Fprintf(b, "%sToggle(%q, isOn: .constant(false))\n", pad, firstNonEmpty(node.VisibleLabel, unquote(node.Arguments), "Option"))
	case KindSpacer:
		fmt.Fprintf(b, "%sSpacer()\n", pad)
	case KindDivider:
		fmt.Fprintf(b, "%sDivider()\n", pad)
	case KindColor:
		fmt.Fprintf(b, "%sColor.clear\n", pad)
	case KindComponent, KindUnsupported:
		fmt.Fprintf(b, "%s// Unsupported component boundary: %s\n", pad, node.Expression)
		fmt.Fprintf(b, "%sEmptyView()\n", pad)
	default:
		fmt.Fprintf(b, "%s// Unsupported layout node: %s / %s\n", pad, node.Kind, node.Expression)
		fmt.Fprintf(b, "%sEmptyView()\n", pad)
	}
}

func writeSwiftUIChildren(b *strings.Builder, node Node, indent int) {
	if len(node.Children) == 0 {
		fmt.Fprintf(b, "%sEmptyView()\n", strings.Repeat("    ", indent))
		return
	}
	for _, child := range node.Children {
		writeSwiftUINode(b, child, indent)
	}
}

func generatorDiagnostics(analysis Analysis) []Diagnostic {
	diagnostics := nonNilDiagnostics(analysis.Diagnostics)
	var walk func(Node)
	walk = func(node Node) {
		if node.Kind == KindComponent || node.Kind == KindUnsupported {
			diagnostics = append(diagnostics, Diagnostic{Severity: SeverityWarning, Code: "GENERATE.COMPONENT_BOUNDARY", Message: "Component boundary requires review before generated output is considered complete: " + node.Expression + "."})
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(analysis.Layout)
	return diagnostics
}

func generatedStatus(diagnostics []Diagnostic) string {
	status := "ok"
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == SeverityError {
			return "error"
		}
		if diagnostic.Severity == SeverityWarning {
			status = "review"
		}
	}
	return status
}

func sourceDialect(analysis Analysis) string {
	if analysis.Layout.Properties != nil && analysis.Layout.Properties["sourceDialect"] != "" {
		return analysis.Layout.Properties["sourceDialect"]
	}
	return InferSourcePlatform(analysis.SourcePath)
}

func xmlEscape(value string) string {
	value = strings.ReplaceAll(value, "&", "&amp;")
	value = strings.ReplaceAll(value, "<", "&lt;")
	value = strings.ReplaceAll(value, ">", "&gt;")
	value = strings.ReplaceAll(value, "\"", "&quot;")
	return value
}

func unquote(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
		return value[1 : len(value)-1]
	}
	return value
}

func sanitizeIdentifier(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "GeneratedView"
	}
	var b strings.Builder
	for i, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || (i > 0 && r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "GeneratedView"
	}
	return b.String()
}

func ReplaceOwnedRegion(path, generated, regionID string, initRegion, overwrite bool) error {
	regionID = firstNonEmpty(regionID, "default")
	begin := fmt.Sprintf("<!-- loom:begin %s -->", regionID)
	end := fmt.Sprintf("<!-- loom:end %s -->", regionID)
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) || !initRegion {
			return err
		}
		data = []byte("<Grid xmlns=\"http://schemas.microsoft.com/winfx/2006/xaml/presentation\">\n  " + begin + "\n  " + end + "\n</Grid>\n")
	} else if !overwrite {
		return fmt.Errorf("refusing to replace existing region in %s without --overwrite", path)
	}
	text := string(data)
	start := strings.Index(text, begin)
	stop := strings.Index(text, end)
	if start < 0 || stop < 0 || stop < start {
		if !initRegion {
			return fmt.Errorf("owned region %s not found in %s", regionID, path)
		}
		text = strings.TrimRight(text, "\n") + "\n" + begin + "\n" + end + "\n"
		start = strings.Index(text, begin)
		stop = strings.Index(text, end)
	}
	replacement := begin + "\n" + strings.TrimRight(generated, "\n") + "\n" + end
	text = text[:start] + replacement + text[stop+len(end):]
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(text), 0644)
}

func ContractsText(report ContractReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "loom contracts\nstatus: %s\nsource: %s\ntarget: %s\ncontracts: %d\n", report.Status, report.SourcePath, report.Target, len(report.Contracts))
	for _, item := range report.Contracts {
		fmt.Fprintf(&b, "- %s %s / %s\n", item.Path, item.Kind, item.Expression)
		if len(item.Contracts) > 0 {
			fmt.Fprintf(&b, "  contracts: %s\n", strings.Join(item.Contracts, ", "))
		}
		if len(item.Policies) > 0 {
			fmt.Fprintf(&b, "  policies: %s\n", strings.Join(item.Policies, "; "))
		}
	}
	return b.String()
}
