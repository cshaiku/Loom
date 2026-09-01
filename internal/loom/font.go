package loom

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf16"
)

const MaxFontBytes int64 = 64 * 1024 * 1024

type FontNameSet struct {
	Family            string   `json:"family,omitempty"`
	Subfamily         string   `json:"subfamily,omitempty"`
	FullName          string   `json:"fullName,omitempty"`
	PostScriptName    string   `json:"postScriptName,omitempty"`
	TypographicFamily string   `json:"typographicFamily,omitempty"`
	TypographicStyle  string   `json:"typographicStyle,omitempty"`
	FallbackNames     []string `json:"fallbackNames,omitempty"`
}

type FontBounds struct {
	XMin int16 `json:"xMin"`
	YMin int16 `json:"yMin"`
	XMax int16 `json:"xMax"`
	YMax int16 `json:"yMax"`
}

type FontMetricSet struct {
	UnitsPerEm           uint16      `json:"unitsPerEm,omitempty"`
	Bounds               *FontBounds `json:"bounds,omitempty"`
	Ascender             int16       `json:"ascender,omitempty"`
	Descender            int16       `json:"descender,omitempty"`
	LineGap              int16       `json:"lineGap,omitempty"`
	TypographicAscender  int16       `json:"typographicAscender,omitempty"`
	TypographicDescender int16       `json:"typographicDescender,omitempty"`
	TypographicLineGap   int16       `json:"typographicLineGap,omitempty"`
	WinAscent            uint16      `json:"winAscent,omitempty"`
	WinDescent           uint16      `json:"winDescent,omitempty"`
	CapHeight            int16       `json:"capHeight,omitempty"`
	XHeight              int16       `json:"xHeight,omitempty"`
	WeightClass          uint16      `json:"weightClass,omitempty"`
	WidthClass           uint16      `json:"widthClass,omitempty"`
	ItalicAngle          float64     `json:"italicAngle,omitempty"`
	FixedPitch           bool        `json:"fixedPitch,omitempty"`
	KerningPairs         int         `json:"kerningPairs,omitempty"`
}

type FontNormalizedMetrics struct {
	AscenderRatio             float64 `json:"ascenderRatio,omitempty"`
	DescenderRatio            float64 `json:"descenderRatio,omitempty"`
	LineGapRatio              float64 `json:"lineGapRatio,omitempty"`
	TypographicAscenderRatio  float64 `json:"typographicAscenderRatio,omitempty"`
	TypographicDescenderRatio float64 `json:"typographicDescenderRatio,omitempty"`
	CapHeightRatio            float64 `json:"capHeightRatio,omitempty"`
	XHeightRatio              float64 `json:"xHeightRatio,omitempty"`
	DefaultLineHeightRatio    float64 `json:"defaultLineHeightRatio,omitempty"`
	BaselineRatio             float64 `json:"baselineRatio,omitempty"`
}

type FontFaceReport struct {
	Index             int                     `json:"index"`
	Path              string                  `json:"path"`
	Format            string                  `json:"format"`
	Names             FontNameSet             `json:"names"`
	Metrics           FontMetricSet           `json:"metrics"`
	NormalizedMetrics FontNormalizedMetrics   `json:"normalizedMetrics"`
	ProfileTypography VisualTypographyProfile `json:"profileTypography"`
}

type FontInspectionReport struct {
	SchemaVersion string           `json:"schema_version"`
	Status        string           `json:"status"`
	Source        string           `json:"source"`
	ResolvedPath  string           `json:"resolvedPath,omitempty"`
	FamilyQuery   string           `json:"familyQuery,omitempty"`
	Faces         []FontFaceReport `json:"faces"`
	Diagnostics   []Diagnostic     `json:"diagnostics,omitempty"`
}

type fontTable struct {
	tag    string
	offset uint32
	length uint32
}

type fontNameCandidate struct {
	value string
	score int
}

