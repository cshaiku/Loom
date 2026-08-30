package loom

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

func AnalyzeXAML(path string) (Analysis, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Analysis{}, fmt.Errorf("could not read XAML source at %s", path)
	}
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var stack []Node
	var roots []Node
	elementCount := 0
	var diagnostics []Diagnostic
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
				continue
			}
			elementCount++
			stack = append(stack, makeXAMLNode(name, attrs(t.Attr), &diagnostics))
		case xml.EndElement:
			if strings.Contains(localXMLName(t.Name.Local), ".") || len(stack) == 0 {
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

func makeXAMLNode(name string, attributes map[string]string, diagnostics *[]Diagnostic) Node {
	props := map[string]string{"xamlElement": name}
	for key, value := range attributes {
		props["xaml."+key] = value
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
	case "Button", "AppBarButton", "HyperlinkButton":
		node.Kind = KindButton
		node.Arguments = quote(firstNonEmpty(attributes["Content"], attributes["Label"]))
	case "TextBox", "PasswordBox":
		node.Kind = KindTextField
		node.Arguments = quote(firstNonEmpty(attributes["PlaceholderText"], attributes["Header"]))
	case "Image":
		node.Kind = KindImage
		node.Arguments = quote(attributes["Source"])
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
	for xaml, swift := range map[string]string{"Width": "width", "Height": "height", "MinWidth": "minWidth", "MinHeight": "minHeight"} {
		if value := attributes[xaml]; value != "" {
			parts = append(parts, swift+": "+value)
		}
	}
	if len(parts) == 0 {
		return nil
	}
	return []Modifier{{Name: "frame", Arguments: strings.Join(parts, ", ")}}
}

func accessibilityModifiers(attributes map[string]string) []Modifier {
	if value := firstNonEmpty(attributes["AutomationId"], attributes["Name"]); value != "" {
		return []Modifier{{Name: "accessibilityIdentifier", Arguments: quote(value)}}
	}
	return nil
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
