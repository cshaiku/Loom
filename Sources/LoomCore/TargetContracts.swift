import Foundation

public enum LoomTargetContractKind: String, Codable, Sendable {
  case action
  case binding
  case lifecycle
  case accessibility
  case themeResource
  case visibility
  case collection
  case component
  case unsupported
}

public struct LoomTargetContractItem: Codable, Equatable, Sendable {
  public var kind: LoomTargetContractKind
  public var name: String
  public var source: String
  public var detail: String
  public var required: Bool
  public var targetHint: String
}

public struct LoomTargetContractReport: Codable, Sendable {
  public var schemaVersion = "1"
  public var target: String
  public var sourcePath: String
  public var rootView: String
  public var component: String
  public var items: [LoomTargetContractItem]
  public var diagnostics: [LoomDiagnostic]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case target, sourcePath, rootView, component, items, diagnostics
  }
}

public struct LoomTargetContractOptions: Sendable {
  public var target: String
  public var themeResourcePrefix: String?

  public init(target: String = "winui3", themeResourcePrefix: String? = nil) {
    self.target = target
    self.themeResourcePrefix = themeResourcePrefix
  }
}

public struct LoomTargetContractGenerator: Sendable {
  public init() {}

  public func generate(
    analysis: LoomAnalysis,
    options: LoomTargetContractOptions = .init()
  ) throws -> LoomTargetContractReport {
    guard options.target.lowercased() == "winui3" else {
      throw LoomError.unsupportedTarget(options.target)
    }

    var items: [LoomTargetContractItem] = []
    if let prefix = options.themeResourcePrefix?.trimmingCharacters(in: .whitespacesAndNewlines),
      !prefix.isEmpty
    {
      items.append(
        item(
          .themeResource,
          name: "\(prefix).*",
          source: "theme-prefix",
          detail: "Generated XAML references project-scoped theme resources.",
          required: true,
          hint: "Define \(prefix)CanvasBrush and \(prefix)BorderBrush in WinUI resources."
        ))
    }

    var stack = analysis.layout.children.reversed().map { ($0, $0.expression) }
    while let (node, path) = stack.popLast() {
      collect(node, path: path, into: &items)
      for child in node.children.reversed() {
        stack.append((child, "\(path)/\(child.expression)"))
      }
    }

    let unique = uniqued(items)
    return LoomTargetContractReport(
      target: options.target.lowercased(),
      sourcePath: analysis.sourcePath,
      rootView: analysis.rootView,
      component: analysis.component,
      items: unique.sorted {
        $0.kind.rawValue == $1.kind.rawValue ? $0.name < $1.name : $0.kind.rawValue < $1.kind.rawValue
      },
      diagnostics: analysis.diagnostics
    )
  }

  public func text(_ report: LoomTargetContractReport) -> String {
    var lines = [
      "Loom target contracts",
      "Target: \(report.target)",
      "Source: \(report.sourcePath)",
      "View: \(report.rootView).\(report.component)",
      "Contracts: \(report.items.count)",
      "",
    ]
    if report.items.isEmpty {
      lines.append("  none")
    } else {
      for item in report.items {
        let requirement = item.required ? "required" : "advisory"
        lines.append("[\(item.kind.rawValue)] \(item.name) (\(requirement))")
        lines.append("  source: \(item.source)")
        lines.append("  detail: \(item.detail)")
        lines.append("  target: \(item.targetHint)")
      }
    }
    return lines.joined(separator: "\n") + "\n"
  }

  public func json(_ report: LoomTargetContractReport) throws -> String {
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
    return String(decoding: try encoder.encode(report), as: UTF8.self) + "\n"
  }

