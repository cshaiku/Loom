import Foundation
import SwiftParser
import SwiftSyntax

struct CollectedView {
  var name: String
  var source: String
  var offset: Int
}

final class ViewDeclarationCollector: SyntaxVisitor {
  var views: [CollectedView] = []

  init() {
    super.init(viewMode: .sourceAccurate)
  }

  override func visit(_ node: StructDeclSyntax) -> SyntaxVisitorContinueKind {
    let inheritsView =
      node.inheritanceClause?.inheritedTypes.contains {
        $0.type.trimmedDescription == "View" || $0.type.trimmedDescription.hasSuffix(".View")
      } ?? false

    if inheritsView {
      views.append(
        CollectedView(
          name: node.name.text,
          source: node.trimmedDescription,
          offset: node.positionAfterSkippingLeadingTrivia.utf8Offset
        )
      )
    }
    return .visitChildren
  }
}

public struct SwiftUIFrontend: Sendable {
  public init() {}

  public func analyze(
    sourcePath: String,
    rootView: String,
    component: String = "body"
  ) throws -> LoomAnalysis {
    guard let source = try? String(contentsOfFile: sourcePath, encoding: .utf8) else {
      throw LoomError.unreadableSource(sourcePath)
    }
    return try analyze(
      source: source,
      sourcePath: sourcePath,
      rootView: rootView,
      component: component
    )
  }

  public func analyze(
    source: String,
    sourcePath: String = "<memory>",
    rootView: String,
    component: String = "body"
  ) throws -> LoomAnalysis {
    let syntax = Parser.parse(source: source)
    let collector = ViewDeclarationCollector()
    collector.walk(syntax)

    guard let view = collector.views.first(where: { $0.name == rootView }) else {
      throw LoomError.viewNotFound(rootView)
    }

    guard
      let body = ComputedPropertyExtractor.extract(
        property: component,
        from: view.source
      )
    else {
      throw LoomError.componentNotFound(view: rootView, component: component)
    }

    var parser = SwiftUIViewExpressionParser(source: body, baseOffset: view.offset)
    let children = parser.parseTopLevel()
    let root = LoomNode(
      kind: .root,
      expression: component,
      properties: ["declaringView": rootView],
      children: children
    )

    var diagnostics = parser.diagnostics
    if root.count(kind: .geometryReader) > 0 {
      diagnostics.append(
        LoomDiagnostic(
          severity: .warning,
          code: "LOOM101",
          message:
            "GeometryReader depends on runtime size proposals; generated XAML preserves structure but requires an explicit sizing rule."
        )
      )
    }
    if root.count(kind: .unsupported) > 0 {
      diagnostics.append(
        LoomDiagnostic(
          severity: .warning,
          code: "LOOM102",
          message: "Unsupported expressions were preserved as diagnostics and XAML comments."
        )
      )
    }

    return LoomAnalysis(
      sourcePath: sourcePath,
      rootView: rootView,
      component: component,
      syntaxNodeCount: Array(syntax.tokens(viewMode: .sourceAccurate)).count,
      layout: root,
      diagnostics: diagnostics
    )
  }
}

enum ComputedPropertyExtractor {
  static func extract(property: String, from source: String) -> String? {
    let characters = Array(source)
    var index = 0

    while index < characters.count {
      guard let varRange = findKeyword("var", in: characters, startingAt: index) else {
        return nil
      }
      var cursor = varRange.upperBound
      skipTrivia(in: characters, cursor: &cursor)
      let name = readIdentifier(in: characters, cursor: &cursor)
      guard name == property else {
        index = varRange.upperBound
        continue
      }

      guard let openingBrace = findOpeningBrace(in: characters, startingAt: cursor),
        let closingBrace = matchingBrace(in: characters, openingAt: openingBrace)
      else {
        return nil
      }
      return String(characters[(openingBrace + 1)..<closingBrace])
    }
    return nil
  }

