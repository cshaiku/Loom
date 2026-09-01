package loom

import (
	"fmt"
	"sort"
	"strings"
)

const Version = "0.24.0"

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
	{"accessibility:audit", "audit accessibility and layout", "audit accessible names, target sizes, redundant layouts, malformed nodes, and layout design risks", "accessibility", AccessConditionalWrite, []string{"--output"}, []string{"a11y"}, []string{"loom accessibility:audit <xaml-file> [--format text|json] [--fail-on none|error|warning] [--output path]"}, []string{"loom accessibility:audit mainwindow.xaml --format json --fail-on warning"}},
	{"checks:command-catalog", "check command catalog", "audit command metadata, aliases, synopsis, and access flags", "diagnostics", AccessRead, nil, nil, []string{"loom checks:command-catalog [--json]"}, nil},
	{"config:validate", "validate project manifest", "validate a loom manifest and referenced layout source files", "setup", AccessRead, nil, nil, []string{"loom config:validate <loom.json> [--project-root path] [--format text|json]"}, nil},
	{"config:schema", "project manifest schema", "print the supported loom project manifest schema", "setup", AccessRead, nil, nil, []string{"loom config:schema"}, nil},
	{"generate:contracts", "generate target contracts", "reserved for future go work: inspectable contract report for native target implementation", "generation", AccessConditionalWrite, []string{"--output"}, []string{"contracts"}, []string{"loom generate:contracts <source-file> [--root-view name] [--component name] [--theme-prefix prefix] [--format text|json] [--output path]"}, nil},
	{"generate:swiftui", "generate swiftui scaffold", "reserved for future go work: reviewable swiftui scaffold from winui xaml", "generation", AccessConditionalWrite, []string{"--output"}, []string{"swiftui"}, []string{"loom generate:swiftui <xaml-file> [--view-name name] [--output path]"}, []string{"loom generate:swiftui mainwindow.xaml --view-name mainwindowscaffold"}},
	{"generate:xaml", "generate winui xaml", "reserved for future go work: reviewable winui 3 xaml fragment", "generation", AccessConditionalWrite, []string{"--output", "--replace-region"}, []string{"generate"}, []string{"loom generate:xaml <source-file> [--root-view name] [--component name] [--theme-prefix prefix] [--patterns-dir path] [--pattern-comments] [--output path]", "loom generate:xaml <source-file> --replace-region <xaml-file> [--region-id id] [--init-region]"}, nil},
	{"graph:components", "graph layout components", "discover layout components and custom dependencies across source files", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"graph"}, []string{"loom graph:components <source-file-or-directory> [--root-view name] [--component name] [--format text|json|dot] [--include glob] [--exclude glob] [--output path]"}, []string{"loom graph:components examples/sampleapp --format dot --output graph.dot"}},
	{"guards:summary", "guards summary", "show commands capable of writing and the flags that authorize writes", "diagnostics", AccessRead, nil, nil, []string{"loom guards:summary [--json]"}, nil},
	{"inspect:ascii", "inspect ascii pattern", "render a winui layout as a plain ascii pattern tree", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"ascii"}, []string{"loom inspect:ascii <xaml-file> [--root-view name] [--component name] [--output path]"}, []string{"loom inspect:ascii mainwindow.xaml"}},
	{"inspect:errors", "inspect errors", "report source, xaml, qt, manifest, or pattern errors", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"errors"}, []string{"loom inspect:errors <path> [--kind swift|xaml|qt|manifest|patterns] [--root-view name] [--component name] [--format text|json] [--fail-on none|error|warning] [--output path]"}, []string{"loom inspect:errors mainwindow.xaml --kind xaml --format json"}},
	{"inspect:font", "inspect font material", "extract intrinsic font names, metrics, normalized ratios, and profile-ready typography", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"font"}, []string{"loom inspect:font <font-file> [--format text|json] [--output path]", "loom inspect:font --family <installed-family-name> [--format text|json] [--output path]"}, []string{"loom inspect:font Inter.ttf --json", "loom inspect:font --family \"Segoe UI\" --json"}},
	{"inspect:parity", "inspect layout parity", "compare source and target layout structure across supported UI dialects", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"parity"}, []string{"loom inspect:parity <source-file> --target <target-file> [--from swiftui|winui3|qt] [--to swiftui|winui3|qt] [--format text|json] [--output path]"}, []string{"loom inspect:parity contentview.swift --target mainwindow.qml --from swiftui --to qt --json"}},
	{"inspect:visual-parity", "inspect visual parity", "compare profile-normalized visual metrics across supported UI dialects", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"visual-parity"}, []string{"loom inspect:visual-parity <source-file> --target <target-file> [--from swiftui|winui3|qt] [--to swiftui|winui3|qt] [--profile visual-profile.json] [--source-font font.ttf] [--target-font font.ttf] [--source-font-family name] [--target-font-family name] [--format text|json] [--output path]"}, []string{"loom inspect:visual-parity contentview.swift --target mainwindow.xaml --from swiftui --to winui3 --profile visual-profile.json --json"}},
	{"inspect:qt", "inspect qt", "normalize Qt QML, Qt Designer UI, or common Qt C++ layout constructs into loom's OS-agnostic layout tree", "inspection", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom inspect:qt <qml-ui-or-cpp-file> [--format text|json] [--output path]"}, []string{"loom inspect:qt mainwindow.qml --format json"}},
	{"inspect:source", "inspect layout source", "auto-detect SwiftUI, WinUI XAML, or Qt source and report a shared layout tree", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"analyze"}, []string{"loom inspect:source <source-file> [--from swiftui|winui3|qt] [--format text|json] [--output path]"}, []string{"loom inspect:source contentview.swift --json"}},
	{"inspect:swiftui", "inspect swiftui", "normalize common swiftui layout constructs into loom's OS-agnostic layout tree", "inspection", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom inspect:swiftui <swift-file> [--format text|json] [--output path]"}, []string{"loom inspect:swiftui contentview.swift --format json"}},
	{"inspect:xaml", "inspect winui xaml", "normalize winui xaml into loom's OS-agnostic layout tree", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"xaml"}, []string{"loom inspect:xaml <xaml-file> [--format text|json] [--output path]"}, []string{"loom inspect:xaml mainwindow.xaml --format json"}},
	{"patterns:export", "export semantic patterns", "export patterns as loom, DTCG, open UI, ARIA, or style dictionary JSON", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:export [--directory path] [--format loom|dtcg|open-ui|aria|style-dictionary] [--output path]"}, nil},
	{"patterns:list", "list semantic patterns", "list OS-agnostic layout and control patterns", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:list [--directory path] [--json] [--output path]"}, nil},
	{"patterns:show", "show semantic pattern", "print one complete pattern definition", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:show <id> [--directory path] [--output path]"}, nil},
	{"patterns:validate", "validate semantic patterns", "validate pattern metadata, constraints, identity, and uniqueness", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:validate [directory] [--json] [--output path]"}, nil},
	{"patterns:lint", "lint operational patterns", "enforce operational quality rules for bidirectional pattern mappings", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:lint [directory] [--json] [--output path]"}, nil},
	{"patterns:transfer", "plan pattern transfer", "classify how safely layout patterns transfer between OS UI frameworks", "patterns", AccessConditionalWrite, []string{"--output"}, []string{"transfer"}, []string{"loom patterns:transfer <source-file> [--from swiftui|macos|winui3|windows] [--to swiftui|macos|winui3|windows] [--root-view name] [--component name] [--patterns-dir path] [--format text|json] [--output path]"}, []string{"loom patterns:transfer contentview.swift --from swiftui --to windows --format json"}},
	{"project:build", "build project translation", "run manifest-directed analysis, transfer, parity, and component graph artifact workflows", "projects", AccessWrite, []string{"--output-dir", "--overwrite"}, []string{"project"}, []string{"loom project:build <loom.json> [--project-root path] [--output-dir path] [--overwrite] [--json]"}, []string{"loom project:build examples/sampleapp/loom.json --output-dir generated/project-build --overwrite --json"}},
	{"self-heal:plan", "self-heal plan", "show explicit self-healing actions and their guardrails", "diagnostics", AccessRead, nil, nil, []string{"loom self-heal:plan [--json]"}, nil},
	{"status", "status", "show local loom readiness and pattern status", "diagnostics", AccessRead, nil, nil, []string{"loom status [--patterns-dir path] [--json]"}, nil},
	{"suggestions:os-errors", "suggest OS error fixes", "show curated user and AI-agent fixes for swiftui, winui, xaml, macos, and windows errors", "suggestions", AccessConditionalWrite, []string{"--output"}, []string{"os-errors"}, []string{"loom suggestions:os-errors [--platform swiftui|winui3|macos|windows|xaml] [--message text] [--format text|json] [--output path]"}, []string{"loom suggestions:os-errors --platform winui3 --message staticresource"}},
	{"verify", "verify", "run loom's read-only command catalog and pattern checks", "diagnostics", AccessRead, nil, nil, []string{"loom verify [--patterns-dir path] [--json]"}, nil},
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
	b.WriteString("Usage:\n  loom [--quiet|--verbose] [--line-ending lf|crlf|native] <command> [args]\n  loom help <command>\n  loom list [--category NAME] [--json]\n\n")
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
