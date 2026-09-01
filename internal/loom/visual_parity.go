package loom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type VisualTypographyProfile struct {
	FontFamily     string   `json:"fontFamily,omitempty"`
	FallbackFonts  []string `json:"fallbackFonts,omitempty"`
	FontSize       float64  `json:"fontSize,omitempty"`
	Kerning        float64  `json:"kerning,omitempty"`
	LineHeight     float64  `json:"lineHeight,omitempty"`
	BaselineOffset float64  `json:"baselineOffset,omitempty"`
}

type VisualSpacingProfile struct {
	DefaultPadding float64 `json:"defaultPadding,omitempty"`
	DefaultMargin  float64 `json:"defaultMargin,omitempty"`
	StackSpacing   float64 `json:"stackSpacing,omitempty"`
}

type VisualControlProfile struct {
	ButtonMinHeight    float64 `json:"buttonMinHeight,omitempty"`
	TextFieldMinHeight float64 `json:"textFieldMinHeight,omitempty"`
	ToggleMinHeight    float64 `json:"toggleMinHeight,omitempty"`
	ListRowMinHeight   float64 `json:"listRowMinHeight,omitempty"`
}

type VisualPlatformProfile struct {
	Typography       VisualTypographyProfile `json:"typography,omitempty"`
	Spacing          VisualSpacingProfile    `json:"spacing,omitempty"`
	Controls         VisualControlProfile    `json:"controls,omitempty"`
	typographyOrigin string
	spacingOrigin    string
	controlsOrigin   string
}

type VisualToleranceProfile struct {
	Distance   float64 `json:"distance,omitempty"`
	Typography float64 `json:"typography,omitempty"`
}

type VisualProfile struct {
	SchemaVersion string                           `json:"schema_version"`
	Platforms     map[string]VisualPlatformProfile `json:"platforms"`
	Tolerances    VisualToleranceProfile           `json:"tolerances,omitempty"`
}

type VisualMetric struct {
	FontFamily     string   `json:"fontFamily,omitempty"`
	FallbackFonts  []string `json:"fallbackFonts,omitempty"`
	FontSize       float64  `json:"fontSize,omitempty"`
	Kerning        float64  `json:"kerning,omitempty"`
	LineHeight     float64  `json:"lineHeight,omitempty"`
	LineSpacing    float64  `json:"lineSpacing,omitempty"`
	BaselineOffset float64  `json:"baselineOffset,omitempty"`
	Padding        float64  `json:"padding,omitempty"`
	Margin         float64  `json:"margin,omitempty"`
	Spacing        float64  `json:"spacing,omitempty"`
	Width          float64  `json:"width,omitempty"`
	Height         float64  `json:"height,omitempty"`
	MinWidth       float64  `json:"minWidth,omitempty"`
	MinHeight      float64  `json:"minHeight,omitempty"`
}

type VisualProvenance struct {
	Origin     string  `json:"origin"`
	Detail     string  `json:"detail,omitempty"`
	Confidence float64 `json:"confidence"`
}

type VisualParityEntry struct {
	Path       string                      `json:"path"`
	Kind       NodeKind                    `json:"kind"`
	Label      string                      `json:"label,omitempty"`
	Metrics    VisualMetric                `json:"metrics"`
	Provenance map[string]VisualProvenance `json:"provenance,omitempty"`
}

type VisualParityFinding struct {
	Severity         DiagnosticSeverity `json:"severity"`
	Code             string             `json:"code"`
	Path             string             `json:"path"`
	Metric           string             `json:"metric,omitempty"`
	Confidence       float64            `json:"confidence,omitempty"`
	SourceProvenance *VisualProvenance  `json:"sourceProvenance,omitempty"`
	TargetProvenance *VisualProvenance  `json:"targetProvenance,omitempty"`
	Message          string             `json:"message"`
}

type VisualParityReport struct {
	SchemaVersion string                `json:"schema_version"`
	Status        string                `json:"status"`
	SourcePath    string                `json:"sourcePath"`
	TargetPath    string                `json:"targetPath"`
	SourceDialect string                `json:"sourceDialect"`
	TargetDialect string                `json:"targetDialect"`
	SourceCount   int                   `json:"sourceCount"`
	TargetCount   int                   `json:"targetCount"`
	Profile       VisualProfile         `json:"profile"`
	SourceEntries []VisualParityEntry   `json:"sourceEntries,omitempty"`
	TargetEntries []VisualParityEntry   `json:"targetEntries,omitempty"`
	Findings      []VisualParityFinding `json:"findings"`
}

