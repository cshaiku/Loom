package loom

import (
	"fmt"
	"strings"
)

type TransferDisposition string

const (
	TransferDirect              TransferDisposition = "direct"
	TransferNeedsPolicy         TransferDisposition = "needs-policy"
	TransferNeedsNativeContract TransferDisposition = "needs-native-contract"
	TransferLossy               TransferDisposition = "lossy"
	TransferUnsupported         TransferDisposition = "unsupported"
)

type TransferItem struct {
	Path             string              `json:"path"`
	Kind             NodeKind            `json:"kind"`
	Expression       string              `json:"expression"`
	PatternID        string              `json:"patternID,omitempty"`
	PatternName      string              `json:"patternName,omitempty"`
	Disposition      TransferDisposition `json:"disposition"`
	Reason           string              `json:"reason"`
	SourceConstructs []string            `json:"sourceConstructs"`
	TargetConstructs []string            `json:"targetConstructs"`
	Contracts        []string            `json:"contracts"`
	Policies         []string            `json:"policies"`
}

type TransferSummary struct {
	Direct              int `json:"direct"`
	NeedsPolicy         int `json:"needs_policy"`
	NeedsNativeContract int `json:"needs_native_contract"`
	Lossy               int `json:"lossy"`
	Unsupported         int `json:"unsupported"`
}

type TransferReport struct {
	SchemaVersion string          `json:"schema_version"`
	Status        string          `json:"status"`
	SourcePath    string          `json:"sourcePath"`
	From          string          `json:"from"`
	To            string          `json:"to"`
	RootView      string          `json:"rootView"`
	Component     string          `json:"component"`
	ASCIIPattern  string          `json:"asciiPattern"`
	Summary       TransferSummary `json:"summary"`
	Items         []TransferItem  `json:"items"`
	Diagnostics   []Diagnostic    `json:"diagnostics"`
}

func Transfer(analysis Analysis, patterns []Pattern, from, to string) TransferReport {
	byKind := map[NodeKind]Pattern{}
	for _, pattern := range patterns {
		byKind[pattern.Kind] = pattern
	}
	items := []TransferItem{}
	var walk func(Node, string)
	walk = func(node Node, path string) {
		items = append(items, transferItem(node, path, byKind, from, to))
		siblingCounts := map[NodeKind]int{}
		for _, child := range node.Children {
			index := siblingCounts[child.Kind]
			siblingCounts[child.Kind]++
			walk(child, path+"/"+indexedPathPart(child.Kind, index))
		}
	}
	siblingCounts := map[NodeKind]int{}
	for _, child := range analysis.Layout.Children {
		index := siblingCounts[child.Kind]
		siblingCounts[child.Kind]++
		walk(child, indexedPathPart(child.Kind, index))
	}
	summary := TransferSummary{}
	for _, item := range items {
		switch item.Disposition {
		case TransferDirect:
			summary.Direct++
		case TransferNeedsPolicy:
			summary.NeedsPolicy++
		case TransferNeedsNativeContract:
			summary.NeedsNativeContract++
		case TransferLossy:
			summary.Lossy++
		case TransferUnsupported:
			summary.Unsupported++
		}
	}
	status := "ok"
	for _, diagnostic := range analysis.Diagnostics {
		if diagnostic.Severity == SeverityError {
			status = "source-invalid"
			break
		}
	}
	if status == "ok" && (summary.Unsupported > 0 || summary.NeedsNativeContract > 0 || summary.NeedsPolicy > 0 || summary.Lossy > 0) {
		status = "partial"
	}
	return TransferReport{"1", status, analysis.SourcePath, from, to, analysis.RootView, analysis.Component, ASCIIAnalysis(analysis), summary, items, nonNilDiagnostics(analysis.Diagnostics)}
}

