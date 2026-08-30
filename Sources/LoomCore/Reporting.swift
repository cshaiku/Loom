import Foundation

public struct AnalysisReporter: Sendable {
  public init() {}

  public func text(_ analysis: LoomAnalysis) -> String {
    var lines: [String] = []
    lines.append("Loom analysis")
    lines.append("Source: \(analysis.sourcePath)")
    lines.append("View: \(analysis.rootView).\(analysis.component)")
    lines.append("Source nodes: \(analysis.syntaxNodeCount)")
    lines.append("Layout nodes: \(analysis.layout.recursiveNodeCount)")
    lines.append("Component references: \(Set(analysis.layout.componentReferences).count)")
    lines.append("")
    lines.append("Layout tree")
    for child in analysis.layout.children {
      appendTree(child, prefix: "", isLast: child == analysis.layout.children.last, into: &lines)
    }
    lines.append("")
    lines.append("Diagnostics: \(analysis.diagnostics.count)")
    if analysis.diagnostics.isEmpty {
      lines.append("  none")
    } else {
      for diagnostic in analysis.diagnostics {
        lines.append("  [\(diagnostic.severity.rawValue)] \(diagnostic.code) \(diagnostic.message)")
      }
    }
    return lines.joined(separator: "\n") + "\n"
  }

  public func json(_ analysis: LoomAnalysis) throws -> String {
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
    return String(decoding: try encoder.encode(analysis), as: UTF8.self) + "\n"
  }

  private func appendTree(
    _ node: LoomNode,
    prefix: String,
    isLast: Bool,
    into lines: inout [String]
  ) {
    let branch = isLast ? "└─" : "├─"
    let detail = node.arguments.isEmpty ? "" : "(\(oneLine(node.arguments)))"
    lines.append("\(prefix)\(branch) \(node.kind.rawValue): \(node.expression)\(detail)")
    let childPrefix = prefix + (isLast ? "   " : "│  ")
    for (index, child) in node.children.enumerated() {
      appendTree(child, prefix: childPrefix, isLast: index == node.children.count - 1, into: &lines)
    }
  }

  private func oneLine(_ text: String) -> String {
    let normalized = text.split(whereSeparator: \.isWhitespace).joined(separator: " ")
    return normalized.count > 80 ? String(normalized.prefix(77)) + "…" : normalized
  }
}

public struct LoomParityReport: Codable, Sendable {
  public var swiftUILayoutNodes: Int
  public var swiftUIComponents: [String]
  public var xamlElementCount: Int
  public var diagnostics: [LoomDiagnostic]
}

public struct XAMLParityChecker: Sendable {
  public init() {}

  public func check(analysis: LoomAnalysis, xamlPath: String) throws -> LoomParityReport {
    guard let xaml = try? String(contentsOfFile: xamlPath, encoding: .utf8) else {
      throw LoomError.unreadableSource(xamlPath)
    }

    var diagnostics = analysis.diagnostics
    let elementPattern = #"<[A-Z][A-Za-z0-9.:]*(?:\s|>|/)"#
    let fixedDimensionPattern = #"(?:Width|Height)=\"[0-9]+(?:\.[0-9]+)?\""#
    let elements = matches(elementPattern, in: xaml).count
    let fixedDimensions = matches(fixedDimensionPattern, in: xaml).count
    let componentNames = Array(Set(analysis.layout.componentReferences)).sorted()

    if fixedDimensions > 0 {
      diagnostics.append(
        LoomDiagnostic(
          severity: .info,
          code: "LOOM201",
          message:
            "Existing XAML contains \(fixedDimensions) fixed width/height attributes; review them against SwiftUI min/ideal/max constraints."
        )
      )
    }
    if xaml.contains("<ScrollViewer") && analysis.layout.count(kind: .scrollView) == 0 {
      diagnostics.append(
        LoomDiagnostic(
          severity: .warning,
          code: "LOOM202",
          message:
            "Existing XAML introduces ScrollViewer regions not visible in the extracted SwiftUI component. Confirm that they do not change layout allocation."
        )
      )
    }

    let normalizedXAML = xaml.lowercased()
    let missing = componentNames.filter { component in
      let normalized = component.filter(\.isLetter).lowercased()
      return normalized.count >= 4 && !normalizedXAML.contains(normalized)
    }
    if !missing.isEmpty {
      diagnostics.append(
        LoomDiagnostic(
          severity: .info,
          code: "LOOM203",
          message:
            "Component names not directly traceable in XAML: \(missing.prefix(20).joined(separator: ", "))."
        )
      )
    }

    return LoomParityReport(
      swiftUILayoutNodes: analysis.layout.recursiveNodeCount,
      swiftUIComponents: componentNames,
      xamlElementCount: elements,
      diagnostics: diagnostics
    )
  }

  public func text(_ report: LoomParityReport) -> String {
    var lines = [
      "Loom parity report",
      "SwiftUI layout nodes: \(report.swiftUILayoutNodes)",
      "SwiftUI component references: \(report.swiftUIComponents.count)",
      "XAML elements: \(report.xamlElementCount)",
      "",
      "Diagnostics: \(report.diagnostics.count)",
    ]
    for diagnostic in report.diagnostics {
      lines.append("  [\(diagnostic.severity.rawValue)] \(diagnostic.code) \(diagnostic.message)")
    }
    return lines.joined(separator: "\n") + "\n"
  }

  private func matches(_ pattern: String, in text: String) -> [NSTextCheckingResult] {
    guard let expression = try? NSRegularExpression(pattern: pattern) else { return [] }
    return expression.matches(in: text, range: NSRange(text.startIndex..., in: text))
  }
}
