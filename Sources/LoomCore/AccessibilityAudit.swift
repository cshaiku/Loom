import Foundation

public enum LoomAuditCategory: String, Codable, Sendable {
  case accessibility
  case layout
  case design
  case malformed
  case redundant
  case missing
}

public struct LoomAccessibilityAuditFinding: Codable, Equatable, Sendable {
  public var severity: LoomDiagnosticSeverity
  public var category: LoomAuditCategory
  public var code: String
  public var path: String
  public var kind: LoomNodeKind
  public var message: String
  public var recommendation: String
}

public struct LoomAccessibilityAuditSummary: Codable, Equatable, Sendable {
  public var errors: Int
  public var warnings: Int
  public var info: Int
  public var accessibility: Int
  public var layout: Int
  public var design: Int
  public var malformed: Int
  public var redundant: Int
  public var missing: Int
}

public struct LoomAccessibilityAuditReport: Codable, Sendable {
  public var schemaVersion = "1"
  public var status: String
  public var sourcePath: String
  public var rootView: String
  public var component: String
  public var summary: LoomAccessibilityAuditSummary
  public var findings: [LoomAccessibilityAuditFinding]
  public var diagnostics: [LoomDiagnostic]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case status, sourcePath, rootView, component, summary, findings, diagnostics
  }
}

public struct LoomAccessibilityAuditor: Sendable {
  public init() {}

  public func audit(_ analysis: LoomAnalysis) -> LoomAccessibilityAuditReport {
    var findings: [LoomAccessibilityAuditFinding] = []
    for diagnostic in analysis.diagnostics {
      let category: LoomAuditCategory = diagnostic.severity == .error ? .malformed : .design
      findings.append(
        finding(
          severity: diagnostic.severity,
          category: category,
          code: "AUDIT.SOURCE.\(diagnostic.code)",
          path: analysis.component,
          kind: .root,
          message: diagnostic.message,
          recommendation: "Resolve source diagnostics before relying on generated layout."
        ))
    }

    if analysis.layout.children.isEmpty {
      findings.append(
        finding(
          severity: .error,
          category: .missing,
          code: "AUDIT001",
          path: analysis.component,
          kind: .root,
          message: "The selected component has no visible layout elements.",
          recommendation: "Confirm the root view/component selection or add an explicit empty-state layout."
        ))
    }

    var stack = analysis.layout.children.reversed().map { ($0, $0.kind.rawValue, 1, [LoomNodeKind]()) }
    while let (node, path, depth, ancestors) = stack.popLast() {
      audit(node, path: path, depth: depth, ancestors: ancestors, findings: &findings)
      for child in node.children.reversed() {
        stack.append((child, "\(path)/\(child.kind.rawValue)", depth + 1, ancestors + [node.kind]))
      }
    }

    let summary = summarize(findings)
    return LoomAccessibilityAuditReport(
      status: summary.errors > 0 ? "error" : (summary.warnings > 0 ? "warning" : "ok"),
      sourcePath: analysis.sourcePath,
      rootView: analysis.rootView,
      component: analysis.component,
      summary: summary,
      findings: sortedFindings(findings),
      diagnostics: analysis.diagnostics
    )
  }

  private func sortedFindings(_ findings: [LoomAccessibilityAuditFinding])
    -> [LoomAccessibilityAuditFinding]
  {
    findings.sorted {
      if $0.severity.rawValue != $1.severity.rawValue {
        return Self.severityRank($0.severity) > Self.severityRank($1.severity)
      }
      return $0.path < $1.path
    }
  }

  public func text(_ report: LoomAccessibilityAuditReport) -> String {
    var lines = [
      "Loom accessibility audit",
      "Status: \(report.status)",
      "Source: \(report.sourcePath)",
      "View: \(report.rootView).\(report.component)",
      "",
      "Summary",
      "  errors: \(report.summary.errors)",
      "  warnings: \(report.summary.warnings)",
      "  info: \(report.summary.info)",
      "  accessibility: \(report.summary.accessibility)",
      "  layout: \(report.summary.layout)",
      "  design: \(report.summary.design)",
      "  malformed: \(report.summary.malformed)",
      "  redundant: \(report.summary.redundant)",
      "  missing: \(report.summary.missing)",
      "",
      "Findings: \(report.findings.count)",
    ]
    if report.findings.isEmpty {
      lines.append("  none")
    } else {
      for finding in report.findings {
        lines.append("[\(finding.severity.rawValue)] \(finding.code) \(finding.path) \(finding.category.rawValue)")
        lines.append("  issue: \(finding.message)")
        lines.append("  fix: \(finding.recommendation)")
      }
    }
    return lines.joined(separator: "\n") + "\n"
  }