func InspectFontSource(path, family string) FontInspectionReport {
	report := FontInspectionReport{SchemaVersion: "1", Status: "ok", Source: path, FamilyQuery: family}
	if path == "" && family == "" {
		report.Status = "error"
		report.Diagnostics = append(report.Diagnostics, Diagnostic{Severity: SeverityError, Code: "FONT.INPUT", Message: "inspect:font requires a font path or --family name."})
		return report
	}
	if family != "" && path == "" {
		resolved, diagnostics := ResolveFontFamily(family)
		report.Diagnostics = append(report.Diagnostics, diagnostics...)
		if resolved == "" {
			report.Status = "error"
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Severity: SeverityError, Code: "FONT.NOT_FOUND", Message: "could not resolve installed font family " + family + "."})
			return report
		}
		path = resolved
		report.Source = resolved
		report.ResolvedPath = resolved
	}
	faces, diagnostics := InspectFontFile(path)
	report.Diagnostics = append(report.Diagnostics, diagnostics...)
	report.Faces = faces
	if len(faces) == 0 {
		report.Status = "error"
		if len(report.Diagnostics) == 0 {
			report.Diagnostics = append(report.Diagnostics, Diagnostic{Severity: SeverityError, Code: "FONT.PARSE", Message: "no supported font faces found."})
		}
		return report
	}
	for _, diagnostic := range report.Diagnostics {
		if diagnostic.Severity == SeverityError {
			report.Status = "error"
			return report
		}
		if diagnostic.Severity == SeverityWarning {
			report.Status = "warning"
		}
	}
	return report
}

func FontInspectionText(report FontInspectionReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "loom font inspection\nstatus: %s\nsource: %s\n", report.Status, firstNonEmpty(report.Source, report.FamilyQuery))
	if report.FamilyQuery != "" {
		fmt.Fprintf(&b, "family query: %s\n", report.FamilyQuery)
	}
	if report.ResolvedPath != "" {
		fmt.Fprintf(&b, "resolved: %s\n", report.ResolvedPath)
	}
	fmt.Fprintf(&b, "faces: %d\n", len(report.Faces))
	for _, face := range report.Faces {
		fmt.Fprintf(&b, "  face %d: %s", face.Index, firstNonEmpty(face.Names.FullName, face.Names.Family, "unnamed"))
		if face.Names.PostScriptName != "" {
			fmt.Fprintf(&b, " (%s)", face.Names.PostScriptName)
		}
		b.WriteString("\n")
		fmt.Fprintf(&b, "    family: %s\n", firstNonEmpty(face.Names.TypographicFamily, face.Names.Family, "unknown"))
		fmt.Fprintf(&b, "    unitsPerEm: %d ascender: %d descender: %d lineGap: %d\n", face.Metrics.UnitsPerEm, face.Metrics.Ascender, face.Metrics.Descender, face.Metrics.LineGap)
		if face.Metrics.TypographicAscender != 0 || face.Metrics.TypographicDescender != 0 || face.Metrics.TypographicLineGap != 0 {
			fmt.Fprintf(&b, "    typographic: ascender %d descender %d lineGap %d\n", face.Metrics.TypographicAscender, face.Metrics.TypographicDescender, face.Metrics.TypographicLineGap)
		}
		if face.Metrics.CapHeight != 0 || face.Metrics.XHeight != 0 {
			fmt.Fprintf(&b, "    capHeight: %d xHeight: %d\n", face.Metrics.CapHeight, face.Metrics.XHeight)
		}
		fmt.Fprintf(&b, "    normalized: lineHeight %.4f baseline %.4f capHeight %.4f xHeight %.4f\n", face.NormalizedMetrics.DefaultLineHeightRatio, face.NormalizedMetrics.BaselineRatio, face.NormalizedMetrics.CapHeightRatio, face.NormalizedMetrics.XHeightRatio)
	}
	for _, diagnostic := range report.Diagnostics {
		fmt.Fprintf(&b, "  [%s] %s: %s\n", diagnostic.Severity, diagnostic.Code, diagnostic.Message)
	}
	return b.String()
}

