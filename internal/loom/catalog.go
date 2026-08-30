package loom

import (
	"fmt"
	"sort"
	"strings"
)

const Version = "0.17.0"

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
	{"accessibility:audit", "Audit Accessibility and Layout", "Audit accessible names, target sizes, redundant layouts, malformed nodes, and layout design risks", "accessibility", AccessConditionalWrite, []string{"--output"}, []string{"a11y"}, []string{"loom accessibility:audit <xaml-file> [--format text|json] [--fail-on none|error|warning] [--output path]"}, nil},
	{"inspect:ascii", "Inspect ASCII Pattern", "Render a WinUI layout as a plain ASCII Pattern tree", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"ascii"}, []string{"loom inspect:ascii <xaml-file> [--output path]"}, nil},
	{"inspect:xaml", "Inspect WinUI XAML", "Normalize WinUI XAML into Loom's OS-agnostic layout tree", "inspection", AccessConditionalWrite, []string{"--output"}, []string{"xaml"}, []string{"loom inspect:xaml <xaml-file> [--format text|json] [--output path]"}, nil},
	{"patterns:export", "Export Semantic Patterns", "Export Patterns as Loom, DTCG, Open UI, ARIA, or Style Dictionary JSON", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:export [--directory path] [--format loom|dtcg|open-ui|aria|style-dictionary] [--output path]"}, nil},
	{"patterns:list", "List Semantic Patterns", "List OS-agnostic layout and control patterns", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:list [--directory path] [--json] [--output path]"}, nil},
	{"patterns:show", "Show Semantic Pattern", "Print one complete pattern definition", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:show <id> [--directory path] [--output path]"}, nil},
	{"patterns:validate", "Validate Semantic Patterns", "Validate pattern metadata, constraints, identity, and uniqueness", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:validate [directory] [--json] [--output path]"}, nil},
	{"patterns:lint", "Lint Operational Patterns", "Enforce operational quality rules for bidirectional pattern mappings", "patterns", AccessConditionalWrite, []string{"--output"}, nil, []string{"loom patterns:lint [directory] [--json] [--output path]"}, nil},
	{"patterns:transfer", "Plan Pattern Transfer", "Classify how safely layout Patterns transfer between WinUI and SwiftUI", "patterns", AccessConditionalWrite, []string{"--output"}, []string{"transfer"}, []string{"loom patterns:transfer <xaml-file> [--from winui3] [--to swiftui] [--patterns-dir path] [--format text|json] [--output path]"}, nil},
	{"status", "Status", "Show local Loom readiness and pattern status", "diagnostics", AccessRead, nil, nil, []string{"loom status [--patterns-dir path] [--json]"}, nil},
	{"suggestions:os-errors", "Suggest OS Error Fixes", "Show curated user and AI-agent fixes for SwiftUI, WinUI, XAML, macOS, and Windows errors", "suggestions", AccessConditionalWrite, []string{"--output"}, []string{"os-errors"}, []string{"loom suggestions:os-errors [--platform swiftui|winui3|macos|windows|xaml] [--message text] [--format text|json] [--output path]"}, nil},
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
