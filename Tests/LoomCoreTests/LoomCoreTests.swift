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
  #expect(catalog.contains("graph:components"))
  #expect(catalog.contains("inspect:xaml"))
  #expect(catalog.contains("generation\n"))
  #expect(catalog.contains("projects\n"))
  #expect(catalog.contains("patterns\n"))
  #expect(catalog.contains("patterns:lint"))
  #expect(catalog.contains("setup\n"))
  #expect(catalog.contains("r/w"))

  let json = try LoomCommandCatalog.catalogJSON(category: "setup")
  let decoded = try JSONDecoder().decode([LoomCommandInfo].self, from: Data(json.utf8))
  #expect(decoded.allSatisfy { $0.category == "setup" })
  #expect(decoded.map(\.command).contains("config:validate"))
}

@Test func xamlFrontendNormalizesWinUIControlsIntoLoomIR() throws {
  let xaml = """
  <Grid xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation">
    <StackPanel Orientation="Horizontal" AutomationProperties.AutomationId="toolbar">
      <TextBlock Text="Now Playing" />
      <Button Content="Play" />
      <TextBox PlaceholderText="Search" Width="240" />
      <ScrollViewer>
        <ListView />
      </ScrollViewer>
    </StackPanel>
  </Grid>
  """

  let analysis = try XAMLFrontend().analyze(source: xaml, sourcePath: "MainWindow.xaml")
  #expect(analysis.rootView == "XAML")
  #expect(analysis.layout.count(kind: .grid) == 1)
  #expect(analysis.layout.count(kind: .horizontalStack) == 1)
  #expect(analysis.layout.count(kind: .text) == 1)
  #expect(analysis.layout.count(kind: .button) == 1)
  #expect(analysis.layout.count(kind: .textField) == 1)
  #expect(analysis.layout.count(kind: .scrollView) == 1)
  #expect(analysis.layout.count(kind: .list) == 1)
  #expect(analysis.diagnostics.isEmpty)

  let text = AnalysisReporter().text(analysis)
  #expect(text.contains("Source nodes:"))
}

@Test func generatedXAMLCanBeIngestedBackIntoComparableIR() throws {
  let swift = """
  struct ContentView: View {
    var body: some View {
      VStack(spacing: 8) {
        Text("Ready")
        Button("Run") { start() }
      }
    }
  }
  """
  let swiftAnalysis = try SwiftUIFrontend().analyze(source: swift, rootView: "ContentView")
  let xaml = XAMLEmitter().emit(swiftAnalysis)
  let xamlAnalysis = try XAMLFrontend().analyze(source: xaml, sourcePath: "generated.xaml")

  #expect(swiftAnalysis.layout.count(kind: .text) == xamlAnalysis.layout.count(kind: .text))
  #expect(swiftAnalysis.layout.count(kind: .button) == xamlAnalysis.layout.count(kind: .button))
  #expect(xamlAnalysis.layout.count(kind: .grid) >= 1)
}

@Test func componentGraphDiscoversComputedAndCustomViewDependencies() throws {
  let directory = FileManager.default.temporaryDirectory
    .appendingPathComponent(UUID().uuidString, isDirectory: true)
  try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
  defer { try? FileManager.default.removeItem(at: directory) }

  try """
  import SwiftUI

  struct ContentView: View {
    @State private var isRunning = false

    var body: some View {
      VStack {
        header
        DetailsView()
        Button("Run") { performAction() }
      }
    }

    private var header: some View {
      Text("Header")
    }
  }
  """.write(
    to: directory.appendingPathComponent("ContentView.swift"), atomically: true, encoding: .utf8)
  try """
  import SwiftUI

  struct DetailsView: View {
    var body: some View {
      List {
        Text("Detail")
      }
    }
  }
  """.write(
    to: directory.appendingPathComponent("DetailsView.swift"), atomically: true, encoding: .utf8)

  let graph = try LoomComponentGraphBuilder().build(
    sourceRoot: directory.path,
    rootView: "ContentView"
  )
  #expect(graph.status == "ok")
  #expect(graph.nodes.map(\.id.description) == [
    "ContentView.body", "ContentView.header", "DetailsView.body",
  ])
  #expect(
    graph.edges.map { "\($0.from.description)->\($0.to.description)" } == [
      "ContentView.body->ContentView.header",
      "ContentView.body->DetailsView.body",
    ])
  #expect(graph.nodes.allSatisfy { !$0.unresolvedReferences.contains("performAction") })
  #expect(LoomComponentGraphBuilder().dot(graph).contains("\"ContentView.body\" -> \"DetailsView.body\""))
}

