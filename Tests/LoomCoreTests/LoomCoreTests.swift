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

@Test func deepLayoutAggregationDoesNotRequireRecursiveStackFrames() throws {
  var node = LoomNode(kind: .component, expression: "Leaf")
  for index in 0..<2_000 {
    node = LoomNode(
      kind: index.isMultiple(of: 2) ? .verticalStack : .horizontalStack,
      expression: "node\(index)",
      children: [node]
    )
  }

  #expect(node.recursiveNodeCount == 2_001)
  #expect(node.count(kind: .component) == 1)
  #expect(node.componentReferences == ["Leaf"])
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
  #expect(catalog.contains("accessibility\n"))
  #expect(catalog.contains("accessibility:audit"))
  #expect(catalog.contains("inspection\n"))
  #expect(catalog.contains("diagnostics\n"))
  #expect(catalog.contains("checks:command-catalog"))
  #expect(catalog.contains("guards:summary"))
  #expect(catalog.contains("self-heal:plan"))
  #expect(catalog.contains("verify"))
  #expect(catalog.contains("graph:components"))
  #expect(catalog.contains("inspect:ascii"))
  #expect(catalog.contains("inspect:errors"))
  #expect(catalog.contains("inspect:xaml"))
  #expect(catalog.contains("generation\n"))
  #expect(catalog.contains("generate:contracts"))
  #expect(catalog.contains("generate:swiftui"))
  #expect(catalog.contains("projects\n"))
  #expect(catalog.contains("patterns\n"))
  #expect(catalog.contains("patterns:lint"))
  #expect(catalog.contains("patterns:export"))
  #expect(catalog.contains("patterns:transfer"))
  #expect(catalog.contains("setup\n"))
  #expect(catalog.contains("r/w"))

  let json = try LoomCommandCatalog.catalogJSON(category: "setup")
  let decoded = try JSONDecoder().decode([LoomCommandInfo].self, from: Data(json.utf8))
  #expect(decoded.allSatisfy { $0.category == "setup" })
  #expect(decoded.map(\.command).contains("config:validate"))
}

