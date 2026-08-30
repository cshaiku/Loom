import Foundation
import SwiftParser

public struct LoomComponentID: Codable, Hashable, Sendable, CustomStringConvertible {
  public var view: String
  public var component: String

  public init(view: String, component: String) {
    self.view = view
    self.component = component
  }

  public var description: String { "\(view).\(component)" }
}

public struct LoomComponentEdge: Codable, Equatable, Sendable {
  public var from: LoomComponentID
  public var to: LoomComponentID
  public var reference: String
}

public struct LoomComponentGraphNode: Codable, Sendable {
  public var id: LoomComponentID
  public var sourcePath: String
  public var layoutNodeCount: Int
  public var diagnosticCount: Int
  public var references: [String]
  public var unresolvedReferences: [String]
}

public struct LoomComponentGraph: Codable, Sendable {
  public var root: LoomComponentID
  public var sourceRoot: String
  public var sourceFiles: [String]
  public var nodes: [LoomComponentGraphNode]
  public var edges: [LoomComponentEdge]
  public var diagnostics: [LoomDiagnostic]

  public var status: String {
    diagnostics.contains { $0.severity == .error } ? "error" : "ok"
  }
}

public struct LoomComponentGraphOptions: Sendable {
  public var include: [String]
  public var exclude: [String]

  public init(include: [String] = ["*.swift", "**/*.swift"], exclude: [String] = []) {
    self.include = include
    self.exclude = exclude
  }
}

private struct GraphView {
  var declaration: CollectedView
  var sourcePath: String
  var components: Set<String>
}

public struct LoomComponentGraphBuilder: Sendable {
  public init() {}

  public func build(
    sourceRoot: String,
    rootView: String,
    component: String = "body",
    options: LoomComponentGraphOptions = .init()
  ) throws -> LoomComponentGraph {
    let rootURL = URL(fileURLWithPath: sourceRoot).standardizedFileURL
    let files = try swiftFiles(at: rootURL, options: options)
    var viewsByName: [String: GraphView] = [:]

    for file in files {
      guard let source = try? String(contentsOf: file, encoding: .utf8) else { continue }
      let syntax = Parser.parse(source: source)
      let collector = ViewDeclarationCollector()
      collector.walk(syntax)
      for view in collector.views {
        viewsByName[view.name] = GraphView(
          declaration: view,
          sourcePath: file.path,
          components: Set(ComputedPropertyExtractor.properties(from: view.source))
        )
      }
    }

    let rootID = LoomComponentID(view: rootView, component: component)
    var nodesByID: [LoomComponentID: LoomComponentGraphNode] = [:]
    var edges = Set<EdgeKey>()
    var diagnostics: [LoomDiagnostic] = []
    var visiting: [LoomComponentID] = []
    var visited = Set<LoomComponentID>()

    func visit(_ id: LoomComponentID) throws {
      if let cycleStart = visiting.firstIndex(of: id) {
        let cycle = (visiting[cycleStart...] + [id]).map(\.description).joined(separator: " -> ")
        diagnostics.append(
          LoomDiagnostic(
            severity: .error,
            code: "GRAPH003",
            message: "Component dependency cycle detected: \(cycle)."
          ))
        return
      }
      if visited.contains(id) { return }
      guard let view = viewsByName[id.view] else {
        diagnostics.append(
          LoomDiagnostic(
            severity: .error,
            code: "GRAPH001",
            message: "No SwiftUI View named \(id.view) was found."
          ))
        return
      }
      guard view.components.contains(id.component) else {
        diagnostics.append(
          LoomDiagnostic(
            severity: .error,
            code: "GRAPH002",
            message: "\(id.view) has no computed view property named \(id.component)."
          ))
        return
      }

      visiting.append(id)
      let analysis = try SwiftUIFrontend().analyze(
        source: view.declaration.source,
        sourcePath: view.sourcePath,
        rootView: id.view,
        component: id.component
      )
      diagnostics.append(contentsOf: analysis.diagnostics)

      var resolved: [LoomComponentID] = []
      var unresolved: [String] = []
      let references = Array(Set(analysis.layout.componentReferences)).sorted()
      for reference in references {
        if view.components.contains(reference) {
          resolved.append(LoomComponentID(view: id.view, component: reference))
        } else if let targetView = viewsByName[reference], targetView.components.contains("body") {
          resolved.append(LoomComponentID(view: reference, component: "body"))
        } else if ignoredSwiftUIValueReferences.contains(reference) {
          continue
        } else if reference.first?.isUppercase == true {
          unresolved.append(reference)
          diagnostics.append(
            LoomDiagnostic(
              severity: .warning,
              code: "GRAPH004",
              message: "\(id.description) references \(reference), but no matching SwiftUI View was found."
            ))
        }
      }

      nodesByID[id] = LoomComponentGraphNode(
        id: id,
        sourcePath: view.sourcePath,
        layoutNodeCount: analysis.layout.recursiveNodeCount,
        diagnosticCount: analysis.diagnostics.count,
        references: references,
        unresolvedReferences: unresolved
      )

      for target in resolved.sorted(by: { $0.description < $1.description }) {
        edges.insert(EdgeKey(from: id, to: target, reference: target.component))
        try visit(target)
      }

      _ = visiting.popLast()
      visited.insert(id)
    }

    try visit(rootID)

    return LoomComponentGraph(
      root: rootID,
      sourceRoot: rootURL.path,
      sourceFiles: files.map(\.path).sorted(),
      nodes: nodesByID.values.sorted { $0.id.description < $1.id.description },
      edges: edges.map(\.edge).sorted {
        $0.from.description == $1.from.description
          ? $0.to.description < $1.to.description
          : $0.from.description < $1.from.description
      },
      diagnostics: diagnostics
    )
  }

