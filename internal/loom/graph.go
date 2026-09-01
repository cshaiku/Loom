package loom

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ComponentGraphReport struct {
	SchemaVersion string               `json:"schema_version"`
	Status        string               `json:"status"`
	Source        string               `json:"source"`
	Root          string               `json:"root,omitempty"`
	Components    []ComponentGraphNode `json:"components"`
	Edges         []ComponentGraphEdge `json:"edges"`
	Diagnostics   []Diagnostic         `json:"diagnostics"`
}

type ComponentGraphNode struct {
	Name        string `json:"name"`
	Path        string `json:"path,omitempty"`
	Dialect     string `json:"dialect"`
	Kind        string `json:"kind"`
	LayoutNodes int    `json:"layoutNodes"`
	Unresolved  bool   `json:"unresolved,omitempty"`
}

type ComponentGraphEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Reason string `json:"reason"`
}

func GraphComponents(source, rootView, component, includeGlob, excludeGlob string) (ComponentGraphReport, error) {
	source = filepath.Clean(source)
	paths, err := graphSourcePaths(source, includeGlob, excludeGlob)
	if err != nil {
		return ComponentGraphReport{}, err
	}
	nodesByName := map[string]ComponentGraphNode{}
	edgesByKey := map[string]ComponentGraphEdge{}
	diagnostics := []Diagnostic{}
	for _, path := range paths {
		analysis, err := AnalyzeByPlatform(path, "")
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{Severity: SeverityWarning, Code: "GRAPH.ANALYZE", Message: fmt.Sprintf("%s could not be analyzed: %v.", path, err)})
			continue
		}
		name := firstNonEmpty(analysis.Component, componentName(path))
		if component != "" && name != component {
			continue
		}
		nodesByName[name] = ComponentGraphNode{
			Name:        name,
			Path:        analysis.SourcePath,
			Dialect:     firstNonEmpty(analysis.Layout.Properties["sourceDialect"], canonicalPatternPlatform(InferSourcePlatform(path))),
			Kind:        "component",
			LayoutNodes: analysis.Layout.RecursiveNodeCount(),
		}
		diagnostics = append(diagnostics, nonNilDiagnostics(analysis.Diagnostics)...)
		for _, target := range graphComponentReferences(analysis.Layout) {
			if target == "" || target == name {
				continue
			}
			key := name + "\x00" + target
			edgesByKey[key] = ComponentGraphEdge{From: name, To: target, Reason: "source layout references a component boundary"}
		}
	}
	for _, edge := range edgesByKey {
		if _, ok := nodesByName[edge.To]; !ok {
			nodesByName[edge.To] = ComponentGraphNode{Name: edge.To, Dialect: "unknown", Kind: "component", Unresolved: true}
			diagnostics = append(diagnostics, Diagnostic{Severity: SeverityWarning, Code: "GRAPH.UNRESOLVED", Message: fmt.Sprintf("Component %s is referenced by %s but was not found in scanned sources.", edge.To, edge.From)})
		}
	}
	components := make([]ComponentGraphNode, 0, len(nodesByName))
	for _, node := range nodesByName {
		components = append(components, node)
	}
	sort.Slice(components, func(i, j int) bool { return components[i].Name < components[j].Name })
	edges := make([]ComponentGraphEdge, 0, len(edgesByKey))
	for _, edge := range edgesByKey {
		edges = append(edges, edge)
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].From == edges[j].From {
			return edges[i].To < edges[j].To
		}
		return edges[i].From < edges[j].From
	})
	status := "ok"
	for _, diagnostic := range diagnostics {
		if diagnostic.Severity == SeverityError {
			status = "error"
			break
		}
		if diagnostic.Severity == SeverityWarning {
			status = "warning"
		}
	}
	root := firstNonEmpty(component, rootView)
	if root == "" && len(components) > 0 {
		root = components[0].Name
	}
	return ComponentGraphReport{"1", status, source, root, components, edges, diagnostics}, nil
}

func graphSourcePaths(source, includeGlob, excludeGlob string) ([]string, error) {
	info, err := os.Stat(source)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if graphPathIncluded(filepath.Base(source), includeGlob, excludeGlob) {
			return []string{source}, nil
		}
		return []string{}, nil
	}
	paths := []string{}
	err = filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "generated" || entry.Name() == "tmp" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isGraphSourceFile(path) {
			return nil
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			rel = path
		}
		if graphPathIncluded(rel, includeGlob, excludeGlob) {
			paths = append(paths, path)
		}
		return nil
	})
	sort.Strings(paths)
	return paths, err
}

func isGraphSourceFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".swift", ".xaml", ".xml", ".qml", ".ui", ".cpp", ".cc", ".cxx", ".hpp", ".hh", ".h":
		return true
	default:
		return false
	}
}

func graphPathIncluded(path, includeGlob, excludeGlob string) bool {
	if excludeGlob != "" {
		if ok, _ := filepath.Match(excludeGlob, path); ok {
			return false
		}
		if ok, _ := filepath.Match(excludeGlob, filepath.Base(path)); ok {
			return false
		}
	}
	if includeGlob == "" {
		return true
	}
	if ok, _ := filepath.Match(includeGlob, path); ok {
		return true
	}
	ok, _ := filepath.Match(includeGlob, filepath.Base(path))
	return ok
}

func graphComponentReferences(root Node) []string {
	seen := map[string]struct{}{}
	out := []string{}
	var walk func(Node)
	walk = func(node Node) {
		if node.Kind == KindComponent {
			name := strings.TrimSpace(node.Expression)
			if name != "" {
				if _, ok := seen[name]; !ok {
					seen[name] = struct{}{}
					out = append(out, name)
				}
			}
		}
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	sort.Strings(out)
	return out
}

func ComponentGraphText(report ComponentGraphReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "loom component graph\nstatus: %s\nsource: %s\nroot: %s\ncomponents: %d\nedges: %d\n\n", report.Status, report.Source, report.Root, len(report.Components), len(report.Edges))
	for _, component := range report.Components {
		state := ""
		if component.Unresolved {
			state = " unresolved"
		}
		fmt.Fprintf(&b, "- %s [%s]%s", component.Name, component.Dialect, state)
		if component.Path != "" {
			fmt.Fprintf(&b, " %s", component.Path)
		}
		if component.LayoutNodes > 0 {
			fmt.Fprintf(&b, " (%d layout nodes)", component.LayoutNodes)
		}
		b.WriteString("\n")
	}
	if len(report.Edges) > 0 {
		b.WriteString("\nedges\n")
		for _, edge := range report.Edges {
			fmt.Fprintf(&b, "- %s -> %s: %s\n", edge.From, edge.To, edge.Reason)
		}
	}
	return b.String()
}

func ComponentGraphDOT(report ComponentGraphReport) string {
	var b strings.Builder
	b.WriteString("digraph loom_components {\n")
	b.WriteString("  rankdir=LR;\n")
	for _, component := range report.Components {
		attrs := []string{fmt.Sprintf("label=%q", component.Name)}
		if component.Unresolved {
			attrs = append(attrs, "style=dashed")
		}
		fmt.Fprintf(&b, "  %q [%s];\n", component.Name, strings.Join(attrs, ", "))
	}
	for _, edge := range report.Edges {
		fmt.Fprintf(&b, "  %q -> %q [label=%q];\n", edge.From, edge.To, edge.Reason)
	}
	b.WriteString("}\n")
	return b.String()
}
