package loom

import (
	"fmt"
	"sort"
	"strings"
)

const Version = "0.18.0"

type CommandAccess string

const (
	AccessRead             CommandAccess = "read"
	AccessWrite            CommandAccess = "write"
	AccessConditionalWrite CommandAccess = "conditional-write"
)

type CommandInfo struct {
	Command     string        `json:"command"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Category    string        `json:"category"`
	Access      CommandAccess `json:"access"`
	WriteFlags  []string      `json:"writeFlags,omitempty"`
	Aliases     []string      `json:"aliases,omitempty"`
	Synopsis    []string      `json:"synopsis"`
	Examples    []string      `json:"examples,omitempty"`
}

var Commands = []CommandInfo{
	{"accessibility:audit", "Audit Accessibility and Layout", "Audit accessible names, target sizes, redundant layouts, malformed nodes, and layout design risks", "accessibility", AccessConditionalWrite, []string{"--output"}, []string{"a11y"}, []string{"loom accessibility:audit <xaml-file> [--format text|json] [--fail-on none|error|warning] [--output path]"}, []string{"loom accessibility:audit MainWindow.xaml --format json --fail-on warning"}},
	{"checks:command-catalog", "Check Command Catalog", "Audit command metadata, aliases, synopsis, and access flags", "diagnostics", AccessRead, nil, nil, []string{"loom checks:command-catalog [--json]"}, nil},
	{"config:validate", "Validate Project Manifest", "Validate a Loom manifest and its referenced SwiftUI components", "setup", AccessRead, nil, nil, []string{"loom config:validate <loom.json> [--project-root path] [--format text|json]"}, nil},
	{"config:schema", "Project Manifest Schema", "Print the supported Loom project manifest schema", "setup", AccessRead, nil, nil, []string{"loom config:schema"}, nil},
	{"generate:contracts", "Generate Target Contracts", "Reserved for future Go work: inspectable contract report for native WinUI implementation", "generation", AccessConditionalWrite, []string{"--output"}, []string{"contracts"}, []string{"loom generate:contracts <swift-file> [--root-view Name] [--component name] [--theme-prefix Prefix] [--format text|json] [--output path]"}, []string{"loom generate:contracts ContentView.swift --format json --output contracts.json"}},
	{"generate:swiftui", "Generate SwiftUI Scaffold", "Reserved for future Go work: reviewable SwiftUI scaffold from WinUI XAML", "generation", AccessConditionalWrite, []string{"--output"}, []string{"swiftui"}, []string{"loom generate:swiftui <xaml-file> [--view-name Name] [--output path]"}, []string{"loom generate:swiftui MainWindow.xaml --view-name MainWindowScaffold"}},
	{"generate:xaml", "Generate WinUI XAML", "Reserved for future Go work: reviewable WinUI 3 XAML fragment", "generation", AccessConditionalWrite, []string{"--output", "--replace-region"}, []string{"generate"}, []string{"loom generate:xaml <swift-file> [--root-view Name] [--component name] [--theme-prefix Prefix] [--patterns-dir path] [--pattern-comments] [--output path]", "loom generate:xaml <swift-file> --replace-region <xaml-file> [--region-id id] [--init-region]"}, nil},
	{"graph:components", "Graph SwiftUI Components", "Reserved for future Go work: discover reachable SwiftUI layout components and custom view dependencies", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"graph"}, []string{"loom graph:components <swift-file-or-directory> [--root-view Name] [--component name] [--format text|json|dot] [--include glob] [--exclude glob]"}, []string{"loom graph:components Sources/App --root-view ContentView"}},
	{"guards:summary", "Guards Summary", "Show commands capable of writing and the flags that authorize writes", "diagnostics", AccessRead, nil, nil, []string{"loom guards:summary [--json]"}, nil},
	{"inspect:ascii", "Inspect ASCII Pattern", "Render a SwiftUI or WinUI layout as a plain ASCII Pattern tree", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"ascii"}, []string{"loom inspect:ascii <swift-file-or-xaml-file> [--root-view Name] [--component name] [--output path]"}, []string{"loom inspect:ascii MainWindow.xaml"}},
	{"inspect:errors", "Inspect Errors", "Report Swift syntax, Loom analysis, XAML, manifest, or Pattern errors", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"errors"}, []string{"loom inspect:errors <path> [--kind swift|xaml|manifest|patterns] [--root-view Name] [--component name] [--format text|json] [--fail-on none|error|warning] [--output path]"}, []string{"loom inspect:errors MainWindow.xaml --kind xaml --format json"}},
	{"inspect:parity", "Inspect XAML Parity", "Reserved for future Go work: compare SwiftUI structure with existing WinUI XAML", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"parity"}, []string{"loom inspect:parity <swift-file> --xaml <xaml-file> [--root-view Name] [--format text|json]"}, nil},
	{"inspect:source", "Inspect SwiftUI Source", "Reserved for future Go work: extract and report a SwiftUI layout tree", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"analyze"}, []string{"loom inspect:source <swift-file> [--root-view Name] [--component name] [--format text|json]"}, nil},
	{"inspect:xaml", "Inspect WinUI XAML", "Normalize WinUI XAML into Loom's OS-agnostic layout tree", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"xaml"}, []string{"loom inspect:xaml <xaml-file> [--format text|json] [--output path]"}, []string{"loom inspect:xaml MainWindow.xaml --format json"}},
	{"patterns:export", "Export Semantic Patterns", "Export Patterns as Loom, DTCG, Open UI, ARIA, or Style Dictionary JSON", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:export [--directory path] [--format loom|dtcg|open-ui|aria|style-dictionary] [--output path]"}, nil},
	{"patterns:list", "List Semantic Patterns", "List OS-agnostic layout and control patterns", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:list [--directory path] [--json] [--output path]"}, nil},
	{"patterns:show", "Show Semantic Pattern", "Print one complete pattern definition", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:show <id> [--directory path] [--output path]"}, nil},
	{"patterns:validate", "Validate Semantic Patterns", "Validate pattern metadata, constraints, identity, and uniqueness", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:validate [directory] [--json] [--output path]"}, nil},
	{"patterns:lint", "Lint Operational Patterns", "Enforce operational quality rules for bidirectional pattern mappings", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:lint [directory] [--json] [--output path]"}, nil},
	{"patterns:transfer", "Plan Pattern Transfer", "Classify how safely layout patterns transfer between WinUI and SwiftUI", "patterns", AccessConditionalWrite, []string{"--output"}, []string{"transfer"}, []string{"loom patterns:transfer <swift-file-or-xaml-file> [--from swiftui|winui3] [--to swiftui|winui3] [--root-view Name] [--component name] [--patterns-dir path] [--format text|json] [--output path]"}, []string{"loom patterns:transfer MainWindow.xaml --from winui3 --to swiftui --format json"}},
	{"project:build", "Build Project Translation", "Reserved for future Go work: run manifest-directed build workflows (analyses, fragments, parity)", "projects", AccessWrite, nil, []string{"project"}, []string{"loom project:build <loom.json> [--project-root path] [--output-dir path]"}, nil},
	{"self-heal:plan", "Self-Heal Plan", "Show explicit self-healing actions and their guardrails", "diagnostics", AccessRead, nil, nil, []string{"loom self-heal:plan [--json]"}, nil},
	{"status", "Status", "Show local Loom readiness and pattern status", "diagnostics", AccessRead, nil, nil, []string{"loom status [--patterns-dir path] [--json]"}, nil},
	{"suggestions:os-errors", "Suggest OS Error Fixes", "Show curated user and AI-agent fixes for SwiftUI, WinUI, XAML, macOS, and Windows errors", "suggestions", AccessConditionalWrite, []string{"--output"}, []string{"os-errors"}, []string{"loom suggestions:os-errors [--platform swiftui|winui3|macos|windows|xaml] [--message text] [--format text|json] [--output path]"}, []string{"loom suggestions:os-errors --platform winui3 --message StaticResource"}},
	{"verify", "Verify", "Run Loom's read-only command catalog and pattern checks", "diagnostics", AccessRead, nil, nil, []string{"loom verify [--patterns-dir path] [--json]"}, nil},
}

func resolveCommand(name string) (CommandInfo, bool) {
	for _, c := range Commands {
		if c.Command == name {
			return c, true
		}
		for _, alias := range c.Aliases {
			if alias == name {
				return c, true
			}
		}
	}
	return CommandInfo{}, false
}

func catalogText(category string) string {
	selected := Commands
	if category != "" {
		selected = nil
		for _, c := range Commands {
			if c.Category == category {
				selected = append(selected, c)
			}
		}
	}
	if len(selected) == 0 {
		return ""
	}
	groups := map[string][]CommandInfo{}
	width := 0
	for _, c := range selected {
		groups[c.Category] = append(groups[c.Category], c)
		if len(c.Command) > width {
			width = len(c.Command)
		}
	}
	categories := make([]string, 0, len(groups))
	for category := range groups {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	var b strings.Builder
	b.WriteString("Usage:\n  loom [--quiet|--verbose] <command> [args]\n  loom help <command>\n  loom list [--category NAME] [--json]\n\n")
	for _, category := range categories {
		sort.Slice(groups[category], func(i, j int) bool { return groups[category][i].Command < groups[category][j].Command })
		b.WriteString(category + "\n")
		for _, c := range groups[category] {
			fmt.Fprintf(&b, "    %-3s %-*s   %s\n", accessMarker(c.Access), width, c.Command, c.Description)
		}
	}
	return b.String()
}

func manual(command string) string {
	c, ok := resolveCommand(command)
	if !ok {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "loom %s\n%s\n\nDESCRIPTION\n  %s\n\nCATEGORY\n  %s\n\nACCESS\n  %s", c.Command, strings.Repeat("=", len(c.Command)+5), c.Description, c.Category, c.Access)
	if len(c.WriteFlags) > 0 {
		b.WriteString(" via " + strings.Join(c.WriteFlags, ", "))
	}
	b.WriteString("\n\nSYNOPSIS\n")
	for _, s := range c.Synopsis {
		b.WriteString("  " + s + "\n")
	}
	if len(c.Aliases) > 0 {
		b.WriteString("\nALIASES\n  " + strings.Join(c.Aliases, ", ") + "\n")
	}
	return b.String()
}

func accessMarker(access CommandAccess) string {
	switch access {
	case AccessRead:
		return "r"
	case AccessWrite:
		return "w"
	default:
		return "r/w"
	}
}
