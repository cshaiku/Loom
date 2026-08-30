import Foundation

public struct LoomProjectManifest: Codable, Sendable {
  public var schemaVersion: String?
  public var project: String
  public var source: String
  public var rootView: String
  public var target: String
  public var existingXaml: String?
  public var referenceLayout: String?
  public var translationGuide: String?
  public var components: [String]
  public var themeResourcePrefix: String?

  public init(
    schemaVersion: String? = "1",
    project: String,
    source: String,
    rootView: String,
    target: String = "winui3",
    existingXaml: String? = nil,
    referenceLayout: String? = nil,
    translationGuide: String? = nil,
    components: [String] = ["body"],
    themeResourcePrefix: String? = nil
  ) {
    self.schemaVersion = schemaVersion
    self.project = project
    self.source = source
    self.rootView = rootView
    self.target = target
    self.existingXaml = existingXaml
    self.referenceLayout = referenceLayout
    self.translationGuide = translationGuide
    self.components = components
    self.themeResourcePrefix = themeResourcePrefix
  }

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case project
    case source
    case rootView
    case target
    case existingXaml
    case referenceLayout
    case translationGuide
    case components
    case themeResourcePrefix
  }
}

public struct LoomProjectComponentResult: Codable, Sendable {
  public var component: String
  public var layoutNodeCount: Int
  public var diagnosticCount: Int
  public var analysisPath: String
  public var xamlPath: String
}

public struct LoomProjectSummary: Codable, Sendable {
  public var project: String
  public var target: String
  public var projectRoot: String
  public var sourcePath: String
  public var rootView: String
  public var components: [LoomProjectComponentResult]
  public var parityPath: String?
  public var totalLayoutNodes: Int
  public var totalDiagnostics: Int
}

public struct LoomProjectRun: Sendable {
  public var summary: LoomProjectSummary
  public var summaryPath: String
}

public struct LoomProjectRunner: Sendable {
  public init() {}

