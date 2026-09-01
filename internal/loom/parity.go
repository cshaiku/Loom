package loom

import (
	"fmt"
	"reflect"
	"strings"
)

type ParityFinding struct {
	Severity DiagnosticSeverity `json:"severity"`
	Code     string             `json:"code"`
	Path     string             `json:"path"`
	Message  string             `json:"message"`
}

type ParityReport struct {
	SchemaVersion string          `json:"schema_version"`
	Status        string          `json:"status"`
	SourcePath    string          `json:"sourcePath"`
	TargetPath    string          `json:"targetPath"`
	SourceDialect string          `json:"sourceDialect"`
	TargetDialect string          `json:"targetDialect"`
	SourceCount   int             `json:"sourceCount"`
	TargetCount   int             `json:"targetCount"`
	Findings      []ParityFinding `json:"findings"`
}

func InspectParity(source, target, sourcePlatform, targetPlatform string) (ParityReport, error) {
	sourceAnalysis, err := AnalyzeByPlatform(source, sourcePlatform)
	if err != nil {
		return ParityReport{}, err
	}
	targetAnalysis, err := AnalyzeByPlatform(target, targetPlatform)
	if err != nil {
		return ParityReport{}, err
	}
	sourceNodes := parityEntries(sourceAnalysis.Layout)
	targetNodes := parityEntries(targetAnalysis.Layout)
	findings := diagnosticParityFindings("source", sourceAnalysis.Diagnostics)
	findings = append(findings, diagnosticParityFindings("target", targetAnalysis.Diagnostics)...)
	findings = append(findings, compareParityEntries(sourceNodes, targetNodes)...)
	status := "ok"
	for _, finding := range findings {
		if finding.Severity == SeverityError {
			status = "error"
			break
		}
		if finding.Severity == SeverityWarning {
			status = "warning"
		}
	}
	return ParityReport{"1", status, source, target, sourceAnalysis.Layout.Properties["sourceDialect"], targetAnalysis.Layout.Properties["sourceDialect"], len(sourceNodes), len(targetNodes), findings}, nil
}

func flattenedKinds(root Node) []NodeKind {
	out := []NodeKind{}
	stack := append([]Node{}, root.Children...)
	for i, j := 0, len(stack)-1; i < j; i, j = i+1, j-1 {
		stack[i], stack[j] = stack[j], stack[i]
	}
	for len(stack) > 0 {
		last := len(stack) - 1
		node := stack[last]
		stack = stack[:last]
		out = append(out, node.Kind)
		for i := len(node.Children) - 1; i >= 0; i-- {
			stack = append(stack, node.Children[i])
		}
	}
	return out
}

func compareKindSequences(source, target []NodeKind) []ParityFinding {
	findings := []ParityFinding{}
	if len(source) != len(target) {
		findings = append(findings, ParityFinding{Severity: SeverityWarning, Code: "PARITY.COUNT", Path: "/", Message: fmt.Sprintf("source has %d layout nodes; target has %d.", len(source), len(target))})
	}
	limit := len(source)
	if len(target) < limit {
		limit = len(target)
	}
	for i := 0; i < limit; i++ {
		if source[i] != target[i] {
			findings = append(findings, ParityFinding{Severity: SeverityWarning, Code: "PARITY.KIND", Path: fmt.Sprintf("/%d", i), Message: fmt.Sprintf("source kind %s differs from target kind %s.", source[i], target[i])})
		}
	}
	if len(findings) == 0 {
		return []ParityFinding{}
	}
	return findings
}

type parityEntry struct {
	Path           string
	Kind           NodeKind
	ChildCount     int
	Arguments      string
	VisibleLabel   string
	AccessibleName string
	Identifier     string
	Placeholder    string
	Resource       string
	Decorative     bool
	Properties     map[string]string
}

func parityEntries(root Node) []parityEntry {
	entries := []parityEntry{}
	var walk func(Node, string)
	walk = func(node Node, path string) {
		entries = append(entries, parityEntry{
			Path:           path,
			Kind:           node.Kind,
			ChildCount:     len(node.Children),
			Arguments:      node.Arguments,
			VisibleLabel:   node.VisibleLabel,
			AccessibleName: node.AccessibleName,
			Identifier:     node.Identifier,
			Placeholder:    node.Placeholder,
			Resource:       node.Resource,
			Decorative:     node.Decorative,
			Properties:     parityProperties(node.Properties),
		})
		siblingCounts := map[NodeKind]int{}
		for _, child := range node.Children {
			index := siblingCounts[child.Kind]
			siblingCounts[child.Kind]++
			walk(child, path+"/"+indexedPathPart(child.Kind, index))
		}
	}
	siblingCounts := map[NodeKind]int{}
	for _, child := range root.Children {
		index := siblingCounts[child.Kind]
		siblingCounts[child.Kind]++
		walk(child, "/"+indexedPathPart(child.Kind, index))
	}
	return entries
}

