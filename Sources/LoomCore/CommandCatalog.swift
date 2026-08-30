import Foundation

public enum LoomCommandAccess: String, Codable, Sendable {
  case read
  case write
  case conditionalWrite = "conditional-write"

  public var compactMarker: String {
    switch self {
    case .read: "r"
    case .write: "w"
    case .conditionalWrite: "r/w"
    }
  }
}

public struct LoomCommandInfo: Codable, Sendable {
  public var command: String
  public var name: String
  public var description: String
  public var category: String
  public var access: LoomCommandAccess
  public var writeFlags: [String]
  public var aliases: [String]
  public var synopsis: [String]
  public var examples: [String]

  public init(
    command: String,
    name: String,
    description: String,
    category: String,
    access: LoomCommandAccess,
    writeFlags: [String] = [],
    aliases: [String] = [],
    synopsis: [String],
    examples: [String] = []
  ) {
    self.command = command
    self.name = name
    self.description = description
    self.category = category
    self.access = access
    self.writeFlags = writeFlags
    self.aliases = aliases
    self.synopsis = synopsis
    self.examples = examples
  }
}

public enum LoomCommandCatalog {
  public static let commands: [LoomCommandInfo] = [
    LoomCommandInfo(
      command: "inspect:source",
      name: "Inspect SwiftUI Source",
      description: "Extract and report a SwiftUI layout tree",
      category: "inspection",
      access: .conditionalWrite,
      writeFlags: ["--output"],
      aliases: ["analyze"],
      synopsis: [
        "loom inspect:source <swift-file> [--root-view Name] [--component name] [--format text|json]"
      ],
      examples: ["loom inspect:source ContentView.swift --root-view ContentView"]
    ),
    LoomCommandInfo(
      command: "inspect:parity",
      name: "Inspect XAML Parity",
      description: "Compare SwiftUI structure with existing WinUI XAML",
      category: "inspection",
      access: .conditionalWrite,
      writeFlags: ["--output"],
      aliases: ["parity"],
      synopsis: [
        "loom inspect:parity <swift-file> --xaml <xaml-file> [--root-view Name] [--format text|json]"
      ]
    ),
    LoomCommandInfo(
      command: "generate:xaml",
      name: "Generate WinUI XAML",
      description: "Emit a reviewable WinUI 3 XAML fragment",
      category: "generation",
      access: .conditionalWrite,
      writeFlags: ["--output"],
      aliases: ["generate"],
      synopsis: [
        "loom generate:xaml <swift-file> [--root-view Name] [--component name] [--theme-prefix Prefix] [--output path]"
      ]
    ),
    LoomCommandInfo(
      command: "project:build",
      name: "Build Project Translation",
      description: "Generate all manifest-declared analyses, XAML fragments, and parity reports",
      category: "projects",
      access: .write,
      aliases: ["project"],
      synopsis: [
        "loom project:build <loom.json> [--project-root path] [--output-dir path]"
      ]
    ),
    LoomCommandInfo(
      command: "config:validate",
      name: "Validate Project Manifest",
      description: "Validate a Loom manifest and its referenced SwiftUI components",
      category: "setup",
      access: .read,
      synopsis: [
        "loom config:validate <loom.json> [--project-root path] [--format text|json]"
      ]
    ),
    LoomCommandInfo(
      command: "config:schema",
      name: "Project Manifest Schema",
      description: "Print the supported Loom project manifest schema",
      category: "setup",
      access: .read,
      synopsis: ["loom config:schema"]
    ),
  ]

  public static func resolve(_ command: String) -> LoomCommandInfo? {
    commands.first { $0.command == command || $0.aliases.contains(command) }
  }

  public static func catalogText(category: String? = nil) -> String {
    let selected = commands.filter { category == nil || $0.category == category }
    guard !selected.isEmpty else { return "" }
    let grouped = Dictionary(grouping: selected, by: \.category)
    var categories = grouped.keys.sorted()
    if let setupIndex = categories.firstIndex(of: "setup") {
      categories.append(categories.remove(at: setupIndex))
    }
    let commandWidth = selected.map(\.command.count).max() ?? 0
    var lines = [
      "Usage:",
      "  loom <command> [args]",
      "  loom help <command>",
      "  loom list [--category NAME] [--json]",
      "",
    ]
    for category in categories {
      lines.append(category)
      for info in grouped[category, default: []].sorted(by: { $0.command < $1.command }) {
        lines.append(
          "    \(padded(info.access.compactMarker, width: 3)) \(padded(info.command, width: commandWidth))   \(info.description)"
        )
      }
    }
    return lines.joined(separator: "\n") + "\n"
  }

  public static func catalogJSON(category: String? = nil) throws -> String {
    let selected = commands.filter { category == nil || $0.category == category }
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
    return String(decoding: try encoder.encode(selected), as: UTF8.self) + "\n"
  }

  public static func manual(_ command: String) -> String? {
    guard let info = resolve(command) else { return nil }
    var lines = [
      "loom \(info.command)",
      String(repeating: "=", count: info.command.count + 5),
      "",
      "DESCRIPTION",
      "  \(info.description)",
      "",
      "CATEGORY",
      "  \(info.category)",
      "",
      "ACCESS",
      "  \(info.access.rawValue)",
    ]
    if !info.writeFlags.isEmpty {
      lines[lines.count - 1] += " via \(info.writeFlags.joined(separator: ", "))"
    }
    lines.append("")
    lines.append("SYNOPSIS")
    lines.append(contentsOf: info.synopsis.map { "  \($0)" })
    if !info.examples.isEmpty {
      lines.append("")
      lines.append("EXAMPLES")
      lines.append(contentsOf: info.examples.map { "  \($0)" })
    }
    if !info.aliases.isEmpty {
      lines.append("")
      lines.append("ALIASES")
      lines.append("  \(info.aliases.joined(separator: ", "))")
    }
    return lines.joined(separator: "\n") + "\n"
  }

  private static func padded(_ value: String, width: Int) -> String {
    value + String(repeating: " ", count: max(0, width - value.count))
  }
}
