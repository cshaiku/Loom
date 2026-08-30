import Foundation
import LoomCore

#if os(Linux)
  import Glibc
#elseif os(Windows)
  import CRT
#else
  import Darwin
#endif

private struct CLIOptions {
  var command: String
  var sourcePath: String
  var rootView = "ContentView"
  var component = "body"
  var format = "text"
  var outputPath: String?
  var xamlPath: String?
  var themeResourcePrefix: String?
}

@main
private enum LoomCommand {
  static func main() {
    do {
      let arguments = Array(CommandLine.arguments.dropFirst())
      if arguments == ["--version"] {
        print("loom 0.1.0")
        return
      }
      let options = try parse(arguments)
      let frontend = SwiftUIFrontend()
      let analysis = try frontend.analyze(
        sourcePath: options.sourcePath,
        rootView: options.rootView,
        component: options.component
      )

      let output: String
      switch options.command {
      case "analyze":
        output =
          options.format == "json"
          ? try AnalysisReporter().json(analysis)
          : AnalysisReporter().text(analysis)
      case "generate":
        output = XAMLEmitter(
          options: XAMLEmissionOptions(themeResourcePrefix: options.themeResourcePrefix)
        ).emit(analysis)
      case "parity":
        guard let xamlPath = options.xamlPath else {
          throw LoomError.invalidArguments("The parity command requires --xaml <path>.")
        }
        let report = try XAMLParityChecker().check(analysis: analysis, xamlPath: xamlPath)
        if options.format == "json" {
          let encoder = JSONEncoder()
          encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
          output = String(decoding: try encoder.encode(report), as: UTF8.self) + "\n"
        } else {
          output = XAMLParityChecker().text(report)
        }
      default:
        throw LoomError.invalidArguments("Unknown command \(options.command).")
      }

      if let outputPath = options.outputPath {
        let url = URL(fileURLWithPath: outputPath)
        try FileManager.default.createDirectory(
          at: url.deletingLastPathComponent(),
          withIntermediateDirectories: true
        )
        try output.write(to: url, atomically: true, encoding: .utf8)
        print("Wrote \(outputPath)")
      } else {
        print(output, terminator: "")
      }
    } catch {
      fputs("loom: \(error)\n", stderr)
      fputs(usage, stderr)
      exit(2)
    }
  }

  private static func parse(_ arguments: [String]) throws -> CLIOptions {
    if arguments.isEmpty || arguments.contains("--help") || arguments.contains("-h") {
      print(usage)
      exit(0)
    }
    guard arguments.count >= 2 else {
      throw LoomError.invalidArguments("A command and Swift source path are required.")
    }
    let supported = ["analyze", "generate", "parity"]
    guard supported.contains(arguments[0]) else {
      throw LoomError.invalidArguments("Expected analyze, generate, or parity.")
    }

    var options = CLIOptions(command: arguments[0], sourcePath: arguments[1])
    var index = 2
    while index < arguments.count {
      let flag = arguments[index]
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
      default: throw LoomError.invalidArguments("Unknown option \(flag).")
      }
      index += 2
    }
    return options
  }
}

private let usage = """
  Usage:
  loom --version
  loom analyze <swift-file> [--root-view Name] [--component name] [--format text|json] [--output path]
  loom generate <swift-file> [--root-view Name] [--component name] [--theme-prefix Prefix] [--output path]
  loom parity <swift-file> --xaml <xaml-file> [--root-view Name] [--component name] [--format text|json]
  """