  static func properties(from source: String) -> [String] {
    let characters = Array(source)
    var index = 0
    var names: [String] = []

    while index < characters.count {
      guard let varRange = findKeyword("var", in: characters, startingAt: index) else {
        break
      }
      var cursor = varRange.upperBound
      skipTrivia(in: characters, cursor: &cursor)
      let name = readIdentifier(in: characters, cursor: &cursor)
      guard !name.isEmpty else {
        index = varRange.upperBound
        continue
      }

      guard let openingBrace = findOpeningBraceBeforeLineBreak(in: characters, startingAt: cursor),
        let closingBrace = matchingBrace(in: characters, openingAt: openingBrace)
      else {
        index = varRange.upperBound
        continue
      }

      names.append(name)
      index = closingBrace + 1
    }
    return names
  }

  private static func findKeyword(
    _ keyword: String,
    in characters: [Character],
    startingAt start: Int
  ) -> Range<Int>? {
    let target = Array(keyword)
    guard !target.isEmpty else { return nil }
    var index = start
    var state = ScanState.code

    while index + target.count <= characters.count {
      updateState(characters, index: &index, state: &state)
      guard state == .code else { continue }

      if Array(characters[index..<(index + target.count)]) == target {
        let beforeIsIdentifier = index > 0 && isIdentifierCharacter(characters[index - 1])
        let afterIndex = index + target.count
        let afterIsIdentifier =
          afterIndex < characters.count && isIdentifierCharacter(characters[afterIndex])
        if !beforeIsIdentifier && !afterIsIdentifier {
          return index..<afterIndex
        }
      }
      index += 1
    }
    return nil
  }

  private static func findOpeningBrace(in characters: [Character], startingAt start: Int) -> Int? {
    var index = start
    var state = ScanState.code
    while index < characters.count {
      updateState(characters, index: &index, state: &state)
      guard state == .code else { continue }
      if characters[index] == "{" { return index }
      index += 1
    }
    return nil
  }

  private static func findOpeningBraceBeforeLineBreak(
    in characters: [Character],
    startingAt start: Int
  ) -> Int? {
    var index = start
    var state = ScanState.code
    while index < characters.count {
      updateState(characters, index: &index, state: &state)
      guard state == .code else { continue }
      if characters[index] == "\n" || characters[index] == "\r" { return nil }
      if characters[index] == "{" { return index }
      index += 1
    }
    return nil
  }

  private static func matchingBrace(in characters: [Character], openingAt opening: Int) -> Int? {
    var index = opening
    var depth = 0
    var state = ScanState.code
    while index < characters.count {
      updateState(characters, index: &index, state: &state)
      guard state == .code else { continue }
      if characters[index] == "{" { depth += 1 }
      if characters[index] == "}" {
        depth -= 1
        if depth == 0 { return index }
      }
      index += 1
    }
    return nil
  }

  private static func skipTrivia(in characters: [Character], cursor: inout Int) {
    while cursor < characters.count && characters[cursor].isWhitespace { cursor += 1 }
  }

  private static func readIdentifier(in characters: [Character], cursor: inout Int) -> String {
    let start = cursor
    while cursor < characters.count && isIdentifierCharacter(characters[cursor]) { cursor += 1 }
    return String(characters[start..<cursor])
  }

  private static func isIdentifierCharacter(_ character: Character) -> Bool {
    character == "_" || character.isLetter || character.isNumber
  }

  private enum ScanState: Equatable {
    case code
    case string
    case lineComment
    case blockComment(Int)
  }

  private static func updateState(
    _ characters: [Character],
    index: inout Int,
    state: inout ScanState
  ) {
    let current = characters[index]
    let next = index + 1 < characters.count ? characters[index + 1] : "\0"

    switch state {
    case .code:
      if current == "\"" {
        state = .string
      } else if current == "/" && next == "/" {
        state = .lineComment
        index += 2
      } else if current == "/" && next == "*" {
        state = .blockComment(1)
        index += 2
      }
    case .string:
      if current == "\\" {
        index = min(index + 2, characters.count)
      } else if current == "\"" {
        state = .code
        index += 1
      } else {
        index += 1
      }
    case .lineComment:
      index += 1
      if current == "\n" { state = .code }
    case .blockComment(let depth):
      if current == "/" && next == "*" {
        state = .blockComment(depth + 1)
        index += 2
      } else if current == "*" && next == "/" {
        state = depth == 1 ? .code : .blockComment(depth - 1)
        index += 2
      } else {
        index += 1
      }
    }
  }
}

