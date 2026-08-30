import Foundation
import LoomCore

#if os(Linux)
  import Glibc
#elseif os(Windows)
  import CRT
#else
  import Darwin
#endif

private struct SourceOptions {
  var command: String
  var sourcePath: String
  var rootView = "ContentView"
  var component = "body"
  var format = "text"
  var outputPath: String?
  var xamlPath: String?
  var themeResourcePrefix: String?
  var patternsDirectory: String?
  var includePatternComments = false
  var replaceRegionPath: String?
  var regionID: String?
  var initRegion = false
}

private struct GraphOptions {
  var sourceRoot: String
  var rootView = "ContentView"
  var component = "body"
  var format = "text"
  var outputPath: String?
  var include: [String] = []
  var exclude: [String] = []
}

private struct SwiftUIOptions {
  var xamlPath: String
  var viewName = "GeneratedView"
  var outputPath: String?
}

private struct ProjectOptions {
  var manifestPath: String
  var projectRoot: String?
  var outputDirectory: String?
}

private struct ValidationOptions {
  var manifestPath: String
  var projectRoot: String?
  var format = "text"
}

private struct PatternOptions {
  var directory = "Patterns"
  var json = false
  var format = LoomPatternExportFormat.loom
  var outputPath: String?
  var positional: [String] = []
}

private struct PatternTransferCommandOptions {
  var sourcePath: String
  var from: LoomTransferPlatform?
  var to: LoomTransferPlatform?
  var rootView = "ContentView"
  var component = "body"
  var patternsDirectory = "Patterns"
  var format = "text"
  var outputPath: String?
}

private struct RuntimeOptions {
  var quiet = false
  var verbose = false
}

private struct ErrorInspectionOptions {
  var path: String
  var kind: LoomErrorInspectionKind?
  var rootView = "ContentView"
  var component = "body"
  var format = "text"
  var failOn = LoomErrorFailMode.none
  var outputPath: String?
}

private struct AccessibilityAuditOptions {
  var path: String
  var rootView = "ContentView"
  var component = "body"
  var format = "text"
  var failOn = LoomErrorFailMode.none
  var outputPath: String?
}

@main
private enum LoomCommand {
  static func main() {
    var runtime = RuntimeOptions()
    do {
      var arguments = Array(CommandLine.arguments.dropFirst())
      runtime = parseRuntimeOptions(&arguments)
      try dispatch(arguments, runtime: runtime)
    } catch {
      fputs("[fatal] \(error)\n", stderr)
      fputs("[hint] Run `loom help` or `loom list`.\n", stderr)
      exit(2)
    }
  }

  private static func parseRuntimeOptions(_ arguments: inout [String]) -> RuntimeOptions {
    var options = RuntimeOptions()
    arguments.removeAll { argument in
      switch argument {
      case "--quiet", "-q":
        options.quiet = true
        return true
      case "--verbose", "-v":
        options.verbose = true
        return true
      default:
        return false
      }
    }
    if options.quiet { options.verbose = false }
    return options
  }