func InspectFontFile(path string) ([]FontFaceReport, []Diagnostic) {
	data, err := readFontFile(path)
	if err != nil {
		return nil, []Diagnostic{{Severity: SeverityError, Code: "FONT.READ", Message: err.Error()}}
	}
	formatOverride := ""
	if len(data) >= 4 && string(data[:4]) == "wOFF" {
		decoded, format, err := decodeWOFF(data)
		if err != nil {
			return nil, []Diagnostic{{Severity: SeverityError, Code: "FONT.WOFF", Message: err.Error()}}
		}
		data = decoded
		formatOverride = format
	}
	offsets, format, err := fontFaceOffsets(data)
	if err != nil {
		return nil, []Diagnostic{{Severity: SeverityError, Code: "FONT.FORMAT", Message: err.Error()}}
	}
	if formatOverride != "" {
		format = formatOverride
	}
	reports := []FontFaceReport{}
	diagnostics := []Diagnostic{}
	for i, offset := range offsets {
		face, faceDiagnostics := inspectFontFace(data, path, format, i, offset)
		diagnostics = append(diagnostics, faceDiagnostics...)
		if face.Names.Family != "" || face.Metrics.UnitsPerEm != 0 {
			reports = append(reports, face)
		}
	}
	return reports, diagnostics
}

func decodeWOFF(data []byte) ([]byte, string, error) {
	if len(data) < 44 {
		return nil, "", fmt.Errorf("WOFF header is truncated")
	}
	flavor := binary.BigEndian.Uint32(data[4:8])
	numTables := int(binary.BigEndian.Uint16(data[12:14]))
	totalSize := int(binary.BigEndian.Uint32(data[16:20]))
	if numTables == 0 || numTables > 4096 {
		return nil, "", fmt.Errorf("WOFF has unsupported table count %d", numTables)
	}
	if len(data) < 44+numTables*20 {
		return nil, "", fmt.Errorf("WOFF table directory is truncated")
	}
	if totalSize < 12+numTables*16 || totalSize > int(MaxFontBytes) {
		return nil, "", fmt.Errorf("WOFF total sfnt size %d is unsupported", totalSize)
	}
	out := make([]byte, fontAlign4(12+numTables*16))
	binary.BigEndian.PutUint32(out[0:4], flavor)
	binary.BigEndian.PutUint16(out[4:6], uint16(numTables))
	writeSFNTSearchParams(out, numTables)
	tableOffset := len(out)
	for i := 0; i < numTables; i++ {
		woffEntry := 44 + i*20
		sfntEntry := 12 + i*16
		tag := data[woffEntry : woffEntry+4]
		offset := int(binary.BigEndian.Uint32(data[woffEntry+4 : woffEntry+8]))
		compLength := int(binary.BigEndian.Uint32(data[woffEntry+8 : woffEntry+12]))
		origLength := int(binary.BigEndian.Uint32(data[woffEntry+12 : woffEntry+16]))
		checksum := binary.BigEndian.Uint32(data[woffEntry+16 : woffEntry+20])
		if offset < 0 || compLength < 0 || origLength < 0 || offset+compLength > len(data) {
			return nil, "", fmt.Errorf("WOFF table %q is outside the file", string(tag))
		}
		tableBytes := data[offset : offset+compLength]
		if compLength < origLength {
			decoded, err := inflateWOFFTable(tableBytes, origLength)
			if err != nil {
				return nil, "", fmt.Errorf("could not inflate WOFF table %q: %w", string(tag), err)
			}
			tableBytes = decoded
		}
		if len(tableBytes) != origLength {
			return nil, "", fmt.Errorf("WOFF table %q length mismatch", string(tag))
		}
		copy(out[sfntEntry:sfntEntry+4], tag)
		binary.BigEndian.PutUint32(out[sfntEntry+4:sfntEntry+8], checksum)
		binary.BigEndian.PutUint32(out[sfntEntry+8:sfntEntry+12], uint32(tableOffset))
		binary.BigEndian.PutUint32(out[sfntEntry+12:sfntEntry+16], uint32(origLength))
		out = append(out, tableBytes...)
		for len(out)%4 != 0 {
			out = append(out, 0)
		}
		tableOffset = len(out)
	}
	format := "woff-" + sfntFormat(string(data[4:8]), flavor)
	return out, format, nil
}

func inflateWOFFTable(data []byte, expectedLength int) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if len(decoded) != expectedLength {
		return nil, fmt.Errorf("inflated length %d does not match expected %d", len(decoded), expectedLength)
	}
	return decoded, nil
}