@Test func patternCatalogIsValidAndCoversSemanticNodeKinds() throws {
  let repository = URL(fileURLWithPath: #filePath)
    .deletingLastPathComponent()
    .deletingLastPathComponent()
    .deletingLastPathComponent()
  let directory = repository.appendingPathComponent("Patterns").path
  let catalog = LoomPatternCatalog()
  let report = catalog.validate(directory: directory)
  #expect(report.status == "ok")
  #expect(report.issues.isEmpty)
  let lint = catalog.lint(directory: directory)
  #expect(lint.status == "ok")
  #expect(lint.issues.isEmpty)

  let patterns = try catalog.load(directory: directory)
  #expect(patterns.count == 20)
  let intentionallyNonSemantic: Set<LoomNodeKind> = [.root, .unsupported]
  #expect(
    Set(patterns.map(\.kind)) == Set(LoomNodeKind.allCases).subtracting(intentionallyNonSemantic))
  #expect(patterns.allSatisfy { !$0.intent.summary.isEmpty })
  #expect(patterns.allSatisfy { !$0.mappings.isEmpty })
}

@Test func xamlEmitterCanTracePatternMappings() throws {
  let repository = URL(fileURLWithPath: #filePath)
    .deletingLastPathComponent()
    .deletingLastPathComponent()
    .deletingLastPathComponent()
  let registry = try LoomPatternRegistry(
    directory: repository.appendingPathComponent("Patterns").path)
  let analysis = try SwiftUIFrontend().analyze(
    source: "struct ContentView: View { var body: some View { VStack { Text(\"Ready\") } } }",
    rootView: "ContentView"
  )

  let stable = XAMLEmitter().emit(analysis)
  #expect(!stable.contains("Loom pattern:"))

  let traced = XAMLEmitter(
    options: .init(includePatternComments: true),
    patternRegistry: registry
  ).emit(analysis)
  #expect(traced.contains("Loom pattern: vertical-stack -> winui3 Grid, StackPanel"))
  #expect(traced.contains("Loom pattern: text -> winui3 TextBlock"))
}

@Test func patternValidationRejectsInvalidRangesAndDuplicateKinds() throws {
  let directory = FileManager.default.temporaryDirectory
    .appendingPathComponent(UUID().uuidString, isDirectory: true)
  try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
  defer { try? FileManager.default.removeItem(at: directory) }
  let repository = URL(fileURLWithPath: #filePath)
    .deletingLastPathComponent()
    .deletingLastPathComponent()
    .deletingLastPathComponent()
  let source = repository.appendingPathComponent("Patterns/text.pattern.json")
  var text = try String(contentsOf: source, encoding: .utf8)
  try text.write(
    to: directory.appendingPathComponent("text.pattern.json"), atomically: true, encoding: .utf8)
  text = text.replacingOccurrences(of: "\"id\":\"text\"", with: "\"id\":\"text-copy\"")
    .replacingOccurrences(
      of: "\"minimum\":1,\"maximum\":1000", with: "\"minimum\":10,\"maximum\":1")
  try text.write(
    to: directory.appendingPathComponent("text-copy.pattern.json"), atomically: true,
    encoding: .utf8)

  let report = LoomPatternCatalog().validate(directory: directory.path)
  #expect(report.status == "error")
  #expect(report.issues.contains { $0.code == "PATTERN008" })
  #expect(report.issues.contains { $0.code == "PATTERN012" })
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
