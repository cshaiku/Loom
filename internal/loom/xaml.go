package loom

import (
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func AnalyzeXAML(path string) (Analysis, error) {
	data, err := readSourceFile(path, "XAML source")
	if err != nil {
		return Analysis{}, fmt.Errorf("could not read XAML source at %s", path)
	}
	resources, resourceDiagnostics := collectXAMLResources(path, data, map[string]bool{})
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var stack []Node
	var propertyStack []string
	roots := []Node{}
	elementCount := 0
	diagnostics := append([]Diagnostic{}, resourceDiagnostics...)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Analysis{}, fmt.Errorf("could not extract the XAML view body: %w", err)
		}
		switch t := token.(type) {
		case xml.StartElement:
			name := localXMLName(t.Name.Local)
			if strings.Contains(name, ".") {
				propertyStack = append(propertyStack, name)
				continue
			}
			if len(propertyStack) > 0 {
				applyXAMLPropertyElement(&stack, propertyStack[len(propertyStack)-1], name, attrs(t.Attr))
				continue
			}
			elementCount++
			stack = append(stack, makeXAMLNode(name, attrs(t.Attr), resources, &diagnostics))
		case xml.EndElement:
			name := localXMLName(t.Name.Local)
			if strings.Contains(name, ".") {
				if len(propertyStack) > 0 {
					propertyStack = propertyStack[:len(propertyStack)-1]
				}
				continue
			}
			if len(propertyStack) > 0 || len(stack) == 0 {
				continue
			}
			node := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				roots = append(roots, node)
			} else {
				parent := &stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}
		case xml.CharData:
			if len(stack) > 0 {
				text := strings.TrimSpace(string(t))
				if text != "" && stack[len(stack)-1].Arguments == "" {
					stack[len(stack)-1].Arguments = quote(text)
				}
			}
		}
	}
	return Analysis{
		SourcePath:      path,
		RootView:        "XAML",
		Component:       "root",
		SyntaxNodeCount: elementCount,
		Layout:          Node{Kind: KindRoot, Expression: "xaml", Properties: map[string]string{"sourceDialect": "winui3"}, Children: roots},
		Diagnostics:     diagnostics,
	}, nil
}

func applyXAMLPropertyElement(stack *[]Node, propertyName, elementName string, attributes map[string]string) {
	if len(*stack) == 0 {
		return
	}
	parent := &(*stack)[len(*stack)-1]
	switch propertyName {
	case "Grid.RowDefinitions":
		if elementName == "RowDefinition" {
			appendIndexedProperty(parent.Properties, "xaml.Grid.RowDefinitions", firstNonEmpty(attributes["Height"], "*"))
		}
	case "Grid.ColumnDefinitions":
		if elementName == "ColumnDefinition" {
			appendIndexedProperty(parent.Properties, "xaml.Grid.ColumnDefinitions", firstNonEmpty(attributes["Width"], "*"))
		}
	}
}

func appendIndexedProperty(props map[string]string, key, value string) {
	if props[key] == "" {
		props[key] = value
		return
	}
	props[key] += "," + value
}

type xamlResourceIndex struct {
	Values         map[string]string
	ImplicitStyles map[string]xamlStyle
	NamedStyles    map[string]xamlStyle
}

type xamlCollectFrame struct {
	Name        string
	Attrs       map[string]string
	Text        string
	ObjectValue string
	Setters     []xamlSetter
}

type xamlSetter struct {
	Property string
	Value    string
}

type xamlStyle struct {
	TargetType string
	BasedOn    string
	Setters    map[string]string
}