func transferItem(node Node, path string, patterns map[NodeKind]Pattern, from, to string) TransferItem {
	pattern := patterns[node.Kind]
	source := mapping(pattern, from)
	target := mapping(pattern, to)
	contracts := contractsFor(node)
	policies := policiesFor(node)
	disposition := TransferDirect
	reason := "the source pattern has a target mapping and no additional transfer risk was detected."
	if node.Properties["componentBoundary"] == "native-winui-control" {
		disposition = TransferUnsupported
		reason = "native WinUI control was preserved as an unsupported component boundary and needs an explicit target mapping or handwritten implementation."
	} else if pattern.ID == "" || len(target.Constructs) == 0 {
		disposition = TransferUnsupported
		reason = "no target pattern mapping exists, or the node is explicitly unsupported."
	} else if len(contracts) > 0 {
		disposition = TransferNeedsNativeContract
		reason = "the visual element transfers, but behavior, state, or accessibility wiring must remain native."
	} else if len(policies) > 0 {
		disposition = TransferNeedsPolicy
		reason = "the element transfers after project policy decisions such as sizing, spacing, or token selection."
	}
	return TransferItem{
		Path:             path,
		Kind:             node.Kind,
		Expression:       node.Expression,
		PatternID:        pattern.ID,
		PatternName:      pattern.Name,
		Disposition:      disposition,
		Reason:           reason,
		SourceConstructs: nonNilStrings(source.Constructs),
		TargetConstructs: nonNilStrings(target.Constructs),
		Contracts:        nonNilStrings(contracts),
		Policies:         nonNilStrings(policies),
	}
}

func TransferText(report TransferReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "loom pattern transfer\nstatus: %s\nsource: %s\nroute: %s -> %s\nview: %s.%s\n\nsummary\n  direct: %d\n  needs-policy: %d\n  needs-native-contract: %d\n  lossy: %d\n  unsupported: %d\n\nascii pattern\n%s\ntransfer items\n", report.Status, report.SourcePath, report.From, report.To, report.RootView, report.Component, report.Summary.Direct, report.Summary.NeedsPolicy, report.Summary.NeedsNativeContract, report.Summary.Lossy, report.Summary.Unsupported, report.ASCIIPattern)
	for _, item := range report.Items {
		fmt.Fprintf(&b, "[%s] %s %s pattern=%s\n  target: %s\n  reason: %s\n", item.Disposition, item.Path, item.Kind, item.PatternID, strings.Join(item.TargetConstructs, ", "), item.Reason)
	}
	return b.String()
}

func mapping(pattern Pattern, platform string) PatternMapping {
	platform = canonicalPatternPlatform(platform)
	for _, mapping := range pattern.Mappings {
		if mapping.Platform == platform {
			return mapping
		}
	}
	return PatternMapping{}
}

func canonicalPatternPlatform(platform string) string {
	switch strings.ToLower(strings.TrimSpace(platform)) {
	case "macos", "mac", "apple":
		return "swiftui"
	case "windows", "winui", "xaml":
		return "winui3"
	case "linux", "qtquick", "qtwidgets", "qml":
		return "qt"
	default:
		return strings.ToLower(strings.TrimSpace(platform))
	}
}

func validPatternPlatform(platform string) bool {
	switch canonicalPatternPlatform(platform) {
	case "swiftui", "winui3", "qt":
		return true
	default:
		return false
	}
}

func contractsFor(node Node) []string {
	if node.Properties["componentBoundary"] == "native-winui-control" {
		return []string{"unsupported native WinUI component boundary"}
	}
	switch node.Kind {
	case KindButton:
		return []string{"action"}
	case KindTextField, KindSlider, KindToggle:
		return []string{"state binding"}
	case KindList:
		return []string{"collection source/template"}
	case KindComponent:
		return []string{"component boundary"}
	default:
		return []string{}
	}
}

func policiesFor(node Node) []string {
	switch node.Kind {
	case KindGrid:
		policies := []string{"Confirm spacing, available-size, and adaptive breakpoint policy."}
		if node.Properties["xaml.Grid.RowDefinitions"] != "" || node.Properties["xaml.Grid.ColumnDefinitions"] != "" {
			policies = append(policies, "Translate WinUI Grid row/column tracks into target stack, grid, split, or adaptive layout policy.")
		}
		return policies
	case KindVerticalStack, KindHorizontalStack, KindScrollView, KindList:
		return []string{"Confirm spacing, available-size, and adaptive breakpoint policy."}
	case KindText, KindTextField, KindButton:
		return []string{"Confirm typography, minimum target size, and platform theme token policy."}
	default:
		return []string{}
	}
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func nonNilDiagnostics(values []Diagnostic) []Diagnostic {
	if values == nil {
		return []Diagnostic{}
	}
	return values
}
