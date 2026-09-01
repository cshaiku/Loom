package loom

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func fixtureXAML(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "mainwindow.xaml")
	source := `<Grid xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation">` + body + `</Grid>`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func fixtureSwiftUI(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "contentview.swift")
	source := `import SwiftUI

struct ContentView: View {
  var body: some View {
` + body + `
  }
}
`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func fixtureQML(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "mainwindow.qml")
	source := `import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

` + body + `
`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func fixtureQtCPP(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "mainwindow.cpp")
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func fixtureFont(t *testing.T, family, subfamily string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "loom-sans.ttf")
	fullName := family + " " + subfamily
	postScriptName := strings.ReplaceAll(family, " ", "") + "-" + subfamily
	tables := map[string][]byte{
		"name": fontNameTable(map[uint16]string{1: family, 2: subfamily, 4: fullName, 6: postScriptName, 16: family, 17: subfamily}),
		"head": fontHeadTable(1000, -50, -250, 1050, 900),
		"hhea": fontHheaTable(800, -200, 200),
		"OS/2": fontOS2Table(400, 5, 800, -200, 200, 900, 300, 500, 700),
		"post": fontPostTable(0, false),
		"kern": fontKernTable(3),
	}
	data := buildSFNT(tables)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func fixtureWOFF(t *testing.T, family, subfamily string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "loom-sans.woff")
	fullName := family + " " + subfamily
	postScriptName := strings.ReplaceAll(family, " ", "") + "-" + subfamily
	tables := map[string][]byte{
		"name": fontNameTable(map[uint16]string{1: family, 2: subfamily, 4: fullName, 6: postScriptName}),
		"head": fontHeadTable(1000, -50, -250, 1050, 900),
		"hhea": fontHheaTable(800, -200, 200),
	}
	if err := os.WriteFile(path, buildWOFF(tables), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func buildSFNT(tables map[string][]byte) []byte {
	tags := make([]string, 0, len(tables))
	for tag := range tables {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	headerSize := 12 + len(tags)*16
	offset := align4(headerSize)
	out := make([]byte, offset)
	binary.BigEndian.PutUint32(out[0:4], 0x00010000)
	binary.BigEndian.PutUint16(out[4:6], uint16(len(tags)))
	writeSFNTSearchParams(out, len(tags))
	for i, tag := range tags {
		entry := 12 + i*16
		copy(out[entry:entry+4], []byte(tag))
		binary.BigEndian.PutUint32(out[entry+8:entry+12], uint32(offset))
		binary.BigEndian.PutUint32(out[entry+12:entry+16], uint32(len(tables[tag])))
		out = append(out, tables[tag]...)
		for len(out)%4 != 0 {
			out = append(out, 0)
		}
		offset = len(out)
	}
	return out
}

func buildWOFF(tables map[string][]byte) []byte {
	tags := make([]string, 0, len(tables))
	for tag := range tables {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	headerSize := 44 + len(tags)*20
	offset := align4(headerSize)
	out := make([]byte, offset)
	copy(out[0:4], "wOFF")
	binary.BigEndian.PutUint32(out[4:8], 0x00010000)
	binary.BigEndian.PutUint16(out[12:14], uint16(len(tags)))
	totalSFNTSize := align4(12 + len(tags)*16)
	for _, tag := range tags {
		totalSFNTSize += align4(len(tables[tag]))
	}
	binary.BigEndian.PutUint32(out[16:20], uint32(totalSFNTSize))
	for i, tag := range tags {
		entry := 44 + i*20
		copy(out[entry:entry+4], []byte(tag))
		binary.BigEndian.PutUint32(out[entry+4:entry+8], uint32(offset))
		binary.BigEndian.PutUint32(out[entry+8:entry+12], uint32(len(tables[tag])))
		binary.BigEndian.PutUint32(out[entry+12:entry+16], uint32(len(tables[tag])))
		out = append(out, tables[tag]...)
		for len(out)%4 != 0 {
			out = append(out, 0)
		}
		offset = len(out)
	}
	binary.BigEndian.PutUint32(out[8:12], uint32(len(out)))
	return out
}

func fontNameTable(values map[uint16]string) []byte {
	ids := make([]int, 0, len(values))
	for id := range values {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	count := len(ids)
	stringOffset := 6 + count*12
	out := make([]byte, stringOffset)
	binary.BigEndian.PutUint16(out[2:4], uint16(count))
	binary.BigEndian.PutUint16(out[4:6], uint16(stringOffset))
	stringsData := []byte{}
	for i, id := range ids {
		encoded := utf16BE(values[uint16(id)])
		entry := 6 + i*12
		binary.BigEndian.PutUint16(out[entry:entry+2], 3)
		binary.BigEndian.PutUint16(out[entry+2:entry+4], 1)
		binary.BigEndian.PutUint16(out[entry+4:entry+6], 0x0409)
		binary.BigEndian.PutUint16(out[entry+6:entry+8], uint16(id))
		binary.BigEndian.PutUint16(out[entry+8:entry+10], uint16(len(encoded)))
		binary.BigEndian.PutUint16(out[entry+10:entry+12], uint16(len(stringsData)))
		stringsData = append(stringsData, encoded...)
	}
	return append(out, stringsData...)
}

func utf16BE(value string) []byte {
	out := []byte{}
	for _, r := range value {
		if r > 0xffff {
			continue
		}
		out = binary.BigEndian.AppendUint16(out, uint16(r))
	}
	return out
}

func fontHeadTable(unitsPerEm uint16, xMin, yMin, xMax, yMax int16) []byte {
	out := make([]byte, 54)
	binary.BigEndian.PutUint32(out[0:4], 0x00010000)
	binary.BigEndian.PutUint16(out[18:20], unitsPerEm)
	putInt16(out[36:38], xMin)
	putInt16(out[38:40], yMin)
	putInt16(out[40:42], xMax)
	putInt16(out[42:44], yMax)
	return out
}

func fontHheaTable(ascender, descender, lineGap int16) []byte {
	out := make([]byte, 36)
	binary.BigEndian.PutUint32(out[0:4], 0x00010000)
	putInt16(out[4:6], ascender)
	putInt16(out[6:8], descender)
	putInt16(out[8:10], lineGap)
	return out
}

func fontOS2Table(weight, width uint16, typoAscender, typoDescender, typoLineGap int16, winAscent, winDescent uint16, xHeight, capHeight int16) []byte {
	out := make([]byte, 96)
	binary.BigEndian.PutUint16(out[0:2], 2)
	binary.BigEndian.PutUint16(out[4:6], weight)
	binary.BigEndian.PutUint16(out[6:8], width)
	putInt16(out[68:70], typoAscender)
	putInt16(out[70:72], typoDescender)
	putInt16(out[72:74], typoLineGap)
	binary.BigEndian.PutUint16(out[74:76], winAscent)
	binary.BigEndian.PutUint16(out[76:78], winDescent)
	putInt16(out[86:88], xHeight)
	putInt16(out[88:90], capHeight)
	return out
}

func fontPostTable(italicAngle float64, fixedPitch bool) []byte {
	out := make([]byte, 32)
	binary.BigEndian.PutUint32(out[0:4], 0x00030000)
	binary.BigEndian.PutUint32(out[4:8], uint32(int32(italicAngle*65536)))
	if fixedPitch {
		binary.BigEndian.PutUint32(out[12:16], 1)
	}
	return out
}

func fontKernTable(pairCount uint16) []byte {
	out := make([]byte, 18)
	binary.BigEndian.PutUint16(out[2:4], 1)
	binary.BigEndian.PutUint16(out[6:8], 14)
	binary.BigEndian.PutUint16(out[10:12], pairCount)
	return out
}

func putInt16(out []byte, value int16) {
	binary.BigEndian.PutUint16(out, uint16(value))
}

func align4(value int) int {
	for value%4 != 0 {
		value++
	}
	return value
}

func TestLoadAndValidatePatterns(t *testing.T) {
	patterns, err := LoadPatterns("../../patterns")
	if err != nil {
		t.Fatal(err)
	}
	if len(patterns) < 20 {
		t.Fatalf("expected at least 20 patterns, got %d", len(patterns))
	}
	report := ValidatePatterns("../../patterns")
	if report.Status != "ok" {
		t.Fatalf("expected valid patterns, got %#v", report.Issues)
	}
}

func TestAnalyzeQtQMLCommonLayout(t *testing.T) {
	path := fixtureQML(t, `ColumnLayout {
  Text { text: "Title" }
  RowLayout {
    TextField { placeholderText: "Name" }
    Button { text: "Save" }
  }
}`)
	analysis, err := AnalyzeQt(path)
	if err != nil {
		t.Fatal(err)
	}
	if analysis.Layout.Properties["sourceDialect"] != "qt" {
		t.Fatalf("expected qt dialect, got %#v", analysis.Layout.Properties)
	}
	if len(analysis.Layout.Children) != 1 || analysis.Layout.Children[0].Kind != KindVerticalStack {
		t.Fatalf("expected root ColumnLayout, got %#v", analysis.Layout.Children)
	}
	stack := analysis.Layout.Children[0]
	if len(stack.Children) != 2 || stack.Children[1].Kind != KindHorizontalStack {
		t.Fatalf("expected parsed Qt child layout, got %#v", stack.Children)
	}
}

func TestAnalyzeQtCPPHeuristicLayout(t *testing.T) {
	path := fixtureQtCPP(t, `auto *layout = new QVBoxLayout();
auto *title = new QLabel("Title");
auto *name = new QLineEdit();
auto *save = new QPushButton("Save");`)
	analysis, err := AnalyzeQt(path)
	if err != nil {
		t.Fatal(err)
	}
	if analysis.Layout.Properties["sourceDialect"] != "qt" {
		t.Fatalf("expected qt dialect, got %#v", analysis.Layout.Properties)
	}
	kinds := flattenedKinds(analysis.Layout)
	for _, expected := range []NodeKind{KindVerticalStack, KindText, KindTextField, KindButton} {
		if !containsNodeKind(kinds, expected) {
			t.Fatalf("expected Qt C++ kind %s in %#v", expected, kinds)
		}
	}
}

func containsNodeKind(values []NodeKind, expected NodeKind) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func TestQtToWindowsTransferUsesWinUIMappings(t *testing.T) {
	path := fixtureQML(t, `ColumnLayout {
  Text { text: "Title" }
  Button { text: "Save" }
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"patterns:transfer", path, "--from", "qt", "--to", "windows", "--format", "json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var report TransferReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.From != "qt" || report.To != "windows" || report.Summary.Unsupported != 0 {
		t.Fatalf("expected Qt to Windows transfer report, got %#v", report)
	}
}

func TestInspectParityComparesSwiftUIAndQt(t *testing.T) {
	swiftPath := fixtureSwiftUI(t, `VStack {
  Text("Title")
  Button("Save") {}
}`)
	qtPath := fixtureQML(t, `ColumnLayout {
  Text { text: "Title" }
  Button { text: "Save" }
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"inspect:parity", swiftPath, "--target", qtPath, "--from", "swiftui", "--to", "qt", "--json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var report ParityReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "ok" || report.SourceDialect != "swiftui" || report.TargetDialect != "qt" {
		t.Fatalf("expected clean SwiftUI/Qt parity report, got %#v", report)
	}
}

func TestInspectVisualParityReportsDefaultProfileDifferences(t *testing.T) {
	swiftPath := fixtureSwiftUI(t, `VStack(spacing: 8) {
  Text("Title")
  Button("Save") {}
}`)
	xamlPath := filepath.Join(t.TempDir(), "mainwindow.xaml")
	if err := os.WriteFile(xamlPath, []byte(`<StackPanel xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation" Spacing="8">
  <TextBlock Text="Title" />
  <Button Content="Save" />
</StackPanel>`), 0644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"inspect:visual-parity", swiftPath, "--target", xamlPath, "--from", "swiftui", "--to", "winui3", "--json"}, &stdout, &stderr); err == nil {
		t.Fatal("expected default platform visual profile differences to fail")
	}
	var report VisualParityReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "warning" || len(report.Findings) == 0 {
		t.Fatalf("expected warning visual parity report, got %#v", report)
	}
	if !strings.Contains(stdout.String(), "VISUAL.METRIC") {
		t.Fatalf("expected metric findings, got %s", stdout.String())
	}
	if len(report.SourceEntries) == 0 || report.SourceEntries[0].Provenance["spacing"].Origin != "source" {
		t.Fatalf("expected source spacing provenance, got %#v", report.SourceEntries)
	}
	foundConfidence := false
	for _, finding := range report.Findings {
		if finding.Code == "VISUAL.METRIC" && finding.Confidence > 0 && finding.SourceProvenance != nil && finding.TargetProvenance != nil {
			foundConfidence = true
		}
	}
	if !foundConfidence {
		t.Fatalf("expected metric finding confidence and provenance, got %#v", report.Findings)
	}
}

func TestInspectVisualParityAcceptsNormalizingProfile(t *testing.T) {
	swiftPath := fixtureSwiftUI(t, `VStack(spacing: 8) {
  Text("Title")
  Button("Save") {}
}`)
	xamlPath := filepath.Join(t.TempDir(), "mainwindow.xaml")
	if err := os.WriteFile(xamlPath, []byte(`<StackPanel xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation" Spacing="8">
  <TextBlock Text="Title" />
  <Button Content="Save" />
</StackPanel>`), 0644); err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(t.TempDir(), "visual-profile.json")
	profile := `{
  "schema_version": "1",
  "platforms": {
    "swiftui": {
      "typography": { "fontFamily": "Inter", "fallbackFonts": ["Segoe UI"], "fontSize": 14, "kerning": 0, "lineHeight": 20, "baselineOffset": 0 },
      "spacing": { "defaultPadding": 0, "defaultMargin": 0, "stackSpacing": 8 },
      "controls": { "buttonMinHeight": 32, "textFieldMinHeight": 32, "toggleMinHeight": 32, "listRowMinHeight": 32 }
    },
    "winui3": {
      "typography": { "fontFamily": "Inter", "fallbackFonts": ["Segoe UI"], "fontSize": 14, "kerning": 0, "lineHeight": 20, "baselineOffset": 0 },
      "spacing": { "defaultPadding": 0, "defaultMargin": 0, "stackSpacing": 8 },
      "controls": { "buttonMinHeight": 32, "textFieldMinHeight": 32, "toggleMinHeight": 32, "listRowMinHeight": 32 }
    }
  },
  "tolerances": { "distance": 0.01, "typography": 0.01 }
}`
	if err := os.WriteFile(profilePath, []byte(profile), 0644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"inspect:visual-parity", swiftPath, "--target", xamlPath, "--from", "swiftui", "--to", "winui3", "--profile", profilePath, "--json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var report VisualParityReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "ok" || len(report.Findings) != 0 {
		t.Fatalf("expected normalized visual profile to pass, got %#v", report)
	}
}

func TestInspectVisualParityUsesFontMaterialInputs(t *testing.T) {
	swiftPath := fixtureSwiftUI(t, `VStack(spacing: 8) {
  Text("Title")
}`)
	xamlPath := filepath.Join(t.TempDir(), "mainwindow.xaml")
	if err := os.WriteFile(xamlPath, []byte(`<StackPanel xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation" Spacing="8">
  <TextBlock Text="Title" />
</StackPanel>`), 0644); err != nil {
		t.Fatal(err)
	}
	fontPath := fixtureFont(t, "Loom Sans", "Regular")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := Run([]string{"inspect:visual-parity", swiftPath, "--target", xamlPath, "--from", "swiftui", "--to", "winui3", "--source-font", fontPath, "--target-font", fontPath, "--json"}, &stdout, &stderr)
	if err != nil {
		t.Fatal(err)
	}
	var report VisualParityReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Profile.Platforms["swiftui"].Typography.FontFamily != "Loom Sans" || report.Profile.Platforms["winui3"].Typography.FontFamily != "Loom Sans" {
		t.Fatalf("expected font material to populate visual profile, got %#v", report.Profile.Platforms)
	}
	if len(report.SourceEntries) == 0 || report.SourceEntries[0].Provenance["fontFamily"].Origin != "font-material" {
		t.Fatalf("expected font material provenance, got %#v", report.SourceEntries)
	}
}

func TestVisualParityExtractsSwiftUIModifierMetrics(t *testing.T) {
	path := fixtureSwiftUI(t, `Text("Title")
  .font(.system(size: 21))
  .kerning(1.5)
  .lineSpacing(3)
  .baselineOffset(2)
  .padding(12)
  .frame(width: 120, height: 44, minWidth: 80, minHeight: 36)`)
	analysis, err := AnalyzeSwiftUI(path)
	if err != nil {
		t.Fatal(err)
	}
	entries := visualParityEntries(analysis.Layout, DefaultVisualProfile().Platforms["swiftui"])
	if len(entries) != 1 {
		t.Fatalf("expected one entry, got %#v", entries)
	}
	metrics := entries[0].Metrics
	if metrics.FontSize != 21 || metrics.Kerning != 1.5 || metrics.LineSpacing != 3 || metrics.BaselineOffset != 2 || metrics.Padding != 12 || metrics.Width != 120 || metrics.Height != 44 || metrics.MinWidth != 80 || metrics.MinHeight != 36 {
		t.Fatalf("expected SwiftUI modifier metrics, got %#v", metrics)
	}
	for _, name := range []string{"fontSize", "kerning", "lineSpacing", "baselineOffset", "padding", "width", "height", "minWidth", "minHeight"} {
		if entries[0].Provenance[name].Origin != "source" {
			t.Fatalf("expected %s source provenance, got %#v", name, entries[0].Provenance[name])
		}
	}
}

func TestVisualParityExtractsXAMLVisualAttributesAndResourceReferences(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mainwindow.xaml")
	if err := os.WriteFile(path, []byte(`<TextBlock xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation"
  Text="Title"
  FontFamily="{StaticResource TitleFont}"
  FontSize="19"
  CharacterSpacing="20"
  LineHeight="24"
  Padding="8"
  Margin="4"
  Width="200"
  Height="32"
  MinWidth="120"
  MinHeight="28" />`), 0644); err != nil {
		t.Fatal(err)
	}
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	entries := visualParityEntries(analysis.Layout, DefaultVisualProfile().Platforms["winui3"])
	if len(entries) != 1 {
		t.Fatalf("expected one entry, got %#v", entries)
	}
	metrics := entries[0].Metrics
	if metrics.FontFamily != "{StaticResource TitleFont}" || metrics.FontSize != 19 || metrics.Kerning != 20 || metrics.LineHeight != 24 || metrics.Padding != 8 || metrics.Margin != 4 || metrics.Width != 200 || metrics.Height != 32 || metrics.MinWidth != 120 || metrics.MinHeight != 28 {
		t.Fatalf("expected XAML visual metrics, got %#v", metrics)
	}
	if entries[0].Provenance["fontFamily"].Origin != "resource-reference" {
		t.Fatalf("expected font family resource provenance, got %#v", entries[0].Provenance["fontFamily"])
	}
}

func TestVisualParityResolvesXAMLResourcesAndImplicitStyles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mainwindow.xaml")
	if err := os.WriteFile(path, []byte(`<StackPanel xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation"
  xmlns:x="http://schemas.microsoft.com/winfx/2006/xaml">
  <StackPanel.Resources>
    <FontFamily x:Key="TitleFont">Inter</FontFamily>
    <x:Double x:Key="TitleSize">21</x:Double>
    <x:Double x:Key="StyledLineHeight">25</x:Double>
    <Style TargetType="TextBlock">
      <Setter Property="FontSize" Value="{StaticResource TitleSize}" />
      <Setter Property="LineHeight" Value="{StaticResource StyledLineHeight}" />
      <Setter Property="Padding" Value="10" />
    </Style>
  </StackPanel.Resources>
  <TextBlock Text="Title" FontFamily="{StaticResource TitleFont}" />
</StackPanel>`), 0644); err != nil {
		t.Fatal(err)
	}
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Layout.Children) != 1 || len(analysis.Layout.Children[0].Children) != 1 {
		t.Fatalf("expected resource dictionary to stay out of visual tree, got %#v", analysis.Layout)
	}
	entries := visualParityEntries(analysis.Layout, DefaultVisualProfile().Platforms["winui3"])
	if len(entries) != 2 {
		t.Fatalf("expected stack and text entries, got %#v", entries)
	}
	text := entries[1]
	if text.Metrics.FontFamily != "Inter" || text.Metrics.FontSize != 21 || text.Metrics.LineHeight != 25 || text.Metrics.Padding != 10 {
		t.Fatalf("expected resolved resource/style metrics, got %#v", text.Metrics)
	}
	if text.Provenance["fontFamily"].Origin != "resolved-resource" {
		t.Fatalf("expected resolved font family provenance, got %#v", text.Provenance["fontFamily"])
	}
	for _, name := range []string{"fontSize", "lineHeight", "padding"} {
		if text.Provenance[name].Origin != "style-setter" {
			t.Fatalf("expected %s style provenance, got %#v", name, text.Provenance[name])
		}
	}
}

func TestVisualParityLoadsMergedXAMLResourceDictionaries(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "Styles"), 0755); err != nil {
		t.Fatal(err)
	}
	dictionaryPath := filepath.Join(dir, "Styles", "Typography.xaml")
	if err := os.WriteFile(dictionaryPath, []byte(`<ResourceDictionary
  xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation"
  xmlns:x="http://schemas.microsoft.com/winfx/2006/xaml">
  <FontFamily x:Key="TitleFont">Aptos</FontFamily>
  <x:Double x:Key="TitleSize">22</x:Double>
  <Style TargetType="TextBlock">
    <Setter Property="FontSize" Value="{StaticResource TitleSize}" />
    <Setter Property="Padding" Value="6" />
  </Style>
</ResourceDictionary>`), 0644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "mainwindow.xaml")
	if err := os.WriteFile(path, []byte(`<StackPanel
  xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation"
  xmlns:x="http://schemas.microsoft.com/winfx/2006/xaml">
  <StackPanel.Resources>
    <ResourceDictionary>
      <ResourceDictionary.MergedDictionaries>
        <ResourceDictionary Source="Styles/Typography.xaml" />
      </ResourceDictionary.MergedDictionaries>
    </ResourceDictionary>
  </StackPanel.Resources>
  <TextBlock Text="Title" FontFamily="{StaticResource TitleFont}" />
</StackPanel>`), 0644); err != nil {
		t.Fatal(err)
	}
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Diagnostics) != 0 {
		t.Fatalf("expected merged dictionary to load without diagnostics, got %#v", analysis.Diagnostics)
	}
	entries := visualParityEntries(analysis.Layout, DefaultVisualProfile().Platforms["winui3"])
	text := entries[1]
	if text.Metrics.FontFamily != "Aptos" || text.Metrics.FontSize != 22 || text.Metrics.Padding != 6 {
		t.Fatalf("expected merged dictionary values, got %#v", text.Metrics)
	}
	if text.Provenance["fontFamily"].Origin != "resolved-resource" || text.Provenance["fontSize"].Origin != "style-setter" {
		t.Fatalf("expected merged dictionary provenance, got %#v", text.Provenance)
	}
}