func collectXAMLResources(path string, data []byte, visited map[string]bool) (xamlResourceIndex, []Diagnostic) {
	index := emptyXAMLResourceIndex()
	diagnostics := []Diagnostic{}
	if abs, err := filepath.Abs(path); err == nil {
		if visited[abs] {
			return index, diagnostics
		}
		visited[abs] = true
	}
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	stack := []xamlCollectFrame{}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return index, diagnostics
		}
		switch t := token.(type) {
		case xml.StartElement:
			stack = append(stack, xamlCollectFrame{Name: localXMLName(t.Name.Local), Attrs: attrs(t.Attr)})
		case xml.CharData:
			if len(stack) > 0 {
				stack[len(stack)-1].Text += string(t)
			}
		case xml.EndElement:
			if len(stack) == 0 {
				continue
			}
			frame := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			switch frame.Name {
			case "Setter":
				if len(stack) > 0 {
					setter := xamlSetter{Property: frame.Attrs["Property"], Value: xamlFrameValue(frame)}
					if setter.Property != "" && setter.Value != "" {
						stack[len(stack)-1].Setters = append(stack[len(stack)-1].Setters, setter)
					}
				}
			case "Style":
				target := normalizeXAMLTargetType(frame.Attrs["TargetType"])
				style := xamlStyle{TargetType: target, BasedOn: frame.Attrs["BasedOn"], Setters: map[string]string{}}
				for _, setter := range frame.Setters {
					style.Setters[setter.Property] = setter.Value
				}
				if key := frame.Attrs["Key"]; key != "" {
					index.NamedStyles[key] = style
				} else if target != "" {
					if existing := index.ImplicitStyles[target]; len(existing.Setters) > 0 {
						for property, value := range existing.Setters {
							style.Setters[property] = value
						}
					}
					index.ImplicitStyles[target] = style
				}
			case "ResourceDictionary":
				if source := frame.Attrs["Source"]; source != "" {
					external, externalDiagnostics := collectXAMLResourceDictionary(path, source, visited)
					diagnostics = append(diagnostics, externalDiagnostics...)
					mergeXAMLResourceIndex(&index, external)
				}
			default:
				if key := frame.Attrs["Key"]; key != "" {
					value := xamlFrameValue(frame)
					if value != "" {
						index.Values[key] = value
					}
				} else if len(stack) > 0 {
					value := xamlFrameValue(frame)
					if value != "" && stack[len(stack)-1].ObjectValue == "" {
						stack[len(stack)-1].ObjectValue = value
					}
				}
			}
		}
	}
	return index, diagnostics
}

func xamlFrameValue(frame xamlCollectFrame) string {
	if value := strings.TrimSpace(frame.Attrs["Value"]); value != "" {
		return value
	}
	if value := xamlObjectAttributeValue(frame.Name, frame.Attrs); value != "" {
		return value
	}
	if value := strings.TrimSpace(frame.Text); value != "" {
		return value
	}
	return strings.TrimSpace(frame.ObjectValue)
}

func xamlObjectAttributeValue(name string, attrs map[string]string) string {
	switch name {
	case "Thickness":
		if value := strings.TrimSpace(attrs["Value"]); value != "" {
			return value
		}
		left := firstNonEmpty(attrs["Left"], attrs["left"])
		top := firstNonEmpty(attrs["Top"], attrs["top"])
		right := firstNonEmpty(attrs["Right"], attrs["right"])
		bottom := firstNonEmpty(attrs["Bottom"], attrs["bottom"])
		if left != "" || top != "" || right != "" || bottom != "" {
			return strings.Join([]string{firstNonEmpty(left, "0"), firstNonEmpty(top, "0"), firstNonEmpty(right, "0"), firstNonEmpty(bottom, "0")}, ",")
		}
	case "FontFamily":
		return strings.TrimSpace(attrs["Source"])
	case "SolidColorBrush", "AcrylicBrush":
		return strings.TrimSpace(attrs["Color"])
	}
	return ""
}

func collectXAMLResourceDictionary(sourcePath, reference string, visited map[string]bool) (xamlResourceIndex, []Diagnostic) {
	index := emptyXAMLResourceIndex()
	resolvedPath, ok := resolveXAMLDictionaryPath(sourcePath, reference)
	if !ok {
		return index, []Diagnostic{{Severity: SeverityWarning, Code: "XAML.RESOURCE_DICTIONARY_UNRESOLVED", Message: "Could not resolve local XAML resource dictionary: " + reference + "."}}
	}
	data, err := readSourceFile(resolvedPath, "XAML resource dictionary")
	if err != nil {
		return index, []Diagnostic{{Severity: SeverityWarning, Code: "XAML.RESOURCE_DICTIONARY_UNREADABLE", Message: "Could not read XAML resource dictionary: " + reference + "."}}
	}
	return collectXAMLResources(resolvedPath, data, visited)
}

