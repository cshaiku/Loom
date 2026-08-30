import Foundation

public enum LoomOSErrorPlatform: String, Codable, Sendable, CaseIterable {
  case swiftui
  case winui3
  case macos
  case windows
  case xaml
}

public struct LoomOSErrorSuggestion: Codable, Equatable, Sendable {
  public var platform: LoomOSErrorPlatform
  public var category: String
  public var matcher: String
  public var issue: String
  public var suggestedFixes: [LoomAuditSuggestedFix]
  public var reference: String

  enum CodingKeys: String, CodingKey {
    case platform, category, matcher, issue, reference
    case suggestedFixes = "suggested_fixes"
  }
}

public struct LoomOSErrorSuggestionReport: Codable, Sendable {
  public var schemaVersion = "1"
  public var status: String
  public var platform: LoomOSErrorPlatform?
  public var query: String?
  public var suggestions: [LoomOSErrorSuggestion]

  enum CodingKeys: String, CodingKey {
    case schemaVersion = "schema_version"
    case status, platform, query, suggestions
  }
}

public struct LoomOSErrorSuggester: Sendable {
  public init() {}

  public func report(
    platform: LoomOSErrorPlatform? = nil,
    query: String? = nil
  ) -> LoomOSErrorSuggestionReport {
    let normalizedQuery = query?.trimmingCharacters(in: .whitespacesAndNewlines)
    let suggestions = Self.catalog.filter { suggestion in
      if let platform, suggestion.platform != platform { return false }
      guard let normalizedQuery, !normalizedQuery.isEmpty else { return true }
      let haystack = [
        suggestion.platform.rawValue,
        suggestion.category,
        suggestion.matcher,
        suggestion.issue,
        suggestion.reference,
      ].joined(separator: " ").lowercased()
      return haystack.contains(normalizedQuery.lowercased())
        || normalizedQuery.lowercased().contains(suggestion.matcher.lowercased())
    }
    return LoomOSErrorSuggestionReport(
      status: suggestions.isEmpty ? "empty" : "ok",
      platform: platform,
      query: normalizedQuery?.isEmpty == false ? normalizedQuery : nil,
      suggestions: suggestions
    )
  }

  public func suggestions(
    for finding: LoomErrorInspectionFinding,
    inspectedKind: LoomErrorInspectionKind
  ) -> [LoomAuditSuggestedFix] {
    var platforms = platformsFor(kind: inspectedKind, code: finding.code)
    if finding.source.lowercased().hasSuffix(".xaml") { platforms.insert(.winui3) }
    let message = "\(finding.code) \(finding.message)".lowercased()
    let matches = Self.catalog.filter { suggestion in
      platforms.contains(suggestion.platform)
        && (
          message.contains(suggestion.matcher.lowercased())
            || suggestion.matcher.lowercased().contains(finding.code.lowercased())
        )
    }
    let fallback = fallbackSuggestion(for: finding, inspectedKind: inspectedKind)
    return uniqued((matches.flatMap(\.suggestedFixes)) + fallback)
  }

  public func text(_ report: LoomOSErrorSuggestionReport) -> String {
    var lines = [
      "Loom OS error suggestions",
      "Status: \(report.status)",
      "Platform: \(report.platform?.rawValue ?? "all")",
      "Query: \(report.query ?? "none")",
      "Suggestions: \(report.suggestions.count)",
      "",
    ]
    if report.suggestions.isEmpty {
      lines.append("  none")
    } else {
      for suggestion in report.suggestions {
        lines.append("[\(suggestion.platform.rawValue)] \(suggestion.category): \(suggestion.matcher)")
        lines.append("  issue: \(suggestion.issue)")
        lines.append("  reference: \(suggestion.reference)")
        lines.append("  fixes:")
        for fix in suggestion.suggestedFixes {
          lines.append("    - \(fix.audience.rawValue): \(fix.action) — \(fix.detail)")
          if let command = fix.command {
            lines.append("      command: \(command)")
          }
        }
      }
    }
    return lines.joined(separator: "\n") + "\n"
  }

  public func json<T: Encodable>(_ value: T) throws -> String {
    let encoder = JSONEncoder()
    encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
    return String(decoding: try encoder.encode(value), as: UTF8.self) + "\n"
  }

  private func fallbackSuggestion(
    for finding: LoomErrorInspectionFinding,
    inspectedKind: LoomErrorInspectionKind
  ) -> [LoomAuditSuggestedFix] {
    switch inspectedKind {
    case .swift:
      return [
        fix(.agent, "Localize Swift syntax issue", "Open the reported line/column, reduce the enclosing SwiftUI result-builder expression, and rerun inspect:errors."),
        fix(.agent, "Run Loom source inspection after syntax repair", "Confirm Loom can extract the selected View/component.", command: "loom inspect:source <source> --root-view <RootView> --json")
      ]
    case .xaml:
      return [
        fix(.agent, "Validate XAML structure", "Check element nesting, namespace declarations, property-element syntax, and resource references, then rerun inspect:xaml."),
        fix(.agent, "Inspect native component boundaries", "Preserve unknown WinUI controls as native boundaries or add Pattern mappings.", command: "loom accessibility:audit <source.xaml> --format json")
      ]
    case .manifest:
      return [
        fix(.agent, "Validate manifest inputs", "Check schema_version, source path, target, rootView, and component names.", command: "loom config:validate <loom.json> --format json")
      ]
    case .patterns:
      return [
        fix(.agent, "Validate Pattern catalog", "Fix schema, id, mapping, range, accessibility, and lint errors.", command: "loom patterns:validate --json && loom patterns:lint --json")
      ]
    }
  }