func InspectVisualParity(source, target, sourcePlatform, targetPlatform, profilePath, sourceFontPath, targetFontPath, sourceFontFamily, targetFontFamily string) (VisualParityReport, error) {
	sourceAnalysis, err := AnalyzeByPlatform(source, sourcePlatform)
	if err != nil {
		return VisualParityReport{}, err
	}
	targetAnalysis, err := AnalyzeByPlatform(target, targetPlatform)
	if err != nil {
		return VisualParityReport{}, err
	}
	profile := DefaultVisualProfile()
	if profilePath != "" {
		loaded, err := LoadVisualProfile(profilePath)
		if err != nil {
			return VisualParityReport{}, err
		}
		profile = MergeVisualProfile(profile, loaded)
	}
	sourceDialect := sourceAnalysis.Layout.Properties["sourceDialect"]
	targetDialect := targetAnalysis.Layout.Properties["sourceDialect"]
	if sourceFontPath != "" || sourceFontFamily != "" {
		typography, err := inspectFontTypography(sourceFontPath, sourceFontFamily)
		if err != nil {
			return VisualParityReport{}, err
		}
		profile = applyFontTypography(profile, sourceDialect, typography)
	}
	if targetFontPath != "" || targetFontFamily != "" {
		typography, err := inspectFontTypography(targetFontPath, targetFontFamily)
		if err != nil {
			return VisualParityReport{}, err
		}
		profile = applyFontTypography(profile, targetDialect, typography)
	}
	sourceEntries := visualParityEntries(sourceAnalysis.Layout, visualPlatformProfile(profile, sourceDialect))
	targetEntries := visualParityEntries(targetAnalysis.Layout, visualPlatformProfile(profile, targetDialect))
	findings := visualDiagnosticFindings("source", sourceAnalysis.Diagnostics)
	findings = append(findings, visualDiagnosticFindings("target", targetAnalysis.Diagnostics)...)
	findings = append(findings, compareVisualParityEntries(sourceEntries, targetEntries, profile.Tolerances)...)
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
	return VisualParityReport{"1", status, source, target, sourceDialect, targetDialect, len(sourceEntries), len(targetEntries), profile, sourceEntries, targetEntries, findings}, nil
}

func inspectFontTypography(path, family string) (VisualTypographyProfile, error) {
	report := InspectFontSource(path, family)
	if report.Status == "error" || len(report.Faces) == 0 {
		return VisualTypographyProfile{}, fmt.Errorf("could not inspect font material for visual parity")
	}
	return report.Faces[0].ProfileTypography, nil
}

func applyFontTypography(profile VisualProfile, platform string, typography VisualTypographyProfile) VisualProfile {
	key := canonicalPatternPlatform(platform)
	if key == "" {
		key = platform
	}
	if profile.Platforms == nil {
		profile.Platforms = map[string]VisualPlatformProfile{}
	}
	platformProfile := profile.Platforms[key]
	platformProfile.Typography = typography
	platformProfile.typographyOrigin = "font-material"
	profile.Platforms[key] = platformProfile
	return profile
}

func DefaultVisualProfile() VisualProfile {
	return VisualProfile{
		SchemaVersion: "1",
		Platforms: map[string]VisualPlatformProfile{
			"swiftui": {
				Typography:       VisualTypographyProfile{FontFamily: ".SF NS", FallbackFonts: []string{"SF Pro", "Helvetica Neue"}, FontSize: 17, Kerning: 0, LineHeight: 22, BaselineOffset: 0},
				Spacing:          VisualSpacingProfile{DefaultPadding: 0, DefaultMargin: 0, StackSpacing: 8},
				Controls:         VisualControlProfile{ButtonMinHeight: 32, TextFieldMinHeight: 32, ToggleMinHeight: 31, ListRowMinHeight: 44},
				typographyOrigin: "default-profile",
				spacingOrigin:    "default-profile",
				controlsOrigin:   "default-profile",
			},
			"winui3": {
				Typography:       VisualTypographyProfile{FontFamily: "Segoe UI", FallbackFonts: []string{"Segoe UI Variable", "Arial"}, FontSize: 14, Kerning: 0, LineHeight: 20, BaselineOffset: 0},
				Spacing:          VisualSpacingProfile{DefaultPadding: 0, DefaultMargin: 0, StackSpacing: 8},
				Controls:         VisualControlProfile{ButtonMinHeight: 32, TextFieldMinHeight: 32, ToggleMinHeight: 32, ListRowMinHeight: 40},
				typographyOrigin: "default-profile",
				spacingOrigin:    "default-profile",
				controlsOrigin:   "default-profile",
			},
			"qt": {
				Typography:       VisualTypographyProfile{FontFamily: "system", FallbackFonts: []string{"Sans Serif"}, FontSize: 14, Kerning: 0, LineHeight: 20, BaselineOffset: 0},
				Spacing:          VisualSpacingProfile{DefaultPadding: 0, DefaultMargin: 0, StackSpacing: 6},
				Controls:         VisualControlProfile{ButtonMinHeight: 30, TextFieldMinHeight: 30, ToggleMinHeight: 30, ListRowMinHeight: 32},
				typographyOrigin: "default-profile",
				spacingOrigin:    "default-profile",
				controlsOrigin:   "default-profile",
			},
		},
		Tolerances: VisualToleranceProfile{Distance: 0.5, Typography: 0.25},
	}
}