  private static func dispatch(_ arguments: [String], runtime: RuntimeOptions) throws {
    if arguments == ["--version"] || arguments == ["version"] {
      print("loom 0.15.0")
      return
    }
    if arguments.isEmpty || arguments == ["help"] || arguments == ["--help"] || arguments == ["-h"]
    {
      print("Loom: SwiftUI to WinUI Layout Compiler\n")
      print(LoomCommandCatalog.catalogText(), terminator: "")
      return
    }
    if arguments[0] == "list" || arguments[0] == "commands" {
      try printCatalog(Array(arguments.dropFirst()))
      return
    }
    if arguments[0] == "help" || arguments[0] == "man" || arguments[0] == "explain" {
      guard arguments.count == 2, let manual = LoomCommandCatalog.manual(arguments[1]) else {
        throw LoomError.invalidArguments("Unknown command manual request.")
      }
      print(manual, terminator: "")
      return
    }
    if arguments[0] == "config:schema" {
      guard arguments.count == 1 else {
        throw LoomError.invalidArguments("Usage: loom config:schema")
      }
      print(LoomProjectSchema.json)
      return
    }

    if arguments[0].hasPrefix("patterns:") {
      try runPatternCommand(arguments, runtime: runtime)
      return
    }

    guard let command = LoomCommandCatalog.resolve(arguments[0]) else {
      throw LoomError.invalidArguments("Unknown command \(arguments[0]).")
    }
    if arguments.count > 1 && (arguments[1] == "--help" || arguments[1] == "-h") {
      print(LoomCommandCatalog.manual(command.command) ?? "", terminator: "")
      return
    }

    switch command.command {
    case "accessibility:audit":
      try runAccessibilityAuditCommand(arguments, runtime: runtime)
    case "inspect:ascii":
      try runASCIICommand(arguments, runtime: runtime)
    case "inspect:errors":
      try runErrorInspectionCommand(arguments, runtime: runtime)
    case "inspect:source", "inspect:parity", "generate:xaml", "generate:contracts":
      try runSourceCommand(command.command, arguments: arguments, runtime: runtime)
    case "inspect:xaml":
      try runXAMLCommand(arguments, runtime: runtime)
    case "generate:swiftui":
      try runSwiftUICommand(arguments, runtime: runtime)
    case "graph:components":
      try runGraphCommand(arguments, runtime: runtime)
    case "project:build":
      let options = try parseProjectOptions(arguments)
      let run = try LoomProjectRunner().run(
        manifestPath: options.manifestPath,
        projectRoot: options.projectRoot,
        outputDirectory: options.outputDirectory
      )
      print(LoomProjectRunner().text(run), terminator: "")
    case "status", "verify", "checks:command-catalog", "guards:summary", "self-heal:plan":
      try runDiagnosticCommand(command.command, arguments: arguments)
    case "config:validate":
      let options = try parseValidationOptions(arguments)
      let validator = LoomProjectValidator()
      let report = validator.validate(
        manifestPath: options.manifestPath,
        projectRoot: options.projectRoot
      )
      let output = options.format == "json" ? try validator.json(report) : validator.text(report)
      print(output, terminator: "")
      if report.status != "ok" { exit(1) }
    default:
      throw LoomError.invalidArguments(
        "Command \(command.command) is registered but has no dispatcher.")
    }
  }

  private static func runSourceCommand(
    _ command: String,
    arguments: [String],
    runtime: RuntimeOptions
  ) throws {
    let options = try parseSourceOptions(command, arguments: arguments)
    let analysis = try SwiftUIFrontend().analyze(
      sourcePath: options.sourcePath,
      rootView: options.rootView,
      component: options.component
    )

    let output: String
    switch command {
    case "inspect:source":
      output =
        options.format == "json"
        ? try AnalysisReporter().json(analysis)
        : AnalysisReporter().text(analysis)
    case "generate:xaml":
      let registry = try options.patternsDirectory.map { try LoomPatternRegistry(directory: $0) }
      output = XAMLEmitter(
        options: XAMLEmissionOptions(
          themeResourcePrefix: options.themeResourcePrefix,
          includePatternComments: options.includePatternComments
        ),
        patternRegistry: registry
      ).emit(analysis)
    case "inspect:parity":
      guard let xamlPath = options.xamlPath else {
        throw LoomError.invalidArguments("inspect:parity requires --xaml <path>.")
      }
      let report = try XAMLParityChecker().check(analysis: analysis, xamlPath: xamlPath)
      if options.format == "json" {
        let encoder = JSONEncoder()
        encoder.outputFormatting = [.prettyPrinted, .sortedKeys, .withoutEscapingSlashes]
        output = String(decoding: try encoder.encode(report), as: UTF8.self) + "\n"
      } else {
        output = XAMLParityChecker().text(report)
      }
    case "generate:contracts":
      let generator = LoomTargetContractGenerator()
      let report = try generator.generate(
        analysis: analysis,
        options: LoomTargetContractOptions(themeResourcePrefix: options.themeResourcePrefix)
      )
      output = options.format == "json" ? try generator.json(report) : generator.text(report)
    default:
      throw LoomError.invalidArguments("Unknown source command \(command).")
    }
    if let replaceRegionPath = options.replaceRegionPath {
      guard command == "generate:xaml" else {
        throw LoomError.invalidArguments("--replace-region is only supported by generate:xaml.")
      }
      guard options.outputPath == nil else {
        throw LoomError.invalidArguments("--output cannot be combined with --replace-region.")
      }
      guard let regionID = options.regionID else {
        throw LoomError.invalidArguments("--replace-region requires --region-id <id>.")
      }
      let update = try LoomOwnedRegionUpdater().replaceXAMLRegion(
        path: replaceRegionPath,
        regionID: regionID,
        content: output,
        createIfMissing: options.initRegion
      )
      if runtime.verbose {
        fputs(
          "[info] region \(update.regionID) in \(update.path): \(update.changed ? "updated" : "unchanged")\n",
          stderr
        )
      }
      if !runtime.quiet {
        print(update.changed ? "Updated \(update.path)" : "No changes for \(update.path)")
      }
    } else {
      try writeOrPrint(output, path: options.outputPath, runtime: runtime)
    }
  }

