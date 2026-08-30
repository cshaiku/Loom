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

public struct LoomPatternKeyboard: Codable, Sendable {
  public var activation: [String]?
  public var navigation: [String]?
  public var escapeBehavior: String?

  enum CodingKeys: String, CodingKey {
    case activation, navigation
    case escapeBehavior = "escape_behavior"
  }
}

public struct LoomPatternMinimumTargetSize: Codable, Sendable {
  public var width: Double
  public var height: Double
  public var unit: String
}

public struct LoomPatternAccessibility: Codable, Sendable {
  public var role: String
  public var nameSource: String
  public var focusBehavior: String
  public var notes: [String]
  public var keyboard: LoomPatternKeyboard?
  public var states: [String]?
  public var properties: [String]?
  public var minimumTargetSize: LoomPatternMinimumTargetSize?

  enum CodingKeys: String, CodingKey {
    case role, notes, keyboard, states, properties
    case nameSource = "nameSource"
    case focusBehavior = "focusBehavior"
    case minimumTargetSize = "minimumTargetSize"
  }
}

public struct LoomPatternMapping: Codable, Sendable {
  public var platform: String
  public var constructs: [String]
  public var strategy: String
  public var caveats: [String]
}

public struct LoomPatternVariant: Codable, Sendable {
  public var name: String
  public var conditions: [String]
  public var layoutPolicy: String
  public var intent: String

  enum CodingKeys: String, CodingKey {
    case name, conditions, intent
    case layoutPolicy = "layout_policy"
  }
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
  public var variants: [LoomPatternVariant]?
  public var tags: [String]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case id, version, name, kind, status, category, intent, semantics, attributes, constraints
    case accessibility, mappings, variants, tags
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

public enum LoomPatternExportFormat: String, CaseIterable, Sendable {
  case loom
  case dtcg
  case openUI = "open-ui"
  case aria
  case styleDictionary = "style-dictionary"
}

public struct LoomPatternRegistry: Sendable {
  private let patternsByKind: [LoomNodeKind: LoomPattern]

  public init(patterns: [LoomPattern]) {
    self.patternsByKind = Dictionary(uniqueKeysWithValues: patterns.map { ($0.kind, $0) })
  }

  public init(directory: String) throws {
    self.init(patterns: try LoomPatternCatalog().load(directory: directory))
  }

  public func pattern(for kind: LoomNodeKind) -> LoomPattern? {
    patternsByKind[kind]
  }