func TestAnalyzeXAMLReportsMissingMergedDictionary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mainwindow.xaml")
	if err := os.WriteFile(path, []byte(`<StackPanel
  xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation">
  <StackPanel.Resources>
    <ResourceDictionary>
      <ResourceDictionary.MergedDictionaries>
        <ResourceDictionary Source="Styles/Missing.xaml" />
      </ResourceDictionary.MergedDictionaries>
    </ResourceDictionary>
  </StackPanel.Resources>
  <TextBlock Text="Title" FontFamily="{StaticResource TitleFont}" />
</StackPanel>`), 0644); err != nil {
		t.Fatal(err)
	}
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Diagnostics) != 1 || analysis.Diagnostics[0].Code != "XAML.RESOURCE_DICTIONARY_UNRESOLVED" {
		t.Fatalf("expected missing dictionary diagnostic, got %#v", analysis.Diagnostics)
	}
	entries := visualParityEntries(analysis.Layout, DefaultVisualProfile().Platforms["winui3"])
	if entries[1].Provenance["fontFamily"].Origin != "resource-reference" {
		t.Fatalf("expected unresolved font family reference, got %#v", entries[1].Provenance["fontFamily"])
	}
}

func TestVisualParityResolvesExplicitXAMLStylesAndBasedOn(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mainwindow.xaml")
	if err := os.WriteFile(path, []byte(`<StackPanel
  xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation"
  xmlns:x="http://schemas.microsoft.com/winfx/2006/xaml">
  <StackPanel.Resources>
    <x:Double x:Key="BaseSize">18</x:Double>
    <Style x:Key="BaseTextStyle" TargetType="TextBlock">
      <Setter Property="FontSize" Value="{StaticResource BaseSize}" />
      <Setter Property="Padding" Value="4" />
    </Style>
    <Style x:Key="TitleTextBlockStyle" TargetType="TextBlock" BasedOn="BaseTextStyle">
      <Setter Property="LineHeight" Value="24" />
      <Setter Property="Padding" Value="9" />
    </Style>
  </StackPanel.Resources>
  <TextBlock Text="Title" Style="{StaticResource TitleTextBlockStyle}" />
</StackPanel>`), 0644); err != nil {
		t.Fatal(err)
	}
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	entries := visualParityEntries(analysis.Layout, DefaultVisualProfile().Platforms["winui3"])
	text := entries[1]
	if text.Metrics.FontSize != 18 || text.Metrics.LineHeight != 24 || text.Metrics.Padding != 9 {
		t.Fatalf("expected explicit based-on style metrics, got %#v", text.Metrics)
	}
	for _, name := range []string{"fontSize", "lineHeight", "padding"} {
		if text.Provenance[name].Origin != "explicit-style-setter" {
			t.Fatalf("expected %s explicit style provenance, got %#v", name, text.Provenance[name])
		}
	}
}

