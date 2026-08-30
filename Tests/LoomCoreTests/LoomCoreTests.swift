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