  public func text(_ graph: LoomComponentGraph) -> String {
    var lines = [
      "Loom component graph",
      "Status: \(graph.status)",
      "Source root: \(graph.sourceRoot)",
      "Root: \(graph.root.description)",
      "Source files: \(graph.sourceFiles.count)",
      "Components: \(graph.nodes.count)",
      "Edges: \(graph.edges.count)",
      "",
      "Components",
    ]
    for node in graph.nodes {
      lines.append(
        "  \(node.id.description): \(node.layoutNodeCount) layout nodes, \(node.diagnosticCount) diagnostics"
      )
      if !node.unresolvedReferences.isEmpty {
        lines.append("    unresolved: \(node.unresolvedReferences.joined(separator: ", "))")
      }
    }

    lines.append("")
    lines.append("Dependencies")
    if graph.edges.isEmpty {
      lines.append("  none")
    } else {
      for edge in graph.edges {
        lines.append("  \(edge.from.description) -> \(edge.to.description)")
      }
    }

    lines.append("")
    lines.append("Diagnostics: \(graph.diagnostics.count)")
    if graph.diagnostics.isEmpty {
      lines.append("  none")
    } else {
      for diagnostic in graph.diagnostics {
        lines.append("  [\(diagnostic.severity.rawValue)] \(diagnostic.code) \(diagnostic.message)")
      }
    }
    return lines.joined(separator: "\n") + "\n"
  }

  public func json(_ graph: LoomComponentGraph) throws -> String {
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
    return String(decoding: try encoder.encode(graph), as: UTF8.self) + "\n"
  }

  public func dot(_ graph: LoomComponentGraph) -> String {
    var lines = ["digraph LoomComponentGraph {"]
    lines.append("  rankdir=LR;")
    for node in graph.nodes {
      lines.append("  \"\(escapeDOT(node.id.description))\";")
    }
    for edge in graph.edges {
      lines.append(
        "  \"\(escapeDOT(edge.from.description))\" -> \"\(escapeDOT(edge.to.description))\";")
    }
    lines.append("}")
    return lines.joined(separator: "\n") + "\n"
  }

