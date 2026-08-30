import Foundation
import SwiftDiagnostics
import SwiftParser
import SwiftParserDiagnostics
import SwiftSyntax

public enum LoomErrorInspectionKind: String, Codable, Sendable {
  case swift
  case xaml
  case manifest
  case patterns
}

public enum LoomErrorFailMode: String, Sendable {
  case none
  case error
  case warning
}

public struct LoomErrorInspectionOptions: Sendable {
  public var kind: LoomErrorInspectionKind?
  public var rootView: String
  public var component: String

  public init(
    kind: LoomErrorInspectionKind? = nil,
    rootView: String = "ContentView",
    component: String = "body"
  ) {
    self.kind = kind
    self.rootView = rootView
    self.component = component
  }
}

public struct LoomErrorInspectionFinding: Codable, Equatable, Sendable {
  public var severity: LoomDiagnosticSeverity
  public var code: String
  public var source: String
  public var message: String
  public var offset: Int?
  public var line: Int?
  public var column: Int?
}

public struct LoomErrorInspectionReport: Codable, Sendable {
  public var schemaVersion = "1"
  public var status: String
  public var inspectedKind: LoomErrorInspectionKind
  public var source: String
  public var findings: [LoomErrorInspectionFinding]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case status, inspectedKind, source, findings
  }
}

public struct LoomStatusReport: Codable, Sendable {
  public var schemaVersion = "1"
  public var version: String
  public var workingDirectory: String
  public var commands: Int
  public var patternDirectory: String
  public var patternStatus: String
  public var patternCount: Int
  public var issues: [LoomDiagnostic]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case version, workingDirectory, commands, patternDirectory, patternStatus, patternCount, issues
  }
}

public struct LoomCommandCatalogCheckReport: Codable, Sendable {
  public var schemaVersion = "1"
  public var status: String
  public var commands: Int
  public var aliases: Int
  public var issues: [LoomDiagnostic]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case status, commands, aliases, issues
  }
}

public struct LoomVerifyReport: Codable, Sendable {
  public var schemaVersion = "1"
  public var status: String
  public var commandCatalog: LoomCommandCatalogCheckReport
  public var patterns: LoomPatternValidationReport
  public var patternLint: LoomPatternValidationReport

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case status, commandCatalog, patterns, patternLint
  }
}

public struct LoomGuardEntry: Codable, Sendable {
  public var command: String
  public var access: LoomCommandAccess
  public var writeFlags: [String]
  public var description: String
}

public struct LoomGuardsReport: Codable, Sendable {
  public var schemaVersion = "1"
  public var status: String
  public var entries: [LoomGuardEntry]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case status, entries
  }
}

public struct LoomSelfHealEntry: Codable, Sendable {
  public var command: String
  public var flag: String
  public var scope: String
  public var guardrail: String
}

public struct LoomSelfHealPlan: Codable, Sendable {
  public var schemaVersion = "1"
  public var status: String
  public var entries: [LoomSelfHealEntry]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case status, entries
  }
}

public struct LoomDiagnostics: Sendable {
  public static let version = "0.15.0"

  public init() {}

  public func status(patternDirectory: String = "Patterns") -> LoomStatusReport {
    let patternReport = LoomPatternCatalog().validate(directory: patternDirectory)
    return LoomStatusReport(
      version: Self.version,
      workingDirectory: FileManager.default.currentDirectoryPath,
      commands: LoomCommandCatalog.commands.count,
      patternDirectory: patternReport.directory,
      patternStatus: patternReport.status,
      patternCount: patternReport.patternCount,
      issues: patternReport.issues.map {
        LoomDiagnostic(severity: $0.severity, code: $0.code, message: "\($0.path): \($0.detail)")
      }
    )
  }