  private func collect(_ node: LoomNode, path: String, into items: inout [LoomTargetContractItem]) {
    switch node.kind {
    case .button:
      let label = firstStringLiteral(in: node.arguments) ?? "Action"
      let actionNames = node.children.filter { $0.kind == .component }.map(\.expression)
      let action = actionNames.first ?? normalizedIdentifier(label, fallback: "buttonAction")
      items.append(
        item(
          .action,
          name: action,
          source: path,
          detail: "SwiftUI Button requires explicit WinUI Click or command wiring.",
          required: true,
          hint: "Create a Click handler or ICommand/x:Bind target for \(action)."
        ))
    case .text:
      if let binding = bindingExpression(from: node.arguments) {
        items.append(
          item(
            .binding,
            name: binding,
            source: path,
            detail: "Dynamic SwiftUI text requires a WinUI view-model property.",
            required: true,
            hint: "Expose \(binding) as a bindable property and raise change notifications."
          ))
      }
    case .textField:
      items.append(
        item(
          .binding,
          name: normalizedIdentifier(firstStringLiteral(in: node.arguments) ?? "textValue", fallback: "textValue"),
          source: path,
          detail: "Text input requires a mutable state binding on WinUI.",
          required: true,
          hint: "Bind Text to a view-model property with the desired update trigger."
        ))
    case .slider:
      items.append(
        item(
          .binding,
          name: "numericValue",
          source: path,
          detail: "Slider/progress value requires a numeric WinUI binding and range policy.",
          required: true,
          hint: "Bind Value and set Minimum/Maximum from the product contract."
        ))
    case .toggle:
      items.append(
        item(
          .binding,
          name: "booleanValue",
          source: path,
          detail: "Toggle state requires a boolean WinUI binding.",
          required: true,
          hint: "Bind IsOn or IsChecked to a view-model boolean."
        ))
    case .conditional:
      let condition = node.properties["condition"] ?? "condition"
      items.append(
        item(
          .visibility,
          name: normalizedIdentifier(condition, fallback: "condition"),
          source: path,
          detail: "SwiftUI conditional layout requires an explicit WinUI visibility/state rule.",
          required: true,
          hint: "Bind Visibility or model this with VisualStateManager."
        ))
    case .loop:
      items.append(
        item(
          .collection,
          name: normalizedIdentifier(node.properties["collection"] ?? node.arguments, fallback: "items"),
          source: path,
          detail: "SwiftUI ForEach requires a WinUI item source and item template.",
          required: true,
          hint: "Bind ItemsSource and translate child layout into ItemTemplate/DataTemplate."
        ))
    case .component:
      items.append(
        item(
          .component,
          name: node.expression,
          source: path,
          detail: "Component reference should map to a generated region, UserControl, or native view.",
          required: false,
          hint: "Generate separately or register a project-specific component mapping."
        ))
    case .unsupported:
      items.append(
        item(
          .unsupported,
          name: node.expression,
          source: path,
          detail: "Unsupported SwiftUI expression requires a native target decision.",
          required: true,
          hint: "Implement manually or add a Loom pattern/emitter mapping."
        ))
    default:
      break
    }

    for modifier in node.modifiers {
      collect(modifier, node: node, path: path, into: &items)
    }
  }

  private func collect(
    _ modifier: LoomModifier,
    node: LoomNode,
    path: String,
    into items: inout [LoomTargetContractItem]
  ) {
    switch modifier.name {
    case "onAppear", "onDisappear", "onChange", "task":
      items.append(
        item(
          .lifecycle,
          name: modifier.name,
          source: path,
          detail: "SwiftUI lifecycle modifier is behavior, not layout.",
          required: true,
          hint: "Wire equivalent page/control lifecycle handling in WinUI."
        ))
    case "accessibilityIdentifier", "accessibilityLabel":
      let value = firstStringLiteral(in: modifier.arguments) ?? node.expression
      items.append(
        item(
          .accessibility,
          name: value,
          source: path,
          detail: "Accessibility metadata should be preserved in target automation properties.",
          required: true,
          hint: "Set AutomationProperties.AutomationId or AutomationProperties.Name."
        ))
    default:
      break
    }
  }

  private func item(
    _ kind: LoomTargetContractKind,
    name: String,
    source: String,
    detail: String,
    required: Bool,
    hint: String
  ) -> LoomTargetContractItem {
    LoomTargetContractItem(
      kind: kind,
      name: name,
      source: source,
      detail: detail,
      required: required,
      targetHint: hint
    )
  }

  private func uniqued(_ items: [LoomTargetContractItem]) -> [LoomTargetContractItem] {
    var seen = Set<String>()
    var result: [LoomTargetContractItem] = []
    for item in items {
      let key = "\(item.kind.rawValue)\u{1F}\(item.name)\u{1F}\(item.source)"
      if seen.insert(key).inserted { result.append(item) }
    }
    return result
  }

  private func bindingExpression(from arguments: String) -> String? {
    guard firstStringLiteral(in: arguments) == nil else { return nil }
    let expression = arguments.trimmingCharacters(in: .whitespacesAndNewlines)
    guard expression.range(of: #"^[A-Za-z_][A-Za-z0-9_.]*$"#, options: .regularExpression) != nil
    else { return nil }
    return expression
  }

  private func normalizedIdentifier(_ value: String, fallback: String) -> String {
    let scalars = value.unicodeScalars.map { scalar -> Character in
      CharacterSet.alphanumerics.contains(scalar) ? Character(String(scalar)) : "_"
    }
    let collapsed = String(scalars)
      .split(separator: "_")
      .joined(separator: "_")
      .trimmingCharacters(in: CharacterSet(charactersIn: "_"))
    return collapsed.isEmpty ? fallback : collapsed
  }

  private func firstStringLiteral(in text: String) -> String? {
    guard let first = text.firstIndex(of: "\"") else { return nil }
    var index = text.index(after: first)
    var result = ""
    var escaping = false
    while index < text.endIndex {
      let character = text[index]
      if escaping {
        result.append(character)
        escaping = false
      } else if character == "\\" {
        escaping = true
      } else if character == "\"" {
        return result
      } else {
        result.append(character)
      }
      index = text.index(after: index)
    }
    return nil
  }
}