private struct SwiftUIViewExpressionParser {
  private let characters: [Character]
  private let baseOffset: Int
  private var cursor = 0
  var diagnostics: [LoomDiagnostic] = []

  init(source: String, baseOffset: Int) {
    self.characters = Array(source)
    self.baseOffset = baseOffset
  }

  mutating func parseTopLevel() -> [LoomNode] {
    parseSequence(untilClosingBrace: false)
  }

  private mutating func parseSequence(untilClosingBrace: Bool) -> [LoomNode] {
    var nodes: [LoomNode] = []
    while cursor < characters.count {
      skipTrivia()
      guard cursor < characters.count else { break }
      if untilClosingBrace && characters[cursor] == "}" {
        cursor += 1
        break
      }
      if characters[cursor] == ";" {
        cursor += 1
        continue
      }

      let before = cursor
      if startsWithKeyword("if") {
        nodes.append(parseConditional())
      } else if startsWithKeyword("switch") || startsWithKeyword("guard") {
        nodes.append(parseUnsupportedStatement())
      } else if let node = parseExpression() {
        nodes.append(node)
      }

      if cursor <= before {
        diagnostics.append(
          LoomDiagnostic(
            severity: .warning,
            code: "LOOM001",
            message: "Skipped an expression that could not be advanced safely.",
            sourceOffset: baseOffset + cursor
          )
        )
        cursor += 1
      }
    }
    return nodes
  }

  private mutating func parseConditional() -> LoomNode {
    let start = cursor
    cursor += 2
    let condition = readUntilTopLevelDelimiter("{").trimmingCharacters(in: .whitespacesAndNewlines)
    guard consume("{") else {
      return unsupported(expressionFrom: start, reason: "conditional without a body")
    }
    let trueChildren = parseSequence(untilClosingBrace: true)
    skipTrivia()

    var children = trueChildren
    var hasElse = false
    if startsWithKeyword("else") {
      hasElse = true
      cursor += 4
      skipTrivia()
      if consume("{") {
        children.append(
          LoomNode(
            kind: .root,
            expression: "else",
            children: parseSequence(untilClosingBrace: true)
          )
        )
      } else if startsWithKeyword("if") {
        children.append(parseConditional())
      }
    }

    return LoomNode(
      kind: .conditional,
      expression: "if",
      properties: ["condition": condition, "hasElse": hasElse ? "true" : "false"],
      children: children
    )
  }

  private mutating func parseUnsupportedStatement() -> LoomNode {
    let start = cursor
    let head = readUntilTopLevelDelimiter("{")
    var raw = head
    if consume("{") {
      raw += "{" + readBalancedBody() + "}"
    }
    diagnostics.append(
      LoomDiagnostic(
        severity: .warning,
        code: "LOOM002",
        message: "Statement requires a manual WinUI translation: \(oneLine(raw)).",
        sourceOffset: baseOffset + start
      )
    )
    return LoomNode(kind: .unsupported, expression: oneLine(raw))
  }

  private mutating func parseExpression() -> LoomNode? {
    let start = cursor
    let head = readExpressionHead().trimmingCharacters(in: .whitespacesAndNewlines)
    if head.isEmpty {
      if cursor < characters.count && characters[cursor] == "}" { return nil }
      skipToLineEnd()
      return nil
    }

    let parsedHead = parseHead(head)
    var children: [LoomNode] = []
    if cursor < characters.count && characters[cursor] == "{" {
      cursor += 1
      consumeClosurePreambleIfPresent()
      children = parseSequence(untilClosingBrace: true)
    }

    var modifiers: [LoomModifier] = parsedHead.inlineModifiers
    modifiers.append(contentsOf: parseTrailingModifiers())
    let kind = classify(parsedHead.name)
    var properties: [String: String] = [:]
    if kind == .conditional { properties["condition"] = parsedHead.arguments }
    if kind == .loop { properties["collection"] = parsedHead.arguments }

    if kind == .unsupported {
      diagnostics.append(
        LoomDiagnostic(
          severity: .warning,
          code: "LOOM003",
          message: "Unsupported SwiftUI expression: \(oneLine(head)).",
          sourceOffset: baseOffset + start
        )
      )
    }

    return LoomNode(
      kind: kind,
      expression: parsedHead.name,
      arguments: parsedHead.arguments,
      properties: properties,
      modifiers: modifiers,
      children: children
    )
  }