func compareParityEntries(source, target []parityEntry) []ParityFinding {
	findings := []ParityFinding{}
	if len(source) != len(target) {
		findings = append(findings, ParityFinding{Severity: SeverityWarning, Code: "PARITY.COUNT", Path: "/", Message: fmt.Sprintf("source has %d layout nodes; target has %d.", len(source), len(target))})
	}
	targetByPath := map[string]parityEntry{}
	for _, entry := range target {
		targetByPath[entry.Path] = entry
	}
	for _, sourceEntry := range source {
		targetEntry, ok := targetByPath[sourceEntry.Path]
		if !ok {
			findings = append(findings, ParityFinding{Severity: SeverityWarning, Code: "PARITY.PATH", Path: sourceEntry.Path, Message: "source node has no matching target node at the same tree path."})
			continue
		}
		if sourceEntry.Kind != targetEntry.Kind {
			findings = append(findings, ParityFinding{Severity: SeverityWarning, Code: "PARITY.KIND", Path: sourceEntry.Path, Message: fmt.Sprintf("source kind %s differs from target kind %s.", sourceEntry.Kind, targetEntry.Kind)})
		}
		if sourceEntry.ChildCount != targetEntry.ChildCount {
			findings = append(findings, ParityFinding{Severity: SeverityWarning, Code: "PARITY.CHILDREN", Path: sourceEntry.Path, Message: fmt.Sprintf("source has %d children; target has %d.", sourceEntry.ChildCount, targetEntry.ChildCount)})
		}
		if !reflect.DeepEqual(sourceEntry.Properties, targetEntry.Properties) {
			findings = append(findings, ParityFinding{Severity: SeverityWarning, Code: "PARITY.PROPERTIES", Path: sourceEntry.Path, Message: "semantic properties differ."})
		}
		if sourceEntry.VisibleLabel != targetEntry.VisibleLabel || sourceEntry.AccessibleName != targetEntry.AccessibleName || sourceEntry.Placeholder != targetEntry.Placeholder || sourceEntry.Resource != targetEntry.Resource || sourceEntry.Decorative != targetEntry.Decorative {
			findings = append(findings, ParityFinding{Severity: SeverityWarning, Code: "PARITY.SEMANTICS", Path: sourceEntry.Path, Message: "semantic label, accessibility, placeholder, resource, or decorative metadata differs."})
		}
	}
	sourceByPath := map[string]bool{}
	for _, entry := range source {
		sourceByPath[entry.Path] = true
	}
	for _, targetEntry := range target {
		if !sourceByPath[targetEntry.Path] {
			findings = append(findings, ParityFinding{Severity: SeverityWarning, Code: "PARITY.PATH", Path: targetEntry.Path, Message: "target node has no matching source node at the same tree path."})
		}
	}
	return findings
}

func diagnosticParityFindings(side string, diagnostics []Diagnostic) []ParityFinding {
	findings := []ParityFinding{}
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == SeverityError {
			findings = append(findings, ParityFinding{Severity: SeverityError, Code: "PARITY.INVALID_SOURCE", Path: "/", Message: fmt.Sprintf("%s analysis has error diagnostic %s: %s", side, diagnostic.Code, diagnostic.Message)})
		}
	}
	return findings
}

func parityProperties(properties map[string]string) map[string]string {
	out := map[string]string{}
	for _, key := range []string{"componentBoundary", "requiresNativeImplementation", "xaml.Grid.RowDefinitions", "xaml.Grid.ColumnDefinitions"} {
		if value := properties[key]; value != "" {
			out[key] = value
		}
	}
	return out
}

func ParityText(report ParityReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "loom layout parity\nstatus: %s\nsource: %s (%s)\ntarget: %s (%s)\nsource nodes: %d\ntarget nodes: %d\nfindings: %d\n", report.Status, report.SourcePath, report.SourceDialect, report.TargetPath, report.TargetDialect, report.SourceCount, report.TargetCount, len(report.Findings))
	if len(report.Findings) == 0 {
		b.WriteString("  none\n")
		return b.String()
	}
	for _, finding := range report.Findings {
		fmt.Fprintf(&b, "  [%s] %s %s: %s\n", finding.Severity, finding.Code, finding.Path, finding.Message)
	}
	return b.String()
}
