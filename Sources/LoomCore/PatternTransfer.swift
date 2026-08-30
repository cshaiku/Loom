import Foundation

public enum LoomTransferPlatform: String, Codable, Sendable, Equatable {
  case swiftui
  case winui3
}

public enum LoomTransferDisposition: String, Codable, Sendable, CaseIterable, Equatable {
  case direct
  case needsPolicy = "needs-policy"
  case needsNativeContract = "needs-native-contract"
  case lossy
  case unsupported
}

public struct LoomPatternTransferOptions: Sendable {
  public var from: LoomTransferPlatform
  public var to: LoomTransferPlatform
  public var patternsDirectory: String

  public init(
    from: LoomTransferPlatform,
    to: LoomTransferPlatform,
    patternsDirectory: String = "Patterns"
  ) {
    self.from = from
    self.to = to
    self.patternsDirectory = patternsDirectory
  }
}

public struct LoomPatternTransferItem: Codable, Equatable, Sendable {
  public var path: String
  public var kind: LoomNodeKind
  public var expression: String
  public var patternID: String?
  public var patternName: String?
  public var disposition: LoomTransferDisposition
  public var reason: String
  public var sourceConstructs: [String]
  public var targetConstructs: [String]
  public var contracts: [String]
  public var policies: [String]
}

public struct LoomPatternTransferSummary: Codable, Equatable, Sendable {
  public var direct: Int
  public var needsPolicy: Int
  public var needsNativeContract: Int
  public var lossy: Int
  public var unsupported: Int

  enum CodingKeys: String, CodingKey {
    case direct, lossy, unsupported
    case needsPolicy = "needs_policy"
    case needsNativeContract = "needs_native_contract"
  }
}

public struct LoomPatternTransferReport: Codable, Sendable {
  public var schemaVersion = "1"
  public var sourcePath: String
  public var from: LoomTransferPlatform
  public var to: LoomTransferPlatform
  public var rootView: String
  public var component: String
  public var asciiPattern: String
  public var summary: LoomPatternTransferSummary
  public var items: [LoomPatternTransferItem]
  public var diagnostics: [LoomDiagnostic]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case sourcePath, from, to, rootView, component, asciiPattern, summary, items, diagnostics
  }
}

public struct LoomASCIIOptions: Sendable {
  public var includeRoot: Bool

  public init(includeRoot: Bool = false) {
    self.includeRoot = includeRoot
  }
}

public struct LoomASCIIPatternRenderer: Sendable {
  public init() {}

  public func render(_ analysis: LoomAnalysis, options: LoomASCIIOptions = .init()) -> String {
    render(analysis.layout, title: "\(analysis.rootView).\(analysis.component)", options: options)
  }

  public func render(
    _ node: LoomNode,
    title: String = "layout",
    options: LoomASCIIOptions = .init()
  ) -> String {
    var lines = ["= \(title)"]
    if options.includeRoot {
      append(node, prefix: "", isLast: true, into: &lines)
    } else {
      for (index, child) in node.children.enumerated() {
        append(child, prefix: "", isLast: index == node.children.count - 1, into: &lines)
      }
    }
    return lines.joined(separator: "\n") + "\n"
  }

  private func append(
    _ node: LoomNode,
    prefix: String,
    isLast: Bool,
    into lines: inout [String]
  ) {
    let branch = isLast ? "\\-- " : "|-- "
    lines.append("\(prefix)\(branch)\(patternName(for: node.kind)) / \(oneLine(node.expression))")
    let childPrefix = prefix + (isLast ? "    " : "|   ")
    for (index, child) in node.children.enumerated() {
      append(child, prefix: childPrefix, isLast: index == node.children.count - 1, into: &lines)
    }
  }

  private func oneLine(_ text: String) -> String {
    let normalized = text.split(whereSeparator: \.isWhitespace).joined(separator: " ")
    return normalized.count > 72 ? String(normalized.prefix(69)) + "..." : normalized
  }

  private func patternName(for kind: LoomNodeKind) -> String {
    var result = ""
    for character in kind.rawValue {
      if character.isUppercase {
        result.append("-")
        result.append(character.lowercased())
      } else {
        result.append(character)
      }
    }
    return result
  }
}