  public func commandCatalogCheck() -> LoomCommandCatalogCheckReport {
    let commands = LoomCommandCatalog.commands
    var issues: [LoomDiagnostic] = []
    var commandNames = Set<String>()
    var symbols = Set<String>()
    var aliasCount = 0

    for command in commands {
      if !commandNames.insert(command.command).inserted {
        issues.append(issue("CATALOG001", "Duplicate command \(command.command)."))
      }
      if !symbols.insert(command.command).inserted {
        issues.append(issue("CATALOG002", "Command collides with another command or alias: \(command.command)."))
      }
      if command.name.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        || command.description.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
        || command.category.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty
      {
        issues.append(issue("CATALOG003", "Command \(command.command) has incomplete metadata."))
      }
      if command.synopsis.isEmpty {
        issues.append(issue("CATALOG004", "Command \(command.command) has no synopsis."))
      }
      for synopsis in command.synopsis where !synopsis.hasPrefix("loom \(command.command)") {
        issues.append(issue("CATALOG005", "Synopsis for \(command.command) does not start with the canonical command."))
      }
      if command.access == .read && !command.writeFlags.isEmpty {
        issues.append(issue("CATALOG006", "Read-only command \(command.command) declares write flags."))
      }
      for alias in command.aliases {
        aliasCount += 1
        if alias.contains(":") {
          issues.append(issue("CATALOG007", "Alias \(alias) should remain short and non-namespaced."))
        }
        if !symbols.insert(alias).inserted {
          issues.append(issue("CATALOG008", "Alias collides with another command or alias: \(alias)."))
        }
      }
    }

    return LoomCommandCatalogCheckReport(
      status: issues.isEmpty ? "ok" : "error",
      commands: commands.count,
      aliases: aliasCount,
      issues: issues
    )
  }

  public func verify(patternDirectory: String = "Patterns") -> LoomVerifyReport {
    let commandCatalog = commandCatalogCheck()
    let patterns = LoomPatternCatalog().validate(directory: patternDirectory)
    let lint = LoomPatternCatalog().lint(directory: patternDirectory)
    let status =
      commandCatalog.status == "ok" && patterns.status == "ok" && lint.status == "ok"
      ? "ok" : "error"
    return LoomVerifyReport(
      status: status,
      commandCatalog: commandCatalog,
      patterns: patterns,
      patternLint: lint
    )
  }

  public func guardsSummary() -> LoomGuardsReport {
    let entries = LoomCommandCatalog.commands
      .filter { $0.access != .read || !$0.writeFlags.isEmpty }
      .map {
        LoomGuardEntry(
          command: $0.command,
          access: $0.access,
          writeFlags: $0.writeFlags,
          description: $0.description
        )
      }
      .sorted { $0.command < $1.command }
    return LoomGuardsReport(status: "ok", entries: entries)
  }

  public func selfHealPlan() -> LoomSelfHealPlan {
    LoomSelfHealPlan(
      status: "ok",
      entries: [
        LoomSelfHealEntry(
          command: "generate:xaml",
          flag: "--init-region",
          scope: "Create a missing generated XAML host file with one Loom-owned region.",
          guardrail: "Existing files still require explicit LOOM-BEGIN / LOOM-END markers."
        )
      ]
    )
  }

  public func inspectErrors(
    path: String,
    options: LoomErrorInspectionOptions = .init()
  ) -> LoomErrorInspectionReport {
    let url = URL(fileURLWithPath: path).standardizedFileURL
    let kind = options.kind ?? inferredKind(for: url)
    let findings: [LoomErrorInspectionFinding]
    switch kind {
    case .swift:
      findings = inspectSwiftErrors(
        path: url.path,
        rootView: options.rootView,
        component: options.component
      )
    case .xaml:
      findings = inspectXAMLErrors(path: url.path)
    case .manifest:
      findings = inspectManifestErrors(path: url.path)
    case .patterns:
      findings = inspectPatternErrors(directory: url.path)
    }
    return LoomErrorInspectionReport(
      status: findings.contains { $0.severity == .error } ? "error" : "ok",
      inspectedKind: kind,
      source: url.path,
      findings: findings.sorted {
        if ($0.offset ?? Int.max) != ($1.offset ?? Int.max) {
          return ($0.offset ?? Int.max) < ($1.offset ?? Int.max)
        }
        return $0.code < $1.code
      }
    )
  }