  private func swiftFiles(at root: URL, options: LoomComponentGraphOptions) throws -> [URL] {
    let fileManager = FileManager.default
    let includeExpressions = options.include.map(glob)
    let excludeExpressions = options.exclude.map(glob)
    var isDirectory: ObjCBool = false
    guard fileManager.fileExists(atPath: root.path, isDirectory: &isDirectory) else {
      throw LoomError.unreadableSource(root.path)
    }
    let files: [URL]
    if isDirectory.boolValue {
      guard
        let enumerator = fileManager.enumerator(
          at: root,
          includingPropertiesForKeys: [.isRegularFileKey, .isDirectoryKey],
          options: [.skipsHiddenFiles]
        )
      else {
        throw LoomError.unreadableSource(root.path)
      }
      var collected: [URL] = []
      for case let url as URL in enumerator {
        if url.hasDirectoryPath && defaultExcludedDirectoryNames.contains(url.lastPathComponent) {
          enumerator.skipDescendants()
          continue
        }
        let values = try? url.resourceValues(forKeys: [.isRegularFileKey])
        if values?.isRegularFile == true && url.pathExtension == "swift" {
          collected.append(url.standardizedFileURL)
        }
      }
      files = collected
    } else {
      files = root.pathExtension == "swift" ? [root] : []
    }

    return files.filter { file in
      let relative = relativePath(file.path, root: root)
      return matchesAny(relative, expressions: includeExpressions)
        && !matchesAny(relative, expressions: excludeExpressions)
    }.sorted { $0.path < $1.path }
  }

  private var defaultExcludedDirectoryNames: Set<String> {
    [".build", "Build", "DerivedData", ".git"]
  }

  private var ignoredSwiftUIValueReferences: Set<String> {
    [
      "AngularGradient",
      "AnyShape",
      "Capsule",
      "Circle",
      "ContainerRelativeShape",
      "Ellipse",
      "LinearGradient",
      "Path",
      "RadialGradient",
      "Rectangle",
      "RoundedRectangle",
      "Section",
    ]
  }

  private func matchesAny(_ path: String, expressions: [NSRegularExpression]) -> Bool {
    expressions.contains { $0.matches(path) }
  }

  private func glob(_ pattern: String) -> NSRegularExpression {
    let escaped = NSRegularExpression.escapedPattern(for: pattern)
      .replacingOccurrences(of: "\\*\\*", with: ".*")
      .replacingOccurrences(of: "\\*", with: "[^/]*")
      .replacingOccurrences(of: "\\?", with: "[^/]")
    return (try? NSRegularExpression(pattern: "^\(escaped)$")) ?? self.neverMatchingExpression
  }

  private func relativePath(_ path: String, root: URL) -> String {
    if root.hasDirectoryPath {
      let prefix = root.path.hasSuffix("/") ? root.path : root.path + "/"
      return path.hasPrefix(prefix) ? String(path.dropFirst(prefix.count)) : path
    }
    return URL(fileURLWithPath: path).lastPathComponent
  }

  private func escapeDOT(_ value: String) -> String {
    value.replacingOccurrences(of: "\\", with: "\\\\").replacingOccurrences(of: "\"", with: "\\\"")
  }

  private var neverMatchingExpression: NSRegularExpression {
    try! NSRegularExpression(pattern: #"a\Ab"#)
  }
}

private struct EdgeKey: Hashable {
  var from: LoomComponentID
  var to: LoomComponentID
  var reference: String

  var edge: LoomComponentEdge {
    LoomComponentEdge(from: from, to: to, reference: reference)
  }
}

extension NSRegularExpression {
  fileprivate func matches(_ value: String) -> Bool {
    firstMatch(in: value, range: NSRange(value.startIndex..., in: value)) != nil
  }
}