@Test func patternCatalogExportsExternalMetadataShapes() throws {
  let repository = URL(fileURLWithPath: #filePath)
    .deletingLastPathComponent()
    .deletingLastPathComponent()
    .deletingLastPathComponent()
  let directory = repository.appendingPathComponent("Patterns").path
  let catalog = LoomPatternCatalog()
  let patterns = try catalog.load(directory: directory)

  let dtcg = try catalog.export(patterns, format: .dtcg)
  #expect(dtcg.contains("\"$type\" : \"other\""))
  #expect(dtcg.contains("\"$value\""))
  #expect(dtcg.contains("\"$extensions\""))
  #expect(dtcg.contains("\"text\""))

  let openUI = try catalog.export(patterns, format: .openUI)
  #expect(openUI.contains("\"source\" : \"loom\""))
  #expect(openUI.contains("\"components\""))
  #expect(openUI.contains("\"mappings\""))

  let aria = try catalog.export(patterns, format: .aria)
  #expect(aria.contains("\"patterns\""))
  #expect(aria.contains("\"role\""))
  #expect(aria.contains("\"focusBehavior\""))

  let styleDictionary = try catalog.export(patterns, format: .styleDictionary)
  #expect(styleDictionary.contains("\"type\" : \"loom.pattern\""))
  #expect(styleDictionary.contains("\"value\" : \"text\""))
  #expect(styleDictionary.contains("\"platforms\""))
}

@Test func diagnosticsCommandsReportReadinessAndGuards() throws {
  let repository = URL(fileURLWithPath: #filePath)
    .deletingLastPathComponent()
    .deletingLastPathComponent()
    .deletingLastPathComponent()
  let patterns = repository.appendingPathComponent("Patterns").path
  let diagnostics = LoomDiagnostics()

  let catalog = diagnostics.commandCatalogCheck()
  #expect(catalog.status == "ok")
  #expect(catalog.commands == LoomCommandCatalog.commands.count)

  let verify = diagnostics.verify(patternDirectory: patterns)
  #expect(verify.status == "ok")
  #expect(verify.commandCatalog.status == "ok")
  #expect(verify.patterns.status == "ok")
  #expect(verify.patternLint.status == "ok")

  let status = diagnostics.status(patternDirectory: patterns)
  #expect(status.version == LoomDiagnostics.version)
  #expect(status.patternStatus == "ok")
  #expect(status.patternCount == 20)

  let guards = diagnostics.guardsSummary()
  #expect(guards.entries.contains { $0.command == "generate:xaml" && $0.writeFlags.contains("--replace-region") })
  #expect(guards.entries.contains { $0.command == "project:build" && $0.writeFlags.isEmpty })

  let plan = diagnostics.selfHealPlan()
  #expect(plan.entries.contains { $0.command == "generate:xaml" && $0.flag == "--init-region" })

  let json = try diagnostics.json(catalog)
  let decoded = try JSONDecoder().decode(LoomCommandCatalogCheckReport.self, from: Data(json.utf8))
  #expect(decoded.status == "ok")
}

@Test func asciiPatternRendererUsesPlainTextTree() throws {
  let analysis = try SwiftUIFrontend().analyze(source: sample, rootView: "ContentView")
  let ascii = LoomASCIIPatternRenderer().render(analysis)
  #expect(ascii.contains("= ContentView.body"))
  #expect(ascii.contains("|--") || ascii.contains("\\--"))
  #expect(ascii.contains("horizontal-stack / HStack"))
  #expect(ascii.contains("button / Button"))
}

@Test func patternTransferClassifiesLayoutBehaviorAndLoss() throws {
  let repository = URL(fileURLWithPath: #filePath)
    .deletingLastPathComponent()
    .deletingLastPathComponent()
    .deletingLastPathComponent()
  let directory = repository.appendingPathComponent("Patterns").path
  let analysis = try SwiftUIFrontend().analyze(source: sample, rootView: "ContentView")
  let report = try LoomPatternTransferAnalyzer().analyze(
    analysis: analysis,
    options: .init(from: .swiftui, to: .winui3, patternsDirectory: directory)
  )

  #expect(report.from == .swiftui)
  #expect(report.to == .winui3)
  #expect(report.asciiPattern.contains("= ContentView.body"))
  #expect(report.items.contains { $0.kind == .geometryReader && $0.disposition == .lossy })
  #expect(report.items.contains { $0.kind == .button && $0.disposition == .needsNativeContract })
  #expect(report.items.contains { $0.kind == .horizontalStack && $0.disposition == .needsPolicy })
  #expect(report.summary.lossy >= 1)
  #expect(report.summary.needsNativeContract >= 1)
  #expect(report.summary.needsPolicy >= 1)

  let text = LoomPatternTransferAnalyzer().text(report)
  #expect(text.contains("Loom pattern transfer"))
  #expect(text.contains("ASCII Pattern"))

  let json = try LoomPatternTransferAnalyzer().json(report)
  let decoded = try JSONDecoder().decode(LoomPatternTransferReport.self, from: Data(json.utf8))
  #expect(decoded.items.count == report.items.count)
}

@Test func accessibilityAuditFindsA11yAndLayoutDesignIssues() throws {
  let badLayout = LoomNode(
    kind: .root,
    expression: "body",
    children: [
      LoomNode(
        kind: .verticalStack,
        expression: "VStack",
        children: [
          LoomNode(
            kind: .verticalStack,
            expression: "VStack",
            children: [LoomNode(kind: .text, expression: "Text", arguments: "\"Nested\"")]
          ),
          LoomNode(
            kind: .button,
            expression: "Button",
            modifiers: [LoomModifier(name: "frame", arguments: "width: 20, height: 20")]
          ),
          LoomNode(kind: .image, expression: "Image", arguments: "\"icon\""),
          LoomNode(kind: .textField, expression: "TextField", arguments: "\"\""),
          LoomNode(kind: .color, expression: "Color.red"),
          LoomNode(
            kind: .scrollView,
            expression: "ScrollView",
            children: [
              LoomNode(kind: .scrollView, expression: "ScrollView")
            ]
          ),
          LoomNode(kind: .unsupported, expression: "CustomLayout")
        ]
      )
    ]
  )
  let analysis = LoomAnalysis(
    sourcePath: "Bad.swift",
    rootView: "BadView",
    component: "body",
    syntaxNodeCount: 0,
    layout: badLayout,
    diagnostics: []
  )

  let auditor = LoomAccessibilityAuditor()
  let report = auditor.audit(analysis)
  #expect(report.status == "error")
  #expect(report.findings.contains { $0.code == "AUDIT020" })
  #expect(report.findings.contains { $0.code == "AUDIT030" })
  #expect(report.findings.contains { $0.code == "AUDIT031" })
  #expect(report.findings.contains { $0.code == "AUDIT040" })
  #expect(report.findings.contains { $0.code == "AUDIT052" })
  #expect(report.findings.contains { $0.code == "AUDIT014" })
  #expect(report.findings.contains { $0.code == "AUDIT060" })
  #expect(report.findings.contains { $0.code == "AUDIT061" })
  #expect(report.findings.contains { $0.category == .redundant })
  #expect(auditor.shouldFail(report, mode: .error))
  #expect(auditor.shouldFail(report, mode: .warning))
  #expect(!auditor.shouldFail(report, mode: .none))

  let text = auditor.text(report)
  #expect(text.contains("Loom accessibility audit"))
  #expect(text.contains("AUDIT020"))

  let json = try auditor.json(report)
  let decoded = try JSONDecoder().decode(LoomAccessibilityAuditReport.self, from: Data(json.utf8))
  #expect(decoded.findings.count == report.findings.count)
}

@Test func errorInspectionReportsSwiftSyntaxAndLoomDiagnostics() throws {
  let directory = FileManager.default.temporaryDirectory
    .appendingPathComponent(UUID().uuidString, isDirectory: true)
  try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
  defer { try? FileManager.default.removeItem(at: directory) }
  let sourceURL = directory.appendingPathComponent("Broken.swift")
  try """
  import SwiftUI

  struct BrokenView: View {
    var body: some View {
      VStack {
        Text("Broken"
      }
    }
  }
  """.write(to: sourceURL, atomically: true, encoding: .utf8)

  let diagnostics = LoomDiagnostics()
  let report = diagnostics.inspectErrors(
    path: sourceURL.path,
    options: .init(kind: .swift, rootView: "BrokenView")
  )

  #expect(report.status == "error")
  #expect(report.inspectedKind == .swift)
  #expect(report.findings.contains { $0.code.hasPrefix("SWIFT.") && $0.severity == .error })
  #expect(diagnostics.shouldFail(report, mode: .error))
  #expect(diagnostics.shouldFail(report, mode: .warning))
  #expect(!diagnostics.shouldFail(report, mode: .none))
}

@Test func errorInspectionReportsXAMLAndPatternErrors() throws {
  let directory = FileManager.default.temporaryDirectory
    .appendingPathComponent(UUID().uuidString, isDirectory: true)
  try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
  defer { try? FileManager.default.removeItem(at: directory) }

  let xamlURL = directory.appendingPathComponent("Broken.xaml")
  try "<Grid><TextBlock Text=\"Broken\"></Grid>".write(
    to: xamlURL, atomically: true, encoding: .utf8)
  let xaml = LoomDiagnostics().inspectErrors(path: xamlURL.path, options: .init(kind: .xaml))
  #expect(xaml.status == "error")
  #expect(xaml.findings.contains { $0.code == "LOOM.XAML" })

  let patterns = LoomDiagnostics().inspectErrors(path: directory.path, options: .init(kind: .patterns))
  #expect(patterns.status == "error")
  #expect(patterns.findings.contains { $0.code == "PATTERN003" })
}

@Test func targetContractsExtractBindingsBehaviorAndResources() throws {
  let source = """
  import SwiftUI

  struct ContractView: View {
    var body: some View {
      VStack {
        Text(statusText)
        TextField("Search", text: $query)
          .accessibilityIdentifier("searchBox")
        if isReady {
          Button("Run") { startRun() }
        }
      }
      .onAppear { load() }
    }
  }
  """
  let analysis = try SwiftUIFrontend().analyze(source: source, rootView: "ContractView")
  let generator = LoomTargetContractGenerator()
  let report = try generator.generate(
    analysis: analysis,
    options: .init(themeResourcePrefix: "Loom")
  )

  #expect(report.target == "winui3")
  #expect(report.items.contains { $0.kind == .binding && $0.name == "statusText" })
  #expect(report.items.contains { $0.kind == .binding && $0.source.contains("TextField") })
  #expect(report.items.contains { $0.kind == .accessibility && $0.name == "searchBox" })
  #expect(report.items.contains { $0.kind == .visibility && $0.name == "isReady" })
  #expect(report.items.contains { $0.kind == .action && $0.name == "startRun" })
  #expect(report.items.contains { $0.kind == .lifecycle && $0.name == "onAppear" })
  #expect(report.items.contains { $0.kind == .themeResource && $0.name == "Loom.*" })

  let text = generator.text(report)
  #expect(text.contains("Loom target contracts"))
  #expect(text.contains("[action] startRun"))

  let json = try generator.json(report)
  let decoded = try JSONDecoder().decode(LoomTargetContractReport.self, from: Data(json.utf8))
  #expect(decoded.items.count == report.items.count)
}

@Test func swiftUIEmitterGeneratesScaffoldFromXAMLIR() throws {
  let xaml = """
  <StackPanel xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation" Orientation="Vertical" Spacing="8">
    <TextBlock Text="Ready" />
    <Button Content="Run" />
    <TextBox PlaceholderText="Search" Width="240" />
    <ScrollViewer>
      <ListView />
    </ScrollViewer>
  </StackPanel>
  """
  let analysis = try XAMLFrontend().analyze(source: xaml, sourcePath: "Panel.xaml")
  let swift = SwiftUIEmitter(options: .init(viewName: "PanelScaffold")).emit(analysis)
  #expect(swift.contains("struct PanelScaffold: View"))
  #expect(swift.contains("VStack(spacing: 8)"))
  #expect(swift.contains("Text(\"Ready\")"))
  #expect(swift.contains("Button(\"Run\")"))
  #expect(swift.contains("TextField(\"Search\", text: .constant(\"\"))"))
  #expect(swift.contains("ScrollView"))
  #expect(swift.contains("List"))
  #expect(swift.contains(".frame(width: 240)"))
}

@Test func swiftUIEmitterPreservesUnsupportedXAMLAsPlaceholder() throws {
  let analysis = try XAMLFrontend().analyze(
    source: """
    <CommandBar xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation">
      <AppBarButton Label="Refresh" />
    </CommandBar>
    """,
    sourcePath: "Toolbar.xaml"
  )
  let swift = SwiftUIEmitter().emit(analysis)
  #expect(swift.contains("// Unsupported XAML component: CommandBar"))
  #expect(swift.contains("Button(\"Refresh\")"))
}

@Test func swiftUIEmitterPreservesChildrenOfUnsupportedXAMLContainers() throws {
  let analysis = try XAMLFrontend().analyze(
    source: """
    <Window xmlns="http://schemas.microsoft.com/winfx/2006/xaml/presentation">
      <Grid>
        <TextBlock Text="Inside" />
      </Grid>
    </Window>
    """,
    sourcePath: "Window.xaml"
  )
  let swift = SwiftUIEmitter().emit(analysis)
  #expect(swift.contains("// Unsupported XAML component: Window"))
  #expect(swift.contains("Text(\"Inside\")"))
}

@Test func ownedRegionUpdaterReplacesOnlyMarkedXAMLRegion() throws {
  let existing = """
  <Grid>
    <TextBlock Text="Handwritten" />
    <!-- LOOM-BEGIN shell.main -->
    <TextBlock Text="Old generated" />
    <!-- LOOM-END shell.main -->
    <Button Content="Handwritten action" />
  </Grid>
  """
  let generated = "<TextBlock Text=\"New generated\" />\n"

  let updated = try LoomOwnedRegionUpdater().replaceXAMLRegion(
    in: existing,
    regionID: "shell.main",
    content: generated
  )
  #expect(updated.contains("<TextBlock Text=\"Handwritten\" />"))
  #expect(updated.contains("<TextBlock Text=\"New generated\" />"))
  #expect(!updated.contains("Old generated"))
  #expect(updated.contains("<Button Content=\"Handwritten action\" />"))
}

@Test func ownedRegionUpdaterRejectsMissingOrDuplicateMarkers() throws {
  #expect(throws: LoomError.self) {
    _ = try LoomOwnedRegionUpdater().replaceXAMLRegion(
      in: "<Grid />",
      regionID: "missing",
      content: "<TextBlock />"
    )
  }
  #expect(throws: LoomError.self) {
    _ = try LoomOwnedRegionUpdater().replaceXAMLRegion(
      in: """
      <!-- LOOM-BEGIN duplicate -->
      <!-- LOOM-BEGIN duplicate -->
      <!-- LOOM-END duplicate -->
      """,
      regionID: "duplicate",
      content: "<TextBlock />"
    )
  }
}

@Test func ownedRegionUpdaterRejectsMalformedOrderAndUnsafeIDs() throws {
  #expect(throws: LoomError.self) {
    _ = try LoomOwnedRegionUpdater().replaceXAMLRegion(
      in: """
      <!-- LOOM-END shell.main -->
      <!-- LOOM-BEGIN shell.main -->
      """,
      regionID: "shell.main",
      content: "<TextBlock />"
    )
  }
  #expect(throws: LoomError.self) {
    _ = try LoomOwnedRegionUpdater().replaceXAMLRegion(
      in: "<Grid />",
      regionID: "../shell",
      content: "<TextBlock />"
    )
  }
}