func fontAlign4(value int) int {
	for value%4 != 0 {
		value++
	}
	return value
}

func readFontFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("could not read font at %s", path)
	}
	if info.Size() > MaxFontBytes {
		return nil, fmt.Errorf("font exceeds maximum supported input size of %d bytes", MaxFontBytes)
	}
	return os.ReadFile(path)
}

func fontFaceOffsets(data []byte) ([]uint32, string, error) {
	if len(data) < 12 {
		return nil, "", fmt.Errorf("font file is too small")
	}
	tag := string(data[:4])
	if tag == "ttcf" {
		if len(data) < 12 {
			return nil, "", fmt.Errorf("font collection header is truncated")
		}
		count := binary.BigEndian.Uint32(data[8:12])
		if count == 0 || count > 1024 {
			return nil, "", fmt.Errorf("font collection has unsupported face count %d", count)
		}
		if uint64(len(data)) < uint64(12)+uint64(count)*4 {
			return nil, "", fmt.Errorf("font collection offset table is truncated")
		}
		offsets := make([]uint32, 0, count)
		for i := uint32(0); i < count; i++ {
			offset := binary.BigEndian.Uint32(data[12+i*4 : 16+i*4])
			if uint64(offset)+12 > uint64(len(data)) {
				return nil, "", fmt.Errorf("font collection face offset %d is outside the file", offset)
			}
			offsets = append(offsets, offset)
		}
		return offsets, "ttc", nil
	}
	if isSFNTVersion(tag, binary.BigEndian.Uint32(data[:4])) {
		return []uint32{0}, sfntFormat(tag, binary.BigEndian.Uint32(data[:4])), nil
	}
	return nil, "", fmt.Errorf("unsupported font signature %q", tag)
}

func isSFNTVersion(tag string, version uint32) bool {
	return version == 0x00010000 || tag == "OTTO" || tag == "true" || tag == "typ1"
}

func sfntFormat(tag string, version uint32) string {
	switch {
	case version == 0x00010000:
		return "truetype"
	case tag == "OTTO":
		return "opentype-cff"
	default:
		return strings.TrimSpace(tag)
	}
}

func writeSFNTSearchParams(out []byte, numTables int) {
	maxPower := 1
	entrySelector := 0
	for maxPower*2 <= numTables {
		maxPower *= 2
		entrySelector++
	}
	searchRange := maxPower * 16
	rangeShift := numTables*16 - searchRange
	binary.BigEndian.PutUint16(out[6:8], uint16(searchRange))
	binary.BigEndian.PutUint16(out[8:10], uint16(entrySelector))
	binary.BigEndian.PutUint16(out[10:12], uint16(rangeShift))
}

func inspectFontFace(data []byte, path, format string, index int, offset uint32) (FontFaceReport, []Diagnostic) {
	report := FontFaceReport{Index: index, Path: path, Format: format}
	tables, err := sfntTables(data, offset)
	if err != nil {
		return report, []Diagnostic{{Severity: SeverityError, Code: "FONT.TABLES", Message: err.Error()}}
	}
	diagnostics := []Diagnostic{}
	report.Names = parseFontNames(tableData(data, tables["name"]))
	report.Metrics.UnitsPerEm, report.Metrics.Bounds = parseHead(tableData(data, tables["head"]))
	hheaAscender, hheaDescender, hheaLineGap, ok := parseHhea(tableData(data, tables["hhea"]))
	if ok {
		report.Metrics.Ascender = hheaAscender
		report.Metrics.Descender = hheaDescender
		report.Metrics.LineGap = hheaLineGap
	}
	mergeOS2Metrics(&report.Metrics, tableData(data, tables["OS/2"]))
	mergePostMetrics(&report.Metrics, tableData(data, tables["post"]))
	report.Metrics.KerningPairs = parseKernPairCount(tableData(data, tables["kern"]))
	report.NormalizedMetrics = normalizeFontMetrics(report.Metrics)
	report.ProfileTypography = fontProfileTypography(report)
	if report.Names.Family == "" {
		diagnostics = append(diagnostics, Diagnostic{Severity: SeverityWarning, Code: "FONT.NAME", Message: "font face has no family name."})
	}
	if report.Metrics.UnitsPerEm == 0 {
		diagnostics = append(diagnostics, Diagnostic{Severity: SeverityWarning, Code: "FONT.METRICS", Message: "font face has no head unitsPerEm metric."})
	}
	return report, diagnostics
}

