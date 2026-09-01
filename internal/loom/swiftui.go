package loom

import (
	"fmt"
	"strings"
	"unicode"
)

type swiftTokenKind int

const (
	swiftTokenIdentifier swiftTokenKind = iota
	swiftTokenString
	swiftTokenSymbol
)

type swiftToken struct {
	kind   swiftTokenKind
	value  string
	offset int
}

type swiftParser struct {
	tokens      []swiftToken
	index       int
	diagnostics []Diagnostic
}

func AnalyzeSwiftUI(path string) (Analysis, error) {
	data, err := readSourceFile(path, "SwiftUI source")
	if err != nil {
		return Analysis{}, fmt.Errorf("could not read SwiftUI source at %s: %w", path, err)
	}
	tokens, tokenDiagnostics := tokenizeSwift(string(data))
	parser := &swiftParser{tokens: tokens}
	parser.diagnostics = append(parser.diagnostics, tokenDiagnostics...)
	parser.diagnostics = append(parser.diagnostics, swiftDelimiterDiagnostics(tokens)...)
	children := parser.parseBlock("")
	return Analysis{
		SourcePath:      path,
		RootView:        "swiftui",
		Component:       componentNameFromSwiftSource(tokens),
		SyntaxNodeCount: countSwiftConstructTokens(tokens),
		Layout:          Node{Kind: KindRoot, Expression: "swiftui", Properties: map[string]string{"sourceDialect": "swiftui"}, Children: children},
		Diagnostics:     nonNilDiagnostics(parser.diagnostics),
	}, nil
}

func tokenizeSwift(source string) ([]swiftToken, []Diagnostic) {
	tokens := []swiftToken{}
	diagnostics := []Diagnostic{}
	runes := []rune(source)
	for i := 0; i < len(runes); {
		r := runes[i]
		if unicode.IsSpace(r) {
			i++
			continue
		}
		if r == '/' && i+1 < len(runes) && runes[i+1] == '/' {
			for i < len(runes) && runes[i] != '\n' {
				i++
			}
			continue
		}
		if r == '/' && i+1 < len(runes) && runes[i+1] == '*' {
			start := i
			i += 2
			for i+1 < len(runes) && !(runes[i] == '*' && runes[i+1] == '/') {
				i++
			}
			if i+1 < len(runes) {
				i += 2
			} else {
				diagnostics = append(diagnostics, Diagnostic{Severity: SeverityError, Code: "SWIFTUI.PARSE", Message: "unterminated Swift block comment.", SourceOffset: &start})
			}
			continue
		}
		if r == '"' {
			start := i
			i++
			var b strings.Builder
			closed := false
			for i < len(runes) {
				if runes[i] == '\\' && i+1 < len(runes) {
					b.WriteRune(runes[i+1])
					i += 2
					continue
				}
				if runes[i] == '"' {
					i++
					closed = true
					break
				}
				b.WriteRune(runes[i])
				i++
			}
			if !closed {
				diagnostics = append(diagnostics, Diagnostic{Severity: SeverityError, Code: "SWIFTUI.PARSE", Message: "unterminated Swift string literal.", SourceOffset: &start})
			}
			tokens = append(tokens, swiftToken{kind: swiftTokenString, value: b.String(), offset: start})
			continue
		}
		if isSwiftIdentifierStart(r) {
			start := i
			var b strings.Builder
			for i < len(runes) && isSwiftIdentifierPart(runes[i]) {
				b.WriteRune(runes[i])
				i++
			}
			tokens = append(tokens, swiftToken{kind: swiftTokenIdentifier, value: b.String(), offset: start})
			continue
		}
		if unicode.IsDigit(r) {
			start := i
			var b strings.Builder
			for i < len(runes) && (unicode.IsDigit(runes[i]) || runes[i] == '.') {
				b.WriteRune(runes[i])
				i++
			}
			tokens = append(tokens, swiftToken{kind: swiftTokenSymbol, value: b.String(), offset: start})
			continue
		}
		tokens = append(tokens, swiftToken{kind: swiftTokenSymbol, value: string(r), offset: i})
		i++
	}
	return tokens, diagnostics
}

func isSwiftIdentifierStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isSwiftIdentifierPart(r rune) bool {
	return isSwiftIdentifierStart(r) || unicode.IsDigit(r)
}