public struct LoomPatternTransferAnalyzer: Sendable {
  public init() {}

  public func analyze(
    analysis: LoomAnalysis,
    options: LoomPatternTransferOptions
  ) throws -> LoomPatternTransferReport {
    let registry = try LoomPatternRegistry(directory: options.patternsDirectory)
    let items = transferItems(for: analysis.layout, registry: registry, options: options)
    return LoomPatternTransferReport(
      sourcePath: analysis.sourcePath,
      from: options.from,
      to: options.to,
      rootView: analysis.rootView,
      component: analysis.component,
      asciiPattern: LoomASCIIPatternRenderer().render(analysis),
      summary: summarize(items),
      items: items,
      diagnostics: analysis.diagnostics
    )
  }

  public func text(_ report: LoomPatternTransferReport) -> String {
    var lines = [
      "Loom pattern transfer",
      "Source: \(report.sourcePath)",
      "Route: \(report.from.rawValue) -> \(report.to.rawValue)",
      "View: \(report.rootView).\(report.component)",
      "",
      "Summary",
      "  direct: \(report.summary.direct)",
      "  needs-policy: \(report.summary.needsPolicy)",
      "  needs-native-contract: \(report.summary.needsNativeContract)",
      "  lossy: \(report.summary.lossy)",
      "  unsupported: \(report.summary.unsupported)",
      "",
      "ASCII Pattern",
    ]
    lines.append(contentsOf: report.asciiPattern.split(separator: "\n").map(String.init))
    lines.append("")
    lines.append("Transfer items")
    if report.items.isEmpty {
      lines.append("  none")
    } else {
      for item in report.items {
        let pattern = item.patternID ?? "none"
        let targets = item.targetConstructs.isEmpty ? "none" : item.targetConstructs.joined(separator: ", ")
        lines.append("[\(item.disposition.rawValue)] \(item.path) \(item.kind.rawValue) pattern=\(pattern)")
        lines.append("  target: \(targets)")
        lines.append("  reason: \(item.reason)")
        if !item.policies.isEmpty {
          lines.append("  policies: \(item.policies.joined(separator: "; "))")
        }
        if !item.contracts.isEmpty {
          lines.append("  native contracts: \(item.contracts.joined(separator: "; "))")
        }
      }
    }
    if !report.diagnostics.isEmpty {
      lines.append("")
      lines.append("Diagnostics: \(report.diagnostics.count)")
      for diagnostic in report.diagnostics {
        lines.append("  [\(diagnostic.severity.rawValue)] \(diagnostic.code) \(diagnostic.message)")
      }
    }
    return lines.joined(separator: "\n") + "\n"
  }

  public func json(_ report: LoomPatternTransferReport) throws -> String {
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
    return String(decoding: try encoder.encode(report), as: UTF8.self) + "\n"
  }

  private func transferItems(
    for root: LoomNode,
    registry: LoomPatternRegistry,
    options: LoomPatternTransferOptions
  ) -> [LoomPatternTransferItem] {
    var items: [LoomPatternTransferItem] = []
    var stack = root.children.reversed().map { ($0, $0.kind.rawValue) }
    while let (node, path) = stack.popLast() {
      items.append(transferItem(for: node, path: path, registry: registry, options: options))
      for child in node.children.reversed() {
        stack.append((child, "\(path)/\(child.kind.rawValue)"))
      }
    }
    return items
  }