  private static func runXAMLCommand(_ arguments: [String], runtime: RuntimeOptions) throws {
    let options = try parseXAMLOptions(arguments)
    let analysis = try XAMLFrontend().analyze(sourcePath: options.sourcePath)
    let output =
      options.format == "json"
      ? try AnalysisReporter().json(analysis)
      : AnalysisReporter().text(analysis)
    try writeOrPrint(output, path: options.outputPath, runtime: runtime)
  }

  private static func runErrorInspectionCommand(_ arguments: [String], runtime: RuntimeOptions)
    throws
  {
    let options = try parseErrorInspectionOptions(arguments)
    let diagnostics = LoomDiagnostics()
    let report = diagnostics.inspectErrors(
      path: options.path,
      options: LoomErrorInspectionOptions(
        kind: options.kind,
        rootView: options.rootView,
        component: options.component
      )
    )
    let output = options.format == "json" ? try diagnostics.json(report) : diagnostics.text(report)
    try writeOrPrint(output, path: options.outputPath, runtime: runtime)
    if diagnostics.shouldFail(report, mode: options.failOn) { exit(1) }
  }

  private static func runAccessibilityAuditCommand(_ arguments: [String], runtime: RuntimeOptions)
    throws
  {
    let options = try parseAccessibilityAuditOptions(arguments)
    let analysis: LoomAnalysis
    if options.path.lowercased().hasSuffix(".xaml") {
      analysis = try XAMLFrontend().analyze(sourcePath: options.path)
    } else {
      analysis = try SwiftUIFrontend().analyze(
        sourcePath: options.path,
        rootView: options.rootView,
        component: options.component
      )
    }
    let auditor = LoomAccessibilityAuditor()
    let report = auditor.audit(analysis)
    let output = options.format == "json" ? try auditor.json(report) : auditor.text(report)
    try writeOrPrint(output, path: options.outputPath, runtime: runtime)
    if auditor.shouldFail(report, mode: options.failOn) { exit(1) }
  }

  private static func runSwiftUICommand(_ arguments: [String], runtime: RuntimeOptions) throws {
    let options = try parseSwiftUIOptions(arguments)
    let analysis = try XAMLFrontend().analyze(sourcePath: options.xamlPath)
    let output = SwiftUIEmitter(options: .init(viewName: options.viewName)).emit(analysis)
    try writeOrPrint(output, path: options.outputPath, runtime: runtime)
  }

  private static func runGraphCommand(_ arguments: [String], runtime: RuntimeOptions) throws {
    let options = try parseGraphOptions(arguments)
    let graph = try LoomComponentGraphBuilder().build(
      sourceRoot: options.sourceRoot,
      rootView: options.rootView,
      component: options.component,
      options: LoomComponentGraphOptions(
        include: options.include.isEmpty ? ["*.swift", "**/*.swift"] : options.include,
        exclude: options.exclude
      )
    )
    let output: String
    switch options.format {
    case "text": output = LoomComponentGraphBuilder().text(graph)
    case "json": output = try LoomComponentGraphBuilder().json(graph)
    case "dot": output = LoomComponentGraphBuilder().dot(graph)
    default: throw LoomError.invalidArguments("--format must be text, json, or dot.")
    }
    try writeOrPrint(output, path: options.outputPath, runtime: runtime)
    if graph.status != "ok" { exit(1) }
  }