  public func mapping(for kind: LoomNodeKind, platform: String) -> LoomPatternMapping? {
    pattern(for: kind)?.mappings.first { $0.platform == platform }
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
      if let minimumTargetSize = pattern.accessibility.minimumTargetSize,
        minimumTargetSize.width <= 0 || minimumTargetSize.height <= 0
          || minimumTargetSize.unit.isEmpty
      {
        issues.append(
          issue(
            "PATTERN018", "\(path)#accessibility.minimumTargetSize",
            "Minimum target size width, height, and unit must be positive/non-empty."))
      }
      if let keyboard = pattern.accessibility.keyboard,
        keyboard.activation?.contains(where: { $0.isEmpty }) == true
          || keyboard.navigation?.contains(where: { $0.isEmpty }) == true
          || keyboard.escapeBehavior == ""
      {
        issues.append(
          issue(
            "PATTERN019", "\(path)#accessibility.keyboard",
            "Keyboard accessibility entries cannot be empty."))
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
      for variant in pattern.variants ?? [] {
        let variantPath = "\(path)#variants.\(variant.name)"
        if variant.name.isEmpty || variant.layoutPolicy.isEmpty || variant.intent.isEmpty {
          issues.append(
            issue("PATTERN016", variantPath, "Variant name, layout_policy, and intent are required."))
        }
        if variant.conditions.contains(where: { $0.isEmpty }) {
          issues.append(issue("PATTERN017", variantPath, "Variant conditions cannot be empty."))
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

  public func lint(directory: String) -> LoomPatternValidationReport {
    var report = validate(directory: directory)
    guard report.status == "ok" else { return report }
    let patterns: [LoomPattern]
    do {
      patterns = try load(directory: directory)
    } catch {
      return validate(directory: directory)
    }

    var issues = report.issues
    for pattern in patterns {
      let path = "\(pattern.id).pattern.json"
      if pattern.status != .stable {
        issues.append(issue("PATTERN101", path, "Operational patterns must be stable."))
      }
      let platforms = Set(pattern.mappings.map(\.platform))
      if !platforms.contains("swiftui") {
        issues.append(issue("PATTERN102", path, "Pattern must include a swiftui mapping."))
      }
      if !platforms.contains("winui3") {
        issues.append(issue("PATTERN103", path, "Pattern must include a winui3 mapping."))
      }
      if pattern.tags.isEmpty {
        issues.append(issue("PATTERN104", path, "Pattern must include at least one tag."))
      }
      for mapping in pattern.mappings {
        let mappingPath = "\(path)#mappings.\(mapping.platform)"
        if mapping.constructs.isEmpty || mapping.strategy.isEmpty {
          issues.append(
            issue("PATTERN105", mappingPath, "Mapping constructs and strategy are required."))
        }
      }
      for attribute in pattern.attributes where attribute.required {
        let attributePath = "\(path)#attributes.\(attribute.name)"
        if attribute.defaultValue != nil {
          issues.append(
            issue("PATTERN106", attributePath, "Required attributes must not declare defaults."))
        }
      }
    }
    report.issues = issues
    report.status = issues.isEmpty ? "ok" : "error"
    return report
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

  public func export(_ patterns: [LoomPattern], format: LoomPatternExportFormat) throws -> String {
    switch format {
    case .loom:
      return try json(patterns)
    case .dtcg:
      return try json(DTCGPatternExport(patterns: patterns))
    case .openUI:
      return try json(OpenUIPatternExport(patterns: patterns))
    case .aria:
      return try json(ARIAPatternExport(patterns: patterns))
    case .styleDictionary:
      return try json(StyleDictionaryPatternExport(patterns: patterns))
    }
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

private struct DTCGPatternExport: Codable {
  var schemaVersion = "1"
  var description = "Loom OS-agnostic UI Pattern catalog exported as Design Tokens Community Group-compatible token objects."
  var loom: [String: DTCGPatternToken]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case description = "$description"
    case loom
  }

  init(patterns: [LoomPattern]) {
    self.loom = Dictionary(
      uniqueKeysWithValues: patterns.map {
        (
          $0.id,
          DTCGPatternToken(
            value: DTCGPatternValue(pattern: $0),
            description: $0.intent.summary,
            extensions: DTCGPatternExtensions(loom: $0)
          )
        )
      })
  }
}

private struct DTCGPatternToken: Codable {
  var type = "other"
  var value: DTCGPatternValue
  var description: String
  var extensions: DTCGPatternExtensions

  enum CodingKeys: String, CodingKey {
    case type = "$type"
    case value = "$value"
    case description = "$description"
    case extensions = "$extensions"
  }
}

private struct DTCGPatternValue: Codable {
  var id: String
  var name: String
  var kind: String
  var category: String
  var status: String
  var attributes: [DTCGPatternAttribute]
  var platforms: [String]

  init(pattern: LoomPattern) {
    id = pattern.id
    name = pattern.name
    kind = pattern.kind.rawValue
    category = pattern.category
    status = pattern.status.rawValue
    attributes = pattern.attributes.map(DTCGPatternAttribute.init(attribute:))
    platforms = pattern.mappings.map(\.platform).sorted()
  }
}

private struct DTCGPatternAttribute: Codable {
  var name: String
  var type: String
  var required: Bool
  var defaultValue: String?
  var allowedValues: [String]?
  var minimum: Double?
  var maximum: Double?
  var units: [String]?

  init(attribute: LoomPatternAttribute) {
    name = attribute.name
    type = attribute.valueType.rawValue
    required = attribute.required
    defaultValue = attribute.defaultValue
    allowedValues = attribute.allowedValues
    minimum = attribute.minimum
    maximum = attribute.maximum
    units = attribute.units
  }
}

private struct DTCGPatternExtensions: Codable {
  var loom: LoomPattern
}

private struct OpenUIPatternExport: Codable {
  var schemaVersion = "1"
  var source = "loom"
  var components: [OpenUIComponent]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case source, components
  }

  init(patterns: [LoomPattern]) {
    components = patterns.map(OpenUIComponent.init(pattern:))
  }
}

private struct OpenUIComponent: Codable {
  var id: String
  var name: String
  var category: String
  var status: String
  var intent: LoomPatternIntent
  var semantics: LoomPatternSemantics
  var attributes: [LoomPatternAttribute]
  var mappings: [LoomPatternMapping]
  var tags: [String]

  init(pattern: LoomPattern) {
    id = pattern.id
    name = pattern.name
    category = pattern.category
    status = pattern.status.rawValue
    intent = pattern.intent
    semantics = pattern.semantics
    attributes = pattern.attributes
    mappings = pattern.mappings
    tags = pattern.tags
  }
}

private struct ARIAPatternExport: Codable {
  var schemaVersion = "1"
  var source = "loom"
  var patterns: [ARIAPattern]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case source, patterns
  }

  init(patterns: [LoomPattern]) {
    self.patterns = patterns.map(ARIAPattern.init(pattern:))
  }
}

private struct ARIAPattern: Codable {
  var id: String
  var name: String
  var role: String
  var nameSource: String
  var focusBehavior: String
  var notes: [String]
  var useWhen: [String]
  var avoidWhen: [String]

  init(pattern: LoomPattern) {
    id = pattern.id
    name = pattern.name
    role = pattern.accessibility.role
    nameSource = pattern.accessibility.nameSource
    focusBehavior = pattern.accessibility.focusBehavior
    notes = pattern.accessibility.notes
    useWhen = pattern.intent.useWhen
    avoidWhen = pattern.intent.avoidWhen
  }
}

private struct StyleDictionaryPatternExport: Codable {
  var loom: [String: StyleDictionaryPatternToken]

  init(patterns: [LoomPattern]) {
    loom = Dictionary(
      uniqueKeysWithValues: patterns.map { ($0.id, StyleDictionaryPatternToken(pattern: $0)) })
  }
}

private struct StyleDictionaryPatternToken: Codable {
  var value: String
  var type = "loom.pattern"
  var comment: String
  var attributes: StyleDictionaryPatternAttributes

  init(pattern: LoomPattern) {
    value = pattern.id
    comment = pattern.intent.summary
    attributes = StyleDictionaryPatternAttributes(pattern: pattern)
  }
}

private struct StyleDictionaryPatternAttributes: Codable {
  var name: String
  var kind: String
  var category: String
  var status: String
  var role: String
  var childPolicy: String
  var sizing: String
  var ordering: String
  var tags: [String]
  var platforms: [String]

  init(pattern: LoomPattern) {
    name = pattern.name
    kind = pattern.kind.rawValue
    category = pattern.category
    status = pattern.status.rawValue
    role = pattern.semantics.role
    childPolicy = pattern.semantics.childPolicy
    sizing = pattern.semantics.sizing
    ordering = pattern.semantics.ordering
    tags = pattern.tags
    platforms = pattern.mappings.map(\.platform).sorted()
  }
}