  private func transferItem(
    for node: LoomNode,
    path: String,
    registry: LoomPatternRegistry,
    options: LoomPatternTransferOptions
  ) -> LoomPatternTransferItem {
    let pattern = registry.pattern(for: node.kind)
    let sourceMapping = pattern?.mappings.first { $0.platform == options.from.rawValue }
    let targetMapping = pattern?.mappings.first { $0.platform == options.to.rawValue }
    let policies = policyNotes(for: node, pattern: pattern)
    let contracts = nativeContracts(for: node)
    let disposition: LoomTransferDisposition
    let reason: String

    if node.kind == .unsupported || targetMapping == nil {
      disposition = .unsupported
      reason = "No target Pattern mapping exists for \(options.to.rawValue), or the node is explicitly unsupported."
    } else if isLossy(node) {
      disposition = .lossy
      reason = "The structure can be represented, but platform-specific layout behavior may degrade without manual review."
    } else if !contracts.isEmpty {
      disposition = .needsNativeContract
      reason = "The visual element transfers, but behavior, state, or accessibility wiring must remain native."
    } else if !policies.isEmpty {
      disposition = .needsPolicy
      reason = "The element transfers after project policy decisions such as sizing, spacing, or token selection."
    } else {
      disposition = .direct
      reason = "The source Pattern has a target mapping and no additional transfer risk was detected."
    }

    return LoomPatternTransferItem(
      path: path,
      kind: node.kind,
      expression: node.expression,
      patternID: pattern?.id,
      patternName: pattern?.name,
      disposition: disposition,
      reason: reason,
      sourceConstructs: sourceMapping?.constructs ?? [],
      targetConstructs: targetMapping?.constructs ?? [],
      contracts: contracts,
      policies: policies
    )
  }

  private func policyNotes(for node: LoomNode, pattern: LoomPattern?) -> [String] {
    var notes: [String] = []
    if [.verticalStack, .horizontalStack, .grid, .splitView, .spacer, .scrollView, .list].contains(node.kind) {
      notes.append("Confirm spacing, available-size, and adaptive breakpoint policy.")
    }
    if [.text, .textField, .button, .toggle, .slider].contains(node.kind) {
      notes.append("Confirm typography, minimum target size, and platform theme token policy.")
    }
    for modifier in node.modifiers where ["padding", "frame", "background", "foregroundStyle", "font"].contains(modifier.name) {
      notes.append("Translate .\(modifier.name) through project design tokens or target layout policy.")
    }
    if pattern?.variants?.isEmpty == false {
      notes.append("Pattern declares variants; select the variant matching target size class, density, or accessibility mode.")
    }
    return Array(Set(notes)).sorted()
  }

  private func nativeContracts(for node: LoomNode) -> [String] {
    var contracts: [String] = []
    switch node.kind {
    case .button:
      contracts.append("action")
    case .textField, .slider, .toggle:
      contracts.append("state binding")
    case .conditional:
      contracts.append("visibility/state rule")
    case .loop, .list:
      contracts.append("collection source/template")
    case .component:
      contracts.append("component boundary")
    case .text:
      if bindingExpression(from: node.arguments) != nil {
        contracts.append("text binding")
      }
    default:
      break
    }
    for modifier in node.modifiers {
      switch modifier.name {
      case "onAppear", "onDisappear", "onChange", "task":
        contracts.append("lifecycle")
      case "accessibilityIdentifier", "accessibilityLabel":
        contracts.append("accessibility metadata")
      default:
        break
      }
    }
    return Array(Set(contracts)).sorted()
  }

  private func isLossy(_ node: LoomNode) -> Bool {
    if [.geometryReader, .overlayStack].contains(node.kind) { return true }
    if node.modifiers.contains(where: { ["alignmentGuide", "gesture", "animation", "transition"].contains($0.name) }) {
      return true
    }
    return false
  }

  private func summarize(_ items: [LoomPatternTransferItem]) -> LoomPatternTransferSummary {
    LoomPatternTransferSummary(
      direct: items.filter { $0.disposition == .direct }.count,
      needsPolicy: items.filter { $0.disposition == .needsPolicy }.count,
      needsNativeContract: items.filter { $0.disposition == .needsNativeContract }.count,
      lossy: items.filter { $0.disposition == .lossy }.count,
      unsupported: items.filter { $0.disposition == .unsupported }.count
    )
  }

  private func bindingExpression(from arguments: String) -> String? {
    let trimmed = arguments.trimmingCharacters(in: .whitespacesAndNewlines)
    guard !trimmed.isEmpty else { return nil }
    if trimmed.hasPrefix("\"") { return nil }
    return trimmed
  }
}