  private static func printCatalog(_ arguments: [String]) throws {
    var category: String?
    var json = false
    var index = 0
    while index < arguments.count {
      let argument = arguments[index]
      switch argument {
      case "--json":
        json = true
        index += 1
      case "--category":
        guard index + 1 < arguments.count else {
          throw LoomError.invalidArguments("Missing value for --category.")
        }
        category = arguments[index + 1]
        index += 2
      default:
        if argument.hasPrefix("--category=") {
          category = String(argument.dropFirst("--category=".count))
          index += 1
        } else {
          throw LoomError.invalidArguments("Unknown list option \(argument).")
        }
      }
    }
    let output =
      json
      ? try LoomCommandCatalog.catalogJSON(category: category)
      : LoomCommandCatalog.catalogText(category: category)
    guard !output.isEmpty else {
      throw LoomError.invalidArguments("Unknown or empty command category \(category ?? "").")
    }
    print(output, terminator: "")
  }

  private static func runPatternCommand(_ arguments: [String], runtime: RuntimeOptions) throws {
    guard let command = LoomCommandCatalog.resolve(arguments[0]) else {
      throw LoomError.invalidArguments("Unknown pattern command \(arguments[0]).")
    }
    if arguments.count > 1 && (arguments[1] == "--help" || arguments[1] == "-h") {
      print(LoomCommandCatalog.manual(command.command) ?? "", terminator: "")
      return
    }
    if command.command == "patterns:transfer" {
      let transferOptions = try parsePatternTransferOptions(arguments)
      let analysis = try analyzeForTransfer(transferOptions)
      let report = try LoomPatternTransferAnalyzer().analyze(
        analysis: analysis,
        options: LoomPatternTransferOptions(
          from: transferOptions.from ?? inferredPlatform(for: transferOptions.sourcePath),
          to: transferOptions.to ?? oppositePlatform(of: inferredPlatform(for: transferOptions.sourcePath)),
          patternsDirectory: transferOptions.patternsDirectory
        )
      )
      let output =
        transferOptions.format == "json"
        ? try LoomPatternTransferAnalyzer().json(report)
        : LoomPatternTransferAnalyzer().text(report)
      try writeOrPrint(output, path: transferOptions.outputPath, runtime: runtime)
      return
    }
    var options = try parsePatternOptions(arguments)

    let catalog = LoomPatternCatalog()
    switch command.command {
    case "patterns:list":
      guard options.positional.isEmpty else {
        throw LoomError.invalidArguments("patterns:list accepts no positional arguments.")
      }
      let patterns = try catalog.load(directory: options.directory)
      let output = options.json ? try catalog.json(patterns) : catalog.listText(patterns)
      try writeOrPrint(output, path: options.outputPath, runtime: runtime)
    case "patterns:show":
      guard options.positional.count == 1 else {
        throw LoomError.invalidArguments("patterns:show requires one pattern id.")
      }
      guard let pattern = try catalog.find(options.positional[0], directory: options.directory)
      else {
        throw LoomError.invalidPattern(
          "No pattern named \(options.positional[0]) in \(options.directory).")
      }
      try writeOrPrint(try catalog.json(pattern), path: options.outputPath, runtime: runtime)
    case "patterns:validate":
      guard options.positional.count <= 1 else {
        throw LoomError.invalidArguments("patterns:validate accepts at most one directory.")
      }
      if let suppliedDirectory = options.positional.first { options.directory = suppliedDirectory }
      let report = catalog.validate(directory: options.directory)
      let output = options.json ? try catalog.json(report) : catalog.text(report)
      try writeOrPrint(output, path: options.outputPath, runtime: runtime)
      if report.status != "ok" { exit(1) }
    case "patterns:lint":
      guard options.positional.count <= 1 else {
        throw LoomError.invalidArguments("patterns:lint accepts at most one directory.")
      }
      if let suppliedDirectory = options.positional.first { options.directory = suppliedDirectory }
      let report = catalog.lint(directory: options.directory)
      let output = options.json ? try catalog.json(report) : catalog.text(report)
      try writeOrPrint(output, path: options.outputPath, runtime: runtime)
      if report.status != "ok" { exit(1) }
    case "patterns:export":
      guard options.positional.isEmpty else {
        throw LoomError.invalidArguments("patterns:export accepts no positional arguments.")
      }
      let patterns = try catalog.load(directory: options.directory)
      let output = try catalog.export(patterns, format: options.format)
      try writeOrPrint(output, path: options.outputPath, runtime: runtime)
    default:
      throw LoomError.invalidArguments("Unknown pattern command \(command.command).")
    }
  }