func emptyXAMLResourceIndex() xamlResourceIndex {
	return xamlResourceIndex{Values: map[string]string{}, ImplicitStyles: map[string]xamlStyle{}, NamedStyles: map[string]xamlStyle{}}
}

func mergeXAMLResourceIndex(dst *xamlResourceIndex, src xamlResourceIndex) {
	for key, value := range src.Values {
		dst.Values[key] = value
	}
	for target, style := range src.ImplicitStyles {
		if existing := dst.ImplicitStyles[target]; len(existing.Setters) > 0 {
			for property, value := range existing.Setters {
				style.Setters[property] = value
			}
		}
		dst.ImplicitStyles[target] = style
	}
	for key, style := range src.NamedStyles {
		dst.NamedStyles[key] = style
	}
}

func resolveXAMLDictionaryPath(sourcePath, reference string) (string, bool) {
	reference = strings.TrimSpace(reference)
	if reference == "" || strings.HasPrefix(reference, "http://") || strings.HasPrefix(reference, "https://") {
		return "", false
	}
	if strings.HasPrefix(reference, "ms-appx:///") {
		return findExistingXAMLPathUpward(filepath.Dir(sourcePath), strings.TrimPrefix(reference, "ms-appx:///"))
	}
	if strings.HasPrefix(reference, "ms-appx://") {
		return "", false
	}
	if filepath.IsAbs(reference) {
		if fileExists(reference) {
			return reference, true
		}
		return "", false
	}
	candidate := filepath.Join(filepath.Dir(sourcePath), filepath.FromSlash(reference))
	if fileExists(candidate) {
		return candidate, true
	}
	return findExistingXAMLPathUpward(filepath.Dir(sourcePath), filepath.FromSlash(reference))
}