@Test func ownedRegionUpdaterReportsUnchangedAndCanInitializeMissingFile() throws {
  let updater = LoomOwnedRegionUpdater()
  let existing = """
  <Grid>
    <!-- LOOM-BEGIN shell.main -->
  <TextBlock Text="Same" />
  <!-- LOOM-END shell.main -->
  </Grid>
  """
  let unchanged = try updater.replaceXAMLRegion(
    in: existing,
    regionID: "shell.main",
    content: "<TextBlock Text=\"Same\" />"
  )
  #expect(unchanged == existing)

  let initialized = updater.initializedXAMLRegion(
    regionID: "shell.main",
    content: "<TextBlock Text=\"Generated\" />"
  )
  #expect(initialized.contains("<!-- LOOM-BEGIN shell.main -->"))
  #expect(initialized.contains("<TextBlock Text=\"Generated\" />"))
  #expect(initialized.contains("<!-- LOOM-END shell.main -->"))

  let directory = FileManager.default.temporaryDirectory
    .appendingPathComponent(UUID().uuidString, isDirectory: true)
  defer { try? FileManager.default.removeItem(at: directory) }
  let missingPath = directory.appendingPathComponent("Generated/Shell.xaml").path
  let update = try updater.replaceXAMLRegion(
    path: missingPath,
    regionID: "shell.main",
    content: "<TextBlock Text=\"Generated\" />",
    createIfMissing: true
  )
  #expect(update.changed)
  #expect(FileManager.default.fileExists(atPath: missingPath))

  let unmarkedPath = directory.appendingPathComponent("Unmarked.xaml")
  try "<Grid />".write(to: unmarkedPath, atomically: true, encoding: .utf8)
  #expect(throws: LoomError.self) {
    _ = try updater.replaceXAMLRegion(
      path: unmarkedPath.path,
      regionID: "shell.main",
      content: "<TextBlock />",
      createIfMissing: true
    )
  }
}