  private func platformsFor(
    kind: LoomErrorInspectionKind,
    code: String
  ) -> Set<LoomOSErrorPlatform> {
    switch kind {
    case .swift:
      return [.swiftui, .macos]
    case .xaml:
      return [.xaml, .winui3, .windows]
    case .manifest, .patterns:
      if code.lowercased().contains("xaml") { return [.xaml, .winui3, .windows] }
      return [.swiftui, .winui3, .macos, .windows]
    }
  }

  private func uniqued(_ fixes: [LoomAuditSuggestedFix]) -> [LoomAuditSuggestedFix] {
    var seen = Set<String>()
    var result: [LoomAuditSuggestedFix] = []
    for fix in fixes {
      let key = "\(fix.audience.rawValue)|\(fix.action)|\(fix.detail)|\(fix.command ?? "")"
      guard seen.insert(key).inserted else { continue }
      result.append(fix)
    }
    return result
  }

  private static func fix(
    _ audience: LoomAuditFixAudience,
    _ action: String,
    _ detail: String,
    command: String? = nil
  ) -> LoomAuditSuggestedFix {
    LoomAuditSuggestedFix(
      audience: audience,
      action: action,
      detail: detail,
      command: command
    )
  }

  private func fix(
    _ audience: LoomAuditFixAudience,
    _ action: String,
    _ detail: String,
    command: String? = nil
  ) -> LoomAuditSuggestedFix {
    Self.fix(audience, action, detail, command: command)
  }