  private func classify(_ rawName: String) -> LoomNodeKind {
    let name = rawName.split(separator: ".").last.map(String.init) ?? rawName
    switch name {
    case "GeometryReader": return .geometryReader
    case "VStack", "LazyVStack": return .verticalStack
    case "HStack", "LazyHStack": return .horizontalStack
    case "ZStack": return .overlayStack
    case "HSplitView", "VSplitView", "NavigationSplitView": return .splitView
    case "Grid", "LazyVGrid", "LazyHGrid": return .grid
    case "ScrollView": return .scrollView
    case "List", "Table": return .list
    case "Text", "Label": return .text
    case "TextField", "SecureField": return .textField
    case "Button": return .button
    case "Image": return .image
    case "Slider", "ProgressView": return .slider
    case "Toggle", "Picker": return .toggle
    case "Spacer": return .spacer
    case "Divider": return .divider
    case "ForEach": return .loop
    case "Color": return .color
    case "Group", "AnyView": return .root
    default:
      if rawName.hasPrefix("Color.") || rawName.contains("Theme") { return .color }
      if rawName.first?.isLowercase == true { return .component }
      if rawName.first?.isUppercase == true { return .component }
      return .unsupported
    }
  }

  private mutating func readExpressionHead() -> String {
    let start = cursor
    var parentheses = 0
    var brackets = 0
    var inString = false
    var escaping = false

    while cursor < characters.count {
      let character = characters[cursor]
      if inString {
        if escaping {
          escaping = false
        } else if character == "\\" {
          escaping = true
        } else if character == "\"" {
          inString = false
        }
        cursor += 1
        continue
      }
      if character == "\"" {
        inString = true
        cursor += 1
        continue
      }
      if character == "(" {
        parentheses += 1
      } else if character == ")" {
        parentheses = max(0, parentheses - 1)
      } else if character == "[" {
        brackets += 1
      } else if character == "]" {
        brackets = max(0, brackets - 1)
      }

      if parentheses == 0 && brackets == 0 {
        if character == "{" || character == "}" || character == ";" { break }
        if character == "\n" { break }
      }
      cursor += 1
    }
    return String(characters[start..<cursor])
  }

  private func parseHead(_ head: String) -> (
    name: String, arguments: String, inlineModifiers: [LoomModifier]
  ) {
    let trimmed = head.trimmingCharacters(in: .whitespacesAndNewlines)
    guard let opening = firstTopLevelParenthesis(in: trimmed),
      let closing = matchingParenthesis(in: trimmed, openingAt: opening)
    else {
      return (trimmed, "", [])
    }
    let name = String(trimmed[..<opening]).trimmingCharacters(in: .whitespaces)
    let arguments = String(trimmed[trimmed.index(after: opening)..<closing])
    let suffix = String(trimmed[trimmed.index(after: closing)...])
    return (name, arguments, parseInlineModifiers(suffix))
  }

  private func firstTopLevelParenthesis(in text: String) -> String.Index? {
    var inString = false
    var escaping = false
    for index in text.indices {
      let character = text[index]
      if inString {
        if escaping {
          escaping = false
        } else if character == "\\" {
          escaping = true
        } else if character == "\"" {
          inString = false
        }
      } else if character == "\"" {
        inString = true
      } else if character == "(" {
        return index
      }
    }
    return nil
  }