func (p *swiftParser) parseBlock(end string) []Node {
	nodes := []Node{}
	for p.index < len(p.tokens) {
		if end != "" && p.peekValue(end) {
			p.index++
			break
		}
		if !p.peekKind(swiftTokenIdentifier) {
			p.index++
			continue
		}
		node, ok := p.parseNode()
		if ok {
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func (p *swiftParser) parseNode() (Node, bool) {
	token := p.tokens[p.index]
	name := token.value
	if p.index > 0 && (p.tokens[p.index-1].value == "struct" || p.tokens[p.index-1].value == "import") {
		p.index++
		return Node{}, false
	}
	if shouldSkipSwiftIdentifier(name) {
		p.index++
		return Node{}, false
	}
	p.index++
	args := ""
	hadCallOrBlock := false
	if p.peekValue("(") {
		hadCallOrBlock = true
		args = p.consumeBalanced("(", ")")
	}
	children := []Node{}
	if p.peekValue("{") {
		hadCallOrBlock = true
		p.index++
		children = p.parseBlock("}")
	}
	if swiftUIConstructKind(name) == "" && !hadCallOrBlock {
		return Node{}, false
	}
	node, ok := makeSwiftUINode(name, args, children, token.offset, &p.diagnostics)
	if !ok {
		return Node{}, false
	}
	node.Modifiers = append(node.Modifiers, p.parseModifiers()...)
	return node, true
}

func (p *swiftParser) parseModifiers() []Modifier {
	modifiers := []Modifier{}
	for p.peekValue(".") {
		p.index++
		if !p.peekKind(swiftTokenIdentifier) {
			continue
		}
		name := p.tokens[p.index].value
		p.index++
		args := ""
		if p.peekValue("(") {
			args = p.consumeBalanced("(", ")")
		}
		modifiers = append(modifiers, Modifier{Name: name, Arguments: strings.TrimSpace(args)})
	}
	return modifiers
}

func (p *swiftParser) consumeBalanced(open, close string) string {
	if !p.peekValue(open) {
		return ""
	}
	depth := 0
	parts := []string{}
	for p.index < len(p.tokens) {
		token := p.tokens[p.index]
		p.index++
		if token.value == open {
			depth++
			if depth == 1 {
				continue
			}
		}
		if token.value == close {
			depth--
			if depth == 0 {
				break
			}
		}
		if token.kind == swiftTokenString {
			parts = append(parts, quote(token.value))
		} else {
			parts = append(parts, token.value)
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func (p *swiftParser) peekValue(value string) bool {
	return p.index < len(p.tokens) && p.tokens[p.index].value == value
}

func (p *swiftParser) peekKind(kind swiftTokenKind) bool {
	return p.index < len(p.tokens) && p.tokens[p.index].kind == kind
}

func makeSwiftUINode(name, args string, children []Node, offset int, diagnostics *[]Diagnostic) (Node, bool) {
	props := map[string]string{"swiftuiConstruct": name}
	node := Node{Expression: name, Properties: props, Children: children}
	firstString := firstSwiftStringArgument(args)
	switch name {
	case "VStack", "LazyVStack":
		node.Kind = KindVerticalStack
	case "HStack", "LazyHStack":
		node.Kind = KindHorizontalStack
	case "ZStack":
		node.Kind = KindOverlayStack
	case "Grid", "LazyVGrid", "LazyHGrid":
		node.Kind = KindGrid
	case "ScrollView":
		node.Kind = KindScrollView
	case "List", "Table":
		node.Kind = KindList
	case "Text", "Label":
		node.Kind = KindText
		node.Arguments = quote(firstString)
		node.VisibleLabel = firstString
	case "TextField", "SecureField", "TextEditor":
		node.Kind = KindTextField
		node.Arguments = quote(firstString)
		node.Placeholder = firstString
	case "Button":
		node.Kind = KindButton
		node.Arguments = quote(firstString)
		node.VisibleLabel = firstString
	case "Image", "AsyncImage":
		node.Kind = KindImage
		node.Resource = firstString
		node.Arguments = quote("")
	case "Slider", "ProgressView", "Gauge":
		node.Kind = KindSlider
	case "Toggle", "Picker":
		node.Kind = KindToggle
		node.Arguments = quote(firstString)
		node.VisibleLabel = firstString
	case "Spacer":
		node.Kind = KindSpacer
	case "Divider":
		node.Kind = KindDivider
	case "Color", "Material":
		node.Kind = KindColor
	case "GeometryReader":
		node.Kind = KindGeometryReader
	case "NavigationSplitView", "HSplitView", "VSplitView":
		node.Kind = KindSplitView
	case "ForEach":
		node.Kind = KindLoop
	case "if":
		node.Kind = KindConditional
	default:
		if isLikelySwiftUIView(name) {
			props["componentBoundary"] = "swiftui-view"
			props["requiresNativeImplementation"] = "true"
			*diagnostics = append(*diagnostics, Diagnostic{Severity: SeverityInfo, Code: "SWIFTUI.COMPONENT_BOUNDARY", Message: "custom SwiftUI view preserved as a component boundary: " + name + ".", SourceOffset: &offset})
			node.Kind = KindComponent
		} else {
			return Node{}, false
		}
	}
	if args != "" {
		props["swiftui.arguments"] = args
	}
	return node, true
}

func shouldSkipSwiftIdentifier(name string) bool {
	switch name {
	case "import", "struct", "class", "enum", "extension", "var", "let", "func", "some", "View", "body", "return", "private", "public", "internal", "fileprivate", "State", "Binding", "ObservedObject", "StateObject", "Environment", "EnvironmentObject":
		return true
	default:
		return false
	}
}

func isLikelySwiftUIView(name string) bool {
	if name == "" {
		return false
	}
	r := []rune(name)[0]
	return unicode.IsUpper(r)
}

func firstSwiftStringArgument(args string) string {
	tokens, _ := tokenizeSwift(args)
	for _, token := range tokens {
		if token.kind == swiftTokenString {
			return token.value
		}
	}
	return ""
}

func componentNameFromSwiftSource(tokens []swiftToken) string {
	for i := 0; i+2 < len(tokens); i++ {
		if tokens[i].value == "struct" && tokens[i+1].kind == swiftTokenIdentifier && tokens[i+2].value == ":" {
			return tokens[i+1].value
		}
	}
	return "root"
}

func countSwiftConstructTokens(tokens []swiftToken) int {
	count := 0
	for i, token := range tokens {
		if token.kind != swiftTokenIdentifier || shouldSkipSwiftIdentifier(token.value) || token.value == "View" {
			continue
		}
		hasCallOrBlock := i+1 < len(tokens) && (tokens[i+1].value == "(" || tokens[i+1].value == "{")
		if swiftUIConstructKind(token.value) != "" || (isLikelySwiftUIView(token.value) && hasCallOrBlock) {
			count++
		}
	}
	return count
}

func swiftUIConstructKind(name string) NodeKind {
	switch name {
	case "VStack", "LazyVStack":
		return KindVerticalStack
	case "HStack", "LazyHStack":
		return KindHorizontalStack
	case "ZStack":
		return KindOverlayStack
	case "Grid", "LazyVGrid", "LazyHGrid":
		return KindGrid
	case "ScrollView":
		return KindScrollView
	case "List", "Table":
		return KindList
	case "Text", "Label":
		return KindText
	case "TextField", "SecureField", "TextEditor":
		return KindTextField
	case "Button":
		return KindButton
	case "Image", "AsyncImage":
		return KindImage
	case "Slider", "ProgressView", "Gauge":
		return KindSlider
	case "Toggle", "Picker":
		return KindToggle
	case "Spacer":
		return KindSpacer
	case "Divider":
		return KindDivider
	case "Color", "Material":
		return KindColor
	case "GeometryReader":
		return KindGeometryReader
	case "NavigationSplitView", "HSplitView", "VSplitView":
		return KindSplitView
	case "ForEach":
		return KindLoop
	case "if":
		return KindConditional
	default:
		return ""
	}
}

func swiftDelimiterDiagnostics(tokens []swiftToken) []Diagnostic {
	stack := []swiftToken{}
	pairs := map[string]string{"}": "{", ")": "(", "]": "["}
	for _, token := range tokens {
		switch token.value {
		case "{", "(", "[":
			stack = append(stack, token)
		case "}", ")", "]":
			if len(stack) == 0 || stack[len(stack)-1].value != pairs[token.value] {
				offset := token.offset
				return []Diagnostic{{Severity: SeverityError, Code: "SWIFTUI.PARSE", Message: "unbalanced SwiftUI source delimiters.", SourceOffset: &offset}}
			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) > 0 {
		offset := stack[len(stack)-1].offset
		return []Diagnostic{{Severity: SeverityError, Code: "SWIFTUI.PARSE", Message: "unbalanced SwiftUI source delimiters.", SourceOffset: &offset}}
	}
	return []Diagnostic{}
}
