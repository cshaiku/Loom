import Foundation

public enum LoomNodeKind: String, Codable, Sendable, CaseIterable {
  case root
  case geometryReader
  case verticalStack
  case horizontalStack
  case overlayStack
  case splitView
  case grid
  case scrollView
  case list
  case text
  case textField
  case button
  case image
  case slider
  case toggle
  case spacer
  case divider
  case conditional
  case loop
  case color
  case component
  case unsupported
}

public struct LoomModifier: Codable, Equatable, Sendable {
  public var name: String
  public var arguments: String

  public init(name: String, arguments: String = "") {
    self.name = name
    self.arguments = arguments
  }
}

public struct LoomNode: Codable, Equatable, Sendable {
  public var kind: LoomNodeKind
  public var expression: String
  public var arguments: String
  public var properties: [String: String]
  public var modifiers: [LoomModifier]
  public var children: [LoomNode]

  public init(
    kind: LoomNodeKind,
    expression: String,
    arguments: String = "",
    properties: [String: String] = [:],
    modifiers: [LoomModifier] = [],
    children: [LoomNode] = []
  ) {
    self.kind = kind
    self.expression = expression
    self.arguments = arguments
    self.properties = properties
    self.modifiers = modifiers
    self.children = children
  }

  public var recursiveNodeCount: Int {
    1 + children.reduce(0) { $0 + $1.recursiveNodeCount }
  }

  public func count(kind target: LoomNodeKind) -> Int {
    (kind == target ? 1 : 0) + children.reduce(0) { $0 + $1.count(kind: target) }
  }

  public var componentReferences: [String] {
    let local = kind == .component ? [expression] : []
    return local + children.flatMap(\.componentReferences)
  }
}

public enum LoomDiagnosticSeverity: String, Codable, Sendable {
  case info
  case warning
  case error
}

public struct LoomDiagnostic: Codable, Equatable, Sendable {
  public var severity: LoomDiagnosticSeverity
  public var code: String
  public var message: String
  public var sourceOffset: Int?

  public init(
    severity: LoomDiagnosticSeverity,
    code: String,
    message: String,
    sourceOffset: Int? = nil
  ) {
    self.severity = severity
    self.code = code
    self.message = message
    self.sourceOffset = sourceOffset
  }
}

public struct LoomAnalysis: Codable, Sendable {
  public var sourcePath: String
  public var rootView: String
  public var component: String
  public var syntaxNodeCount: Int
  public var layout: LoomNode
  public var diagnostics: [LoomDiagnostic]

  public init(
    sourcePath: String,
    rootView: String,
    component: String,
    syntaxNodeCount: Int,
    layout: LoomNode,
    diagnostics: [LoomDiagnostic]
  ) {
    self.sourcePath = sourcePath
    self.rootView = rootView
    self.component = component
    self.syntaxNodeCount = syntaxNodeCount
    self.layout = layout
    self.diagnostics = diagnostics
  }
}

public enum LoomError: Error, CustomStringConvertible {
  case unreadableSource(String)
  case viewNotFound(String)
  case componentNotFound(view: String, component: String)
  case malformedViewBody(String)
  case invalidArguments(String)

  public var description: String {
    switch self {
    case .unreadableSource(let path):
      return "Could not read Swift source at \(path)."
    case .viewNotFound(let name):
      return "No Swift struct named \(name) conforming to View was found."
    case .componentNotFound(let view, let component):
      return "View \(view) has no extractable computed view property named \(component)."
    case .malformedViewBody(let detail):
      return "Could not extract the SwiftUI view body: \(detail)"
    case .invalidArguments(let detail):
      return detail
    }
  }
}