func LoadVisualProfile(path string) (VisualProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return VisualProfile{}, fmt.Errorf("could not read visual profile at %s", path)
	}
	var profile VisualProfile
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&profile); err != nil {
		return VisualProfile{}, fmt.Errorf("could not parse visual profile at %s: %w", path, err)
	}
	if profile.SchemaVersion != "" && profile.SchemaVersion != "1" {
		return VisualProfile{}, fmt.Errorf("unsupported visual profile schema_version %q", profile.SchemaVersion)
	}
	if err := ValidateVisualProfile(profile); err != nil {
		return VisualProfile{}, err
	}
	return profile, nil
}

func ValidateVisualProfile(profile VisualProfile) error {
	if profile.Platforms == nil {
		return fmt.Errorf("visual profile must include platforms")
	}
	for platform, platformProfile := range profile.Platforms {
		if !validPatternPlatform(platform) {
			return fmt.Errorf("visual profile platform %q is not supported", platform)
		}
		if platformProfile.Typography.FontSize < 0 || platformProfile.Typography.LineHeight < 0 {
			return fmt.Errorf("visual profile platform %s has negative typography metric", platform)
		}
		if platformProfile.Spacing.DefaultPadding < 0 || platformProfile.Spacing.DefaultMargin < 0 || platformProfile.Spacing.StackSpacing < 0 {
			return fmt.Errorf("visual profile platform %s has negative spacing metric", platform)
		}
		if platformProfile.Controls.ButtonMinHeight < 0 || platformProfile.Controls.TextFieldMinHeight < 0 || platformProfile.Controls.ToggleMinHeight < 0 || platformProfile.Controls.ListRowMinHeight < 0 {
			return fmt.Errorf("visual profile platform %s has negative control metric", platform)
		}
	}
	if profile.Tolerances.Distance < 0 || profile.Tolerances.Typography < 0 {
		return fmt.Errorf("visual profile tolerances cannot be negative")
	}
	return nil
}

func MergeVisualProfile(base, override VisualProfile) VisualProfile {
	if override.SchemaVersion != "" {
		base.SchemaVersion = override.SchemaVersion
	}
	if base.Platforms == nil {
		base.Platforms = map[string]VisualPlatformProfile{}
	}
	for platform, profile := range override.Platforms {
		key := canonicalPatternPlatform(platform)
		if key == "" {
			key = platform
		}
		existing := base.Platforms[key]
		base.Platforms[key] = mergeVisualPlatformProfile(existing, profile, "profile")
	}
	if override.Tolerances.Distance != 0 {
		base.Tolerances.Distance = override.Tolerances.Distance
	}
	if override.Tolerances.Typography != 0 {
		base.Tolerances.Typography = override.Tolerances.Typography
	}
	return base
}