  public func run(
    manifestPath: String,
    projectRoot: String? = nil,
    outputDirectory: String? = nil
  ) throws -> LoomProjectRun {
    let manifestURL = URL(fileURLWithPath: manifestPath).standardizedFileURL
    let manifest = try LoomProjectManifestLoader.load(path: manifestPath)

    guard manifest.schemaVersion == "1" else {
      throw LoomError.invalidProjectManifest("schema_version must be 1")
    }
    guard manifest.target.lowercased() == "winui3" else {
      throw LoomError.unsupportedTarget(manifest.target)
    }
    let components = LoomProjectManifestLoader.uniqueComponents(manifest.components)
    guard !components.isEmpty else {
      throw LoomError.invalidProjectManifest(
        "components must contain at least one computed view property")
    }

    let manifestDirectory = manifestURL.deletingLastPathComponent()
    let rootURL = projectRoot.map { URL(fileURLWithPath: $0) } ?? manifestDirectory
    let standardizedRoot = rootURL.standardizedFileURL
    let sourceURL = resolve(manifest.source, relativeTo: standardizedRoot)
    let outputURL =
      outputDirectory.map { URL(fileURLWithPath: $0) }
      ?? manifestDirectory.appendingPathComponent("Generated", isDirectory: true)

    let frontend = SwiftUIFrontend()
    let analyses = try components.map { component in
      try frontend.analyze(
        sourcePath: sourceURL.path,
        rootView: manifest.rootView,
        component: component
      )
    }

    try FileManager.default.createDirectory(
      at: outputURL,
      withIntermediateDirectories: true
    )

    let reporter = AnalysisReporter()
    let emitter = XAMLEmitter(
      options: XAMLEmissionOptions(themeResourcePrefix: manifest.themeResourcePrefix)
    )
    var componentResults: [LoomProjectComponentResult] = []

    for analysis in analyses {
      let baseName = LoomProjectManifestLoader.safeFileName(analysis.component)
      let analysisURL = outputURL.appendingPathComponent("\(baseName).analysis.json")
      let xamlURL = outputURL.appendingPathComponent("\(baseName).generated.xaml")
      try reporter.json(analysis).write(to: analysisURL, atomically: true, encoding: .utf8)
      try emitter.emit(analysis).write(to: xamlURL, atomically: true, encoding: .utf8)
      componentResults.append(
        LoomProjectComponentResult(
          component: analysis.component,
          layoutNodeCount: analysis.layout.recursiveNodeCount,
          diagnosticCount: analysis.diagnostics.count,
          analysisPath: analysisURL.path,
          xamlPath: xamlURL.path
        )
      )
    }

    var parityPath: String?
    if let existingXaml = manifest.existingXaml,
      let rootAnalysis = analyses.first(where: { $0.component == "body" }) ?? analyses.first
    {
      let xamlURL = resolve(existingXaml, relativeTo: standardizedRoot)
      let parity = try XAMLParityChecker().check(analysis: rootAnalysis, xamlPath: xamlURL.path)
      let parityURL = outputURL.appendingPathComponent("project.parity.json")
      try encoded(parity).write(to: parityURL, atomically: true, encoding: .utf8)
      parityPath = parityURL.path
    }

    let summary = LoomProjectSummary(
      project: manifest.project,
      target: manifest.target,
      projectRoot: standardizedRoot.path,
      sourcePath: sourceURL.path,
      rootView: manifest.rootView,
      components: componentResults,
      parityPath: parityPath,
      totalLayoutNodes: componentResults.reduce(0) { $0 + $1.layoutNodeCount },
      totalDiagnostics: componentResults.reduce(0) { $0 + $1.diagnosticCount }
    )
    let summaryURL = outputURL.appendingPathComponent("project.summary.json")
    try encoded(summary).write(to: summaryURL, atomically: true, encoding: .utf8)
    return LoomProjectRun(summary: summary, summaryPath: summaryURL.path)
  }

  public func text(_ run: LoomProjectRun) -> String {
    let summary = run.summary
    var lines = [
      "Loom project run",
      "Project: \(summary.project)",
      "Target: \(summary.target)",
      "Source: \(summary.sourcePath)",
      "Root view: \(summary.rootView)",
      "Components: \(summary.components.count)",
      "Layout nodes: \(summary.totalLayoutNodes)",
      "Diagnostics: \(summary.totalDiagnostics)",
      "",
    ]
    for component in summary.components {
      lines.append(
        "  \(component.component): \(component.layoutNodeCount) nodes, \(component.diagnosticCount) diagnostics"
      )
    }
    if let parityPath = summary.parityPath {
      lines.append("")
      lines.append("Parity: \(parityPath)")
    }
    lines.append("Summary: \(run.summaryPath)")
    return lines.joined(separator: "\n") + "\n"
  }

  private func resolve(_ path: String, relativeTo root: URL) -> URL {
    if path.hasPrefix("/") {
      return URL(fileURLWithPath: path).standardizedFileURL
    }
    return root.appendingPathComponent(path).standardizedFileURL
  }

  private func encoded<T: Encodable>(_ value: T) throws -> String {
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
    return String(decoding: try encoder.encode(value), as: UTF8.self) + "\n"
  }
}

enum LoomProjectManifestLoader {
  static func load(path: String) throws -> LoomProjectManifest {
    let url = URL(fileURLWithPath: path).standardizedFileURL
    guard let data = try? Data(contentsOf: url) else {
      throw LoomError.unreadableSource(path)
    }
    do {
      return try JSONDecoder().decode(LoomProjectManifest.self, from: data)
    } catch {
      throw LoomError.invalidProjectManifest(error.localizedDescription)
    }
  }

