import Foundation
import Testing

@testable import LoomCore

private let sample = """
  import SwiftUI

  struct ContentView: View {
      var body: some View {
          GeometryReader { geometry in
              ZStack {
                  Color.black
                  HStack(spacing: 12) {
                      Text("Now Playing")
                      Spacer()
                      Button("Play") {
                          play()
                      }
                  }
                  .padding(10)
              }
          }
      }

      private var sidebar: some View {
          VStack {
              Text("Singers")
              List { Text("Ada") }
          }
          .frame(minWidth: 170, idealWidth: 190, maxWidth: 240)
      }
  }
  """

@Test func extractsViewHierarchy() throws {
  let analysis = try SwiftUIFrontend().analyze(source: sample, rootView: "ContentView")
  #expect(analysis.layout.count(kind: .geometryReader) == 1)
  #expect(analysis.layout.count(kind: .overlayStack) == 1)
  #expect(analysis.layout.count(kind: .horizontalStack) == 1)
  #expect(analysis.layout.count(kind: .spacer) == 1)
  #expect(analysis.layout.count(kind: .button) == 1)
}

@Test func extractsComputedSubview() throws {
  let analysis = try SwiftUIFrontend().analyze(
    source: sample,
    rootView: "ContentView",
    component: "sidebar"
  )
  #expect(analysis.layout.count(kind: .verticalStack) == 1)
  #expect(analysis.layout.count(kind: .list) == 1)
  let frame = analysis.layout.children.first?.modifiers.first(where: { $0.name == "frame" })
  #expect(frame != nil)
}

@Test func emitsWinUIGridAndControls() throws {
  let analysis = try SwiftUIFrontend().analyze(source: sample, rootView: "ContentView")
  let xaml = XAMLEmitter().emit(analysis)
  #expect(xaml.contains("<Grid.ColumnDefinitions>"))
  #expect(xaml.contains("<TextBlock Text=\"Now Playing\""))
  #expect(xaml.contains("<Button Content=\"Play\""))
  #expect(xaml.contains("Width=\"*\""))
}

@Test func emitsDynamicTextAsBindingAndSupportsThemePrefix() throws {
  let analysis = try SwiftUIFrontend().analyze(
    source:
      "struct StatusView: View { var body: some View { ZStack { Color.black; Text(statusText) } } }",
    rootView: "StatusView"
  )
  let xaml = XAMLEmitter(options: .init(themeResourcePrefix: "Voci")).emit(analysis)
  #expect(xaml.contains("Text=\"{Binding statusText}\""))
  #expect(xaml.contains("{ThemeResource VociCanvasBrush}"))
}

@Test func missingViewFailsClearly() {
  #expect(throws: LoomError.self) {
    try SwiftUIFrontend().analyze(source: sample, rootView: "MissingView")
  }
}

@Test func lifecycleModifierClosureIsNotAddedToVisualTree() throws {
  let source = """
    struct LifecycleView: View {
      var body: some View {
        Text("Ready")
          .onAppear { startPlayback() }
          .onChange(of: status) { updateStatus() }
      }
    }
    """
  let analysis = try SwiftUIFrontend().analyze(source: source, rootView: "LifecycleView")
  #expect(analysis.layout.recursiveNodeCount == 2)
  #expect(analysis.layout.count(kind: .text) == 1)
  #expect(analysis.diagnostics.filter { $0.code == "LOOM004" }.count == 2)
}

@Test func parityCheckerFindsTargetSpecificLayoutRisk() throws {
  let analysis = try SwiftUIFrontend().analyze(source: sample, rootView: "ContentView")
  let directory = FileManager.default.temporaryDirectory
    .appendingPathComponent(UUID().uuidString, isDirectory: true)
  try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
  defer { try? FileManager.default.removeItem(at: directory) }
  let xamlURL = directory.appendingPathComponent("View.xaml")
  try "<Grid Width=\"280\"><ScrollViewer /></Grid>".write(
    to: xamlURL,
    atomically: true,
    encoding: .utf8
  )

  let report = try XAMLParityChecker().check(analysis: analysis, xamlPath: xamlURL.path)
  #expect(report.diagnostics.contains { $0.code == "LOOM201" })
  #expect(report.diagnostics.contains { $0.code == "LOOM202" })
}