  private static func runASCIICommand(_ arguments: [String], runtime: RuntimeOptions) throws {
    let options = try parseSourceOptions("inspect:ascii", arguments: arguments)
    let analysis: LoomAnalysis
    if options.sourcePath.lowercased().hasSuffix(".xaml") {
      analysis = try XAMLFrontend().analyze(sourcePath: options.sourcePath)
    } else {
      analysis = try SwiftUIFrontend().analyze(
        sourcePath: options.sourcePath,
        rootView: options.rootView,
        component: options.component
      )
    }
    try writeOrPrint(
      LoomASCIIPatternRenderer().render(analysis),
      path: options.outputPath,
      runtime: runtime
    )
  }

  private static func parsePatternOptions(_ arguments: [String]) throws -> PatternOptions {
    var options = PatternOptions()
    var index = 1
    while index < arguments.count {
      let argument = arguments[index]
      switch argument {
      case "--json":
        options.json = true
        index += 1
      case "--directory":
        guard index + 1 < arguments.count else {
          throw LoomError.invalidArguments("Missing value for --directory.")
        }
        options.directory = arguments[index + 1]
        index += 2
      case "--format":
        guard index + 1 < arguments.count else {
          throw LoomError.invalidArguments("Missing value for --format.")
        }
        guard let format = LoomPatternExportFormat(rawValue: arguments[index + 1]) else {
          throw LoomError.invalidArguments(
            "--format must be loom, dtcg, open-ui, aria, or style-dictionary.")
        }
        options.format = format
        index += 2
      case "--output":
        guard index + 1 < arguments.count else {
          throw LoomError.invalidArguments("Missing value for --output.")
        }
        options.outputPath = arguments[index + 1]
        index += 2
      default:
        if argument.hasPrefix("--directory=") {
          options.directory = String(argument.dropFirst("--directory=".count))
          index += 1
        } else if argument.hasPrefix("--format=") {
          let value = String(argument.dropFirst("--format=".count))
          guard let format = LoomPatternExportFormat(rawValue: value) else {
            throw LoomError.invalidArguments(
              "--format must be loom, dtcg, open-ui, aria, or style-dictionary.")
          }
          options.format = format
          index += 1
        } else if argument.hasPrefix("--output=") {
          options.outputPath = String(argument.dropFirst("--output=".count))
          index += 1
        } else {
          options.positional.append(argument)
          index += 1
        }
      }
    }
    return options
  }

  private static func parsePatternTransferOptions(_ arguments: [String]) throws
    -> PatternTransferCommandOptions
  {
    guard arguments.count >= 2 else {
      throw LoomError.invalidArguments("patterns:transfer requires a Swift or XAML source path.")
    }
    var options = PatternTransferCommandOptions(sourcePath: arguments[1])
    var index = 2
    while index < arguments.count {
      let flag = arguments[index]
      if flag == "--json" {
        options.format = "json"
        index += 1
        continue
      }
      guard index + 1 < arguments.count else {
        throw LoomError.invalidArguments("Missing value for \(flag).")
      }
      let value = arguments[index + 1]
      switch flag {
      case "--from":
        guard let platform = LoomTransferPlatform(rawValue: value.lowercased()) else {
          throw LoomError.invalidArguments("--from must be swiftui or winui3.")
        }
        options.from = platform
      case "--to":
        guard let platform = LoomTransferPlatform(rawValue: value.lowercased()) else {
          throw LoomError.invalidArguments("--to must be swiftui or winui3.")
        }
        options.to = platform
      case "--root-view":
        options.rootView = value
      case "--component":
        options.component = value
      case "--patterns-dir":
        options.patternsDirectory = value
      case "--format":
        guard value == "text" || value == "json" else {
          throw LoomError.invalidArguments("--format must be text or json.")
        }
        options.format = value
      case "--output":
        options.outputPath = value
      default:
        throw LoomError.invalidArguments("Unknown transfer option \(flag).")
      }
      index += 2
    }
    let from = options.from ?? inferredPlatform(for: options.sourcePath)
    let to = options.to ?? oppositePlatform(of: from)
    guard from != to else {
      throw LoomError.invalidArguments("--from and --to must describe different platforms.")
    }
    options.from = from
    options.to = to
    return options
  }

