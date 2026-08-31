package loom

import (
	"fmt"
	"strconv"
	"strings"
)

type AuditCategory string

const (
	AuditAccessibility AuditCategory = "accessibility"
	AuditLayout        AuditCategory = "layout"
	AuditDesign        AuditCategory = "design"
	AuditMalformed     AuditCategory = "malformed"
	AuditRedundant     AuditCategory = "redundant"
	AuditMissing       AuditCategory = "missing"
)

type AuditFinding struct {
	Severity       DiagnosticSeverity `json:"severity"`
	Category       AuditCategory      `json:"category"`
	Code           string             `json:"code"`
	Path           string             `json:"path"`
	Kind           NodeKind           `json:"kind"`
	Message        string             `json:"message"`
	Recommendation string             `json:"recommendation"`
	SuggestedFixes []SuggestedFix     `json:"suggested_fixes"`
}

type AuditSummary struct {
	Errors        int `json:"errors"`
	Warnings      int `json:"warnings"`
	Info          int `json:"info"`
	Accessibility int `json:"accessibility"`
	Layout        int `json:"layout"`
	Design        int `json:"design"`
	Malformed     int `json:"malformed"`
	Redundant     int `json:"redundant"`
	Missing       int `json:"missing"`
}

type AuditReport struct {
	SchemaVersion string         `json:"schema_version"`
	Status        string         `json:"status"`
	SourcePath    string         `json:"sourcePath"`
	RootView      string         `json:"rootView"`
	Component     string         `json:"component"`
	Summary       AuditSummary   `json:"summary"`
	Findings      []AuditFinding `json:"findings"`
	Diagnostics   []Diagnostic   `json:"diagnostics"`
}

func Audit(analysis Analysis) AuditReport {
	findings := []AuditFinding{}
	var walk func(Node, string, int, []NodeKind)
	walk = func(node Node, path string, depth int, ancestors []NodeKind) {
		if node.Properties["componentBoundary"] == "native-winui-control" {
			findings = append(findings, auditFinding(SeverityWarning, AuditDesign, "AUDIT070", path, node.Kind, "native WinUI control is preserved as an unsupported component boundary.", "keep it as handwritten native UI, add a project-specific pattern mapping, or declare its transfer contract explicitly."))
		}
		if isContainer(node.Kind) && len(node.Children) == 0 {
			findings = append(findings, auditFinding(SeverityWarning, AuditMissing, "AUDIT011", path, node.Kind, "Container has no visible children.", "Remove the empty container or add an explicit placeholder/empty-state element."))
		}
		if isContainer(node.Kind) && len(node.Children) == 1 && len(node.Modifiers) == 0 && node.Kind != KindScrollView && node.Kind != KindGeometryReader && node.Kind != KindConditional {
			findings = append(findings, auditFinding(SeverityInfo, AuditRedundant, "AUDIT012", path, node.Kind, "Container wraps a single child without modifiers.", "Consider removing the wrapper unless it exists for a named semantic grouping."))
		}
		if len(ancestors) > 0 && ancestors[len(ancestors)-1] == node.Kind && (node.Kind == KindVerticalStack || node.Kind == KindHorizontalStack || node.Kind == KindOverlayStack || node.Kind == KindGrid) {
			findings = append(findings, auditFinding(SeverityWarning, AuditRedundant, "AUDIT013", path, node.Kind, "Nested layout repeats the parent layout kind without a local policy.", "Collapse the nested layout or add explicit spacing/alignment semantics."))
		}
		switch node.Kind {
		case KindButton:
			if accessibleName(node) == "" {
				findings = append(findings, auditFinding(SeverityError, AuditAccessibility, "AUDIT020", path, node.Kind, "Button has no detectable accessible name.", "Add visible text, an accessibility label, or a target AutomationProperties.Name."))
			}
		case KindImage:
			if !hasAccessibility(node) {
				findings = append(findings, auditFinding(SeverityWarning, AuditAccessibility, "AUDIT030", path, node.Kind, "Image has no accessible label or decorative-hidden intent.", "Add an accessibility label for meaningful images or mark decorative images as hidden."))
			}
		case KindTextField:
			if accessibleName(node) == "" {
				findings = append(findings, auditFinding(SeverityError, AuditAccessibility, "AUDIT031", path, node.Kind, "Text input has no label, placeholder, or accessible name.", "Provide a stable label or target accessibility name."))
			}
		case KindColor:
			findings = append(findings, auditFinding(SeverityInfo, AuditAccessibility, "AUDIT040", path, node.Kind, "Color surface has no inherent accessible meaning.", "Ensure color is decorative or that meaning is also conveyed by text, icon label, or state."))
		case KindGeometryReader:
			findings = append(findings, auditFinding(SeverityWarning, AuditLayout, "AUDIT050", path, node.Kind, "Geometry-dependent layout may not transfer deterministically across platforms.", "Replace implicit geometry dependency with named size-class or breakpoint policy."))
		case KindScrollView:
			for _, ancestor := range ancestors {
				if ancestor == KindScrollView {
					findings = append(findings, auditFinding(SeverityWarning, AuditLayout, "AUDIT052", path, node.Kind, "Nested scroll region can create ambiguous gesture and keyboard navigation behavior.", "Avoid nested scrolling or define axis ownership and focus movement policy."))
					break
				}
			}
		}
		auditTargetSize(node, path, &findings)
		for _, child := range node.Children {
			walk(child, path+"/"+string(child.Kind), depth+1, append(ancestors, node.Kind))
		}
	}
	for _, child := range analysis.Layout.Children {
		walk(child, string(child.Kind), 1, nil)
	}
	summary := summarizeAudit(findings)
	status := "ok"
	if summary.Errors > 0 {
		status = "error"
	} else if summary.Warnings > 0 {
		status = "warning"
	}
	return AuditReport{"1", status, analysis.SourcePath, analysis.RootView, analysis.Component, summary, findings, nonNilDiagnostics(analysis.Diagnostics)}
}