  public func shouldFail(_ report: LoomErrorInspectionReport, mode: LoomErrorFailMode) -> Bool {
    switch mode {
    case .none:
      return false
    case .error:
      return report.findings.contains { $0.severity == .error }
    case .warning:
      return report.findings.contains { $0.severity == .error || $0.severity == .warning }
    }
  }

  public func text(_ report: LoomStatusReport) -> String {
    var lines = [
      "Loom status",
      "Version: \(report.version)",
      "Working directory: \(report.workingDirectory)",
      "Commands: \(report.commands)",
      "Patterns: \(report.patternStatus) (\(report.patternCount))",
      "Pattern directory: \(report.patternDirectory)",
      "Issues: \(report.issues.count)",
    ]
    lines.append(contentsOf: report.issues.map { "  [\($0.severity.rawValue)] \($0.code) \($0.message)" })
    return lines.joined(separator: "\n") + "\n"
  }

  public func text(_ report: LoomCommandCatalogCheckReport) -> String {
    var lines = [
      "Loom command catalog check",
      "Status: \(report.status)",
      "Commands: \(report.commands)",
      "Aliases: \(report.aliases)",
      "Issues: \(report.issues.count)",
    ]
    lines.append(contentsOf: report.issues.map { "  [\($0.severity.rawValue)] \($0.code) \($0.message)" })
    return lines.joined(separator: "\n") + "\n"
  }

  public func text(_ report: LoomVerifyReport) -> String {
    [
      "Loom verify",
      "Status: \(report.status)",
      "Command catalog: \(report.commandCatalog.status)",
      "Patterns: \(report.patterns.status)",
      "Pattern lint: \(report.patternLint.status)",
      "Issues: \(report.commandCatalog.issues.count + report.patterns.issues.count + report.patternLint.issues.count)",
    ].joined(separator: "\n") + "\n"
  }

  public func text(_ report: LoomGuardsReport) -> String {
    var lines = ["Loom guards summary", "Status: \(report.status)", "Writing commands: \(report.entries.count)", ""]
    for entry in report.entries {
      let flags = entry.writeFlags.isEmpty ? "always writes" : entry.writeFlags.joined(separator: ", ")
      lines.append("\(entry.access.compactMarker) \(entry.command): \(flags)")
      lines.append("  \(entry.description)")
    }
    return lines.joined(separator: "\n") + "\n"
  }

  public func text(_ report: LoomSelfHealPlan) -> String {
    var lines = ["Loom self-heal plan", "Status: \(report.status)", "Healable actions: \(report.entries.count)", ""]
    for entry in report.entries {
      lines.append("\(entry.command) \(entry.flag)")
      lines.append("  scope: \(entry.scope)")
      lines.append("  guardrail: \(entry.guardrail)")
    }
    return lines.joined(separator: "\n") + "\n"
  }

  public func text(_ report: LoomErrorInspectionReport) -> String {
    var lines = [
      "Loom error inspection",
      "Status: \(report.status)",
      "Kind: \(report.inspectedKind.rawValue)",
      "Source: \(report.source)",
      "Findings: \(report.findings.count)",
      "",
    ]
    if report.findings.isEmpty {
      lines.append("  none")
    } else {
      for finding in report.findings {
        let location: String
        if let line = finding.line, let column = finding.column {
          location = "\(line):\(column)"
        } else if let offset = finding.offset {
          location = "offset \(offset)"
        } else {
          location = "unknown"
        }
        lines.append("[\(finding.severity.rawValue)] \(finding.code) \(location)")
        lines.append("  \(finding.message)")
      }
    }
    return lines.joined(separator: "\n") + "\n"
  }

  public func json<T: Encodable>(_ value: T) throws -> String {
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
    return String(decoding: try encoder.encode(value), as: UTF8.self) + "\n"
  }

