package loom

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type PatternStatus string
type PatternValueType string

const DefaultPatternDirectory = "patterns"

type PatternIntent struct {
	Summary   string   `json:"summary"`
	UseWhen   []string `json:"useWhen"`
	AvoidWhen []string `json:"avoidWhen"`
}

type PatternSemantics struct {
	Role        string `json:"role"`
	ChildPolicy string `json:"childPolicy"`
	Sizing      string `json:"sizing"`
	Ordering    string `json:"ordering"`
}

type PatternAttribute struct {
	Name          string   `json:"name"`
	ValueType     string   `json:"valueType"`
	Required      bool     `json:"required"`
	Description   string   `json:"description"`
	DefaultValue  string   `json:"defaultValue,omitempty"`
	AllowedValues []string `json:"allowedValues,omitempty"`
	Minimum       *float64 `json:"minimum,omitempty"`
	Maximum       *float64 `json:"maximum,omitempty"`
	Units         []string `json:"units,omitempty"`
}

type PatternConstraint struct {
	Expression string `json:"expression"`
	Message    string `json:"message"`
}

type PatternKeyboard struct {
	Activation     []string `json:"activation,omitempty"`
	Navigation     []string `json:"navigation,omitempty"`
	EscapeBehavior string   `json:"escape_behavior,omitempty"`
}

type PatternMinimumTargetSize struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
	Unit   string  `json:"unit"`
}

type PatternAccessibility struct {
	Role              string                    `json:"role"`
	NameSource        string                    `json:"nameSource"`
	FocusBehavior     string                    `json:"focusBehavior"`
	Notes             []string                  `json:"notes"`
	Keyboard          *PatternKeyboard          `json:"keyboard,omitempty"`
	States            []string                  `json:"states,omitempty"`
	Properties        []string                  `json:"properties,omitempty"`
	MinimumTargetSize *PatternMinimumTargetSize `json:"minimumTargetSize,omitempty"`
}

type PatternMapping struct {
	Platform   string   `json:"platform"`
	Constructs []string `json:"constructs"`
	Strategy   string   `json:"strategy"`
	Caveats    []string `json:"caveats"`
}

type PatternVariant struct {
	Name         string   `json:"name"`
	Conditions   []string `json:"conditions"`
	LayoutPolicy string   `json:"layout_policy"`
	Intent       string   `json:"intent"`
}

type Pattern struct {
	SchemaVersion string               `json:"schema_version"`
	ID            string               `json:"id"`
	Version       string               `json:"version"`
	Name          string               `json:"name"`
	Kind          NodeKind             `json:"kind"`
	Status        PatternStatus        `json:"status"`
	Category      string               `json:"category"`
	Intent        PatternIntent        `json:"intent"`
	Semantics     PatternSemantics     `json:"semantics"`
	Attributes    []PatternAttribute   `json:"attributes"`
	Constraints   []PatternConstraint  `json:"constraints"`
	Accessibility PatternAccessibility `json:"accessibility"`
	Mappings      []PatternMapping     `json:"mappings"`
	Variants      []PatternVariant     `json:"variants,omitempty"`
	Tags          []string             `json:"tags"`
}

type PatternIssue struct {
	Severity DiagnosticSeverity `json:"severity"`
	Code     string             `json:"code"`
	Path     string             `json:"path"`
	Detail   string             `json:"detail"`
}

type PatternValidationReport struct {
	SchemaVersion string         `json:"schema_version"`
	Status        string         `json:"status"`
	Directory     string         `json:"directory"`
	PatternCount  int            `json:"patternCount"`
	Issues        []PatternIssue `json:"issues"`
}

func LoadPatterns(directory string) ([]Pattern, error) {
	directory = ResolvePatternDirectory(directory)
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("could not read pattern directory at %s: %w", directory, err)
	}
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".pattern.json") {
			files = append(files, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(files)
	patterns := make([]Pattern, 0, len(files))
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("could not read pattern file %s: %w", file, err)
		}
		var pattern Pattern
		if err := json.Unmarshal(data, &pattern); err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(file), err)
		}
		patterns = append(patterns, pattern)
	}
	return patterns, nil
}

func ResolvePatternDirectory(directory string) string {
	if directory == "" {
		directory = DefaultPatternDirectory
	}
	if directory != DefaultPatternDirectory {
		return directory
	}
	candidates := []string{DefaultPatternDirectory}
	if cwd, err := os.Getwd(); err == nil {
		for dir := cwd; dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
			candidates = append(candidates, filepath.Join(dir, DefaultPatternDirectory))
		}
	}
	if executable, err := os.Executable(); err == nil {
		executableDir := filepath.Dir(executable)
		candidates = append(candidates,
			filepath.Join(executableDir, DefaultPatternDirectory),
			filepath.Join(executableDir, "..", "share", "loom", DefaultPatternDirectory),
		)
	}
	for _, candidate := range candidates {
		if entries, err := os.ReadDir(candidate); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".pattern.json") {
					return candidate
				}
			}
		}
	}
	return directory
}