  static func uniqueComponents(_ components: [String]) -> [String] {
    var seen = Set<String>()
    var result: [String] = []
    for component in components {
      let trimmed = component.trimmingCharacters(in: .whitespacesAndNewlines)
      if !trimmed.isEmpty && seen.insert(trimmed).inserted {
        result.append(trimmed)
      }
    }
    return result
  }

  static func safeFileName(_ component: String) -> String {
    let allowed = CharacterSet.alphanumerics.union(CharacterSet(charactersIn: "-_"))
    let scalars = component.unicodeScalars.map {
      allowed.contains($0) ? Character(String($0)) : "-"
    }
    let result = String(scalars).trimmingCharacters(in: CharacterSet(charactersIn: "-"))
    return result.isEmpty ? "component" : result
  }
}

public struct LoomProjectValidationIssue: Codable, Sendable {
  public var severity: LoomDiagnosticSeverity
  public var code: String
  public var path: String
  public var detail: String
  public var fix: String
}

public struct LoomProjectValidationReport: Codable, Sendable {
  public var schemaVersion = "1"
  public var status: String
  public var project: String?
  public var issues: [LoomProjectValidationIssue]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case status
    case project
    case issues
  }
}

public struct LoomProjectValidator: Sendable {
  public init() {}

  public func validate(manifestPath: String, projectRoot: String? = nil)
    -> LoomProjectValidationReport
  {
    let manifestURL = URL(fileURLWithPath: manifestPath).standardizedFileURL
    let manifest: LoomProjectManifest
    do {
      manifest = try LoomProjectManifestLoader.load(path: manifestPath)
    } catch {
      return report(
        project: nil,
        issues: [
          issue(
            "error", code: "manifest.decode", path: manifestPath, detail: String(describing: error),
            fix: "Correct the JSON syntax and required fields.")
        ])
    }

    var issues: [LoomProjectValidationIssue] = []
    if manifest.schemaVersion != "1" {
      issues.append(
        issue(
          "error", code: "manifest.schema_version", path: "schema_version",
          detail: "Expected schema_version 1.", fix: "Set schema_version to \"1\"."))
    }
    issues.append(contentsOf: unknownKeyIssues(manifestPath: manifestPath))
    if manifest.project.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
      issues.append(
        issue(
          "error", code: "manifest.project.empty", path: "project",
          detail: "Project name is empty.", fix: "Set project to a stable display name."))
    }
    if manifest.target.lowercased() != "winui3" {
      issues.append(
        issue(
          "error", code: "manifest.target.unsupported", path: "target",
          detail: "Unsupported target \(manifest.target).", fix: "Set target to winui3."))
    }
    let components = LoomProjectManifestLoader.uniqueComponents(manifest.components)
    if components.isEmpty {
      issues.append(
        issue(
          "error", code: "manifest.components.empty", path: "components",
          detail: "No components are declared.",
          fix: "Add body or another computed SwiftUI view property."))
    }