@Test func commandCatalogUsesVigilStyleGroupsAndAliases() throws {
  let alias = LoomCommandCatalog.resolve("analyze")
  #expect(alias?.command == "inspect:source")
  #expect(alias?.category == "inspection")
  #expect(alias?.access == .conditionalWrite)
  #expect(alias?.writeFlags == ["--output"])

  let catalog = LoomCommandCatalog.catalogText()
  #expect(catalog.contains("inspection\n"))
  #expect(catalog.contains("generation\n"))
  #expect(catalog.contains("projects\n"))
  #expect(catalog.contains("setup\n"))
  #expect(catalog.contains("r/w"))

  let json = try LoomCommandCatalog.catalogJSON(category: "setup")
  let decoded = try JSONDecoder().decode([LoomCommandInfo].self, from: Data(json.utf8))
  #expect(decoded.allSatisfy { $0.category == "setup" })
  #expect(decoded.map(\.command).contains("config:validate"))
}

@Test func projectWorkflowValidatesAndBuildsAllComponents() throws {
  let directory = FileManager.default.temporaryDirectory
    .appendingPathComponent(UUID().uuidString, isDirectory: true)
  let output = directory.appendingPathComponent("Output", isDirectory: true)
  try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
  defer { try? FileManager.default.removeItem(at: directory) }

  let sourceURL = directory.appendingPathComponent("ContentView.swift")
  let xamlURL = directory.appendingPathComponent("MainWindow.xaml")
  let manifestURL = directory.appendingPathComponent("loom.json")
  try sample.write(to: sourceURL, atomically: true, encoding: .utf8)
  try "<Grid Width=\"280\" />".write(to: xamlURL, atomically: true, encoding: .utf8)
  let manifest = LoomProjectManifest(
    project: "Fixture",
    source: "ContentView.swift",
    rootView: "ContentView",
    existingXaml: "MainWindow.xaml",
    components: ["body", "sidebar", "body"],
    themeResourcePrefix: "Fixture"
  )
  let encoder = JSONEncoder()
  encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
  try encoder.encode(manifest).write(to: manifestURL)

  let validation = LoomProjectValidator().validate(manifestPath: manifestURL.path)
  #expect(validation.status == "ok")
  #expect(validation.issues.isEmpty)

  let run = try LoomProjectRunner().run(
    manifestPath: manifestURL.path,
    outputDirectory: output.path
  )
  #expect(run.summary.components.count == 2)
  #expect(run.summary.components.map(\.component) == ["body", "sidebar"])
  #expect(run.summary.parityPath != nil)
  #expect(FileManager.default.fileExists(atPath: run.summaryPath))
  #expect(
    FileManager.default.fileExists(
      atPath: output.appendingPathComponent("body.generated.xaml").path))
  #expect(
    FileManager.default.fileExists(
      atPath: output.appendingPathComponent("sidebar.analysis.json").path))

  let generated = try String(contentsOf: output.appendingPathComponent("body.generated.xaml"))
  #expect(generated.contains("{ThemeResource FixtureCanvasBrush}"))
}

@Test func manifestValidationReportsUnresolvedComponentsWithoutWriting() throws {
  let directory = FileManager.default.temporaryDirectory
    .appendingPathComponent(UUID().uuidString, isDirectory: true)
  try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
  defer { try? FileManager.default.removeItem(at: directory) }
  try sample.write(
    to: directory.appendingPathComponent("ContentView.swift"),
    atomically: true,
    encoding: .utf8
  )
  let manifest = LoomProjectManifest(
    project: "Fixture",
    source: "ContentView.swift",
    rootView: "ContentView",
    components: ["missingComponent"]
  )
  let manifestURL = directory.appendingPathComponent("loom.json")
  try JSONEncoder().encode(manifest).write(to: manifestURL)

  let report = LoomProjectValidator().validate(manifestPath: manifestURL.path)
  #expect(report.status == "error")
  #expect(report.issues.contains { $0.code == "component.unresolved" })
  #expect(
    !FileManager.default.fileExists(atPath: directory.appendingPathComponent("Generated").path))
}

@Test func manifestValidationRejectsMissingVersionAndUnknownKeys() throws {
  let directory = FileManager.default.temporaryDirectory
    .appendingPathComponent(UUID().uuidString, isDirectory: true)
  try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
  defer { try? FileManager.default.removeItem(at: directory) }
  try sample.write(
    to: directory.appendingPathComponent("ContentView.swift"),
    atomically: true,
    encoding: .utf8
  )
  let manifestURL = directory.appendingPathComponent("loom.json")
  try """
  {
    "project": "Fixture",
    "source": "ContentView.swift",
    "rootView": "ContentView",
    "target": "winui3",
    "components": ["body"],
    "mystery": true
  }
  """.write(to: manifestURL, atomically: true, encoding: .utf8)

  let report = LoomProjectValidator().validate(manifestPath: manifestURL.path)
  #expect(report.status == "error")
  #expect(report.issues.contains { $0.code == "manifest.schema_version" })
  #expect(
    report.issues.contains { $0.code == "manifest.key.unsupported" && $0.path == "mystery" })
}
