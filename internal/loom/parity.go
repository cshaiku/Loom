package loom

import (
	"fmt"
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
	sourceKinds := flattenedKinds(sourceAnalysis.Layout)
	targetKinds := flattenedKinds(targetAnalysis.Layout)
	findings := compareKindSequences(sourceKinds, targetKinds)
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
	return ParityReport{"1", status, source, target, sourceAnalysis.Layout.Properties["sourceDialect"], targetAnalysis.Layout.Properties["sourceDialect"], len(sourceKinds), len(targetKinds), findings}, nil
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
