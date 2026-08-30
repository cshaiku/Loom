import Foundation

public enum LoomPatternStatus: String, Codable, Sendable {
  case draft
  case stable
  case deprecated
}

public enum LoomPatternValueType: String, Codable, Sendable, Equatable {
  case string
  case number
  case boolean
  case enumeration = "enum"
  case length
  case insets
  case alignment
  case binding
  case resource
  case reference
}

public struct LoomPatternIntent: Codable, Sendable {
  public var summary: String
  public var useWhen: [String]
  public var avoidWhen: [String]
}

public struct LoomPatternSemantics: Codable, Sendable {
  public var role: String
  public var childPolicy: String
  public var sizing: String
  public var ordering: String
}

public struct LoomPatternAttribute: Codable, Sendable {
  public var name: String
  public var valueType: LoomPatternValueType
  public var required: Bool
  public var description: String
  public var defaultValue: String?
  public var allowedValues: [String]?
  public var minimum: Double?
  public var maximum: Double?
  public var units: [String]?
}

public struct LoomPatternConstraint: Codable, Sendable {
  public var expression: String
  public var message: String
}

public struct LoomPatternAccessibility: Codable, Sendable {
  public var role: String
  public var nameSource: String
  public var focusBehavior: String
  public var notes: [String]
}

public struct LoomPatternMapping: Codable, Sendable {
  public var platform: String
  public var constructs: [String]
  public var strategy: String
  public var caveats: [String]
}

public struct LoomPattern: Codable, Sendable {
  public var schemaVersion: String
  public var id: String
  public var version: String
  public var name: String
  public var kind: LoomNodeKind
  public var status: LoomPatternStatus
  public var category: String
  public var intent: LoomPatternIntent
  public var semantics: LoomPatternSemantics
  public var attributes: [LoomPatternAttribute]
  public var constraints: [LoomPatternConstraint]
  public var accessibility: LoomPatternAccessibility
  public var mappings: [LoomPatternMapping]
  public var tags: [String]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case id, version, name, kind, status, category, intent, semantics, attributes, constraints
    case accessibility, mappings, tags
  }
}

public struct LoomPatternIssue: Codable, Sendable {
  public var severity: LoomDiagnosticSeverity
  public var code: String
  public var path: String
  public var detail: String
}

public struct LoomPatternValidationReport: Codable, Sendable {
  public var schemaVersion = "1"
  public var status: String
  public var directory: String
  public var patternCount: Int
  public var issues: [LoomPatternIssue]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case status, directory, patternCount, issues
  }
}

public struct LoomPatternCatalog: Sendable {
  public init() {}

  public func load(directory: String) throws -> [LoomPattern] {
    let url = URL(fileURLWithPath: directory).standardizedFileURL
    let files: [URL]
    do {
      files = try FileManager.default.contentsOfDirectory(
        at: url, includingPropertiesForKeys: nil
      ).filter { $0.lastPathComponent.hasSuffix(".pattern.json") }.sorted {
        $0.lastPathComponent < $1.lastPathComponent
      }
    } catch {
      throw LoomError.invalidPattern("Could not read pattern directory at \(url.path).")
    }
    return try files.map { file in
      do {
        return try JSONDecoder().decode(LoomPattern.self, from: Data(contentsOf: file))
      } catch {
        throw LoomError.invalidPattern("\(file.lastPathComponent): \(error.localizedDescription)")
      }
    }
  }

  public func find(_ id: String, directory: String) throws -> LoomPattern? {
    try load(directory: directory).first { $0.id == id }
  }

