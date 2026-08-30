// swift-tools-version: 6.0

import PackageDescription

let package = Package(
  name: "Loom",
  platforms: [
    .macOS(.v13)
  ],
  products: [
    .library(name: "LoomCore", targets: ["LoomCore"]),
    .executable(name: "loom", targets: ["LoomCLI"]),
  ],
  dependencies: [
    .package(
      url: "https://github.com/swiftlang/swift-syntax.git",
      exact: "602.0.0"
    ),
    .package(
      url: "https://github.com/swiftlang/swift-testing.git",
      revision: "48a471ab313e858258ab0b9b0bf2cea55a50cefb"
    ),
  ],
  targets: [
    .target(
      name: "LoomCore",
      dependencies: [
        .product(name: "SwiftParser", package: "swift-syntax"),
        .product(name: "SwiftDiagnostics", package: "swift-syntax"),
        .product(name: "SwiftParserDiagnostics", package: "swift-syntax"),
        .product(name: "SwiftSyntax", package: "swift-syntax"),
      ]
    ),
    .executableTarget(
      name: "LoomCLI",
      dependencies: ["LoomCore"]
    ),
    .testTarget(
      name: "LoomCoreTests",
      dependencies: [
        "LoomCore",
        .product(name: "Testing", package: "swift-testing"),
      ],
      resources: [.copy("Fixtures")]
    ),
  ]
)