  public func json(_ report: LoomAccessibilityAuditReport) throws -> String {
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
    return String(decoding: try encoder.encode(report), as: UTF8.self) + "\n"
  }

  public func shouldFail(_ report: LoomAccessibilityAuditReport, mode: LoomErrorFailMode) -> Bool {
    switch mode {
    case .none:
      return false
    case .error:
      return report.summary.errors > 0
    case .warning:
      return report.summary.errors > 0 || report.summary.warnings > 0
    }
  }

  private func audit(
    _ node: LoomNode,
    path: String,
    depth: Int,
    ancestors: [LoomNodeKind],
    findings: inout [LoomAccessibilityAuditFinding]
  ) {
    if depth > 8 {
      findings.append(
        finding(
          severity: .warning,
          category: .design,
          code: "AUDIT010",
          path: path,
          kind: node.kind,
          message: "Layout nesting depth is \(depth), which is difficult to transfer and review.",
          recommendation: "Flatten repeated wrappers or extract a named component boundary."
        ))
    }

    if isContainer(node.kind) && node.children.isEmpty {
      findings.append(
        finding(
          severity: .warning,
          category: .missing,
          code: "AUDIT011",
          path: path,
          kind: node.kind,
          message: "Container has no visible children.",
          recommendation: "Remove the empty container or add an explicit placeholder/empty-state element."
        ))
    }

    if isContainer(node.kind) && node.children.count == 1 && node.modifiers.isEmpty
      && ![.scrollView, .geometryReader, .conditional].contains(node.kind)
    {
      findings.append(
        finding(
          severity: .info,
          category: .redundant,
          code: "AUDIT012",
          path: path,
          kind: node.kind,
          message: "Container wraps a single child without modifiers.",
          recommendation: "Consider removing the wrapper unless it exists for a named semantic grouping."
        ))
    }

    if ancestors.last == node.kind && [.verticalStack, .horizontalStack, .overlayStack, .grid].contains(node.kind)
      && node.modifiers.isEmpty
    {
      findings.append(
        finding(
          severity: .warning,
          category: .redundant,
          code: "AUDIT013",
          path: path,
          kind: node.kind,
          message: "Nested \(node.kind.rawValue) repeats the parent layout kind without a local policy.",
          recommendation: "Collapse the nested layout or add explicit spacing/alignment semantics."
        ))
    }

    if node.kind == .unsupported {
      findings.append(
        finding(
          severity: .error,
          category: .malformed,
          code: "AUDIT014",
          path: path,
          kind: node.kind,
          message: "Unsupported source expression appears in the layout tree.",
          recommendation: "Add a Pattern/emitter mapping or replace the expression with supported layout syntax."
        ))
    }

    switch node.kind {
    case .button:
      auditButton(node, path: path, findings: &findings)
    case .image:
      auditImage(node, path: path, findings: &findings)
    case .textField:
      auditTextField(node, path: path, findings: &findings)
    case .slider, .toggle:
      auditStatefulControl(node, path: path, findings: &findings)
    case .color:
      findings.append(
        finding(
          severity: .info,
          category: .accessibility,
          code: "AUDIT040",
          path: path,
          kind: node.kind,
          message: "Color surface has no inherent accessible meaning.",
          recommendation: "Ensure color is decorative or that meaning is also conveyed by text, icon label, or state."
        ))
    case .geometryReader:
      findings.append(
        finding(
          severity: .warning,
          category: .layout,
          code: "AUDIT050",
          path: path,
          kind: node.kind,
          message: "Geometry-dependent layout may not transfer deterministically across platforms.",
          recommendation: "Replace implicit geometry dependency with named size-class or breakpoint policy."
        ))
    case .overlayStack:
      if node.children.count > 3 {
        findings.append(
          finding(
            severity: .warning,
            category: .design,
            code: "AUDIT051",
            path: path,
            kind: node.kind,
            message: "Overlay has \(node.children.count) layers, which can obscure hit testing and reading order.",
            recommendation: "Declare z-order, hit-test, and accessibility reading-order policy explicitly."
          ))
      }
    case .scrollView:
      if ancestors.contains(.scrollView) {
        findings.append(
          finding(
            severity: .warning,
            category: .layout,
            code: "AUDIT052",
            path: path,
            kind: node.kind,
            message: "Nested scroll region can create ambiguous gesture and keyboard navigation behavior.",
            recommendation: "Avoid nested scrolling or define axis ownership and focus movement policy."
          ))
      }
    case .list:
      if node.children.isEmpty {
        findings.append(
          finding(
            severity: .warning,
            category: .missing,
            code: "AUDIT053",
            path: path,
            kind: node.kind,
            message: "List has no visible item template or child content.",
            recommendation: "Provide an item template, empty state, or collection source contract."
          ))
      }
    default:
      break
    }

    auditFixedTargetSize(node, path: path, findings: &findings)
  }

