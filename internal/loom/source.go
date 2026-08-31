package loom

import (
	"fmt"
	"path/filepath"
	"strings"
)

func AnalyzeByPlatform(path, platform string) (Analysis, error) {
	switch canonicalPatternPlatform(firstNonEmpty(platform, InferSourcePlatform(path))) {
	case "winui3", "xaml":
		return AnalyzeXAML(path)
	case "swiftui":
		return AnalyzeSwiftUI(path)
	case "qt", "qml", "linux":
		return AnalyzeQt(path)
	default:
		return Analysis{}, fmt.Errorf("unsupported source platform %q for %s", platform, path)
	}
}

func InferSourcePlatform(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".swift":
		return "swiftui"
	case ".xaml", ".xml":
		return "winui3"
	case ".qml", ".ui", ".cpp", ".cc", ".cxx", ".hpp", ".hh", ".h":
		return "qt"
	default:
		return "winui3"
	}
}

func defaultTransferTarget(from string) string {
	switch canonicalPatternPlatform(from) {
	case "swiftui":
		return "winui3"
	default:
		return "macos"
	}
}