  private static func analyzeForTransfer(_ options: PatternTransferCommandOptions) throws
    -> LoomAnalysis
  {
    switch options.from ?? inferredPlatform(for: options.sourcePath) {
    case .swiftui:
      return try SwiftUIFrontend().analyze(
        sourcePath: options.sourcePath,
        rootView: options.rootView,
        component: options.component
      )
    case .winui3:
      return try XAMLFrontend().analyze(sourcePath: options.sourcePath)
    }
  }

  private static func inferredPlatform(for path: String) -> LoomTransferPlatform {
    path.lowercased().hasSuffix(".xaml") ? .winui3 : .swiftui
  }

  private static func oppositePlatform(of platform: LoomTransferPlatform) -> LoomTransferPlatform {
    platform == .swiftui ? .winui3 : .swiftui
  }

  private static func runDiagnosticCommand(_ command: String, arguments: [String]) throws {
    let options = try parseDiagnosticOptions(arguments)
    let diagnostics = LoomDiagnostics()
    let output: String
    let exitStatus: Int32
    switch command {
    case "status":
      let report = diagnostics.status(patternDirectory: options.patternsDirectory)
      output = options.json ? try diagnostics.json(report) : diagnostics.text(report)
      exitStatus = report.patternStatus == "ok" ? 0 : 1
    case "verify":
      let report = diagnostics.verify(patternDirectory: options.patternsDirectory)
      output = options.json ? try diagnostics.json(report) : diagnostics.text(report)
      exitStatus = report.status == "ok" ? 0 : 1
    case "checks:command-catalog":
      let report = diagnostics.commandCatalogCheck()
      output = options.json ? try diagnostics.json(report) : diagnostics.text(report)
      exitStatus = report.status == "ok" ? 0 : 1
    case "guards:summary":
      let report = diagnostics.guardsSummary()
      output = options.json ? try diagnostics.json(report) : diagnostics.text(report)
      exitStatus = 0
    case "self-heal:plan":
      let report = diagnostics.selfHealPlan()
      output = options.json ? try diagnostics.json(report) : diagnostics.text(report)
      exitStatus = 0
    default:
      throw LoomError.invalidArguments("Unknown diagnostic command \(command).")
    }
    print(output, terminator: "")
    if exitStatus != 0 { exit(exitStatus) }
  }

  private static func parseSourceOptions(_ command: String, arguments: [String]) throws
    -> SourceOptions
  {
    guard arguments.count >= 2 else {
      throw LoomError.invalidArguments("\(command) requires a Swift source path.")
    }
    var options = SourceOptions(command: command, sourcePath: arguments[1])
    var index = 2
    while index < arguments.count {
      let flag = arguments[index]
      if flag == "--json" {
        options.format = "json"
        index += 1
        continue
      }
      if flag == "--pattern-comments" {
        options.includePatternComments = true
        index += 1
        continue
      }
      if flag == "--init-region" {
        options.initRegion = true
        index += 1
        continue
      }
      guard index + 1 < arguments.count else {
        throw LoomError.invalidArguments("Missing value for \(flag).")
      }
      let value = arguments[index + 1]
      switch flag {
      case "--root-view": options.rootView = value
      case "--component": options.component = value
      case "--format":
        guard value == "text" || value == "json" else {
          throw LoomError.invalidArguments("--format must be text or json.")
        }
        options.format = value
      case "--output": options.outputPath = value
      case "--xaml": options.xamlPath = value
      case "--theme-prefix": options.themeResourcePrefix = value
      case "--patterns-dir": options.patternsDirectory = value
      case "--replace-region": options.replaceRegionPath = value
      case "--region-id": options.regionID = value
      default: throw LoomError.invalidArguments("Unknown option \(flag).")
      }
      index += 2
    }
    if options.includePatternComments && options.patternsDirectory == nil {
      throw LoomError.invalidArguments("--pattern-comments requires --patterns-dir <path>.")
    }
    if options.regionID != nil && options.replaceRegionPath == nil {
      throw LoomError.invalidArguments("--region-id requires --replace-region <path>.")
    }
    if options.initRegion && options.replaceRegionPath == nil {
      throw LoomError.invalidArguments("--init-region requires --replace-region <path>.")
    }
    return options
  }

  private struct DiagnosticOptions {
    var json = false
    var patternsDirectory = "Patterns"
  }

