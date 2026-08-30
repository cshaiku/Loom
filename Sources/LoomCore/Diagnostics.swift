import Foundation

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
  public static let version = "0.11.0"

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

  public func json<T: Encodable>(_ value: T) throws -> String {
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
    return String(decoding: try encoder.encode(value), as: UTF8.self) + "\n"
  }

  private func issue(_ code: String, _ message: String) -> LoomDiagnostic {
    LoomDiagnostic(severity: .error, code: code, message: message)
  }
}
