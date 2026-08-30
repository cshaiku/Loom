package loom

import "strings"

type FixAudience string

const (
	FixUser  FixAudience = "user"
	FixAgent FixAudience = "agent"
)

type SuggestedFix struct {
	Audience FixAudience `json:"audience"`
	Action   string      `json:"action"`
	Detail   string      `json:"detail"`
	Command  string      `json:"command,omitempty"`
}

type OSErrorPlatform string

const (
	PlatformSwiftUI OSErrorPlatform = "swiftui"
	PlatformWinUI3  OSErrorPlatform = "winui3"
	PlatformMacOS   OSErrorPlatform = "macos"
	PlatformWindows OSErrorPlatform = "windows"
	PlatformXAML    OSErrorPlatform = "xaml"
)

type OSErrorSuggestion struct {
	Platform       OSErrorPlatform `json:"platform"`
	Category       string          `json:"category"`
	Matcher        string          `json:"matcher"`
	Issue          string          `json:"issue"`
	SuggestedFixes []SuggestedFix  `json:"suggested_fixes"`
	Reference      string          `json:"reference"`
}

type OSErrorSuggestionReport struct {
	SchemaVersion string              `json:"schema_version"`
	Status        string              `json:"status"`
	Platform      OSErrorPlatform     `json:"platform,omitempty"`
	Query         string              `json:"query,omitempty"`
	Suggestions   []OSErrorSuggestion `json:"suggestions"`
}

var osSuggestionCatalog = []OSErrorSuggestion{
	{PlatformSwiftUI, "syntax", "SWIFT.PARSE", "Swift parser diagnostics mean the SwiftUI body cannot be reliably extracted.", []SuggestedFix{{FixUser, "Fix source syntax first", "Do not transfer until source parses cleanly.", ""}, {FixAgent, "Reduce failing expression", "Check bracket/brace balance, string literals, trailing closures, and modifier chains.", "loom inspect:errors <source-file> --kind swift --json"}}, "Apple Swift language and SwiftSyntax diagnostics"},
	{PlatformSwiftUI, "accessibility", "accessibilityLabel", "Icon-only or non-text SwiftUI controls need explicit labels.", []SuggestedFix{{FixUser, "Name the control", "Provide the phrase users should hear.", ""}, {FixAgent, "Add SwiftUI accessibility label", "Use .accessibilityLabel and rerun audit after source inspection is available.", ""}}, "Apple SwiftUI accessibilityLabel documentation"},
	{PlatformXAML, "parse", "LOOM.XAML", "XAML parse/load failures usually come from malformed XML, namespace issues, property elements, or unresolved resources.", []SuggestedFix{{FixUser, "Confirm intended XAML structure", "Review whether the element belongs in markup or native code-behind.", ""}, {FixAgent, "Check XAML syntax", "Validate matching tags, namespace declarations, property elements, and attribute quoting.", "loom inspect:errors <source.xaml> --kind xaml --json"}}, "Microsoft WinUI/XAML parse exception guidance"},
	{PlatformWinUI3, "native-boundary", "XAML.UNSUPPORTED_COMPONENT_BOUNDARY", "A native WinUI control has no Loom semantic mapping and was preserved as a component boundary.", []SuggestedFix{{FixUser, "Choose native-boundary strategy", "Keep native, replace with portable layout, or approve a new Pattern mapping.", ""}, {FixAgent, "Document boundary contract", "Keep as handwritten WinUI and record source/target expectations.", "loom patterns:transfer <source.xaml> --from winui3 --to macos --format json"}}, "Loom native WinUI component-boundary policy"},
	{PlatformWinUI3, "resources", "StaticResource", "Unresolved StaticResource or ThemeResource keys can throw XAML parse/runtime exceptions.", []SuggestedFix{{FixUser, "Confirm token/resource ownership", "Identify whether resource comes from app, platform, or generated tokens.", ""}, {FixAgent, "Check resource dictionaries", "Verify key exists and dictionary is loaded before use.", ""}}, "Microsoft StaticResource and resource dictionary documentation"},
	{PlatformWinUI3, "accessibility", "AutomationProperties.Name", "WinUI controls need stable UI Automation names that usually match visible label text.", []SuggestedFix{{FixUser, "Approve accessible name", "Provide the user-facing name assistive tech should announce.", ""}, {FixAgent, "Set AutomationProperties.Name", "Use a localized value consistent with visible label.", "loom accessibility:audit <source.xaml> --format json"}}, "Microsoft AutomationProperties.Name guidance"},
	{PlatformWinUI3, "accessibility", "AccessibilityView", "Composed WinUI UIs can expose duplicate or low-value UIA nodes.", []SuggestedFix{{FixUser, "Decide tree exposure", "Confirm Control, Content, or Raw view.", ""}, {FixAgent, "Set accessibility tree visibility", "Use AutomationProperties.AccessibilityView for decorative or duplicate nodes.", ""}}, "Microsoft basic accessibility information"},
}

func OSErrorSuggestions(platform, query string) OSErrorSuggestionReport {
	out := []OSErrorSuggestion{}
	for _, suggestion := range osSuggestionCatalog {
		if platform != "" && string(suggestion.Platform) != platform {
			continue
		}
		if query != "" {
			haystack := strings.ToLower(string(suggestion.Platform) + " " + suggestion.Category + " " + suggestion.Matcher + " " + suggestion.Issue)
			if !queryMatches(haystack, query) {
				continue
			}
		}
		out = append(out, suggestion)
	}
	status := "ok"
	if len(out) == 0 {
		status = "empty"
	} else if query != "" {
		status = "matched"
	}
	return OSErrorSuggestionReport{"1", status, OSErrorPlatform(platform), query, out}
}

func queryMatches(haystack, query string) bool {
	normalized := strings.ToLower(query)
	if strings.Contains(haystack, normalized) {
		return true
	}
	for _, token := range strings.FieldsFunc(normalized, func(r rune) bool {
		return r < 'a' || r > 'z'
	}) {
		if len(token) >= 4 && strings.Contains(haystack, token) {
			return true
		}
	}
	return false
}