func mergeVisualPlatformProfile(base, override VisualPlatformProfile, origin string) VisualPlatformProfile {
	if override.Typography.FontFamily != "" {
		base.Typography.FontFamily = override.Typography.FontFamily
		base.typographyOrigin = origin
	}
	if override.Typography.FallbackFonts != nil {
		base.Typography.FallbackFonts = override.Typography.FallbackFonts
		base.typographyOrigin = origin
	}
	if override.Typography.FontSize != 0 {
		base.Typography.FontSize = override.Typography.FontSize
		base.typographyOrigin = origin
	}
	if override.Typography.Kerning != 0 {
		base.Typography.Kerning = override.Typography.Kerning
		base.typographyOrigin = origin
	}
	if override.Typography.LineHeight != 0 {
		base.Typography.LineHeight = override.Typography.LineHeight
		base.typographyOrigin = origin
	}
	if override.Typography.BaselineOffset != 0 {
		base.Typography.BaselineOffset = override.Typography.BaselineOffset
		base.typographyOrigin = origin
	}
	if override.Spacing.DefaultPadding != 0 {
		base.Spacing.DefaultPadding = override.Spacing.DefaultPadding
		base.spacingOrigin = origin
	}
	if override.Spacing.DefaultMargin != 0 {
		base.Spacing.DefaultMargin = override.Spacing.DefaultMargin
		base.spacingOrigin = origin
	}
	if override.Spacing.StackSpacing != 0 {
		base.Spacing.StackSpacing = override.Spacing.StackSpacing
		base.spacingOrigin = origin
	}
	if override.Controls.ButtonMinHeight != 0 {
		base.Controls.ButtonMinHeight = override.Controls.ButtonMinHeight
		base.controlsOrigin = origin
	}
	if override.Controls.TextFieldMinHeight != 0 {
		base.Controls.TextFieldMinHeight = override.Controls.TextFieldMinHeight
		base.controlsOrigin = origin
	}
	if override.Controls.ToggleMinHeight != 0 {
		base.Controls.ToggleMinHeight = override.Controls.ToggleMinHeight
		base.controlsOrigin = origin
	}
	if override.Controls.ListRowMinHeight != 0 {
		base.Controls.ListRowMinHeight = override.Controls.ListRowMinHeight
		base.controlsOrigin = origin
	}
	return base
}

func visualPlatformProfile(profile VisualProfile, platform string) VisualPlatformProfile {
	key := canonicalPatternPlatform(platform)
	if p, ok := profile.Platforms[key]; ok {
		return p
	}
	if p, ok := profile.Platforms[platform]; ok {
		return p
	}
	return VisualPlatformProfile{}
}