func findExistingXAMLPathUpward(startDir, relativePath string) (string, bool) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		dir = startDir
	}
	for {
		candidate := filepath.Join(dir, relativePath)
		if fileExists(candidate) {
			return candidate, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

func normalizeXAMLTargetType(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "{}")
	value = strings.TrimPrefix(value, "x:Type ")
	if parts := strings.Split(value, ":"); len(parts) > 1 {
		value = parts[len(parts)-1]
	}
	return strings.TrimSpace(value)
}

func makeXAMLNode(name string, attributes map[string]string, resources xamlResourceIndex, diagnostics *[]Diagnostic) Node {
	attributes, styleOrigins := applyXAMLStyleSetters(name, attributes, resources)
	if style := attributes["Style"]; style != "" {
		if _, ok := resolvedXAMLExplicitStyleValue(style, resources); !ok && isXAMLResourceReference(style) {
			*diagnostics = append(*diagnostics, Diagnostic{Severity: SeverityWarning, Code: "XAML.STYLE_REFERENCE_UNRESOLVED", Message: "Could not resolve XAML style reference on " + name + ": " + style + "."})
		}
	}
	props := map[string]string{"xamlElement": name}
	for key, value := range attributes {
		props["xaml."+key] = value
		if resolved := resolveXAMLResourceReference(value, resources.Values); resolved != value {
			props["xaml.resolved."+key] = resolved
		}
		if origin := styleOrigins[key]; origin != "" {
			props["xaml.origin."+key] = origin
		}
	}
	modifiers := frameModifiers(attributes)
	modifiers = append(modifiers, accessibilityModifiers(attributes)...)
	node := Node{Expression: name, Properties: props, Modifiers: modifiers}
	switch name {
	case "Grid":
		node.Kind = KindGrid
	case "StackPanel":
		if attributes["Orientation"] == "Horizontal" {
			node.Kind = KindHorizontalStack
		} else {
			node.Kind = KindVerticalStack
		}
	case "TextBlock":
		node.Kind = KindText
		node.Arguments = quote(firstNonEmpty(attributes["Text"], ""))
		node.VisibleLabel = firstNonEmpty(attributes["Text"], "")
	case "Button", "AppBarButton", "HyperlinkButton":
		node.Kind = KindButton
		node.VisibleLabel = firstNonEmpty(attributes["Content"], attributes["Label"])
		node.AccessibleName = attributes["AutomationProperties.Name"]
		node.Identifier = firstNonEmpty(attributes["AutomationId"], attributes["Name"])
		node.Arguments = quote(node.VisibleLabel)
	case "TextBox", "PasswordBox":
		node.Kind = KindTextField
		node.VisibleLabel = attributes["Header"]
		node.AccessibleName = attributes["AutomationProperties.Name"]
		node.Identifier = firstNonEmpty(attributes["AutomationId"], attributes["Name"])
		node.Placeholder = attributes["PlaceholderText"]
		node.Arguments = quote(firstNonEmpty(node.VisibleLabel, node.AccessibleName))
	case "Image":
		node.Kind = KindImage
		node.AccessibleName = attributes["AutomationProperties.Name"]
		node.Identifier = firstNonEmpty(attributes["AutomationId"], attributes["Name"])
		node.Resource = attributes["Source"]
		node.Decorative = strings.EqualFold(attributes["AutomationProperties.AccessibilityView"], "Raw") || strings.EqualFold(attributes["IsHitTestVisible"], "False")
		node.Arguments = quote(firstNonEmpty(node.AccessibleName, ""))
	case "ScrollViewer":
		node.Kind = KindScrollView
	case "ListView", "GridView", "ItemsRepeater":
		node.Kind = KindList
	case "Slider", "ProgressBar":
		node.Kind = KindSlider
	case "ToggleSwitch", "CheckBox":
		node.Kind = KindToggle
	case "Rectangle":
		if attributes["Height"] == "1" || attributes["Width"] == "1" {
			node.Kind = KindDivider
		} else {
			return nativeBoundaryNode(name, props, modifiers, diagnostics)
		}
	case "Border":
		if attributes["Background"] != "" {
			node.Kind = KindColor
		} else {
			node.Kind = KindRoot
		}
	default:
		return nativeBoundaryNode(name, props, modifiers, diagnostics)
	}
	return node
}

func applyXAMLStyleSetters(elementName string, attributes map[string]string, resources xamlResourceIndex) (map[string]string, map[string]string) {
	out := map[string]string{}
	for key, value := range attributes {
		out[key] = value
	}
	origins := map[string]string{}
	if style := resolvedXAMLImplicitStyle(elementName, resources); len(style) > 0 && out["Style"] == "" {
		applyResolvedXAMLSetters(out, origins, style, "style-setter")
	}
	if style := resolvedXAMLExplicitStyle(out["Style"], resources); len(style) > 0 {
		applyResolvedXAMLSetters(out, origins, style, "explicit-style-setter")
	}
	return out, origins
}

func applyResolvedXAMLSetters(attributes map[string]string, origins map[string]string, setters map[string]string, origin string) {
	for key, value := range setters {
		if attributes[key] == "" {
			attributes[key] = value
			origins[key] = origin
		}
	}
}

func resolvedXAMLImplicitStyle(elementName string, resources xamlResourceIndex) map[string]string {
	style, ok := resources.ImplicitStyles[elementName]
	if !ok {
		return nil
	}
	return resolveXAMLStyleSetters(style, resources, map[string]bool{})
}

func resolvedXAMLExplicitStyle(styleReference string, resources xamlResourceIndex) map[string]string {
	style, ok := resolvedXAMLExplicitStyleValue(styleReference, resources)
	if !ok {
		return nil
	}
	return resolveXAMLStyleSetters(style, resources, map[string]bool{})
}

func resolvedXAMLExplicitStyleValue(styleReference string, resources xamlResourceIndex) (xamlStyle, bool) {
	key := xamlStyleReferenceKey(styleReference)
	if key == "" {
		return xamlStyle{}, false
	}
	style, ok := resources.NamedStyles[key]
	if !ok {
		return xamlStyle{}, false
	}
	return style, true
}

func resolveXAMLStyleSetters(style xamlStyle, resources xamlResourceIndex, visited map[string]bool) map[string]string {
	out := map[string]string{}
	if baseKey := xamlStyleReferenceKey(style.BasedOn); baseKey != "" && !visited[baseKey] {
		visited[baseKey] = true
		if baseStyle, exists := resources.NamedStyles[baseKey]; exists {
			for key, value := range resolveXAMLStyleSetters(baseStyle, resources, visited) {
				out[key] = value
			}
		}
	}
	for key, value := range style.Setters {
		out[key] = resolveXAMLResourceReference(value, resources.Values)
	}
	return out
}

func xamlStyleReferenceKey(value string) string {
	if key, ok := xamlResourceReferenceKey(value); ok {
		return key
	}
	return strings.TrimSpace(value)
}

func resolveXAMLResourceReference(value string, resources map[string]string) string {
	return resolveXAMLResourceReferenceWithVisited(value, resources, map[string]bool{})
}

func resolveXAMLResourceReferenceWithVisited(value string, resources map[string]string, visited map[string]bool) string {
	key, ok := xamlResourceReferenceKey(value)
	if !ok {
		return value
	}
	if visited[key] {
		return value
	}
	visited[key] = true
	if resolved := resources[key]; resolved != "" {
		return resolveXAMLResourceReferenceWithVisited(resolved, resources, visited)
	}
	return value
}

func isXAMLResourceReference(value string) bool {
	_, ok := xamlResourceReferenceKey(value)
	return ok
}

func xamlResourceReferenceKey(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	for _, prefix := range []string{"{StaticResource", "{ThemeResource"} {
		if strings.HasPrefix(trimmed, prefix) && strings.HasSuffix(trimmed, "}") {
			key := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, prefix), "}"))
			return key, key != ""
		}
	}
	return "", false
}