  private func matchingParenthesis(in text: String, openingAt opening: String.Index) -> String
    .Index?
  {
    var depth = 0
    var inString = false
    var escaping = false
    var index = opening
    while index < text.endIndex {
      let character = text[index]
      if inString {
        if escaping {
          escaping = false
        } else if character == "\\" {
          escaping = true
        } else if character == "\"" {
          inString = false
        }
      } else if character == "\"" {
        inString = true
      } else if character == "(" {
        depth += 1
      } else if character == ")" {
        depth -= 1
        if depth == 0 { return index }
      }
      index = text.index(after: index)
    }
    return nil
  }

  private func parseInlineModifiers(_ suffix: String) -> [LoomModifier] {
    var parser = ModifierTextParser(text: suffix)
    return parser.parse()
  }

  private mutating func parseTrailingModifiers() -> [LoomModifier] {
    var result: [LoomModifier] = []
    while true {
      let saved = cursor
      skipTrivia()
      guard cursor < characters.count, characters[cursor] == "." else {
        cursor = saved
        break
      }
      cursor += 1
      let nameStart = cursor
      while cursor < characters.count && isIdentifierCharacter(characters[cursor]) { cursor += 1 }
      let name = String(characters[nameStart..<cursor])
      skipHorizontalWhitespace()
      var arguments = ""
      if cursor < characters.count && characters[cursor] == "(" {
        arguments = readBalanced(open: "(", close: ")")
      }
      if name.isEmpty {
        cursor = saved
        break
      }
      skipTrivia()
      if cursor < characters.count && characters[cursor] == "{" {
        let closureOffset = cursor
        cursor += 1
        let closureBody = readBalancedBody()
        let summary = oneLine(closureBody)
        arguments += arguments.isEmpty ? "closure: \(summary)" : ", closure: \(summary)"
        diagnostics.append(
          LoomDiagnostic(
            severity: .info,
            code: "LOOM004",
            message:
              "Modifier .\(name) has a closure that belongs in WinUI behavior or a template mapping, not the generated visual tree.",
            sourceOffset: baseOffset + closureOffset
          )
        )
      }
      result.append(LoomModifier(name: name, arguments: arguments))
    }
    return result
  }

  private mutating func consumeClosurePreambleIfPresent() {
    let saved = cursor
    var probe = cursor
    var nesting = 0
    while probe < characters.count && characters[probe] != "\n" && characters[probe] != "}" {
      if characters[probe] == "(" || characters[probe] == "[" { nesting += 1 }
      if characters[probe] == ")" || characters[probe] == "]" { nesting = max(0, nesting - 1) }
      if nesting == 0 && isKeywordAt("in", index: probe) {
        cursor = probe + 2
        return
      }
      probe += 1
    }
    cursor = saved
  }

  private mutating func readUntilTopLevelDelimiter(_ delimiter: Character) -> String {
    let start = cursor
    var parentheses = 0
    var brackets = 0
    while cursor < characters.count {
      let character = characters[cursor]
      if character == "(" { parentheses += 1 }
      if character == ")" { parentheses = max(0, parentheses - 1) }
      if character == "[" { brackets += 1 }
      if character == "]" { brackets = max(0, brackets - 1) }
      if character == delimiter && parentheses == 0 && brackets == 0 { break }
      cursor += 1
    }
    return String(characters[start..<cursor])
  }

  private mutating func readBalancedBody() -> String {
    let start = cursor
    var depth = 1
    while cursor < characters.count {
      if characters[cursor] == "{" { depth += 1 }
      if characters[cursor] == "}" {
        depth -= 1
        if depth == 0 {
          let value = String(characters[start..<cursor])
          cursor += 1
          return value
        }
      }
      cursor += 1
    }
    return String(characters[start..<cursor])
  }

  private mutating func readBalanced(open: Character, close: Character) -> String {
    guard cursor < characters.count, characters[cursor] == open else { return "" }
    cursor += 1
    let start = cursor
    var depth = 1
    var inString = false
    var escaping = false
    while cursor < characters.count {
      let character = characters[cursor]
      if inString {
        if escaping {
          escaping = false
        } else if character == "\\" {
          escaping = true
        } else if character == "\"" {
          inString = false
        }
      } else if character == "\"" {
        inString = true
      } else if character == open {
        depth += 1
      } else if character == close {
        depth -= 1
        if depth == 0 {
          let result = String(characters[start..<cursor])
          cursor += 1
          return result
        }
      }
      cursor += 1
    }
    return String(characters[start..<cursor])
  }