  private func auditButton(
    _ node: LoomNode,
    path: String,
    findings: inout [LoomAccessibilityAuditFinding]
  ) {
    if accessibleName(for: node).isEmpty {
      findings.append(
        finding(
          severity: .error,
          category: .accessibility,
          code: "AUDIT020",
          path: path,
          kind: node.kind,
          message: "Button has no detectable accessible name.",
          recommendation: "Add visible text, an accessibility label, or a target AutomationProperties.Name."
        ))
    }
    if !hasExplicitAccessibilityMetadata(node) && node.children.contains(where: { $0.kind == .image }) {
      findings.append(
        finding(
          severity: .warning,
          category: .accessibility,
          code: "AUDIT021",
          path: path,
          kind: node.kind,
          message: "Icon-like button relies on child imagery without explicit accessibility metadata.",
          recommendation: "Add an explicit action name and hint that survives platform transfer."
        ))
    }
  }

  private func auditImage(
    _ node: LoomNode,
    path: String,
    findings: inout [LoomAccessibilityAuditFinding]
  ) {
    guard !hasExplicitAccessibilityMetadata(node) && !hasAccessibilityHidden(node) else { return }
    findings.append(
      finding(
        severity: .warning,
        category: .accessibility,
        code: "AUDIT030",
        path: path,
        kind: node.kind,
        message: "Image has no accessible label or decorative-hidden intent.",
        recommendation: "Add an accessibility label for meaningful images or mark decorative images as hidden."
      ))
  }

  private func auditTextField(
    _ node: LoomNode,
    path: String,
    findings: inout [LoomAccessibilityAuditFinding]
  ) {
    if accessibleName(for: node).isEmpty {
      findings.append(
        finding(
          severity: .error,
          category: .accessibility,
          code: "AUDIT031",
          path: path,
          kind: node.kind,
          message: "Text input has no label, placeholder, or accessible name.",
          recommendation: "Provide a stable label or target accessibility name."
        ))
    } else if !hasExplicitAccessibilityMetadata(node) {
      findings.append(
        finding(
          severity: .warning,
          category: .accessibility,
          code: "AUDIT032",
          path: path,
          kind: node.kind,
          message: "Text input appears to rely on placeholder text as its accessible name.",
          recommendation: "Prefer a persistent label or explicit accessibility name."
        ))
    }
  }

  private func auditStatefulControl(
    _ node: LoomNode,
    path: String,
    findings: inout [LoomAccessibilityAuditFinding]
  ) {
    if accessibleName(for: node).isEmpty {
      findings.append(
        finding(
          severity: .warning,
          category: .accessibility,
          code: "AUDIT033",
          path: path,
          kind: node.kind,
          message: "Stateful control has no detectable accessible name.",
          recommendation: "Add a visible label or explicit accessibility name describing the controlled value."
        ))
    }
  }

  private func auditFixedTargetSize(
    _ node: LoomNode,
    path: String,
    findings: inout [LoomAccessibilityAuditFinding]
  ) {
    guard [.button, .toggle, .slider, .textField].contains(node.kind) else { return }
    guard let frame = node.modifiers.first(where: { $0.name == "frame" }) else { return }
    let labels = labeledArguments(frame.arguments)
    let width = numericValue(labels["width"] ?? labels["minWidth"])
    let height = numericValue(labels["height"] ?? labels["minHeight"])
    if let width, width > 0, width < 44 {
      findings.append(
        finding(
          severity: .warning,
          category: .accessibility,
          code: "AUDIT060",
          path: path,
          kind: node.kind,
          message: "Interactive control width is below the 44-unit minimum target heuristic.",
          recommendation: "Increase width or ensure the effective hit target is at least 44 by 44."
        ))
    }
    if let height, height > 0, height < 44 {
      findings.append(
        finding(
          severity: .warning,
          category: .accessibility,
          code: "AUDIT061",
          path: path,
          kind: node.kind,
          message: "Interactive control height is below the 44-unit minimum target heuristic.",
          recommendation: "Increase height or ensure the effective hit target is at least 44 by 44."
        ))
    }
  }