func sfntTables(data []byte, offset uint32) (map[string]fontTable, error) {
	if uint64(offset)+12 > uint64(len(data)) {
		return nil, fmt.Errorf("sfnt header is outside the file")
	}
	count := binary.BigEndian.Uint16(data[offset+4 : offset+6])
	tableStart := uint64(offset) + 12
	if tableStart+uint64(count)*16 > uint64(len(data)) {
		return nil, fmt.Errorf("sfnt table directory is truncated")
	}
	tables := map[string]fontTable{}
	for i := uint16(0); i < count; i++ {
		entry := tableStart + uint64(i)*16
		tag := string(data[entry : entry+4])
		tableOffset := binary.BigEndian.Uint32(data[entry+8 : entry+12])
		length := binary.BigEndian.Uint32(data[entry+12 : entry+16])
		if uint64(tableOffset)+uint64(length) > uint64(len(data)) {
			continue
		}
		tables[tag] = fontTable{tag: tag, offset: tableOffset, length: length}
	}
	return tables, nil
}

func tableData(data []byte, table fontTable) []byte {
	if table.length == 0 {
		return nil
	}
	return data[table.offset : table.offset+table.length]
}

func parseFontNames(data []byte) FontNameSet {
	names := FontNameSet{}
	if len(data) < 6 {
		return names
	}
	count := int(binary.BigEndian.Uint16(data[2:4]))
	stringOffset := int(binary.BigEndian.Uint16(data[4:6]))
	if len(data) < 6+count*12 || stringOffset > len(data) {
		return names
	}
	selected := map[uint16]fontNameCandidate{}
	for i := 0; i < count; i++ {
		entry := 6 + i*12
		platformID := binary.BigEndian.Uint16(data[entry : entry+2])
		encodingID := binary.BigEndian.Uint16(data[entry+2 : entry+4])
		languageID := binary.BigEndian.Uint16(data[entry+4 : entry+6])
		nameID := binary.BigEndian.Uint16(data[entry+6 : entry+8])
		length := int(binary.BigEndian.Uint16(data[entry+8 : entry+10]))
		offset := int(binary.BigEndian.Uint16(data[entry+10 : entry+12]))
		start := stringOffset + offset
		if start < 0 || length < 0 || start+length > len(data) {
			continue
		}
		value := decodeFontName(data[start:start+length], platformID, encodingID)
		if value == "" {
			continue
		}
		score := fontNameScore(platformID, languageID)
		if current, ok := selected[nameID]; !ok || score > current.score {
			selected[nameID] = fontNameCandidate{value: value, score: score}
		}
	}
	names.Family = selectedName(selected, 1)
	names.Subfamily = selectedName(selected, 2)
	names.FullName = selectedName(selected, 4)
	names.PostScriptName = selectedName(selected, 6)
	names.TypographicFamily = selectedName(selected, 16)
	names.TypographicStyle = selectedName(selected, 17)
	names.FallbackNames = compactUniqueStrings([]string{names.TypographicFamily, names.Family, names.FullName, names.PostScriptName})
	return names
}

func selectedName(selected map[uint16]fontNameCandidate, id uint16) string {
	if value, ok := selected[id]; ok {
		return value.value
	}
	return ""
}

func decodeFontName(data []byte, platformID, encodingID uint16) string {
	if platformID == 0 || platformID == 3 || (platformID == 2 && encodingID == 1) {
		if len(data)%2 != 0 {
			data = data[:len(data)-1]
		}
		chars := make([]uint16, 0, len(data)/2)
		for i := 0; i+1 < len(data); i += 2 {
			chars = append(chars, binary.BigEndian.Uint16(data[i:i+2]))
		}
		return strings.TrimSpace(string(utf16.Decode(chars)))
	}
	return strings.TrimSpace(string(data))
}