func TestVisualParityResolvesObjectValuedXAMLSettersAndResourceChains(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mainwindow.xaml")
	if err := os.WriteFile(path, []byte(`<StackPanel
  xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation"
  xmlns:x="http://schemas.microsoft.com/winfx/2006/xaml">
  <StackPanel.Resources>
    <Thickness x:Key="BasePadding">14</Thickness>
    <Thickness x:Key="AliasPadding">{StaticResource BasePadding}</Thickness>
    <x:Double x:Key="RealTitleSize">23</x:Double>
    <x:Double x:Key="AliasTitleSize">{StaticResource RealTitleSize}</x:Double>
    <Style x:Key="TitleTextBlockStyle" TargetType="TextBlock">
      <Setter Property="FontSize" Value="{StaticResource AliasTitleSize}" />
      <Setter Property="Padding">
        <Setter.Value>
          <Thickness>{StaticResource AliasPadding}</Thickness>
        </Setter.Value>
      </Setter>
      <Setter Property="Margin">
        <Setter.Value>
          <Thickness Left="2" Top="4" Right="6" Bottom="8" />
        </Setter.Value>
      </Setter>
    </Style>
  </StackPanel.Resources>
  <TextBlock Text="Title" Style="{StaticResource TitleTextBlockStyle}" />
</StackPanel>`), 0644); err != nil {
		t.Fatal(err)
	}
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	entries := visualParityEntries(analysis.Layout, DefaultVisualProfile().Platforms["winui3"])
	text := entries[1]
	if text.Metrics.FontSize != 23 || text.Metrics.Padding != 14 || text.Metrics.Margin != 2 {
		t.Fatalf("expected object-valued style metrics, got %#v", text.Metrics)
	}
	for _, name := range []string{"fontSize", "padding", "margin"} {
		if text.Provenance[name].Origin != "explicit-style-setter" {
			t.Fatalf("expected %s explicit style provenance, got %#v", name, text.Provenance[name])
		}
	}
}

