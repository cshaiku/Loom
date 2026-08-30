package loom

import (
	"fmt"
	"strings"
	"unicode"
)

func ASCIIAnalysis(analysis Analysis) string {
	var b strings.Builder
	fmt.Fprintf(&b, "= %s.%s\n", analysis.RootView, analysis.Component)
	for i, child := range analysis.Layout.Children {
		appendASCII(&b, child, "", i == len(analysis.Layout.Children)-1)
	}
	return b.String()
}

func appendASCII(b *strings.Builder, node Node, prefix string, last bool) {
	branch := "|-- "
	childPrefix := prefix + "|   "
	if last {
		branch = "\\-- "
		childPrefix = prefix + "    "
	}
	fmt.Fprintf(b, "%s%s%s / %s\n", prefix, branch, patternCase(string(node.Kind)), node.Expression)
	for i, child := range node.Children {
		appendASCII(b, child, childPrefix, i == len(node.Children)-1)
	}
}

func patternCase(value string) string {
	var b strings.Builder
	for _, r := range value {
		if unicode.IsUpper(r) {
			b.WriteRune('-')
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