func AuditText(report AuditReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "loom accessibility audit\nstatus: %s\nsource: %s\nview: %s.%s\n\nfindings: %d\n", report.Status, report.SourcePath, report.RootView, report.Component, len(report.Findings))
	if len(report.Findings) == 0 {
		b.WriteString("  none\n")
	}
	for _, finding := range report.Findings {
		fmt.Fprintf(&b, "[%s] %s %s %s\n  issue: %s\n  fix: %s\n", finding.Severity, finding.Code, finding.Path, finding.Category, finding.Message, finding.Recommendation)
		if len(finding.SuggestedFixes) > 0 {
			b.WriteString("  suggested fixes:\n")
			for _, fix := range finding.SuggestedFixes {
				fmt.Fprintf(&b, "    - %s: %s - %s\n", fix.Audience, fix.Action, fix.Detail)
			}
		}
	}
	return b.String()
}

func auditFinding(severity DiagnosticSeverity, category AuditCategory, code string, path string, kind NodeKind, message string, recommendation string) AuditFinding {
	return AuditFinding{severity, category, code, path, kind, message, recommendation, auditSuggestedFixes(code, recommendation)}
}

func auditSuggestedFixes(code, recommendation string) []SuggestedFix {
	switch code {
	case "AUDIT070":
		return []SuggestedFix{{FixUser, "choose native-boundary strategy", "keep native, replace with portable layout, or approve a pattern mapping.", ""}, {FixAgent, "declare native component boundary", recommendation, "loom patterns:transfer <source.xaml> --from winui3 --to macos --format json"}}
	case "AUDIT020":
		return []SuggestedFix{{FixUser, "Name the button", recommendation, ""}, {FixAgent, "Add accessible-name metadata", "Prefer visible label text or AutomationProperties.Name.", ""}}
	default:
		return []SuggestedFix{{FixUser, "Review finding", recommendation, ""}, {FixAgent, "Apply recommendation", recommendation, ""}}
	}
}

func summarizeAudit(findings []AuditFinding) AuditSummary {
	var summary AuditSummary
	for _, finding := range findings {
		switch finding.Severity {
		case SeverityError:
			summary.Errors++
		case SeverityWarning:
			summary.Warnings++
		default:
			summary.Info++
		}
		switch finding.Category {
		case AuditAccessibility:
			summary.Accessibility++
		case AuditLayout:
			summary.Layout++
		case AuditDesign:
			summary.Design++
		case AuditMalformed:
			summary.Malformed++
		case AuditRedundant:
			summary.Redundant++
		case AuditMissing:
			summary.Missing++
		}
	}
	return summary
}

func isContainer(kind NodeKind) bool {
	switch kind {
	case KindGeometryReader, KindVerticalStack, KindHorizontalStack, KindOverlayStack, KindSplitView, KindGrid, KindScrollView, KindList, KindConditional, KindLoop:
		return true
	default:
		return false
	}
}

func accessibleName(node Node) string {
	if value := strings.Trim(node.Arguments, `"`); value != "" {
		return value
	}
	for _, child := range node.Children {
		if child.Kind == KindText {
			if value := strings.Trim(child.Arguments, `"`); value != "" {
				return value
			}
		}
	}
	return firstNonEmpty(node.Properties["xaml.Name"], node.Properties["xaml.AutomationId"], node.Properties["xaml.AutomationProperties.Name"])
}

func hasAccessibility(node Node) bool {
	if accessibleName(node) != "" {
		return true
	}
	for _, modifier := range node.Modifiers {
		if strings.HasPrefix(modifier.Name, "accessibility") {
			return true
		}
	}
	return false
}

func auditTargetSize(node Node, path string, findings *[]AuditFinding) {
	if node.Kind != KindButton && node.Kind != KindToggle && node.Kind != KindSlider && node.Kind != KindTextField {
		return
	}
	for _, modifier := range node.Modifiers {
		if modifier.Name != "frame" {
			continue
		}
		for _, part := range strings.Split(modifier.Arguments, ",") {
			pieces := strings.SplitN(part, ":", 2)
			if len(pieces) != 2 {
				continue
			}
			label := strings.TrimSpace(pieces[0])
			value, err := strconv.ParseFloat(strings.TrimSpace(pieces[1]), 64)
			if err == nil && value > 0 && value < 44 && (label == "width" || label == "height" || label == "minWidth" || label == "minHeight") {
				*findings = append(*findings, auditFinding(SeverityWarning, AuditAccessibility, "AUDIT060", path, node.Kind, "Interactive control target is below the 44-unit minimum target heuristic.", "Increase the effective hit target to at least 44 by 44."))
			}
		}
	}
}