func TestAnalyzeXAMLReportsUnresolvedExplicitStyleReference(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mainwindow.xaml")
	if err := os.WriteFile(path, []byte(`<StackPanel
  xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation">
  <TextBlock Text="Title" Style="{StaticResource MissingTextStyle}" />
</StackPanel>`), 0644); err != nil {
		t.Fatal(err)
	}
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, diagnostic := range analysis.Diagnostics {
		if diagnostic.Code == "XAML.STYLE_REFERENCE_UNRESOLVED" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected unresolved style diagnostic, got %#v", analysis.Diagnostics)
	}
}

func TestInspectVisualParityRejectsInvalidProfiles(t *testing.T) {
	swiftPath := fixtureSwiftUI(t, `VStack { Text("Title") }`)
	xamlPath := filepath.Join(t.TempDir(), "mainwindow.xaml")
	if err := os.WriteFile(xamlPath, []byte(`<StackPanel xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation"><TextBlock Text="Title" /></StackPanel>`), 0644); err != nil {
		t.Fatal(err)
	}
	profilePath := filepath.Join(t.TempDir(), "bad-profile.json")
	if err := os.WriteFile(profilePath, []byte(`{"schema_version":"1","platforms":{"swiftui":{"typography":{"fontSize":-1},"mystery":true}}}`), 0644); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"inspect:visual-parity", swiftPath, "--target", xamlPath, "--profile", profilePath, "--json"}, &stdout, &stderr); err == nil {
		t.Fatal("expected invalid visual profile to fail")
	}
}

