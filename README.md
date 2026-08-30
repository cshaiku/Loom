# Loom

Current version: **0.1.0**

Loom is a compiler-assisted bridge from SwiftUI layout source to WinUI 3 XAML.
It extracts a SwiftUI view hierarchy, lowers it into a platform-neutral layout
model, emits reviewable XAML, and reports constructs that need an explicit
Windows implementation.

Loom deliberately does **not** claim to transpile an entire SwiftUI application.
State ownership, media surfaces, gestures, custom layouts, platform services,
and application behavior remain platform code. Loom focuses on repeatable UI
structure, layout constraints, controls, theme references, accessibility, and
parity diagnostics.

## Why the name

The name follows the short, evocative style of Vigil, Clyde, Nora, and Scribe.
Loom weaves one declarative UI description into another while keeping the
platform-specific threads visible.

## Current status

The first usable slice includes:

- Swift parsing through SwiftParser and SwiftSyntax.
- Discovery of Swift structs conforming to `View`.
- Extraction of `body` or another computed view property.
- A Codable platform-neutral layout tree.
- Recognition of common SwiftUI stacks, split views, controls, conditionals,
  loops, modifiers, and custom component references.
- WinUI 3 XAML fragment generation.
- Text and JSON analysis reports.
- A conservative XAML parity scan for fixed layout dimensions, scroll regions,
  generated component coverage, and unsupported SwiftUI constructs.
- Tests and a Voci integration profile.

## Build and test

```sh
swift build
swift test
```

The executable is available as `.build/debug/loom` after a debug build.

## Commands

Analyze a SwiftUI view:

```sh
swift run loom analyze MyView.swift --root-view ContentView
swift run loom analyze MyView.swift --root-view ContentView --format json
```

Analyze a computed subview property:

```sh
swift run loom analyze MyView.swift \
  --root-view ContentView \
  --component operatorTopBar
```

Generate a reviewable WinUI 3 XAML fragment:

```sh
swift run loom generate MyView.swift \
  --root-view ContentView \
  --output Generated/ContentView.xaml
```

Projects with a consistent XAML resource prefix can retain their theme tokens:

```sh
swift run loom generate MyView.swift \
  --root-view ContentView \
  --theme-prefix Voci \
  --output Generated/ContentView.xaml
```

Compare SwiftUI-derived structure with an existing XAML file:

```sh
swift run loom parity MyView.swift \
  --root-view ContentView \
  --xaml MainWindow.xaml
```

Run the Voci example from this repository:

```sh
./Examples/Voci/analyze-voci.sh /private/SDF/Voci
```

## Translation policy

Loom translates intent, not spelling. For example, `HStack` does not always
become `StackPanel`: a stack containing `Spacer`, flexible frames, or competing
minimum widths needs a `Grid` with star-sized columns. `ZStack` naturally maps
to layered Grid content. `ForEach` normally becomes a collection control and
data template rather than repeated literal elements.

Every loss of semantics must result in a diagnostic or an explanatory XAML
comment. Silent approximation is treated as a bug.

## Repository structure

- `Sources/LoomCore`: parser, layout model, XAML emitter, and parity checker.
- `Sources/LoomCLI`: command-line interface.
- `Tests/LoomCoreTests`: unit and integration-style fixtures.
- `Examples/Voci`: configuration and commands for the sibling Voci project.
- `docs`: architecture and translation support matrix.

## Roadmap

1. Resolve custom computed subviews recursively across Swift files.
2. Add a configurable semantic-token and control mapping registry.
3. Generate `x:Bind` view-model contracts and C++/WinRT event stubs.
4. Add syntax-aware incremental XAML regions instead of whole-file replacement.
5. Build Windows-hosted XAML compilation and screenshot comparison adapters.
6. Add equivalent emitters for other declarative desktop UI targets.
