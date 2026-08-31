package loom

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func AnalyzeQt(path string) (Analysis, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".ui" {
		return AnalyzeQtUI(path)
	}
	return AnalyzeQtText(path)
}

func AnalyzeQtUI(path string) (Analysis, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Analysis{}, fmt.Errorf("could not read Qt UI source at %s: %w", path, err)
	}
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	stack := []Node{}
	roots := []Node{}
	elementCount := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return Analysis{}, fmt.Errorf("could not extract the Qt UI body: %w", err)
		}
		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local != "widget" && t.Name.Local != "layout" && t.Name.Local != "spacer" {
				continue
			}
			attrs := attrs(t.Attr)
			className := firstNonEmpty(attrs["class"], t.Name.Local)
			name := firstNonEmpty(attrs["name"], className)
			elementCount++
			stack = append(stack, makeQtNode(className, name, "", nil))
		case xml.EndElement:
			if t.Name.Local != "widget" && t.Name.Local != "layout" && t.Name.Local != "spacer" {
				continue
			}
			if len(stack) == 0 {
				continue
			}
			node := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if len(stack) == 0 {
				roots = append(roots, node)
			} else {
				stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, node)
			}
		}
	}
	return Analysis{SourcePath: path, RootView: "qt", Component: componentName(path), SyntaxNodeCount: elementCount, Layout: Node{Kind: KindRoot, Expression: "qt", Properties: map[string]string{"sourceDialect": "qt"}, Children: roots}, Diagnostics: []Diagnostic{}}, nil
}

func AnalyzeQtText(path string) (Analysis, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Analysis{}, fmt.Errorf("could not read Qt source at %s: %w", path, err)
	}
	source := string(data)
	tokens := tokenizeSwift(source)
	diagnostics := qtDelimiterDiagnostics(tokens)
	children := parseQtNodes(tokens)
	return Analysis{SourcePath: path, RootView: "qt", Component: componentName(path), SyntaxNodeCount: countQtConstructTokens(tokens), Layout: Node{Kind: KindRoot, Expression: "qt", Properties: map[string]string{"sourceDialect": "qt"}, Children: children}, Diagnostics: nonNilDiagnostics(diagnostics)}, nil
}

func parseQtNodes(tokens []swiftToken) []Node {
	stack := []Node{}
	roots := []Node{}
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if token.kind != swiftTokenIdentifier {
			if token.value == "}" && len(stack) > 0 {
				node := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				if len(stack) == 0 {
					roots = append(roots, node)
				} else {
					stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, node)
				}
			}
			continue
		}
		kind := qtConstructKind(token.value)
		if kind == "" {
			continue
		}
		args := ""
		if i+1 < len(tokens) && tokens[i+1].value == "(" {
			parser := &swiftParser{tokens: tokens, index: i + 1}
			args = parser.consumeBalanced("(", ")")
			i = parser.index - 1
		}
		node := makeQtNode(token.value, token.value, args, nil)
		if i+1 < len(tokens) && tokens[i+1].value == "{" {
			stack = append(stack, node)
			i++
		} else if len(stack) > 0 {
			stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, node)
		} else {
			roots = append(roots, node)
		}
	}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if len(stack) == 0 {
			roots = append(roots, node)
		} else {
			stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, node)
		}
	}
	return roots
}

func makeQtNode(className, expression, args string, children []Node) Node {
	props := map[string]string{"qtConstruct": className}
	node := Node{Kind: qtConstructKind(className), Expression: expression, Properties: props, Children: children}
	if node.Kind == "" {
		node.Kind = KindComponent
		props["componentBoundary"] = "qt-component"
		props["requiresNativeImplementation"] = "true"
	}
	if args != "" {
		props["qt.arguments"] = args
		if text := firstSwiftStringArgument(args); text != "" {
			node.Arguments = quote(text)
		}
	}
	return node
}

func qtConstructKind(name string) NodeKind {
	switch name {
	case "Column", "ColumnLayout", "QVBoxLayout", "QBoxLayout":
		return KindVerticalStack
	case "Row", "RowLayout", "QHBoxLayout":
		return KindHorizontalStack
	case "StackLayout", "StackView":
		return KindOverlayStack
	case "Grid", "GridLayout", "QGridLayout":
		return KindGrid
	case "ScrollView", "Flickable", "QScrollArea":
		return KindScrollView
	case "ListView", "TableView", "TreeView", "QListView", "QTableView", "QTreeView":
		return KindList
	case "Text", "Label", "QLabel":
		return KindText
	case "TextField", "TextArea", "TextInput", "QLineEdit", "QTextEdit":
		return KindTextField
	case "Button", "ToolButton", "QPushButton", "QToolButton":
		return KindButton
	case "Image", "QPixmap", "QIcon":
		return KindImage
	case "Slider", "ProgressBar", "QSlider", "QProgressBar":
		return KindSlider
	case "Switch", "CheckBox", "RadioButton", "QCheckBox", "QRadioButton":
		return KindToggle
	case "Spacer", "QSpacerItem":
		return KindSpacer
	case "Rectangle":
		return KindColor
	case "SplitView", "QSplitter":
		return KindSplitView
	case "Repeater":
		return KindLoop
	default:
		return ""
	}
}

func countQtConstructTokens(tokens []swiftToken) int {
	count := 0
	for _, token := range tokens {
		if token.kind == swiftTokenIdentifier && qtConstructKind(token.value) != "" {
			count++
		}
	}
	return count
}

func qtDelimiterDiagnostics(tokens []swiftToken) []Diagnostic {
	for _, diagnostic := range swiftDelimiterDiagnostics(tokens) {
		diagnostic.Code = "QT.PARSE"
		diagnostic.Message = "unbalanced Qt layout source delimiters."
		return []Diagnostic{diagnostic}
	}
	return []Diagnostic{}
}

func componentName(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if base == "" {
		return "root"
	}
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			return r
		}
		return -1
	}, base)
}