func fontNameScore(platformID, languageID uint16) int {
	score := 1
	if platformID == 3 {
		score += 10
	}
	if platformID == 0 {
		score += 8
	}
	switch languageID {
	case 0x0409, 0:
		score += 20
	}
	return score
}

func parseHead(data []byte) (uint16, *FontBounds) {
	if len(data) < 54 {
		return 0, nil
	}
	return binary.BigEndian.Uint16(data[18:20]), &FontBounds{
		XMin: int16(binary.BigEndian.Uint16(data[36:38])),
		YMin: int16(binary.BigEndian.Uint16(data[38:40])),
		XMax: int16(binary.BigEndian.Uint16(data[40:42])),
		YMax: int16(binary.BigEndian.Uint16(data[42:44])),
	}
}

func parseHhea(data []byte) (int16, int16, int16, bool) {
	if len(data) < 10 {
		return 0, 0, 0, false
	}
	return int16(binary.BigEndian.Uint16(data[4:6])), int16(binary.BigEndian.Uint16(data[6:8])), int16(binary.BigEndian.Uint16(data[8:10])), true
}

func mergeOS2Metrics(metrics *FontMetricSet, data []byte) {
	if len(data) < 78 {
		return
	}
	metrics.WeightClass = binary.BigEndian.Uint16(data[4:6])
	metrics.WidthClass = binary.BigEndian.Uint16(data[6:8])
	metrics.TypographicAscender = int16(binary.BigEndian.Uint16(data[68:70]))
	metrics.TypographicDescender = int16(binary.BigEndian.Uint16(data[70:72]))
	metrics.TypographicLineGap = int16(binary.BigEndian.Uint16(data[72:74]))
	metrics.WinAscent = binary.BigEndian.Uint16(data[74:76])
	metrics.WinDescent = binary.BigEndian.Uint16(data[76:78])
	if len(data) >= 90 {
		metrics.XHeight = int16(binary.BigEndian.Uint16(data[86:88]))
		metrics.CapHeight = int16(binary.BigEndian.Uint16(data[88:90]))
	}
}

func mergePostMetrics(metrics *FontMetricSet, data []byte) {
	if len(data) < 16 {
		return
	}
	raw := int32(binary.BigEndian.Uint32(data[4:8]))
	metrics.ItalicAngle = math.Round((float64(raw)/65536)*1000) / 1000
	metrics.FixedPitch = binary.BigEndian.Uint32(data[12:16]) != 0
}

func parseKernPairCount(data []byte) int {
	if len(data) < 4 {
		return 0
	}
	subtableCount := int(binary.BigEndian.Uint16(data[2:4]))
	offset := 4
	total := 0
	for i := 0; i < subtableCount && offset+6 <= len(data); i++ {
		length := int(binary.BigEndian.Uint16(data[offset+2 : offset+4]))
		coverage := binary.BigEndian.Uint16(data[offset+4 : offset+6])
		if length < 6 || offset+length > len(data) {
			break
		}
		format := coverage >> 8
		if format == 0 && length >= 14 {
			total += int(binary.BigEndian.Uint16(data[offset+6 : offset+8]))
		}
		offset += length
	}
	return total
}

func normalizeFontMetrics(metrics FontMetricSet) FontNormalizedMetrics {
	upm := float64(metrics.UnitsPerEm)
	if upm == 0 {
		return FontNormalizedMetrics{}
	}
	ascender := firstInt16(metrics.TypographicAscender, metrics.Ascender)
	descender := firstInt16(metrics.TypographicDescender, metrics.Descender)
	lineGap := firstInt16(metrics.TypographicLineGap, metrics.LineGap)
	return FontNormalizedMetrics{
		AscenderRatio:             ratio(metrics.Ascender, upm),
		DescenderRatio:            ratio(metrics.Descender, upm),
		LineGapRatio:              ratio(metrics.LineGap, upm),
		TypographicAscenderRatio:  ratio(metrics.TypographicAscender, upm),
		TypographicDescenderRatio: ratio(metrics.TypographicDescender, upm),
		CapHeightRatio:            ratio(metrics.CapHeight, upm),
		XHeightRatio:              ratio(metrics.XHeight, upm),
		DefaultLineHeightRatio:    math.Round((float64(ascender-descender+lineGap)/upm)*10000) / 10000,
		BaselineRatio:             math.Round((float64(ascender)/upm)*10000) / 10000,
	}
}