func visualParityEntries(root Node, profile VisualPlatformProfile) []VisualParityEntry {
	entries := []VisualParityEntry{}
	var walk func(Node, string)
	walk = func(node Node, path string) {
		metrics, provenance := visualMetricsWithProvenance(node, profile)
		entries = append(entries, VisualParityEntry{Path: path, Kind: node.Kind, Label: firstNonEmpty(node.VisibleLabel, node.AccessibleName, node.Placeholder, node.Resource), Metrics: metrics, Provenance: provenance})
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

func visualMetrics(node Node, profile VisualPlatformProfile) VisualMetric {
	metrics, _ := visualMetricsWithProvenance(node, profile)
	return metrics
}

func visualMetricsWithProvenance(node Node, profile VisualPlatformProfile) (VisualMetric, map[string]VisualProvenance) {
	metric := VisualMetric{
		FontFamily:     profile.Typography.FontFamily,
		FallbackFonts:  append([]string{}, profile.Typography.FallbackFonts...),
		FontSize:       profile.Typography.FontSize,
		Kerning:        profile.Typography.Kerning,
		LineHeight:     profile.Typography.LineHeight,
		BaselineOffset: profile.Typography.BaselineOffset,
		Padding:        profile.Spacing.DefaultPadding,
		Margin:         profile.Spacing.DefaultMargin,
	}
	provenance := map[string]VisualProvenance{}
	for _, name := range []string{"fontFamily", "fallbackFonts", "fontSize", "kerning", "lineHeight", "baselineOffset"} {
		provenance[name] = visualProvenance(firstNonEmpty(profile.typographyOrigin, "profile"), "platform typography", 0)
	}
	provenance["lineSpacing"] = visualProvenance("unknown", "no platform line spacing default", 0)
	for _, name := range []string{"padding", "margin", "spacing"} {
		provenance[name] = visualProvenance(firstNonEmpty(profile.spacingOrigin, "profile"), "platform spacing", 0)
	}
	for _, name := range []string{"minHeight"} {
		provenance[name] = visualProvenance(firstNonEmpty(profile.controlsOrigin, "profile"), "platform control default", 0)
	}
	switch node.Kind {
	case KindVerticalStack, KindHorizontalStack, KindGrid:
		metric.Spacing = profile.Spacing.StackSpacing
		if value, ok := numberPropertyValue(node, "xaml.Spacing"); ok && value > 0 {
			metric.Spacing = value
			provenance["spacing"] = visualProvenance("source", "xaml.Spacing", 0)
		} else if value, ok := numberPropertyValue(node, "xaml.RowSpacing"); ok && value > 0 {
			metric.Spacing = value
			provenance["spacing"] = visualProvenance("source", "xaml.RowSpacing", 0)
		} else if value, ok := swiftArgumentNumberValue(node, "spacing"); ok && value > 0 {
			metric.Spacing = value
			provenance["spacing"] = visualProvenance("source", "swiftui.arguments spacing", 0)
		}
	case KindButton:
		metric.MinHeight = profile.Controls.ButtonMinHeight
		provenance["minHeight"] = visualProvenance(firstNonEmpty(profile.controlsOrigin, "profile"), "button control default", 0)
	case KindTextField:
		metric.MinHeight = profile.Controls.TextFieldMinHeight
		provenance["minHeight"] = visualProvenance(firstNonEmpty(profile.controlsOrigin, "profile"), "text field control default", 0)
	case KindToggle:
		metric.MinHeight = profile.Controls.ToggleMinHeight
		provenance["minHeight"] = visualProvenance(firstNonEmpty(profile.controlsOrigin, "profile"), "toggle control default", 0)
	case KindList:
		metric.MinHeight = profile.Controls.ListRowMinHeight
		provenance["minHeight"] = visualProvenance(firstNonEmpty(profile.controlsOrigin, "profile"), "list row control default", 0)
	}
	applyNumericOverride := func(name, xamlKey, swiftKey string, positive bool, set func(float64)) {
		if raw, origin, detail, ok := xamlPropertyMetricValue(node, xamlKey); ok {
			value, hasNumber := firstNumericStringValue(raw)
			if hasNumber && (!positive || value > 0) {
				set(value)
				provenance[name] = visualProvenance(origin, detail, 0)
				return
			}
		}
		if value, ok := numberPropertyValue(node, xamlKey); ok && (!positive || value > 0) {
			set(value)
			provenance[name] = visualProvenance("source", xamlKey, 0)
			return
		}
		if swiftKey == "" {
			return
		}
		if value, ok := swiftArgumentNumberValue(node, swiftKey); ok && (!positive || value > 0) {
			set(value)
			provenance[name] = visualProvenance("source", "swiftui.arguments "+swiftKey, 0)
		}
	}
	if value, origin, detail, ok := xamlPropertyMetricValue(node, "xaml.FontFamily"); ok {
		metric.FontFamily = value
		provenance["fontFamily"] = visualProvenance(origin, detail, 0)
	}
	applyNumericOverride("fontSize", "xaml.FontSize", "size", true, func(value float64) { metric.FontSize = value })
	applyNumericOverride("kerning", "xaml.CharacterSpacing", "kerning", false, func(value float64) { metric.Kerning = value })
	applyNumericOverride("lineHeight", "xaml.LineHeight", "", true, func(value float64) { metric.LineHeight = value })
	applyNumericOverride("padding", "xaml.Padding", "padding", false, func(value float64) { metric.Padding = value })
	applyNumericOverride("margin", "xaml.Margin", "margin", false, func(value float64) { metric.Margin = value })
	applyNumericOverride("width", "xaml.Width", "width", true, func(value float64) { metric.Width = value })
	applyNumericOverride("height", "xaml.Height", "height", true, func(value float64) { metric.Height = value })
	applyNumericOverride("minWidth", "xaml.MinWidth", "minWidth", true, func(value float64) { metric.MinWidth = value })
	applyNumericOverride("minHeight", "xaml.MinHeight", "minHeight", true, func(value float64) { metric.MinHeight = value })
	applySwiftUIModifierMetrics(node, &metric, provenance)
	return metric, provenance
}

func applySwiftUIModifierMetrics(node Node, metric *VisualMetric, provenance map[string]VisualProvenance) {
	for _, modifier := range node.Modifiers {
		switch modifier.Name {
		case "font":
			if value, ok := modifierNumberValue(modifier, "size"); ok && value > 0 {
				metric.FontSize = value
				provenance["fontSize"] = visualProvenance("source", "swiftui.modifier font.size", 0)
			}
		case "fontDesign":
			if modifier.Arguments != "" {
				metric.FontFamily = "swiftui.fontDesign:" + compactSwiftExpression(modifier.Arguments)
				provenance["fontFamily"] = visualProvenance("source", "swiftui.modifier fontDesign", 0)
			}
		case "kerning", "tracking":
			if value, ok := modifierFirstNumberValue(modifier); ok {
				metric.Kerning = value
				provenance["kerning"] = visualProvenance("source", "swiftui.modifier "+modifier.Name, 0)
			}
		case "lineSpacing":
			if value, ok := modifierFirstNumberValue(modifier); ok {
				metric.LineSpacing = value
				provenance["lineSpacing"] = visualProvenance("source", "swiftui.modifier lineSpacing", 0)
			}
		case "baselineOffset":
			if value, ok := modifierFirstNumberValue(modifier); ok {
				metric.BaselineOffset = value
				provenance["baselineOffset"] = visualProvenance("source", "swiftui.modifier baselineOffset", 0)
			}
		case "padding":
			if value, ok := modifierFirstNumberValue(modifier); ok {
				metric.Padding = value
				provenance["padding"] = visualProvenance("source", "swiftui.modifier padding", 0)
			}
		case "frame":
			applyModifierNumber(modifier, "width", metric.Width, func(value float64) { metric.Width = value }, provenance)
			applyModifierNumber(modifier, "height", metric.Height, func(value float64) { metric.Height = value }, provenance)
			applyModifierNumber(modifier, "minWidth", metric.MinWidth, func(value float64) { metric.MinWidth = value }, provenance)
			applyModifierNumber(modifier, "minHeight", metric.MinHeight, func(value float64) { metric.MinHeight = value }, provenance)
		case "controlSize":
			size := compactSwiftExpression(modifier.Arguments)
			if value := swiftUIControlMinHeight(size, node.Kind); value > 0 {
				metric.MinHeight = value
				provenance["minHeight"] = visualProvenance("source", "swiftui.modifier controlSize "+size, 0)
			}
		}
	}
}

func applyModifierNumber(modifier Modifier, name string, current float64, set func(float64), provenance map[string]VisualProvenance) {
	_ = current
	if value, ok := modifierNumberValue(modifier, name); ok && value > 0 {
		set(value)
		provenance[name] = visualProvenance("source", "swiftui.modifier frame."+name, 0)
	}
}

func compareVisualParityEntries(source, target []VisualParityEntry, tolerances VisualToleranceProfile) []VisualParityFinding {
	findings := []VisualParityFinding{}
	if len(source) != len(target) {
		findings = append(findings, VisualParityFinding{Severity: SeverityWarning, Code: "VISUAL.COUNT", Path: "/", Confidence: 0.95, Message: fmt.Sprintf("source has %d visual nodes; target has %d.", len(source), len(target))})
	}
	targetByPath := map[string]VisualParityEntry{}
	for _, entry := range target {
		targetByPath[entry.Path] = entry
	}
	for _, sourceEntry := range source {
		targetEntry, ok := targetByPath[sourceEntry.Path]
		if !ok {
			findings = append(findings, VisualParityFinding{Severity: SeverityWarning, Code: "VISUAL.PATH", Path: sourceEntry.Path, Confidence: 0.95, Message: "source visual node has no matching target node at the same tree path."})
			continue
		}
		if sourceEntry.Kind != targetEntry.Kind {
			findings = append(findings, VisualParityFinding{Severity: SeverityWarning, Code: "VISUAL.KIND", Path: sourceEntry.Path, Confidence: 0.95, Message: fmt.Sprintf("source kind %s differs from target kind %s.", sourceEntry.Kind, targetEntry.Kind)})
		}
		findings = append(findings, compareVisualMetrics(sourceEntry.Path, sourceEntry.Metrics, targetEntry.Metrics, sourceEntry.Provenance, targetEntry.Provenance, tolerances)...)
	}
	sourceByPath := map[string]bool{}
	for _, entry := range source {
		sourceByPath[entry.Path] = true
	}
	for _, targetEntry := range target {
		if !sourceByPath[targetEntry.Path] {
			findings = append(findings, VisualParityFinding{Severity: SeverityWarning, Code: "VISUAL.PATH", Path: targetEntry.Path, Confidence: 0.95, Message: "target visual node has no matching source node at the same tree path."})
		}
	}
	return findings
}

func compareVisualMetrics(path string, source, target VisualMetric, sourceProvenance, targetProvenance map[string]VisualProvenance, tolerances VisualToleranceProfile) []VisualParityFinding {
	findings := []VisualParityFinding{}
	if source.FontFamily != target.FontFamily {
		findings = append(findings, visualMetricFinding(path, "fontFamily", sourceProvenance, targetProvenance, fmt.Sprintf("source font family %q differs from target %q.", source.FontFamily, target.FontFamily)))
	}
	if strings.Join(source.FallbackFonts, ",") != strings.Join(target.FallbackFonts, ",") {
		findings = append(findings, visualMetricFinding(path, "fallbackFonts", sourceProvenance, targetProvenance, "font fallback stacks differ."))
	}
	typographyTolerance := firstPositiveNumber(0.25, tolerances.Typography)
	distanceTolerance := firstPositiveNumber(0.5, tolerances.Distance)
	numeric := []struct {
		name      string
		source    float64
		target    float64
		tolerance float64
	}{
		{"fontSize", source.FontSize, target.FontSize, typographyTolerance},
		{"kerning", source.Kerning, target.Kerning, typographyTolerance},
		{"lineHeight", source.LineHeight, target.LineHeight, typographyTolerance},
		{"lineSpacing", source.LineSpacing, target.LineSpacing, typographyTolerance},
		{"baselineOffset", source.BaselineOffset, target.BaselineOffset, typographyTolerance},
		{"padding", source.Padding, target.Padding, distanceTolerance},
		{"margin", source.Margin, target.Margin, distanceTolerance},
		{"spacing", source.Spacing, target.Spacing, distanceTolerance},
		{"width", source.Width, target.Width, distanceTolerance},
		{"height", source.Height, target.Height, distanceTolerance},
		{"minWidth", source.MinWidth, target.MinWidth, distanceTolerance},
		{"minHeight", source.MinHeight, target.MinHeight, distanceTolerance},
	}
	for _, metric := range numeric {
		if math.Abs(metric.source-metric.target) > metric.tolerance {
			findings = append(findings, visualMetricFinding(path, metric.name, sourceProvenance, targetProvenance, fmt.Sprintf("source %s %.2f differs from target %.2f.", metric.name, metric.source, metric.target)))
		}
	}
	return findings
}

func visualMetricFinding(path, metric string, sourceProvenance, targetProvenance map[string]VisualProvenance, message string) VisualParityFinding {
	source := provenanceForMetric(sourceProvenance, metric)
	target := provenanceForMetric(targetProvenance, metric)
	confidence := math.Round(((source.Confidence+target.Confidence)/2)*100) / 100
	return VisualParityFinding{Severity: SeverityWarning, Code: "VISUAL.METRIC", Path: path, Metric: metric, Confidence: confidence, SourceProvenance: &source, TargetProvenance: &target, Message: message}
}

func visualDiagnosticFindings(side string, diagnostics []Diagnostic) []VisualParityFinding {
	findings := []VisualParityFinding{}
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == SeverityError {
			findings = append(findings, VisualParityFinding{Severity: SeverityError, Code: "VISUAL.INVALID_SOURCE", Path: "/", Confidence: 1, Message: fmt.Sprintf("%s analysis has error diagnostic %s: %s", side, diagnostic.Code, diagnostic.Message)})
		}
	}
	return findings
}

func VisualParityText(report VisualParityReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "loom visual parity\nstatus: %s\nsource: %s (%s)\ntarget: %s (%s)\nsource nodes: %d\ntarget nodes: %d\nfindings: %d\n", report.Status, report.SourcePath, report.SourceDialect, report.TargetPath, report.TargetDialect, report.SourceCount, report.TargetCount, len(report.Findings))
	if len(report.Findings) == 0 {
		b.WriteString("  none\n")
		return b.String()
	}
	for _, finding := range report.Findings {
		metric := ""
		if finding.Metric != "" {
			metric = " " + finding.Metric
		}
		confidence := ""
		if finding.Confidence != 0 {
			confidence = fmt.Sprintf(" confidence %.2f", finding.Confidence)
		}
		fmt.Fprintf(&b, "  [%s] %s%s %s%s: %s\n", finding.Severity, finding.Code, metric, finding.Path, confidence, finding.Message)
	}
	return b.String()
}