func nativeBoundaryNode(name string, props map[string]string, modifiers []Modifier, diagnostics *[]Diagnostic) Node {
	props["componentBoundary"] = "native-winui-control"
	props["unsupportedXamlElement"] = name
	props["requiresNativeImplementation"] = "true"
	*diagnostics = append(*diagnostics, Diagnostic{Severity: SeverityWarning, Code: "XAML.UNSUPPORTED_COMPONENT_BOUNDARY", Message: "Unsupported native WinUI control preserved as a component boundary: " + name + "."})
	return Node{Kind: KindComponent, Expression: name, Properties: props, Modifiers: modifiers}
}

func attrs(in []xml.Attr) map[string]string {
	out := map[string]string{}
	for _, attr := range in {
		out[localXMLName(attr.Name.Local)] = attr.Value
	}
	return out
}

func localXMLName(name string) string {
	if parts := strings.Split(name, ":"); len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return name
}

func frameModifiers(attributes map[string]string) []Modifier {
	var parts []string
	for _, field := range []struct {
		xaml string
		ir   string
	}{
		{"Width", "width"},
		{"Height", "height"},
		{"MinWidth", "minWidth"},
		{"MinHeight", "minHeight"},
	} {
		if value := attributes[field.xaml]; value != "" {
			parts = append(parts, field.ir+": "+value)
		}
	}
	if len(parts) == 0 {
		return []Modifier{}
	}
	return []Modifier{{Name: "frame", Arguments: strings.Join(parts, ", ")}}
}

func accessibilityModifiers(attributes map[string]string) []Modifier {
	if value := firstNonEmpty(attributes["AutomationId"], attributes["Name"]); value != "" {
		return []Modifier{{Name: "accessibilityIdentifier", Arguments: quote(value)}}
	}
	return []Modifier{}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func quote(value string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), `"`, `\"`) + `"`
}
