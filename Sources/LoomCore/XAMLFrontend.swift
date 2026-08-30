import Foundation
#if canImport(FoundationXML)
  import FoundationXML
#endif

public struct XAMLFrontend: Sendable {
  public init() {}

  public func analyze(sourcePath: String) throws -> LoomAnalysis {
    guard let source = try? String(contentsOfFile: sourcePath, encoding: .utf8) else {
      throw LoomError.unreadableSource(sourcePath)
    }
    return try analyze(source: source, sourcePath: sourcePath)
  }

  public func analyze(source: String, sourcePath: String = "<memory>") throws -> LoomAnalysis {
    let parser = XMLParser(data: Data(source.utf8))
    let delegate = XAMLParserDelegate()
    parser.delegate = delegate
    guard parser.parse() else {
      let message = parser.parserError?.localizedDescription ?? "XAML could not be parsed as XML."
      throw LoomError.malformedViewBody(message)
    }

    let root = LoomNode(
      kind: .root,
      expression: "xaml",
      properties: ["sourceDialect": "winui3"],
      children: delegate.rootNodes
    )
    return LoomAnalysis(
      sourcePath: sourcePath,
      rootView: "XAML",
      component: "root",
      syntaxNodeCount: delegate.elementCount,
      layout: root,
      diagnostics: delegate.diagnostics
    )
  }
}

private final class XAMLParserDelegate: NSObject, XMLParserDelegate {
  var rootNodes: [LoomNode] = []
  var diagnostics: [LoomDiagnostic] = []
  var elementCount = 0
  private var stack: [PendingXAMLNode] = []
  private var skippedPropertyDepth = 0

  func parser(
    _ parser: XMLParser,
    didStartElement elementName: String,
    namespaceURI: String?,
    qualifiedName qName: String?,
    attributes attributeDict: [String: String] = [:]
  ) {
    elementCount += 1
    let name = qName ?? elementName
    if skippedPropertyDepth > 0 || isPropertyElement(name) {
      skippedPropertyDepth += 1
      return
    }
    stack.append(PendingXAMLNode(name: localName(name), attributes: attributeDict))
  }

  func parser(_ parser: XMLParser, foundCharacters string: String) {
    guard skippedPropertyDepth == 0, !stack.isEmpty else { return }
    stack[stack.count - 1].text += string
  }

  func parser(
    _ parser: XMLParser,
    didEndElement elementName: String,
    namespaceURI: String?,
    qualifiedName qName: String?
  ) {
    if skippedPropertyDepth > 0 {
      skippedPropertyDepth -= 1
      return
    }
    guard let pending = stack.popLast() else { return }
    let node = makeNode(pending)
    if stack.isEmpty {
      rootNodes.append(node)
    } else {
      stack[stack.count - 1].children.append(node)
    }
  }

  private func makeNode(_ pending: PendingXAMLNode) -> LoomNode {
    let attributes = normalizedAttributes(pending.attributes)
    let modifiers = frameModifiers(attributes) + accessibilityModifiers(attributes)
    let properties = attributes.reduce(into: ["xamlElement": pending.name]) { result, pair in
      result["xaml.\(pair.key)"] = pair.value
    }

    switch pending.name {
    case "Grid":
      return LoomNode(
        kind: .grid,
        expression: pending.name,
        properties: properties,
        modifiers: modifiers,
        children: pending.children
      )
    case "StackPanel":
      let orientation = attributes["Orientation"] ?? "Vertical"
      return LoomNode(
        kind: orientation == "Horizontal" ? .horizontalStack : .verticalStack,
        expression: pending.name,
        properties: properties,
        modifiers: modifiers,
        children: pending.children
      )
    case "TextBlock":
      let text = attributes["Text"] ?? pending.text.trimmingCharacters(in: .whitespacesAndNewlines)
      return LoomNode(
        kind: .text,
        expression: pending.name,
        arguments: quoted(text),
        properties: properties,
        modifiers: modifiers
      )
    case "Button", "AppBarButton", "HyperlinkButton":
      let content = attributes["Content"] ?? attributes["Label"] ?? pending.accessibleText
      return LoomNode(
        kind: .button,
        expression: pending.name,
        arguments: quoted(content),
        properties: properties,
        modifiers: modifiers,
        children: pending.children
      )
    case "TextBox", "PasswordBox":
      let placeholder = attributes["PlaceholderText"] ?? attributes["Header"] ?? ""
      return LoomNode(
        kind: .textField,
        expression: pending.name,
        arguments: quoted(placeholder),
        properties: properties,
        modifiers: modifiers
      )
    case "Image":
      return LoomNode(
        kind: .image,
        expression: pending.name,
        arguments: quoted(attributes["Source"] ?? ""),
        properties: properties,
        modifiers: modifiers
      )
    case "ScrollViewer":
      return LoomNode(
        kind: .scrollView,
        expression: pending.name,
        properties: properties,
        modifiers: modifiers,
        children: pending.children
      )
    case "ListView", "GridView", "ItemsRepeater":
      return LoomNode(
        kind: .list,
        expression: pending.name,
        properties: properties,
        modifiers: modifiers,
        children: pending.children
      )
    case "Slider", "ProgressBar":
      return LoomNode(kind: .slider, expression: pending.name, properties: properties, modifiers: modifiers)
    case "ToggleSwitch", "CheckBox":
      return LoomNode(kind: .toggle, expression: pending.name, properties: properties, modifiers: modifiers)
    case "Rectangle":
      if attributes["Height"] == "1" || attributes["Width"] == "1" {
        return LoomNode(kind: .divider, expression: pending.name, properties: properties, modifiers: modifiers)
      }
      return unsupported(pending, properties: properties, modifiers: modifiers)
    case "Border":
      if attributes["Background"] != nil {
        return LoomNode(kind: .color, expression: pending.name, properties: properties, modifiers: modifiers)
      }
      return LoomNode(
        kind: .root,
        expression: pending.name,
        properties: properties,
        modifiers: modifiers,
        children: pending.children
      )
    default:
      return unsupported(pending, properties: properties, modifiers: modifiers)
    }
  }

