package loom

type NodeKind string

const (
	KindRoot            NodeKind = "root"
	KindGeometryReader  NodeKind = "geometryReader"
	KindVerticalStack   NodeKind = "verticalStack"
	KindHorizontalStack NodeKind = "horizontalStack"
	KindOverlayStack    NodeKind = "overlayStack"
	KindSplitView       NodeKind = "splitView"
	KindGrid            NodeKind = "grid"
	KindScrollView      NodeKind = "scrollView"
	KindList            NodeKind = "list"
	KindText            NodeKind = "text"
	KindTextField       NodeKind = "textField"
	KindButton          NodeKind = "button"
	KindImage           NodeKind = "image"
	KindSlider          NodeKind = "slider"
	KindToggle          NodeKind = "toggle"
	KindSpacer          NodeKind = "spacer"
	KindDivider         NodeKind = "divider"
	KindConditional     NodeKind = "conditional"
	KindLoop            NodeKind = "loop"
	KindColor           NodeKind = "color"
	KindComponent       NodeKind = "component"
	KindUnsupported     NodeKind = "unsupported"
)

type Modifier struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Node struct {
	Kind       NodeKind          `json:"kind"`
	Expression string            `json:"expression"`
	Arguments  string            `json:"arguments,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
	Modifiers  []Modifier        `json:"modifiers,omitempty"`
	Children   []Node            `json:"children,omitempty"`
}

type DiagnosticSeverity string

const (
	SeverityInfo    DiagnosticSeverity = "info"
	SeverityWarning DiagnosticSeverity = "warning"
	SeverityError   DiagnosticSeverity = "error"
)

type Diagnostic struct {
	Severity     DiagnosticSeverity `json:"severity"`
	Code         string             `json:"code"`
	Message      string             `json:"message"`
	SourceOffset *int               `json:"sourceOffset,omitempty"`
}

type Analysis struct {
	SourcePath      string       `json:"sourcePath"`
	RootView        string       `json:"rootView"`
	Component       string       `json:"component"`
	SyntaxNodeCount int          `json:"syntaxNodeCount"`
	Layout          Node         `json:"layout"`
	Diagnostics     []Diagnostic `json:"diagnostics"`
}

func (n Node) RecursiveNodeCount() int {
	count := 0
	stack := []Node{n}
	for len(stack) > 0 {
		last := len(stack) - 1
		node := stack[last]
		stack = stack[:last]
		count++
		stack = append(stack, node.Children...)
	}
	return count
}