var visualNumberPattern = regexp.MustCompile(`-?\d+(?:\.\d+)?`)

func numberProperty(node Node, key string) float64 {
	return firstNumericString(node.Properties[key])
}

func numberPropertyValue(node Node, key string) (float64, bool) {
	value := node.Properties[key]
	if value == "" {
		return 0, false
	}
	number, ok := firstNumericStringValue(value)
	return number, ok
}

func swiftArgumentNumber(node Node, name string) float64 {
	args := node.Properties["swiftui.arguments"]
	if args == "" {
		return 0
	}
	pattern := regexp.MustCompile(regexp.QuoteMeta(name) + `\s*:\s*(-?\d+(?:\.\d+)?)`)
	match := pattern.FindStringSubmatch(args)
	if len(match) != 2 {
		return 0
	}
	value, _ := strconv.ParseFloat(match[1], 64)
	return value
}

func swiftArgumentNumberValue(node Node, name string) (float64, bool) {
	args := node.Properties["swiftui.arguments"]
	if args == "" {
		return 0, false
	}
	pattern := regexp.MustCompile(regexp.QuoteMeta(name) + `\s*:\s*(-?\d+(?:\.\d+)?)`)
	match := pattern.FindStringSubmatch(args)
	if len(match) != 2 {
		return 0, false
	}
	value, err := strconv.ParseFloat(match[1], 64)
	return value, err == nil
}