func ValidatePatterns(directory string) PatternValidationReport {
	resolved := ResolvePatternDirectory(directory)
	abs, _ := filepath.Abs(resolved)
	patterns, err := LoadPatterns(resolved)
	if err != nil {
		return PatternValidationReport{"1", "error", abs, 0, []PatternIssue{{SeverityError, "PATTERN001", abs, "pattern directory cannot be read."}}}
	}
	issues := []PatternIssue{}
	if len(patterns) == 0 {
		issues = append(issues, PatternIssue{SeverityError, "PATTERN003", abs, "No .pattern.json files were found."})
	}
	ids := map[string]bool{}
	kinds := map[NodeKind]bool{}
	for _, pattern := range patterns {
		path := pattern.ID + ".pattern.json"
		if pattern.SchemaVersion != "1" {
			issues = append(issues, PatternIssue{SeverityError, "PATTERN004", path, "schema_version must be 1."})
		}
		if pattern.ID == "" || strings.Contains(pattern.ID, "_") {
			issues = append(issues, PatternIssue{SeverityError, "PATTERN005", path, "id must use lowercase letters, numbers, and hyphens."})
		}
		if ids[pattern.ID] {
			issues = append(issues, PatternIssue{SeverityError, "PATTERN007", path, "Duplicate pattern id " + pattern.ID + "."})
		}
		ids[pattern.ID] = true
		if kinds[pattern.Kind] {
			issues = append(issues, PatternIssue{SeverityError, "PATTERN008", path, "Duplicate semantic kind " + string(pattern.Kind) + "."})
		}
		kinds[pattern.Kind] = true
		if pattern.Intent.Summary == "" || pattern.Semantics.Role == "" || pattern.Category == "" {
			issues = append(issues, PatternIssue{SeverityError, "PATTERN009", path, "Intent, semantic role, and category must be non-empty."})
		}
		if len(pattern.Mappings) == 0 {
			issues = append(issues, PatternIssue{SeverityError, "PATTERN015", path, "Mappings must be non-empty."})
		}
		for _, attribute := range pattern.Attributes {
			if attribute.Minimum != nil && attribute.Maximum != nil && *attribute.Minimum > *attribute.Maximum {
				issues = append(issues, PatternIssue{SeverityError, "PATTERN012", path + "#attributes." + attribute.Name, "minimum cannot exceed maximum."})
			}
		}
	}
	status := "ok"
	if len(issues) > 0 {
		status = "error"
	}
	return PatternValidationReport{"1", status, abs, len(patterns), issues}
}

func LintPatterns(directory string) PatternValidationReport {
	report := ValidatePatterns(directory)
	if report.Status != "ok" {
		return report
	}
	patterns, _ := LoadPatterns(directory)
	for _, pattern := range patterns {
		platforms := map[string]bool{}
		for _, mapping := range pattern.Mappings {
			platforms[mapping.Platform] = true
		}
		if !platforms["swiftui"] {
			report.Issues = append(report.Issues, PatternIssue{SeverityError, "PATTERN102", pattern.ID + ".pattern.json", "pattern must include a swiftui mapping."})
		}
		if !platforms["winui3"] {
			report.Issues = append(report.Issues, PatternIssue{SeverityError, "PATTERN103", pattern.ID + ".pattern.json", "pattern must include a winui3 mapping."})
		}
		if !platforms["qt"] {
			report.Issues = append(report.Issues, PatternIssue{SeverityError, "PATTERN104", pattern.ID + ".pattern.json", "pattern must include a qt mapping."})
		}
	}
	if len(report.Issues) > 0 {
		report.Status = "error"
	}
	return report
}

func PatternListText(patterns []Pattern) string {
	width := 0
	for _, pattern := range patterns {
		if len(pattern.ID) > width {
			width = len(pattern.ID)
		}
	}
	var b strings.Builder
	for _, pattern := range patterns {
		fmt.Fprintf(&b, "%-*s  %s  %s\n", width, pattern.ID, pattern.Kind, pattern.Intent.Summary)
	}
	return b.String()
}

func FindPattern(patterns []Pattern, id string) (Pattern, bool) {
	for _, pattern := range patterns {
		if pattern.ID == id {
			return pattern, true
		}
	}
	return Pattern{}, false
}

func ExportPatterns(patterns []Pattern, format string) (string, error) {
	switch format {
	case "", "loom":
		return prettyJSON(patterns)
	case "dtcg":
		type token struct {
			Type        string             `json:"$type"`
			Value       any                `json:"$value"`
			Description string             `json:"$description"`
			Extensions  map[string]Pattern `json:"$extensions"`
		}
		out := map[string]any{"$description": "loom OS-agnostic UI pattern catalog exported as DTCG-compatible token objects.", "loom": map[string]token{}}
		tokens := out["loom"].(map[string]token)
		for _, pattern := range patterns {
			tokens[pattern.ID] = token{"other", map[string]any{"id": pattern.ID, "name": pattern.Name, "kind": pattern.Kind, "category": pattern.Category}, pattern.Intent.Summary, map[string]Pattern{"loom": pattern}}
		}
		return prettyJSON(out)
	case "open-ui":
		return prettyJSON(map[string]any{"schema_version": "1", "source": "loom", "components": patterns})
	case "aria":
		type aria struct {
			ID            string   `json:"id"`
			Name          string   `json:"name"`
			Role          string   `json:"role"`
			NameSource    string   `json:"nameSource"`
			FocusBehavior string   `json:"focusBehavior"`
			Notes         []string `json:"notes"`
		}
		var entries []aria
		for _, pattern := range patterns {
			entries = append(entries, aria{pattern.ID, pattern.Name, pattern.Accessibility.Role, pattern.Accessibility.NameSource, pattern.Accessibility.FocusBehavior, pattern.Accessibility.Notes})
		}
		return prettyJSON(map[string]any{"schema_version": "1", "source": "loom", "patterns": entries})
	case "style-dictionary":
		out := map[string]any{"loom": map[string]any{}}
		items := out["loom"].(map[string]any)
		for _, pattern := range patterns {
			items[pattern.ID] = map[string]any{"value": pattern.ID, "type": "loom.pattern", "comment": pattern.Intent.Summary, "attributes": map[string]any{"name": pattern.Name, "kind": pattern.Kind, "category": pattern.Category}}
		}
		return prettyJSON(out)
	default:
		return "", fmt.Errorf("--format must be loom, dtcg, open-ui, aria, or style-dictionary")
	}
}