  private static func parseDiagnosticOptions(_ arguments: [String]) throws -> DiagnosticOptions {
    var options = DiagnosticOptions()
    var index = 1
    while index < arguments.count {
      let argument = arguments[index]
      switch argument {
      case "--json":
        options.json = true
        index += 1
      case "--patterns-dir":
        guard index + 1 < arguments.count else {
          throw LoomError.invalidArguments("Missing value for --patterns-dir.")
        }
        options.patternsDirectory = arguments[index + 1]
        index += 2
      default:
        if argument.hasPrefix("--patterns-dir=") {
          options.patternsDirectory = String(argument.dropFirst("--patterns-dir=".count))
          index += 1
        } else {
          throw LoomError.invalidArguments("Unknown diagnostic option \(argument).")
        }
      }
    }
    return options
  }

  private static func parseXAMLOptions(_ arguments: [String]) throws -> SourceOptions {
    guard arguments.count >= 2 else {
      throw LoomError.invalidArguments("inspect:xaml requires a XAML path.")
    }
    var options = SourceOptions(command: "inspect:xaml", sourcePath: arguments[1])
    var index = 2
    while index < arguments.count {
      let flag = arguments[index]
      if flag == "--json" {
        options.format = "json"
        index += 1
        continue
      }
      guard index + 1 < arguments.count else {
        throw LoomError.invalidArguments("Missing value for \(flag).")
      }
      let value = arguments[index + 1]
      switch flag {
      case "--format":
        guard value == "text" || value == "json" else {
          throw LoomError.invalidArguments("--format must be text or json.")
        }
        options.format = value
      case "--output": options.outputPath = value
      default: throw LoomError.invalidArguments("Unknown XAML option \(flag).")
      }
      index += 2
    }
    return options
  }

  private static func parseErrorInspectionOptions(_ arguments: [String]) throws
    -> ErrorInspectionOptions
  {
    guard arguments.count >= 2 else {
      throw LoomError.invalidArguments("inspect:errors requires a source path.")
    }
    var options = ErrorInspectionOptions(path: arguments[1])
    var index = 2
    while index < arguments.count {
      let flag = arguments[index]
      if flag == "--json" {
        options.format = "json"
        index += 1
        continue
      }
      guard index + 1 < arguments.count else {
        throw LoomError.invalidArguments("Missing value for \(flag).")
      }
      let value = arguments[index + 1]
      switch flag {
      case "--kind":
        guard let kind = LoomErrorInspectionKind(rawValue: value) else {
          throw LoomError.invalidArguments("--kind must be swift, xaml, manifest, or patterns.")
        }
        options.kind = kind
      case "--root-view": options.rootView = value
      case "--component": options.component = value
      case "--format":
        guard value == "text" || value == "json" else {
          throw LoomError.invalidArguments("--format must be text or json.")
        }
        options.format = value
      case "--fail-on":
        guard let failOn = LoomErrorFailMode(rawValue: value) else {
          throw LoomError.invalidArguments("--fail-on must be none, error, or warning.")
        }
        options.failOn = failOn
      case "--output": options.outputPath = value
      default:
        throw LoomError.invalidArguments("Unknown error inspection option \(flag).")
      }
      index += 2
    }
    return options
  }

  private static func parseAccessibilityAuditOptions(_ arguments: [String]) throws
    -> AccessibilityAuditOptions
  {
    guard arguments.count >= 2 else {
      throw LoomError.invalidArguments("accessibility:audit requires a Swift or XAML source path.")
    }
    var options = AccessibilityAuditOptions(path: arguments[1])
    var index = 2
    while index < arguments.count {
      let flag = arguments[index]
      if flag == "--json" {
        options.format = "json"
        index += 1
        continue
      }
      guard index + 1 < arguments.count else {
        throw LoomError.invalidArguments("Missing value for \(flag).")
      }
      let value = arguments[index + 1]
      switch flag {
      case "--root-view": options.rootView = value
      case "--component": options.component = value
      case "--format":
        guard value == "text" || value == "json" else {
          throw LoomError.invalidArguments("--format must be text or json.")
        }
        options.format = value
      case "--fail-on":
        guard let failOn = LoomErrorFailMode(rawValue: value) else {
          throw LoomError.invalidArguments("--fail-on must be none, error, or warning.")
        }
        options.failOn = failOn
      case "--output": options.outputPath = value
      default:
        throw LoomError.invalidArguments("Unknown accessibility audit option \(flag).")
      }
      index += 2
    }
    return options
  }