  private func accessibleName(for node: LoomNode) -> String {
    if let label = node.modifiers.first(where: { $0.name == "accessibilityLabel" }),
      let value = firstStringLiteral(in: label.arguments)
    {
      return value
    }
    if let name = node.properties["xaml.Name"], !name.isEmpty { return name }
    if let automationName = node.properties["xaml.AutomationProperties.Name"], !automationName.isEmpty {
      return automationName
    }
    if let automationID = node.properties["xaml.AutomationId"], !automationID.isEmpty {
      return automationID
    }
    if let literal = firstStringLiteral(in: node.arguments), !literal.isEmpty { return literal }
    for child in node.children {
      if child.kind == .text, let literal = firstStringLiteral(in: child.arguments), !literal.isEmpty {
        return literal
      }
    }
    return ""
  }

  private func hasExplicitAccessibilityMetadata(_ node: LoomNode) -> Bool {
    node.modifiers.contains {
      ["accessibilityLabel", "accessibilityIdentifier", "accessibilityHint"].contains($0.name)
    } || node.properties.keys.contains {
      ["xaml.AutomationProperties.Name", "xaml.AutomationId", "xaml.Name"].contains($0)
    }
  }

  private func hasAccessibilityHidden(_ node: LoomNode) -> Bool {
    node.modifiers.contains {
      $0.name == "accessibilityHidden" && $0.arguments.trimmingCharacters(in: .whitespacesAndNewlines) == "true"
    }
  }

  private func isContainer(_ kind: LoomNodeKind) -> Bool {
    [.geometryReader, .verticalStack, .horizontalStack, .overlayStack, .splitView, .grid, .scrollView, .list, .conditional, .loop].contains(kind)
  }

  private func summarize(_ findings: [LoomAccessibilityAuditFinding]) -> LoomAccessibilityAuditSummary {
    LoomAccessibilityAuditSummary(
      errors: findings.filter { $0.severity == .error }.count,
      warnings: findings.filter { $0.severity == .warning }.count,
      info: findings.filter { $0.severity == .info }.count,
      accessibility: findings.filter { $0.category == .accessibility }.count,
      layout: findings.filter { $0.category == .layout }.count,
      design: findings.filter { $0.category == .design }.count,
      malformed: findings.filter { $0.category == .malformed }.count,
      redundant: findings.filter { $0.category == .redundant }.count,
      missing: findings.filter { $0.category == .missing }.count
    )
  }

  private func finding(
    severity: LoomDiagnosticSeverity,
    category: LoomAuditCategory,
    code: String,
    path: String,
    kind: LoomNodeKind,
    message: String,
    recommendation: String
  ) -> LoomAccessibilityAuditFinding {
    LoomAccessibilityAuditFinding(
      severity: severity,
      category: category,
      code: code,
      path: path,
      kind: kind,
      message: message,
      recommendation: recommendation
    )
  }

  private func firstStringLiteral(in text: String) -> String? {
    guard let first = text.firstIndex(of: "\"") else { return nil }
    var index = text.index(after: first)
    var value = ""
    var escaped = false
    while index < text.endIndex {
      let character = text[index]
      if escaped {
        value.append(character)
        escaped = false
      } else if character == "\\" {
        escaped = true
      } else if character == "\"" {
        return value
      } else {
        value.append(character)
      }
      index = text.index(after: index)
    }
    return nil
  }

  private func labeledArguments(_ arguments: String) -> [String: String] {
    var result: [String: String] = [:]
    for part in arguments.split(separator: ",") {
      let pieces = part.split(separator: ":", maxSplits: 1)
      guard pieces.count == 2 else { continue }
      result[pieces[0].trimmingCharacters(in: .whitespacesAndNewlines)] =
        pieces[1].trimmingCharacters(in: .whitespacesAndNewlines)
    }
    return result
  }

  private func numericValue(_ value: String?) -> Double? {
    guard let value else { return nil }
    return Double(value.trimmingCharacters(in: .whitespacesAndNewlines))
  }

  private static func severityRank(_ severity: LoomDiagnosticSeverity) -> Int {
    switch severity {
    case .error: 3
    case .warning: 2
    case .info: 1
    }
  }
}