  private func issue(_ code: String, _ message: String) -> LoomDiagnostic {
    LoomDiagnostic(severity: .error, code: code, message: message)
  }

  private func inferredKind(for url: URL) -> LoomErrorInspectionKind {
    if url.hasDirectoryPath { return .patterns }
    switch url.pathExtension.lowercased() {
    case "swift", "txt":
      return .swift
    case "xaml", "xml":
      return .xaml
    case "json":
      return url.lastPathComponent.hasSuffix(".pattern.json") ? .patterns : .manifest
    default:
      return .swift
    }
  }

  private func inspectSwiftErrors(path: String, rootView: String, component: String)
    -> [LoomErrorInspectionFinding]
  {
    guard let source = try? String(contentsOfFile: path, encoding: .utf8) else {
      return [finding(.error, code: "SOURCE001", source: path, message: "Could not read Swift source.")]
    }
    let tree = Parser.parse(source: source)
    let converter = SourceLocationConverter(fileName: path, tree: tree)
    var findings = ParseDiagnosticsGenerator.diagnostics(for: tree).map { diagnostic in
      let location = diagnostic.location(converter: converter)
      return finding(
        severity(from: diagnostic.diagMessage.severity),
        code: "SWIFT.PARSE",
        source: path,
        message: diagnostic.message,
        offset: diagnostic.position.utf8Offset,
        line: location.line,
        column: location.column
      )
    }

    do {
      let analysis = try SwiftUIFrontend().analyze(
        source: source,
        sourcePath: path,
        rootView: rootView,
        component: component
      )
      findings.append(contentsOf: analysis.diagnostics.map { finding($0, source: path) })
    } catch {
      findings.append(
        finding(
          .error,
          code: "LOOM.SWIFTUI",
          source: path,
          message: String(describing: error)
        ))
    }
    return findings
  }

  private func inspectXAMLErrors(path: String) -> [LoomErrorInspectionFinding] {
    do {
      let analysis = try XAMLFrontend().analyze(sourcePath: path)
      return analysis.diagnostics.map { finding($0, source: path) }
    } catch {
      return [
        finding(.error, code: "LOOM.XAML", source: path, message: String(describing: error))
      ]
    }
  }

  private func inspectManifestErrors(path: String) -> [LoomErrorInspectionFinding] {
    LoomProjectValidator().validate(manifestPath: path).issues.map {
      finding(
        $0.severity,
        code: $0.code,
        source: path,
        message: "\($0.path): \($0.detail) Fix: \($0.fix)"
      )
    }
  }

  private func inspectPatternErrors(directory: String) -> [LoomErrorInspectionFinding] {
    let catalog = LoomPatternCatalog()
    let validate = catalog.validate(directory: directory)
    let lint = validate.status == "ok" ? catalog.lint(directory: directory) : validate
    return (validate.issues + lint.issues).map {
      finding($0.severity, code: $0.code, source: directory, message: "\($0.path): \($0.detail)")
    }
  }

  private func finding(_ diagnostic: LoomDiagnostic, source: String) -> LoomErrorInspectionFinding {
    finding(
      diagnostic.severity,
      code: diagnostic.code,
      source: source,
      message: diagnostic.message,
      offset: diagnostic.sourceOffset
    )
  }

  private func finding(
    _ severity: LoomDiagnosticSeverity,
    code: String,
    source: String,
    message: String,
    offset: Int? = nil,
    line: Int? = nil,
    column: Int? = nil
  ) -> LoomErrorInspectionFinding {
    LoomErrorInspectionFinding(
      severity: severity,
      code: code,
      source: source,
      message: message,
      offset: offset,
      line: line,
      column: column
    )
  }

  private func severity(from severity: DiagnosticSeverity) -> LoomDiagnosticSeverity {
    switch severity {
    case .error: return .error
    case .warning: return .warning
    case .note, .remark: return .info
    }
  }
}