  public func validate(directory: String) -> LoomPatternValidationReport {
    let standardized = URL(fileURLWithPath: directory).standardizedFileURL.path
    let files: [URL]
    do {
      files = try FileManager.default.contentsOfDirectory(
        at: URL(fileURLWithPath: standardized), includingPropertiesForKeys: nil
      ).filter { $0.lastPathComponent.hasSuffix(".pattern.json") }.sorted {
        $0.lastPathComponent < $1.lastPathComponent
      }
    } catch {
      return report(
        directory: standardized, count: 0,
        issues: [
          issue("PATTERN001", standardized, "Pattern directory cannot be read.")
        ])
    }

    var patterns: [(URL, LoomPattern)] = []
    var issues: [LoomPatternIssue] = []
    for file in files {
      do {
        patterns.append(
          (file, try JSONDecoder().decode(LoomPattern.self, from: Data(contentsOf: file))))
      } catch {
        issues.append(issue("PATTERN002", file.lastPathComponent, error.localizedDescription))
      }
    }
    if files.isEmpty {
      issues.append(issue("PATTERN003", standardized, "No .pattern.json files were found."))
    }

    var ids = Set<String>()
    var kinds = Set<LoomNodeKind>()
    for (file, pattern) in patterns {
      let path = file.lastPathComponent
      if pattern.schemaVersion != "1" {
        issues.append(issue("PATTERN004", path, "schema_version must be 1."))
      }
      if pattern.id.isEmpty
        || !pattern.id.allSatisfy({ $0.isLowercase || $0.isNumber || $0 == "-" })
      {
        issues.append(
          issue("PATTERN005", path, "id must use lowercase letters, numbers, and hyphens."))
      }
      if file.deletingPathExtension().deletingPathExtension().lastPathComponent != pattern.id {
        issues.append(issue("PATTERN006", path, "Filename must be \(pattern.id).pattern.json."))
      }
      if !ids.insert(pattern.id).inserted {
        issues.append(issue("PATTERN007", path, "Duplicate pattern id \(pattern.id)."))
      }
      if !kinds.insert(pattern.kind).inserted {
        issues.append(
          issue("PATTERN008", path, "Duplicate semantic kind \(pattern.kind.rawValue)."))
      }
      if pattern.intent.summary.isEmpty || pattern.semantics.role.isEmpty
        || pattern.category.isEmpty
      {
        issues.append(
          issue("PATTERN009", path, "Intent, semantic role, and category must be non-empty."))
      }
      var attributeNames = Set<String>()
      for attribute in pattern.attributes {
        let attributePath = "\(path)#attributes.\(attribute.name)"
        if attribute.name.isEmpty || attribute.description.isEmpty {
          issues.append(
            issue("PATTERN010", attributePath, "Attribute name and description are required."))
        }
        if !attributeNames.insert(attribute.name).inserted {
          issues.append(issue("PATTERN011", attributePath, "Duplicate attribute name."))
        }
        if let minimum = attribute.minimum, let maximum = attribute.maximum, minimum > maximum {
          issues.append(issue("PATTERN012", attributePath, "minimum cannot exceed maximum."))
        }
        if attribute.valueType == .enumeration && (attribute.allowedValues?.isEmpty != false) {
          issues.append(
            issue("PATTERN013", attributePath, "enum attributes require allowedValues."))
        }
        if let value = attribute.defaultValue, let allowed = attribute.allowedValues,
          !allowed.contains(value)
        {
          issues.append(
            issue("PATTERN014", attributePath, "defaultValue must be an allowed value."))
        }
      }
      if pattern.mappings.isEmpty
        || Set(pattern.mappings.map(\.platform)).count != pattern.mappings.count
      {
        issues.append(
          issue("PATTERN015", path, "Mappings must be non-empty and platform names unique."))
      }
    }
    return report(directory: standardized, count: patterns.count, issues: issues)
  }

  public func json<T: Encodable>(_ value: T) throws -> String {
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
    return String(decoding: try encoder.encode(value), as: UTF8.self) + "\n"
  }

  public func text(_ report: LoomPatternValidationReport) -> String {
    var lines = [
      "Loom pattern validation", "Status: \(report.status)",
      "Directory: \(report.directory)", "Patterns: \(report.patternCount)",
      "Issues: \(report.issues.count)",
    ]
    lines += report.issues.map { "  [\($0.severity.rawValue)] \($0.code) \($0.path): \($0.detail)" }
    return lines.joined(separator: "\n") + "\n"
  }

  public func listText(_ patterns: [LoomPattern]) -> String {
    let width = patterns.map(\.id.count).max() ?? 0
    return patterns.map {
      "\($0.id + String(repeating: " ", count: max(0, width - $0.id.count)))  \($0.kind.rawValue)  \($0.intent.summary)"
    }.joined(separator: "\n") + (patterns.isEmpty ? "" : "\n")
  }

  private func issue(_ code: String, _ path: String, _ detail: String) -> LoomPatternIssue {
    LoomPatternIssue(severity: .error, code: code, path: path, detail: detail)
  }

  private func report(directory: String, count: Int, issues: [LoomPatternIssue])
    -> LoomPatternValidationReport
  {
    LoomPatternValidationReport(
      status: issues.isEmpty ? "ok" : "error", directory: directory,
      patternCount: count, issues: issues)
  }
}
