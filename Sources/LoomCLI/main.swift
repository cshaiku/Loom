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

@main
private enum LoomCommand {
  static func main() {
    do {
      try dispatch(Array(CommandLine.arguments.dropFirst()))
    } catch {
      fputs("[fatal] \(error)\n", stderr)
      fputs("[hint] Run `loom help` or `loom list`.\n", stderr)
      exit(2)
    }
  }

  private static func dispatch(_ arguments: [String]) throws {
    if arguments == ["--version"] || arguments == ["version"] {
      print("loom 0.8.0")
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
      try runPatternCommand(arguments)
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
    case "inspect:source", "inspect:parity", "generate:xaml":
      try runSourceCommand(command.command, arguments: arguments)
    case "inspect:xaml":
      try runXAMLCommand(arguments)
    case "generate:swiftui":
      try runSwiftUICommand(arguments)
    case "graph:components":
      try runGraphCommand(arguments)
    case "project:build":
      let options = try parseProjectOptions(arguments)
      let run = try LoomProjectRunner().run(
        manifestPath: options.manifestPath,
        projectRoot: options.projectRoot,
        outputDirectory: options.outputDirectory
      )
      print(LoomProjectRunner().text(run), terminator: "")
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

  private static func runSourceCommand(_ command: String, arguments: [String]) throws {
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
        content: output
      )
      print(update.changed ? "Updated \(update.path)" : "No changes for \(update.path)")
    } else {
      try writeOrPrint(output, path: options.outputPath)
    }
  }

  private static func runXAMLCommand(_ arguments: [String]) throws {
    let options = try parseXAMLOptions(arguments)
    let analysis = try XAMLFrontend().analyze(sourcePath: options.sourcePath)
    let output =
      options.format == "json"
      ? try AnalysisReporter().json(analysis)
      : AnalysisReporter().text(analysis)
    try writeOrPrint(output, path: options.outputPath)
  }

  private static func runSwiftUICommand(_ arguments: [String]) throws {
    let options = try parseSwiftUIOptions(arguments)
    let analysis = try XAMLFrontend().analyze(sourcePath: options.xamlPath)
    let output = SwiftUIEmitter(options: .init(viewName: options.viewName)).emit(analysis)
    try writeOrPrint(output, path: options.outputPath)
  }

  private static func runGraphCommand(_ arguments: [String]) throws {
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
    try writeOrPrint(output, path: options.outputPath)
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

  private static func runPatternCommand(_ arguments: [String]) throws {
    guard let command = LoomCommandCatalog.resolve(arguments[0]) else {
      throw LoomError.invalidArguments("Unknown pattern command \(arguments[0]).")
    }
    if arguments.count > 1 && (arguments[1] == "--help" || arguments[1] == "-h") {
      print(LoomCommandCatalog.manual(command.command) ?? "", terminator: "")
      return
    }
    var directory = "Patterns"
    var json = false
    var positional: [String] = []
    var index = 1
    while index < arguments.count {
      switch arguments[index] {
      case "--json":
        json = true
        index += 1
      case "--directory":
        guard index + 1 < arguments.count else {
          throw LoomError.invalidArguments("Missing value for --directory.")
        }
        directory = arguments[index + 1]
        index += 2
      default:
        positional.append(arguments[index])
        index += 1
      }
    }

    let catalog = LoomPatternCatalog()
    switch command.command {
    case "patterns:list":
      guard positional.isEmpty else {
        throw LoomError.invalidArguments("patterns:list accepts no positional arguments.")
      }
      let patterns = try catalog.load(directory: directory)
      print(json ? try catalog.json(patterns) : catalog.listText(patterns), terminator: "")
    case "patterns:show":
      guard positional.count == 1 else {
        throw LoomError.invalidArguments("patterns:show requires one pattern id.")
      }
      guard let pattern = try catalog.find(positional[0], directory: directory) else {
        throw LoomError.invalidPattern("No pattern named \(positional[0]) in \(directory).")
      }
      print(try catalog.json(pattern), terminator: "")
    case "patterns:validate":
      guard positional.count <= 1 else {
        throw LoomError.invalidArguments("patterns:validate accepts at most one directory.")
      }
      if let suppliedDirectory = positional.first { directory = suppliedDirectory }
      let report = catalog.validate(directory: directory)
      print(json ? try catalog.json(report) : catalog.text(report), terminator: "")
      if report.status != "ok" { exit(1) }
    case "patterns:lint":
      guard positional.count <= 1 else {
        throw LoomError.invalidArguments("patterns:lint accepts at most one directory.")
      }
      if let suppliedDirectory = positional.first { directory = suppliedDirectory }
      let report = catalog.lint(directory: directory)
      print(json ? try catalog.json(report) : catalog.text(report), terminator: "")
      if report.status != "ok" { exit(1) }
    default:
      throw LoomError.invalidArguments("Unknown pattern command \(command.command).")
    }
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

  private static func writeOrPrint(_ output: String, path: String?) throws {
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
    print("Wrote \(path)")
  }
}