  private static func parseSwiftUIOptions(_ arguments: [String]) throws -> SwiftUIOptions {
    guard arguments.count >= 2 else {
      throw LoomError.invalidArguments("generate:swiftui requires a XAML path.")
    }
    var options = SwiftUIOptions(xamlPath: arguments[1])
    var index = 2
    while index < arguments.count {
      guard index + 1 < arguments.count else {
        throw LoomError.invalidArguments("Missing value for \(arguments[index]).")
      }
      switch arguments[index] {
      case "--view-name": options.viewName = arguments[index + 1]
      case "--output": options.outputPath = arguments[index + 1]
      default: throw LoomError.invalidArguments("Unknown SwiftUI generation option \(arguments[index]).")
      }
      index += 2
    }
    return options
  }

  private static func parseGraphOptions(_ arguments: [String]) throws -> GraphOptions {
    guard arguments.count >= 2 else {
      throw LoomError.invalidArguments("graph:components requires a Swift source path or directory.")
    }
    var options = GraphOptions(sourceRoot: arguments[1])
    var index = 2
    while index < arguments.count {
      let flag = arguments[index]
      if flag == "--json" {
        options.format = "json"
        index += 1
        continue
      }
      guard index + 1 < arguments.count else {
        throw LoomError.invalidArguments("Missing value for \(flag).")
      }
      let value = arguments[index + 1]
      switch flag {
      case "--root-view": options.rootView = value
      case "--component": options.component = value
      case "--format":
        guard ["text", "json", "dot"].contains(value) else {
          throw LoomError.invalidArguments("--format must be text, json, or dot.")
        }
        options.format = value
      case "--output": options.outputPath = value
      case "--include": options.include.append(value)
      case "--exclude": options.exclude.append(value)
      default: throw LoomError.invalidArguments("Unknown graph option \(flag).")
      }
      index += 2
    }
    return options
  }

  private static func parseProjectOptions(_ arguments: [String]) throws -> ProjectOptions {
    guard arguments.count >= 2 else {
      throw LoomError.invalidArguments("project:build requires a Loom manifest path.")
    }
    var options = ProjectOptions(manifestPath: arguments[1])
    var index = 2
    while index < arguments.count {
      guard index + 1 < arguments.count else {
        throw LoomError.invalidArguments("Missing value for \(arguments[index]).")
      }
      switch arguments[index] {
      case "--project-root": options.projectRoot = arguments[index + 1]
      case "--output-dir": options.outputDirectory = arguments[index + 1]
      default: throw LoomError.invalidArguments("Unknown project option \(arguments[index]).")
      }
      index += 2
    }
    return options
  }

  private static func parseValidationOptions(_ arguments: [String]) throws -> ValidationOptions {
    guard arguments.count >= 2 else {
      throw LoomError.invalidArguments("config:validate requires a Loom manifest path.")
    }
    var options = ValidationOptions(manifestPath: arguments[1])
    var index = 2
    while index < arguments.count {
      if arguments[index] == "--json" {
        options.format = "json"
        index += 1
        continue
      }
      guard index + 1 < arguments.count else {
        throw LoomError.invalidArguments("Missing value for \(arguments[index]).")
      }
      switch arguments[index] {
      case "--project-root": options.projectRoot = arguments[index + 1]
      case "--format": options.format = arguments[index + 1]
      default: throw LoomError.invalidArguments("Unknown validation option \(arguments[index]).")
      }
      index += 2
    }
    guard options.format == "text" || options.format == "json" else {
      throw LoomError.invalidArguments("--format must be text or json.")
    }
    return options
  }

  private static func writeOrPrint(_ output: String, path: String?, runtime: RuntimeOptions) throws {
    guard let path else {
      print(output, terminator: "")
      return
    }
    let url = URL(fileURLWithPath: path)
    try FileManager.default.createDirectory(
      at: url.deletingLastPathComponent(),
      withIntermediateDirectories: true
    )
    try output.write(to: url, atomically: true, encoding: .utf8)
    if runtime.verbose {
      fputs("[info] wrote \(path) (\(output.utf8.count) bytes)\n", stderr)
    }
    if !runtime.quiet {
      print("Wrote \(path)")
    }
  }
}