func TestInspectFontExtractsIntrinsicMetrics(t *testing.T) {
	fontPath := fixtureFont(t, "Loom Sans", "Regular")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"inspect:font", fontPath, "--json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var report FontInspectionReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "ok" || len(report.Faces) != 1 {
		t.Fatalf("expected one clean font face, got %#v", report)
	}
	face := report.Faces[0]
	if face.Names.Family != "Loom Sans" || face.Names.PostScriptName != "LoomSans-Regular" {
		t.Fatalf("expected parsed font names, got %#v", face.Names)
	}
	if face.Metrics.UnitsPerEm != 1000 || face.Metrics.Ascender != 800 || face.Metrics.Descender != -200 || face.Metrics.CapHeight != 700 || face.Metrics.XHeight != 500 {
		t.Fatalf("expected intrinsic metrics, got %#v", face.Metrics)
	}
	if face.NormalizedMetrics.DefaultLineHeightRatio != 1.2 || face.NormalizedMetrics.BaselineRatio != 0.8 || face.ProfileTypography.FontFamily != "Loom Sans" {
		t.Fatalf("expected normalized/profile metrics, got %#v %#v", face.NormalizedMetrics, face.ProfileTypography)
	}
}

func TestInspectFontExtractsWOFFMetrics(t *testing.T) {
	fontPath := fixtureWOFF(t, "Loom Web Sans", "Regular")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"inspect:font", fontPath, "--json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var report FontInspectionReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "ok" || len(report.Faces) != 1 {
		t.Fatalf("expected one WOFF face, got %#v", report)
	}
	if report.Faces[0].Format != "woff-truetype" || report.Faces[0].Names.Family != "Loom Web Sans" {
		t.Fatalf("expected parsed WOFF material, got %#v", report.Faces[0])
	}
}

func TestInspectFontRejectsMissingInput(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"inspect:font", "--json"}, &stdout, &stderr); err == nil {
		t.Fatal("expected missing font input to fail")
	}
	if !strings.Contains(stdout.String(), "FONT.INPUT") {
		t.Fatalf("expected input diagnostic, got %s", stdout.String())
	}
}