  public static let catalog: [LoomOSErrorSuggestion] = [
    LoomOSErrorSuggestion(
      platform: .swiftui,
      category: "syntax",
      matcher: "SWIFT.PARSE",
      issue: "Swift parser diagnostics usually mean the SwiftUI result-builder body cannot be reliably extracted.",
      suggestedFixes: [
        fix(.user, "Fix the Swift syntax first", "Do not attempt UI transfer until the source parses cleanly."),
        fix(.agent, "Reduce the failing expression", "Check bracket/brace balance, string literals, trailing closures, and modifier chains around the reported line/column.", command: "loom inspect:errors <source.swift> --kind swift --json")
      ],
      reference: "Apple Swift language diagnostics and SwiftSyntax parser diagnostics"
    ),
    LoomOSErrorSuggestion(
      platform: .swiftui,
      category: "result-builder",
      matcher: "result builder",
      issue: "SwiftUI ViewBuilder errors often come from non-View statements, mismatched branch return types, or unsupported control flow in a view body.",
      suggestedFixes: [
        fix(.user, "Clarify intended visual branch", "Confirm what should be shown for each condition or loop branch."),
        fix(.agent, "Extract complex branches", "Move complex conditional/loop content into named computed View properties and rerun Loom.", command: "loom graph:components <source-root> --root-view <RootView>")
      ],
      reference: "SwiftUI View and result-builder behavior"
    ),
    LoomOSErrorSuggestion(
      platform: .swiftui,
      category: "type-checking",
      matcher: "unable to type-check",
      issue: "Large SwiftUI expressions can exceed compiler type-checking limits and are also hard for layout transfer tools to reason about.",
      suggestedFixes: [
        fix(.user, "Approve component extraction", "Allow large visual sections to become named subcomponents."),
        fix(.agent, "Split the view body", "Extract subtrees into computed properties or child Views, then rerun component graph inspection.", command: "loom graph:components <source-root> --root-view <RootView>")
      ],
      reference: "Swift compiler diagnostics for complex expressions"
    ),
    LoomOSErrorSuggestion(
      platform: .swiftui,
      category: "accessibility",
      matcher: "accessibilityLabel",
      issue: "Icon-only or non-text SwiftUI controls need explicit labels to expose useful accessibility names.",
      suggestedFixes: [
        fix(.user, "Name the control", "Provide the phrase users should hear for the control."),
        fix(.agent, "Add SwiftUI accessibility label", "Use .accessibilityLabel for non-text controls and rerun accessibility audit.", command: "loom accessibility:audit <source.swift> --root-view <RootView> --json")
      ],
      reference: "Apple SwiftUI accessibilityLabel documentation"
    ),
    LoomOSErrorSuggestion(
      platform: .swiftui,
      category: "accessibility",
      matcher: "accessibilityHidden",
      issue: "Decorative SwiftUI views should be hidden from accessibility rather than exposed as low-value nodes.",
      suggestedFixes: [
        fix(.user, "Classify decorative content", "Decide whether the element conveys meaning or is visual decoration only."),
        fix(.agent, "Hide decorative content", "Use .accessibilityHidden(true) only when equivalent meaning exists elsewhere.")
      ],
      reference: "Apple SwiftUI accessibility(hidden:) documentation"
    ),
    LoomOSErrorSuggestion(
      platform: .xaml,
      category: "parse",
      matcher: "LOOM.XAML",
      issue: "XAML parse/load failures usually come from malformed XML, invalid property-element nesting, namespace issues, or unresolved resources.",
      suggestedFixes: [
        fix(.user, "Confirm intended XAML structure", "Review whether the element belongs in markup or native code-behind."),
        fix(.agent, "Check XML and XAML syntax", "Validate matching tags, namespace declarations, property elements, and attribute quoting.", command: "loom inspect:errors <source.xaml> --kind xaml --json")
      ],
      reference: "Microsoft WinUI/XAML overview: XamlParseException line/position context"
    ),
    LoomOSErrorSuggestion(
      platform: .winui3,
      category: "native-boundary",
      matcher: "XAML.UNSUPPORTED_COMPONENT_BOUNDARY",
      issue: "A native WinUI control has no Loom semantic mapping and was preserved as a component boundary.",
      suggestedFixes: [
        fix(.user, "Choose native-boundary strategy", "Keep the control native, replace it with portable layout, or approve a new Pattern mapping."),
        fix(.agent, "Document the boundary contract", "Keep the control as handwritten WinUI and record its source/target expectations.", command: "loom patterns:transfer <source.xaml> --from winui3 --to swiftui --format json"),
        fix(.agent, "Add Pattern support if needed", "Create or extend a Pattern mapping for this WinUI control before attempting transfer.")
      ],
      reference: "Loom native WinUI component-boundary policy"
    ),
    LoomOSErrorSuggestion(
      platform: .winui3,
      category: "resources",
      matcher: "StaticResource",
      issue: "Unresolved StaticResource or ThemeResource keys can throw XAML parse/runtime exceptions.",
      suggestedFixes: [
        fix(.user, "Confirm token/resource ownership", "Identify whether the resource should come from app theme, platform theme, or generated Loom tokens."),
        fix(.agent, "Check resource dictionaries", "Verify the key exists in local, merged, app, or theme dictionaries and is loaded before use.")
      ],
      reference: "Microsoft StaticResource markup extension and XAML resource dictionary documentation"
    ),
    LoomOSErrorSuggestion(
      platform: .winui3,
      category: "accessibility",
      matcher: "AutomationProperties.Name",
      issue: "WinUI controls need stable UI Automation names that usually match visible label text.",
      suggestedFixes: [
        fix(.user, "Approve accessible name", "Provide the visible/user-facing name that assistive technology should announce."),
        fix(.agent, "Set AutomationProperties.Name", "Use a localized value consistent with the visible label and rerun accessibility audit.", command: "loom accessibility:audit <source.xaml> --format json")
      ],
      reference: "Microsoft AutomationProperties.Name guidance"
    ),
    LoomOSErrorSuggestion(
      platform: .winui3,
      category: "accessibility",
      matcher: "AccessibilityView",
      issue: "Composed WinUI UIs can expose duplicate or low-value UIA nodes unless accessibility tree visibility is controlled.",
      suggestedFixes: [
        fix(.user, "Decide tree exposure", "Confirm whether the element should be in Control, Content, or Raw view."),
        fix(.agent, "Set accessibility tree visibility", "Use AutomationProperties.AccessibilityView to remove decorative or duplicate elements from high-level UIA views.")
      ],
      reference: "Microsoft basic accessibility information for Windows apps"
    ),
    LoomOSErrorSuggestion(
      platform: .winui3,
      category: "custom-control",
      matcher: "AutomationPeer",
      issue: "Custom WinUI controls may require custom automation peers for robust UI Automation support.",
      suggestedFixes: [
        fix(.user, "Confirm custom control semantics", "Define role, name, state, and supported interaction patterns."),
        fix(.agent, "Add custom automation peer", "Implement or verify an AutomationPeer for custom native controls that cannot be represented by stock WinUI automation.")
      ],
      reference: "Microsoft custom automation peers for WinUI apps"
    ),
    LoomOSErrorSuggestion(
      platform: .windows,
      category: "binding",
      matcher: "BindingExpression",
      issue: "Windows binding failures usually indicate a missing property, wrong data context, type mismatch, or notification gap.",
      suggestedFixes: [
        fix(.user, "Confirm data contract", "Identify the source property, value type, and update direction."),
        fix(.agent, "Generate target contracts", "Use Loom target contracts to list required bindings before editing XAML/code-behind.", command: "loom generate:contracts <source.swift> --root-view <RootView> --format json")
      ],
      reference: "WinUI binding and UI Automation implementation guidance"
    ),
  ]
}