func fontProfileTypography(report FontFaceReport) VisualTypographyProfile {
	fontSize := 14.0
	lineHeight := fontSize
	if report.NormalizedMetrics.DefaultLineHeightRatio != 0 {
		lineHeight = math.Round(fontSize*report.NormalizedMetrics.DefaultLineHeightRatio*100) / 100
	}
	return VisualTypographyProfile{
		FontFamily:     firstNonEmpty(report.Names.TypographicFamily, report.Names.Family, report.Names.FullName),
		FallbackFonts:  report.Names.FallbackNames,
		FontSize:       fontSize,
		Kerning:        0,
		LineHeight:     lineHeight,
		BaselineOffset: 0,
	}
}

func ResolveFontFamily(family string) (string, []Diagnostic) {
	query := normalizeFontQuery(family)
	if query == "" {
		return "", []Diagnostic{{Severity: SeverityError, Code: "FONT.FAMILY", Message: "font family query is empty."}}
	}
	paths := InstalledFontPaths()
	diagnostics := []Diagnostic{}
	type match struct {
		path  string
		score int
	}
	matches := []match{}
	for _, path := range paths {
		faces, faceDiagnostics := InspectFontFile(path)
		_ = faceDiagnostics
		for _, face := range faces {
			score := fontFamilyMatchScore(query, face.Names) + fontStylePreferenceScore(face.Names)
			if score > 0 {
				matches = append(matches, match{path: path, score: score})
			}
		}
	}
	if len(matches) == 0 {
		diagnostics = append(diagnostics, Diagnostic{Severity: SeverityWarning, Code: "FONT.NOT_FOUND", Message: fmt.Sprintf("no installed font matched family %q across %d scanned files.", family, len(paths))})
		return "", diagnostics
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score == matches[j].score {
			return matches[i].path < matches[j].path
		}
		return matches[i].score > matches[j].score
	})
	return matches[0].path, diagnostics
}

func InstalledFontPaths() []string {
	dirs := []string{
		"/System/Library/Fonts",
		"/Library/Fonts",
		filepath.Join(os.Getenv("HOME"), "Library/Fonts"),
		"C:/Windows/Fonts",
		"/usr/share/fonts",
		"/usr/local/share/fonts",
		filepath.Join(os.Getenv("HOME"), ".local/share/fonts"),
	}
	seen := map[string]bool{}
	paths := []string{}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		_ = filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry == nil || entry.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			switch ext {
			case ".ttf", ".otf", ".ttc":
				if !seen[path] {
					seen[path] = true
					paths = append(paths, path)
				}
			}
			return nil
		})
	}
	sort.Strings(paths)
	return paths
}

func fontFamilyMatchScore(query string, names FontNameSet) int {
	candidates := []string{names.TypographicFamily, names.Family, names.FullName, names.PostScriptName}
	for i, candidate := range candidates {
		normalized := normalizeFontQuery(candidate)
		switch {
		case normalized == query:
			return 100 - i
		case strings.Contains(normalized, query):
			return 50 - i
		}
	}
	return 0
}

func fontStylePreferenceScore(names FontNameSet) int {
	style := normalizeFontQuery(firstNonEmpty(names.TypographicStyle, names.Subfamily, names.FullName))
	switch {
	case style == "regular" || style == "book" || style == "roman":
		return 20
	case strings.Contains(style, "regular"):
		return 10
	case strings.Contains(style, "bold") || strings.Contains(style, "italic") || strings.Contains(style, "oblique"):
		return -10
	default:
		return 0
	}
}

func normalizeFontQuery(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(" ", "", "-", "", "_", "")
	return replacer.Replace(value)
}

func compactUniqueStrings(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func firstInt16(values ...int16) int16 {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func ratio(value int16, upm float64) float64 {
	if value == 0 || upm == 0 {
		return 0
	}
	return math.Round((float64(value)/upm)*10000) / 10000
}