  private mutating func skipTrivia() {
    while cursor < characters.count {
      if characters[cursor].isWhitespace {
        cursor += 1
        continue
      }
      if cursor + 1 < characters.count && characters[cursor] == "/" && characters[cursor + 1] == "/"
      {
        cursor += 2
        while cursor < characters.count && characters[cursor] != "\n" { cursor += 1 }
        continue
      }
      if cursor + 1 < characters.count && characters[cursor] == "/" && characters[cursor + 1] == "*"
      {
        cursor += 2
        var depth = 1
        while cursor + 1 < characters.count && depth > 0 {
          if characters[cursor] == "/" && characters[cursor + 1] == "*" {
            depth += 1
            cursor += 2
          } else if characters[cursor] == "*" && characters[cursor + 1] == "/" {
            depth -= 1
            cursor += 2
          } else {
            cursor += 1
          }
        }
        continue
      }
      break
    }
  }

  private mutating func skipHorizontalWhitespace() {
    while cursor < characters.count && (characters[cursor] == " " || characters[cursor] == "\t") {
      cursor += 1
    }
  }

  private mutating func skipToLineEnd() {
    while cursor < characters.count && characters[cursor] != "\n" { cursor += 1 }
  }

  private func startsWithKeyword(_ keyword: String) -> Bool {
    isKeywordAt(keyword, index: cursor)
  }

  private func isKeywordAt(_ keyword: String, index: Int) -> Bool {
    let target = Array(keyword)
    guard index + target.count <= characters.count else { return false }
    guard Array(characters[index..<(index + target.count)]) == target else { return false }
    let before = index > 0 ? characters[index - 1] : " "
    let after = index + target.count < characters.count ? characters[index + target.count] : " "
    return !isIdentifierCharacter(before) && !isIdentifierCharacter(after)
  }

  private mutating func consume(_ character: Character) -> Bool {
    skipTrivia()
    guard cursor < characters.count && characters[cursor] == character else { return false }
    cursor += 1
    return true
  }

  private func isIdentifierCharacter(_ character: Character) -> Bool {
    character == "_" || character.isLetter || character.isNumber
  }

  private func oneLine(_ text: String) -> String {
    text.split(whereSeparator: \.isWhitespace).joined(separator: " ").prefix(140).description
  }

  private func unsupported(expressionFrom start: Int, reason: String) -> LoomNode {
    LoomNode(
      kind: .unsupported,
      expression: String(characters[start..<min(cursor, characters.count)]),
      properties: ["reason": reason]
    )
  }
}

private struct ModifierTextParser {
  let characters: [Character]
  var cursor = 0

  init(text: String) {
    self.characters = Array(text)
  }

  mutating func parse() -> [LoomModifier] {
    var result: [LoomModifier] = []
    while cursor < characters.count {
      while cursor < characters.count && characters[cursor].isWhitespace { cursor += 1 }
      guard cursor < characters.count && characters[cursor] == "." else { break }
      cursor += 1
      let start = cursor
      while cursor < characters.count
        && (characters[cursor].isLetter || characters[cursor].isNumber || characters[cursor] == "_")
      { cursor += 1 }
      let name = String(characters[start..<cursor])
      while cursor < characters.count && characters[cursor].isWhitespace { cursor += 1 }
      var arguments = ""
      if cursor < characters.count && characters[cursor] == "(" {
        cursor += 1
        let argumentStart = cursor
        var depth = 1
        while cursor < characters.count && depth > 0 {
          if characters[cursor] == "(" { depth += 1 }
          if characters[cursor] == ")" {
            depth -= 1
            if depth == 0 { break }
          }
          cursor += 1
        }
        arguments = String(characters[argumentStart..<min(cursor, characters.count)])
        if cursor < characters.count { cursor += 1 }
      }
      if !name.isEmpty { result.append(LoomModifier(name: name, arguments: arguments)) }
    }
    return result
  }
}