    let manifestDirectory = manifestURL.deletingLastPathComponent()
    let root = projectRoot.map { URL(fileURLWithPath: $0) } ?? manifestDirectory
    let sourceURL = resolve(manifest.source, relativeTo: root.standardizedFileURL)
    if !FileManager.default.fileExists(atPath: sourceURL.path) {
      issues.append(
        issue(
          "error", code: "source.missing", path: "source",
          detail: "Swift source does not exist at \(sourceURL.path).",
          fix: "Correct source or --project-root."))
    } else {
      for component in components {
        do {
          _ = try SwiftUIFrontend().analyze(
            sourcePath: sourceURL.path, rootView: manifest.rootView, component: component)
        } catch {
          issues.append(
            issue(
              "error", code: "component.unresolved", path: "components.\(component)",
              detail: String(describing: error),
              fix: "Correct rootView/component or remove the stale component entry."))
        }
      }
    }
    if let existingXaml = manifest.existingXaml {
      let xamlURL = resolve(existingXaml, relativeTo: root.standardizedFileURL)
      if !FileManager.default.fileExists(atPath: xamlURL.path) {
        issues.append(
          issue(
            "error", code: "xaml.missing", path: "existingXaml",
            detail: "XAML source does not exist at \(xamlURL.path).",
            fix: "Correct existingXaml or --project-root."))
      }
    }
    for (field, path) in [
      ("referenceLayout", manifest.referenceLayout),
      ("translationGuide", manifest.translationGuide),
    ] {
      if let path {
        let url = resolve(path, relativeTo: root.standardizedFileURL)
        if !FileManager.default.fileExists(atPath: url.path) {
          issues.append(
            issue(
              "warning", code: "reference.missing", path: field,
              detail: "Reference does not exist at \(url.path).",
              fix: "Correct or remove the optional reference path."))
        }
      }
    }
    return report(project: manifest.project, issues: issues)
  }

  public func text(_ report: LoomProjectValidationReport) -> String {
    var lines = [
      "Loom manifest validation",
      "Status: \(report.status)",
      "Project: \(report.project ?? "unknown")",
      "Issues: \(report.issues.count)",
    ]
    for issue in report.issues {
      lines.append("  [\(issue.severity.rawValue)] \(issue.code) \(issue.path): \(issue.detail)")
      lines.append("    fix: \(issue.fix)")
    }
    return lines.joined(separator: "\n") + "\n"
  }

  public func json(_ report: LoomProjectValidationReport) throws -> String {
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
    return String(decoding: try encoder.encode(report), as: UTF8.self) + "\n"
  }

  private func resolve(_ path: String, relativeTo root: URL) -> URL {
    path.hasPrefix("/")
      ? URL(fileURLWithPath: path).standardizedFileURL
      : root.appendingPathComponent(path).standardizedFileURL
  }

  private func report(project: String?, issues: [LoomProjectValidationIssue])
    -> LoomProjectValidationReport
  {
    LoomProjectValidationReport(
      status: issues.contains { $0.severity == .error } ? "error" : "ok", project: project,
      issues: issues)
  }

  private func issue(_ severity: String, code: String, path: String, detail: String, fix: String)
    -> LoomProjectValidationIssue
  {
    LoomProjectValidationIssue(
      severity: severity == "error" ? .error : .warning, code: code, path: path, detail: detail,
      fix: fix)
  }

  private func unknownKeyIssues(manifestPath: String) -> [LoomProjectValidationIssue] {
    guard let data = try? Data(contentsOf: URL(fileURLWithPath: manifestPath)),
      let object = try? JSONSerialization.jsonObject(with: data) as? [String: Any]
    else { return [] }
    let known: Set<String> = [
      "schema_version", "project", "source", "rootView", "target", "existingXaml",
      "referenceLayout", "translationGuide", "components", "themeResourcePrefix",
    ]
    return object.keys.filter { !known.contains($0) }.sorted().map { key in
      issue(
        "error", code: "manifest.key.unsupported", path: key,
        detail: "Unsupported manifest key \(key).",
        fix: "Remove the key or check `loom config:schema`.")
    }
  }
}

public enum LoomProjectSchema {
  public static let json = """
    {
      "$schema": "https://json-schema.org/draft/2020-12/schema",
      "title": "Loom project manifest",
      "type": "object",
      "required": ["schema_version", "project", "source", "rootView", "target", "components"],
      "properties": {
        "schema_version": { "const": "1" },
        "project": { "type": "string", "minLength": 1 },
        "source": { "type": "string", "minLength": 1 },
        "rootView": { "type": "string", "minLength": 1 },
        "target": { "const": "winui3" },
        "existingXaml": { "type": "string" },
        "referenceLayout": { "type": "string" },
        "translationGuide": { "type": "string" },
        "components": { "type": "array", "minItems": 1, "items": { "type": "string", "minLength": 1 } },
        "themeResourcePrefix": { "type": "string" }
      },
      "additionalProperties": false
    }
    """
}