func modifierNumberValue(modifier Modifier, name string) (float64, bool) {
	pattern := regexp.MustCompile(regexp.QuoteMeta(name) + `\s*:\s*(-?\d+(?:\.\d+)?)`)
	match := pattern.FindStringSubmatch(modifier.Arguments)
	if len(match) != 2 {
		return 0, false
	}
	value, err := strconv.ParseFloat(match[1], 64)
	return value, err == nil
}

func modifierFirstNumberValue(modifier Modifier) (float64, bool) {
	return firstNumericStringValue(modifier.Arguments)
}

func swiftUIControlMinHeight(size string, kind NodeKind) float64 {
	if kind != KindButton && kind != KindTextField && kind != KindToggle {
		return 0
	}
	switch size {
	case ".mini", "mini":
		return 24
	case ".small", "small":
		return 28
	case ".regular", "regular", "":
		return 32
	case ".large", "large":
		return 44
	case ".extraLarge", "extraLarge", "extralarge":
		return 52
	default:
		return 0
	}
}

func compactSwiftExpression(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func xamlValueOrigin(value string) string {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "{StaticResource") || strings.HasPrefix(trimmed, "{ThemeResource") || strings.HasPrefix(trimmed, "{Binding") {
		return "resource-reference"
	}
	return "source"
}