@Test func xamlFrontendRejectsMalformedXML() throws {
  #expect(throws: LoomError.self) {
    _ = try XAMLFrontend().analyze(
      source: "<Grid><TextBlock Text=\"Broken\"></Grid>",
      sourcePath: "Broken.xaml"
    )
  }
}

@Test func componentGraphReportsCyclesWithoutRecursingForever() throws {
  let directory = FileManager.default.temporaryDirectory
    .appendingPathComponent(UUID().uuidString, isDirectory: true)
  try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
  defer { try? FileManager.default.removeItem(at: directory) }
  try """
  import SwiftUI

  struct ContentView: View {
    var body: some View { first }
    var first: some View { second }
    var second: some View { first }
  }
  """.write(
    to: directory.appendingPathComponent("ContentView.swift"), atomically: true, encoding: .utf8)

  let graph = try LoomComponentGraphBuilder().build(
    sourceRoot: directory.path,
    rootView: "ContentView"
  )
  #expect(graph.status == "error")
  #expect(graph.diagnostics.contains { $0.code == "GRAPH003" })
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

@Test func patternLintRejectsOperationalMappingGaps() throws {
  let directory = FileManager.default.temporaryDirectory
    .appendingPathComponent(UUID().uuidString, isDirectory: true)
  try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
  defer { try? FileManager.default.removeItem(at: directory) }

  let text = """
  {
    "schema_version": "1",
    "id": "text",
    "version": "1.0.0",
    "name": "Text",
    "kind": "text",
    "status": "stable",
    "category": "content",
    "intent": {
      "summary": "Render read-only text content.",
      "useWhen": ["Static or bound text is needed."],
      "avoidWhen": []
    },
    "semantics": {
      "role": "text",
      "childPolicy": "none",
      "sizing": "intrinsic",
      "ordering": "document"
    },
    "attributes": [
      {
        "name": "content",
        "valueType": "binding",
        "required": true,
        "description": "Text content or binding expression.",
        "defaultValue": "bad"
      }
    ],
    "constraints": [],
    "accessibility": {
      "role": "text",
      "nameSource": "content",
      "focusBehavior": "not focusable",
      "notes": []
    },
    "mappings": [
      {
        "platform": "swiftui",
        "constructs": ["Text"],
        "strategy": "Map content into Text.",
        "caveats": []
      }
    ],
    "tags": ["content"]
  }
  """
  try text.write(
    to: directory.appendingPathComponent("text.pattern.json"),
    atomically: true,
    encoding: .utf8
  )

  let report = LoomPatternCatalog().lint(directory: directory.path)
  #expect(report.status == "error")
  #expect(report.issues.contains { $0.code == "PATTERN103" })
  #expect(report.issues.contains { $0.code == "PATTERN106" })
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