func TestInspectQtReportsDelimiterErrors(t *testing.T) {
	path := fixtureQML(t, `ColumnLayout {
  Text { text: "Broken" }
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := Run([]string{"inspect:errors", path, "--kind", "qt", "--format", "json", "--fail-on", "error"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected malformed Qt source to fail")
	}
	if !strings.Contains(stdout.String(), "QT.PARSE") {
		t.Fatalf("expected Qt parse finding, got %s", stdout.String())
	}
}

func TestAnalyzeSwiftUICommonLayout(t *testing.T) {
	path := fixtureSwiftUI(t, `VStack(spacing: 12) {
  Text("Hello")
  Button("Save") {
    Text("Save")
  }
  TextField("Name", text: $name)
  Spacer()
}`)
	analysis, err := AnalyzeSwiftUI(path)
	if err != nil {
		t.Fatal(err)
	}
	if analysis.Layout.Properties["sourceDialect"] != "swiftui" {
		t.Fatalf("expected swiftui dialect, got %#v", analysis.Layout.Properties)
	}
	if analysis.Component != "ContentView" {
		t.Fatalf("expected component name from struct, got %q", analysis.Component)
	}
	if len(analysis.Layout.Children) != 1 || analysis.Layout.Children[0].Kind != KindVerticalStack {
		t.Fatalf("expected root VStack, got %#v", analysis.Layout.Children)
	}
	stack := analysis.Layout.Children[0]
	if len(stack.Children) < 4 {
		t.Fatalf("expected parsed SwiftUI children, got %#v", stack.Children)
	}
	if stack.Children[0].Kind != KindText || stack.Children[1].Kind != KindButton || stack.Children[2].Kind != KindTextField || stack.Children[3].Kind != KindSpacer {
		t.Fatalf("unexpected SwiftUI child kinds: %#v", stack.Children)
	}
}

func TestAnalyzeSwiftUIIgnoresUncalledTypeReferences(t *testing.T) {
	path := fixtureSwiftUI(t, `let type = SomeModel.self
VStack {
  CustomRow()
  Text("Title")
}`)
	analysis, err := AnalyzeSwiftUI(path)
	if err != nil {
		t.Fatal(err)
	}
	kinds := flattenedKinds(analysis.Layout)
	componentCount := 0
	for _, kind := range kinds {
		if kind == KindComponent {
			componentCount++
		}
	}
	if componentCount != 1 {
		t.Fatalf("expected only called custom view to become component, got %#v", kinds)
	}
}

func TestSwiftUIToWindowsTransferUsesWinUIMappings(t *testing.T) {
	path := fixtureSwiftUI(t, `VStack {
  Text("Title")
  Button("Save") {}
}`)
	analysis, err := AnalyzeSwiftUI(path)
	if err != nil {
		t.Fatal(err)
	}
	patterns, err := LoadPatterns("../../patterns")
	if err != nil {
		t.Fatal(err)
	}
	report := Transfer(analysis, patterns, "swiftui", "windows")
	if report.To != "windows" {
		t.Fatalf("expected public target to remain windows, got %q", report.To)
	}
	foundButton := false
	for _, item := range report.Items {
		if item.Kind == KindButton {
			foundButton = true
			if !contains(item.TargetConstructs, "Button") {
				t.Fatalf("expected WinUI button mapping, got %#v", item)
			}
		}
	}
	if !foundButton {
		t.Fatalf("expected button transfer item, got %#v", report.Items)
	}
}

func TestCLISwiftUIInspectionAndTransfer(t *testing.T) {
	path := fixtureSwiftUI(t, `HStack {
  Image("gear")
  Toggle("Enabled", isOn: $enabled)
}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"inspect:swiftui", path, "--format", "json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var analysis Analysis
	if err := json.Unmarshal(stdout.Bytes(), &analysis); err != nil {
		t.Fatal(err)
	}
	if analysis.Layout.Children[0].Kind != KindHorizontalStack {
		t.Fatalf("expected hstack from CLI, got %#v", analysis.Layout.Children)
	}
	stdout.Reset()
	if err := Run([]string{"patterns:transfer", path, "--from", "macos", "--to", "windows", "--format", "json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var report TransferReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.From != "macos" || report.To != "windows" || report.Summary.Unsupported != 0 {
		t.Fatalf("expected macos to windows transfer report, got %#v", report)
	}
}

func TestInspectSourceAutoDetectsSwiftUI(t *testing.T) {
	path := fixtureSwiftUI(t, `Text("Auto")`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"inspect:source", path, "--json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"sourceDialect": "swiftui"`) {
		t.Fatalf("expected inspect:source to auto-detect SwiftUI, got %s", stdout.String())
	}
}

func TestInspectSwiftUIReportsDelimiterErrors(t *testing.T) {
	path := fixtureSwiftUI(t, `VStack {
  Text("Unclosed")
`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := Run([]string{"inspect:errors", path, "--kind", "swift", "--format", "json", "--fail-on", "error"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected malformed SwiftUI to fail")
	}
	if !strings.Contains(stdout.String(), "SWIFTUI.PARSE") {
		t.Fatalf("expected SwiftUI parse finding, got %s", stdout.String())
	}
}

func TestAnalyzeXAMLUnsupportedNativeBoundary(t *testing.T) {
	path := fixtureXAML(t, `<NavigationView PaneTitle="Shell"><Button Content="Save" /></NavigationView>`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(analysis.Diagnostics) != 1 {
		t.Fatalf("expected one diagnostic, got %d", len(analysis.Diagnostics))
	}
	if analysis.Diagnostics[0].Code != "XAML.UNSUPPORTED_COMPONENT_BOUNDARY" {
		t.Fatalf("unexpected diagnostic: %#v", analysis.Diagnostics[0])
	}
	boundary := analysis.Layout.Children[0].Children[0]
	if boundary.Kind != KindComponent {
		t.Fatalf("expected component boundary, got %s", boundary.Kind)
	}
	if boundary.Properties["componentBoundary"] != "native-winui-control" {
		t.Fatalf("missing native boundary metadata: %#v", boundary.Properties)
	}
}

func TestAnalyzeXAMLGridDefinitionsBecomeMetadata(t *testing.T) {
	path := fixtureXAML(t, `<Grid.RowDefinitions><RowDefinition Height="Auto" /><RowDefinition Height="*" /></Grid.RowDefinitions><Grid.ColumnDefinitions><ColumnDefinition Width="240" /><ColumnDefinition Width="*" /></Grid.ColumnDefinitions><TextBlock Text="Name" />`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	grid := analysis.Layout.Children[0]
	if got := grid.Properties["xaml.Grid.RowDefinitions"]; got != "Auto,*" {
		t.Fatalf("unexpected row definition metadata: %q", got)
	}
	if got := grid.Properties["xaml.Grid.ColumnDefinitions"]; got != "240,*" {
		t.Fatalf("unexpected column definition metadata: %q", got)
	}
	for _, diagnostic := range analysis.Diagnostics {
		if diagnostic.Code == "XAML.UNSUPPORTED_COMPONENT_BOUNDARY" {
			t.Fatalf("grid definitions should not become unsupported boundaries: %#v", diagnostic)
		}
	}
	if len(grid.Children) != 1 || grid.Children[0].Kind != KindText {
		t.Fatalf("expected only the visible TextBlock child, got %#v", grid.Children)
	}
}

func TestAccessibilityAuditSuggestedFixes(t *testing.T) {
	path := fixtureXAML(t, `<NavigationView /><Button Width="20" Height="20" />`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	report := Audit(analysis)
	if report.Summary.Warnings == 0 && report.Summary.Errors == 0 {
		t.Fatal("expected audit findings")
	}
	foundBoundary := false
	foundFixes := false
	for _, finding := range report.Findings {
		if finding.Code == "AUDIT070" {
			foundBoundary = true
		}
		if len(finding.SuggestedFixes) > 0 {
			foundFixes = true
		}
	}
	if !foundBoundary {
		t.Fatal("expected unsupported native boundary audit finding")
	}
	if !foundFixes {
		t.Fatal("expected structured suggested fixes")
	}
}

func TestTransferFlagsUnsupportedNativeBoundary(t *testing.T) {
	path := fixtureXAML(t, `<NavigationView />`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	patterns, err := LoadPatterns("../../patterns")
	if err != nil {
		t.Fatal(err)
	}
	report := Transfer(analysis, patterns, "winui3", "macos")
	if report.Summary.Unsupported == 0 {
		t.Fatalf("expected unsupported transfer item, got %#v", report.Summary)
	}
	if !strings.Contains(report.ASCIIPattern, `\-- grid`) {
		t.Fatalf("expected ASCII tree in transfer report, got %q", report.ASCIIPattern)
	}
}

func TestTransferIncludesGridTrackPolicy(t *testing.T) {
	path := fixtureXAML(t, `<Grid.RowDefinitions><RowDefinition Height="Auto" /><RowDefinition Height="*" /></Grid.RowDefinitions><Button Content="Save" />`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	patterns, err := LoadPatterns("../../patterns")
	if err != nil {
		t.Fatal(err)
	}
	report := Transfer(analysis, patterns, "winui3", "macos")
	found := false
	for _, item := range report.Items {
		if item.Kind == KindGrid && strings.Contains(strings.Join(item.Policies, " "), "row/column tracks") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected grid track transfer policy, got %#v", report.Items)
	}
}

func TestTransferMacOSTargetUsesSwiftUIMappings(t *testing.T) {
	path := fixtureXAML(t, `<Button Content="Save" />`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	patterns, err := LoadPatterns("../../patterns")
	if err != nil {
		t.Fatal(err)
	}
	report := Transfer(analysis, patterns, "winui3", "macos")
	if report.To != "macos" {
		t.Fatalf("expected public route target to remain macos, got %q", report.To)
	}
	for _, item := range report.Items {
		if item.Kind == KindButton && !contains(item.TargetConstructs, "Button") {
			t.Fatalf("expected macos route to use SwiftUI button mapping, got %#v", item)
		}
	}
}

func TestOSErrorSuggestionsMatchStaticResource(t *testing.T) {
	report := OSErrorSuggestions("winui3", "StaticResource not found")
	if report.Status != "matched" {
		t.Fatalf("expected matched report, got %s", report.Status)
	}
	if len(report.Suggestions) == 0 || len(report.Suggestions[0].SuggestedFixes) == 0 {
		t.Fatalf("expected suggestions with fixes, got %#v", report.Suggestions)
	}
}

func TestOSErrorSuggestionsPlatformAndQueryFiltering(t *testing.T) {
	allWinUI := OSErrorSuggestions("winui3", "")
	if allWinUI.Status != "ok" || len(allWinUI.Suggestions) < 3 {
		t.Fatalf("expected all WinUI suggestions, got %#v", allWinUI)
	}
	xamlParse := OSErrorSuggestions("xaml", "")
	if xamlParse.Status != "ok" || len(xamlParse.Suggestions) != 1 || xamlParse.Suggestions[0].Category != "parse" {
		t.Fatalf("expected xaml parse suggestion, got %#v", xamlParse)
	}
	empty := OSErrorSuggestions("windows", "no matching issue")
	if empty.Status != "empty" || len(empty.Suggestions) != 0 {
		t.Fatalf("expected empty windows no-match report, got %#v", empty)
	}
}

func TestQueryMatchesCaseAndTokenizedInput(t *testing.T) {
	haystack := "winui3 resources staticresource unresolved resource dictionaries"
	for _, query := range []string{"StaticResource", "static-resource failed", "RESOURCE_DICTIONARY"} {
		if !queryMatches(haystack, query) {
			t.Fatalf("expected query %q to match %q", query, haystack)
		}
	}
	if queryMatches(haystack, "xml") {
		t.Fatal("short unrelated tokens should not match")
	}
}

func TestUnknownCommandGuidesHumansAndAgents(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := Run([]string{"not:a-command"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected unknown command to fail")
	}
	message := err.Error()
	for _, needle := range []string{"not:a-command", "loom help", "loom list --json"} {
		if !strings.Contains(message, needle) {
			t.Fatalf("expected unknown command guidance to include %q, got %q", needle, message)
		}
	}
}

func TestCLIJSONAndVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"version"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "loom 0.23.0" {
		t.Fatalf("unexpected version output: %q", got)
	}
	stdout.Reset()
	if err := Run([]string{"list", "--category", "patterns", "--json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var commands []CommandInfo
	if err := json.Unmarshal(stdout.Bytes(), &commands); err != nil {
		t.Fatal(err)
	}
	if len(commands) == 0 {
		t.Fatal("expected commands in JSON output")
	}
	if strings.Contains(stdout.String(), `\u003c`) {
		t.Fatalf("command JSON should keep synopsis placeholders readable, got %q", stdout.String())
	}
}

func TestFunctionJSONUsesStableEmptyArrays(t *testing.T) {
	patternReport := ValidatePatterns("../../patterns")
	text, err := prettyJSON(patternReport)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text, `"issues": null`) {
		t.Fatalf("expected empty issues array, got %s", text)
	}
	suggestions := OSErrorSuggestions("windows", "no matching issue")
	text, err = prettyJSON(suggestions)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(text, `"suggestions": null`) {
		t.Fatalf("expected empty suggestions array, got %s", text)
	}
	transfer := Transfer(Analysis{Layout: Node{Children: []Node{{Kind: KindUnsupported, Expression: "Unknown"}}}}, nil, "winui3", "macos")
	text, err = prettyJSON(transfer)
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{`"sourceConstructs": []`, `"targetConstructs": []`, `"contracts": []`, `"policies": []`} {
		if !strings.Contains(text, needle) {
			t.Fatalf("expected stable empty array %s in %s", needle, text)
		}
	}
}

func TestLineEndingOptionControlsStdoutAndFiles(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"--line-ending", "crlf", "version"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != "loom 0.23.0\r\n" {
		t.Fatalf("expected CRLF stdout, got %q", got)
	}

	path := fixtureXAML(t, `<Button Content="Save" />`)
	out := filepath.Join(t.TempDir(), "tree.txt")
	stdout.Reset()
	if err := Run([]string{"inspect:ascii", path, "--output", out, "--line-ending", "crlf"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(stdout.String(), "\r\n") {
		t.Fatalf("expected CRLF write confirmation, got %q", stdout.String())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte("\r\n")) || bytes.Contains(data, []byte("\r\r\n")) {
		t.Fatalf("expected normalized CRLF file output, got %q", string(data))
	}
	if err := Run([]string{"--line-ending", "weird", "version"}, &stdout, &stderr); err == nil {
		t.Fatal("expected invalid line ending to fail")
	}
}

func TestLineEndingPolicyFileExists(t *testing.T) {
	data, err := os.ReadFile("../../.gitattributes")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, rule := range []string{"* text=auto eol=lf", "*.bat text eol=crlf", "*.cmd text eol=crlf", "*.ps1 text eol=crlf"} {
		if !strings.Contains(text, rule) {
			t.Fatalf("missing line-ending policy rule %q in %s", rule, text)
		}
	}
}

func TestCLIQuietOutputWrite(t *testing.T) {
	path := fixtureXAML(t, `<Button Content="Save" />`)
	out := filepath.Join(t.TempDir(), "tree.txt")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"--quiet", "inspect:ascii", path, "--output", out}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("quiet write should not print success chatter, got %q", stdout.String())
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "button / Button") {
		t.Fatalf("unexpected ASCII output: %s", string(data))
	}
}

func TestCommandCatalogCoverage(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"list", "--json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var commands []CommandInfo
	if err := json.Unmarshal(stdout.Bytes(), &commands); err != nil {
		t.Fatal(err)
	}
	if len(commands) < 16 {
		t.Fatalf("expected expanded command coverage, got %d", len(commands))
	}
	found := 0
	required := map[string]struct{}{"config:validate": {}, "config:schema": {}, "checks:command-catalog": {}, "guards:summary": {}, "self-heal:plan": {}, "inspect:errors": {}, "project:build": {}, "generate:xaml": {}}
	for _, command := range commands {
		if _, ok := required[command.Command]; ok {
			found++
			delete(required, command.Command)
		}
	}
	if len(required) > 0 {
		missing := make([]string, 0, len(required))
		for command := range required {
			missing = append(missing, command)
		}
		t.Fatalf("missing required command coverage: %s", strings.Join(missing, ", "))
	}
	if found != 8 {
		t.Fatalf("unexpected required command count %d", found)
	}
}

func TestDiagnosticCommandJSON(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"checks:command-catalog", "--json"}, &stdout, &stderr); err != nil {
		t.Fatal(err)
	}
	var report LoomCommandCatalogCheckReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "ok" {
		t.Fatalf("expected catalog checks ok, got %s", report.Status)
	}
}

func TestInspectErrorsAndFailMode(t *testing.T) {
	path := fixtureXAML(t, `<Grid><TextBlock></Grid>`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := Run([]string{"inspect:errors", path, "--kind", "xaml", "--format", "json", "--fail-on", "error"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected malformed xaml to fail")
	}
	if !strings.Contains(strings.TrimSpace(stdout.String()), "\"status\": \"error\"") && !strings.Contains(stdout.String(), "LOOM.XAML") {
		t.Fatalf("expected error report, got %q", stdout.String())
	}
	var report LoomErrorInspectionReport
	if err2 := json.Unmarshal(stdout.Bytes(), &report); err2 == nil {
		if report.Status != "error" {
			t.Fatalf("expected report status error, got %s", report.Status)
		}
	}
	if !strings.Contains(err.Error(), "command completed") {
		t.Fatal(err)
	}
	_ = stderr
}

func TestOutputRefusesInputCollisionAndExistingFile(t *testing.T) {
	path := fixtureXAML(t, `<TextBlock Text="Keep" />`)
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"inspect:ascii", path, "--output", path}, &stdout, &stderr); err == nil {
		t.Fatal("expected output over input to be refused")
	}
	if after, err := os.ReadFile(path); err != nil || string(after) != string(original) {
		t.Fatalf("source was modified: err=%v content=%q", err, string(after))
	}
	out := filepath.Join(t.TempDir(), "analysis.txt")
	if err := os.WriteFile(out, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"inspect:ascii", path, "--output", out}, &stdout, &stderr); err == nil {
		t.Fatal("expected existing output to be refused without --overwrite")
	}
	if got, _ := os.ReadFile(out); string(got) != "existing" {
		t.Fatalf("existing output was modified: %q", string(got))
	}
}

func TestSourceCommandsRejectUnknownAndMissingFlags(t *testing.T) {
	path := fixtureXAML(t, `<TextBlock Text="Hello" />`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"inspect:xaml", path, "--formt", "json"}, &stdout, &stderr); err == nil {
		t.Fatal("expected unknown option to fail")
	}
	if err := Run([]string{"inspect:xaml", path, "--output", "--json"}, &stdout, &stderr); err == nil {
		t.Fatal("expected missing output value to fail")
	}
	if err := Run([]string{"inspect:xaml", path, "--format", "yaml"}, &stdout, &stderr); err == nil {
		t.Fatal("expected unsupported format to fail")
	}
}

func TestParityFailsOnMalformedSourceAndTreeShape(t *testing.T) {
	malformedSwift := fixtureSwiftUI(t, `VStack {
  Text("Broken")
`)
	qtPath := fixtureQML(t, `ColumnLayout { Text { text: "Broken" } }`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := Run([]string{"inspect:parity", malformedSwift, "--target", qtPath, "--from", "swiftui", "--to", "qt", "--json"}, &stdout, &stderr); err == nil {
		t.Fatal("expected malformed source parity to fail")
	}
	if !strings.Contains(stdout.String(), "PARITY.INVALID_SOURCE") {
		t.Fatalf("expected invalid-source finding, got %s", stdout.String())
	}

	flatA := fixtureSwiftUI(t, `VStack {
  HStack { Text("A") }
  Button("B") {}
}`)
	flatB := fixtureSwiftUI(t, `VStack {
  HStack { }
  Text("A")
  Button("B") {}
}`)
	stdout.Reset()
	if err := Run([]string{"inspect:parity", flatA, "--target", flatB, "--from", "swiftui", "--to", "swiftui", "--json"}, &stdout, &stderr); err == nil {
		t.Fatal("expected tree shape parity mismatch to fail")
	}
	if !strings.Contains(stdout.String(), "PARITY.CHILDREN") && !strings.Contains(stdout.String(), "PARITY.PATH") {
		t.Fatalf("expected structural finding, got %s", stdout.String())
	}
}

func TestPatternValidationRejectsUnknownFieldsAndBadContract(t *testing.T) {
	dir := t.TempDir()
	bad := `{
  "schema_version":"1",
  "id":"Bad_ID",
  "version":"one",
  "name":"bad",
  "kind":"notAKind",
  "status":"unknown",
  "category":"",
  "intent":{"summary":"","useWhen":[],"avoidWhen":[]},
  "semantics":{"role":"","childPolicy":"","sizing":"","ordering":""},
  "attributes":[{"name":"mode","valueType":"enum","required":false,"description":"mode","defaultValue":"wrong","allowedValues":["right"]},{"name":"mode","valueType":"enum","required":false,"description":"duplicate"}],
  "constraints":[],
  "accessibility":{"role":"","nameSource":"","focusBehavior":"","notes":[]},
  "mappings":[{"platform":"swiftui","constructs":[],"strategy":"","caveats":[]},{"platform":"swiftui","constructs":["Button"],"strategy":"duplicate","caveats":[]}],
  "tags":[],
  "unknown":true
}`
	if err := os.WriteFile(filepath.Join(dir, "bad.pattern.json"), []byte(bad), 0644); err != nil {
		t.Fatal(err)
	}
	report := ValidatePatterns(dir)
	if report.Status != "error" || len(report.Issues) == 0 {
		t.Fatalf("expected invalid pattern report, got %#v", report)
	}
	if !strings.Contains(report.Issues[0].Detail, "unknown") {
		t.Fatalf("expected unknown field rejection, got %#v", report.Issues)
	}
}

func TestAccessibilityDoesNotTreatPlaceholdersResourcesOrIdentifiersAsNames(t *testing.T) {
	path := fixtureXAML(t, `<StackPanel>
  <Image Source="icon.png" />
  <Button AutomationId="saveButton" />
  <TextBox PlaceholderText="Name" />
  <Button Content="Tiny" Width="0" Height="-1" />
</StackPanel>`)
	analysis, err := AnalyzeXAML(path)
	if err != nil {
		t.Fatal(err)
	}
	report := Audit(analysis)
	codes := map[string]int{}
	for _, finding := range report.Findings {
		codes[finding.Code]++
	}
	if codes["AUDIT030"] == 0 {
		t.Fatalf("expected image resource not to count as accessible name: %#v", report.Findings)
	}
	if codes["AUDIT020"] == 0 {
		t.Fatalf("expected automation id not to count as button name: %#v", report.Findings)
	}
	if codes["AUDIT031"] == 0 {
		t.Fatalf("expected placeholder not to count as text input label: %#v", report.Findings)
	}
	if codes["AUDIT060"] == 0 {
		t.Fatalf("expected zero/negative target size finding: %#v", report.Findings)
	}
}

func TestFrameModifiersAreDeterministic(t *testing.T) {
	got := frameModifiers(map[string]string{"Height": "20", "Width": "10", "MinHeight": "4", "MinWidth": "3"})
	if len(got) != 1 {
		t.Fatalf("expected frame modifier, got %#v", got)
	}
	want := "width: 10, height: 20, minWidth: 3, minHeight: 4"
	if got[0].Arguments != want {
		t.Fatalf("expected deterministic frame args %q, got %q", want, got[0].Arguments)
	}
}

func TestManifestValidationRequiresSchemaVersion(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "ContentView.swift")
	if err := os.WriteFile(source, []byte("import SwiftUI\n"), 0644); err != nil {
		t.Fatal(err)
	}
	manifest := filepath.Join(dir, "loom.json")
	body := `{"project":"Demo","source":"ContentView.swift","rootView":"ContentView","target":"winui3","components":["ContentView"]}`
	if err := os.WriteFile(manifest, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	report := DiagnosticsProjectConfigValidate(manifest, dir)
	if report.Status != "error" {
		t.Fatalf("expected missing schema_version to fail, got %#v", report)
	}
	found := false
	for _, issue := range report.Issues {
		if issue.Code == "manifest.schema_version.missing" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected missing schema version issue, got %#v", report.Issues)
	}
}

func TestSuggestionsMatchDiagnosticCodeExactly(t *testing.T) {
	swift := OSErrorSuggestions("swiftui", "SWIFTUI.PARSE")
	if swift.Status != "matched" || len(swift.Suggestions) != 1 || swift.Suggestions[0].Matcher != "SWIFTUI.PARSE" {
		t.Fatalf("expected exact SwiftUI parse match, got %#v", swift)
	}
	winui := OSErrorSuggestions("winui3", "SWIFTUI.PARSE")
	if winui.Status != "empty" || len(winui.Suggestions) != 0 {
		t.Fatalf("expected platform mismatch to suppress code match, got %#v", winui)
	}
}

func TestAnalyzeRejectsOversizedSource(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "large.swift")
	data := bytes.Repeat([]byte(" "), int(MaxInputBytes)+1)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := AnalyzeSwiftUI(path); err == nil {
		t.Fatal("expected oversized source to be rejected")
	}
}

func TestUnavailableSwiftOnlyCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := Run([]string{"generate:xaml", "MainView.swift"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected go-only unsupported error")
	}
}