func xamlPropertyMetricValue(node Node, key string) (string, string, string, bool) {
	suffix := strings.TrimPrefix(key, "xaml.")
	if suffix == "" {
		return "", "", "", false
	}
	if resolved := node.Properties["xaml.resolved."+suffix]; resolved != "" {
		return resolved, "resolved-resource", "xaml." + suffix + " resolved resource", true
	}
	if value := node.Properties[key]; value != "" {
		origin := node.Properties["xaml.origin."+suffix]
		if origin == "" {
			origin = xamlValueOrigin(value)
		}
		return value, origin, key, true
	}
	return "", "", "", false
}

func firstNumericString(value string) float64 {
	number, _ := firstNumericStringValue(value)
	return number
}

func firstNumericStringValue(value string) (float64, bool) {
	match := visualNumberPattern.FindString(value)
	if match == "" {
		return 0, false
	}
	number, err := strconv.ParseFloat(match, 64)
	return number, err == nil
}

func visualProvenance(origin, detail string, confidence float64) VisualProvenance {
	if confidence == 0 {
		switch origin {
		case "source":
			confidence = 0.9
		case "font-material":
			confidence = 0.9
		case "resolved-resource":
			confidence = 0.85
		case "style-setter":
			confidence = 0.85
		case "explicit-style-setter":
			confidence = 0.85
		case "resource-reference":
			confidence = 0.8
		case "profile":
			confidence = 0.75
		case "default-profile":
			confidence = 0.55
		default:
			confidence = 0.4
		}
	}
	return VisualProvenance{Origin: origin, Detail: detail, Confidence: confidence}
}

func provenanceForMetric(provenance map[string]VisualProvenance, metric string) VisualProvenance {
	if value, ok := provenance[metric]; ok {
		return value
	}
	return visualProvenance("unknown", "", 0)
}

func firstPositiveNumber(fallback float64, values ...float64) float64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return fallback
}

func firstNumber(fallback float64, values ...float64) float64 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return fallback
}