  private func unsupported(
    _ pending: PendingXAMLNode,
    properties: [String: String],
    modifiers: [LoomModifier]
  ) -> LoomNode {
    var boundaryProperties = properties
    boundaryProperties["componentBoundary"] = "native-winui-control"
    boundaryProperties["unsupportedXamlElement"] = pending.name
    boundaryProperties["requiresNativeImplementation"] = "true"
    diagnostics.append(
      LoomDiagnostic(
        severity: .warning,
        code: "XAML.UNSUPPORTED_COMPONENT_BOUNDARY",
        message:
          "Unsupported native WinUI control preserved as a component boundary: \(pending.name)."
      ))
    return LoomNode(
      kind: .component,
      expression: pending.name,
      properties: boundaryProperties,
      modifiers: modifiers,
      children: pending.children
    )
  }

  private func isPropertyElement(_ name: String) -> Bool {
    localName(name).contains(".")
  }

  private func localName(_ name: String) -> String {
    String(name.split(separator: ":").last.map(String.init) ?? name)
  }

  private func normalizedAttributes(_ attributes: [String: String]) -> [String: String] {
    attributes.reduce(into: [:]) { result, pair in
      result[localName(pair.key)] = pair.value
    }
  }

  private func frameModifiers(_ attributes: [String: String]) -> [LoomModifier] {
    var parts: [String] = []
    for (xaml, swift) in [
      ("Width", "width"), ("Height", "height"), ("MinWidth", "minWidth"),
      ("MaxWidth", "maxWidth"), ("MinHeight", "minHeight"), ("MaxHeight", "maxHeight"),
    ] {
      if let value = attributes[xaml], isNumeric(value) {
        parts.append("\(swift): \(value)")
      }
    }
    return parts.isEmpty ? [] : [LoomModifier(name: "frame", arguments: parts.joined(separator: ", "))]
  }

  private func accessibilityModifiers(_ attributes: [String: String]) -> [LoomModifier] {
    guard let value = attributes["AutomationId"] ?? attributes["Name"] else { return [] }
    return [LoomModifier(name: "accessibilityIdentifier", arguments: quoted(value))]
  }

  private func isNumeric(_ value: String) -> Bool {
    value.range(of: #"^-?[0-9]+(?:\.[0-9]+)?$"#, options: .regularExpression) != nil
  }

  private func quoted(_ value: String) -> String {
    "\"\(value.replacingOccurrences(of: "\\", with: "\\\\").replacingOccurrences(of: "\"", with: "\\\""))\""
  }
}

private struct PendingXAMLNode {
  var name: String
  var attributes: [String: String]
  var children: [LoomNode] = []
  var text = ""

  var accessibleText: String {
    let childText = children.compactMap { child -> String? in
      guard child.kind == .text else { return nil }
      return child.arguments.trimmingCharacters(in: CharacterSet(charactersIn: "\""))
    }
    return childText.first ?? ""
  }
}
