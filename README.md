# Loom

Current version: **0.17.0**

Loom is a compiler-assisted bridge for transferring interface layout design
between SwiftUI and WinUI 3 XAML. It lowers source UI into a platform-neutral
layout model, matches elements to OS-agnostic Patterns, reports transfer risks,
and emits reviewable target scaffolds where support exists.

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

Loom is switching to Go for cross-platform distribution on macOS, Windows, and
Linux. The Go CLI is the forward runtime starting in 0.17.0. The Swift
implementation remains in the repository as the tagged 0.16.0 reference while
the SwiftUI parser and target emitters are ported.

The current Go slice includes:

- Cross-platform command runtime in Go.
- Pattern catalog listing, showing, validation, linting, and export.
- WinUI XAML ingestion into the platform-neutral layout tree.
- ASCII Pattern rendering for XAML input.
- Pattern transfer planning from WinUI XAML to SwiftUI.
- Accessibility and layout design auditing for XAML input, including
  unsupported native WinUI component boundaries.
- Curated OS/framework error suggestions.
- Global `--quiet` and `--verbose` runtime output controls.
- JSON output for automation and AI-agent workflows.

The Swift reference implementation still includes:

- Swift parsing through SwiftParser and SwiftSyntax.
- Discovery of Swift structs conforming to `View`.
- Extraction of `body` or another computed view property.
- A Codable platform-neutral layout tree.
- Recognition of common SwiftUI stacks, split views, controls, conditionals,
  loops, modifiers, and custom component references.
- Recursive component graph discovery for same-view computed subviews and
  custom `View` structs across Swift source directories.
- WinUI 3 XAML fragment generation.
- WinUI target contract generation for native bindings, actions, lifecycle
  hooks, visibility rules, accessibility metadata, theme resources,
  collections, components, and unsupported behavior.
- Text and JSON analysis reports for SwiftUI and WinUI XAML sources.
- WinUI XAML ingestion into the same platform-neutral layout tree used by the
  SwiftUI frontend.
- SwiftUI scaffold generation from WinUI XAML through the shared layout tree.
- Safe owned XAML region replacement for generated output inside existing
  handwritten files.
- Guarded self-healing for missing generated-region host files through explicit
  `--init-region`.
- Global `--quiet` and `--verbose` runtime output controls for automation.
- Vigil-inspired read-only diagnostics: `status`, `verify`,
  `checks:command-catalog`, `guards:summary`, and `self-heal:plan`.
- Error-focused inspection for Swift syntax, Loom analysis, XAML, manifest, and
  Pattern issues through `inspect:errors`.
- A dedicated [AI Agents guide](docs/AI_AGENTS.md) for safe command sequencing,
  JSON output, write guards, and translation boundaries.
- A conservative XAML parity scan for fixed layout dimensions, scroll regions,
  generated component coverage, and unsupported SwiftUI constructs.
- Tests and a Voci integration profile.
- A Vigil-style command registry with grouped namespaced commands, access
  markers, aliases, JSON catalog output, and per-command manuals.
- Manifest validation and one-command multi-component project builds.
- A versioned, validated Patterns catalog defining OS-agnostic layout and
  control semantics independently of SwiftUI and WinUI spellings.
- An operational Pattern registry and stricter lint gate for bidirectional
  platform mappings.
- Pattern export for Loom-native JSON plus DTCG-style tokens, Open UI-style
  component metadata, ARIA accessibility summaries, and Style
  Dictionary-compatible token packages.
- Pattern transfer planning that classifies each layout element as direct,
  policy-dependent, native-contract-dependent, lossy, or unsupported when
  moving between SwiftUI and WinUI.
- ASCII Pattern rendering for a compact plaintext view of the layout shape.
- Optional Pattern variants for compact, dense, accessibility, or adaptive
  layout policies.
- Accessibility and layout design auditing for missing names, unlabeled
  images/inputs, undersized targets, redundant wrappers, malformed nodes,
  nested scroll regions, color-only semantics, and geometry-heavy layouts.
- Optional Pattern accessibility metadata for keyboard behavior, states,
  required accessibility properties, and minimum target sizes.
- Explicit unsupported component-boundary warnings for native WinUI controls
  that Loom preserves but cannot semantically transfer yet.
- Structured suggested fixes in accessibility audit output for both users and
  AI agents.
- Curated OS/framework error suggestions for SwiftUI, WinUI, XAML, macOS, and
  Windows errors, also attached automatically to `inspect:errors` findings.

## Build and test

```sh
go build ./cmd/loom
go test ./...
```

The executable can be built anywhere Go runs:

```sh
go build -o loom ./cmd/loom
```

The Swift reference implementation can still be built and tested:

```sh
swift build
swift test
```

The Swift executable is available as `.build/debug/loom` after a debug build.

## Commands

Loom follows Vigil's command organization: canonical commands use
`type:operation` names, commands are grouped by purpose, and the catalog marks
read-only (`r`), writing (`w`), and conditional-writing (`r/w`) operations.

```sh
go run ./cmd/loom list
go run ./cmd/loom list --category inspection
go run ./cmd/loom list --json
go run ./cmd/loom help patterns:transfer
```

Run read-only diagnostics:

```sh
go run ./cmd/loom status
go run ./cmd/loom verify --json
```

The Swift reference retains additional diagnostics such as
`checks:command-catalog`, `guards:summary`, and `self-heal:plan` until those are
ported.

Inspect errors before generating:

```sh
swift run loom inspect:errors MyView.swift --root-view ContentView
swift run loom inspect:errors MainWindow.xaml --kind xaml --json
swift run loom inspect:errors Patterns --kind patterns --fail-on error
```

Get OS/framework-specific suggestions:

```sh
go run ./cmd/loom suggestions:os-errors --platform winui3 --message StaticResource
go run ./cmd/loom suggestions:os-errors --platform swiftui --format json
```

`inspect:errors` also attaches `suggested_fixes` to findings where Loom can
recognize a relevant SwiftUI, WinUI, XAML, macOS, or Windows error pattern.

Audit accessibility and layout design quality:

```sh
go run ./cmd/loom accessibility:audit MainWindow.xaml --format json
go run ./cmd/loom accessibility:audit MainWindow.xaml --fail-on warning
```

The audit catches missing accessible names, unlabeled images and inputs,
small interactive targets, color-only semantic risks, unsupported/malformed
nodes, empty or redundant containers, nested scroll regions, geometry-dependent
layout, unsupported native WinUI component boundaries, and other
transfer-hostile design issues. Each finding includes a prose recommendation
and structured suggested fixes for users and AI agents.

Global runtime flags can be placed before any command:

```sh
go run ./cmd/loom --quiet inspect:ascii MainWindow.xaml --output Generated/View.txt
go run ./cmd/loom --verbose inspect:xaml MainWindow.xaml --output Generated/View.json --json
```

`--quiet` suppresses successful write chatter. Fatal errors and hints still go
to stderr. `--verbose` adds write details on stderr; `--quiet` wins if both are
provided.

Inspect and validate semantic patterns:

```sh
go run ./cmd/loom patterns:list
go run ./cmd/loom patterns:show split-view
go run ./cmd/loom patterns:validate
go run ./cmd/loom patterns:lint
go run ./cmd/loom patterns:export --format dtcg --output Generated/loom.tokens.json
```

The root [`Patterns`](Patterns) directory contains one precise metadata file
for each meaningful layout or control kind Loom currently recognizes. Each
file defines intent, child and sizing semantics, typed attributes, valid
ranges, constraints, accessibility behavior, and platform mappings. The
pattern remains canonical; platform mappings are implementations of it.
`patterns:export` provides integration-oriented JSON views for design-token
pipelines, component inventories, accessibility review, and design-system
tooling without changing the canonical `.pattern.json` source files.

Analyze a SwiftUI view:

```sh
swift run loom inspect:source MyView.swift --root-view ContentView
swift run loom inspect:source MyView.swift --root-view ContentView --json
```

Render an ASCII Pattern tree:

```sh
go run ./cmd/loom inspect:ascii MainWindow.xaml
```

The ASCII Pattern is a plaintext structural sketch using simple tree
characters such as `|`, `-`, `=`, and `\`. It is useful in reviews, agent logs,
and diffs where screenshots or rich rendering are inappropriate.

Plan interface transfer across platforms:

```sh
go run ./cmd/loom patterns:transfer MainWindow.xaml --from winui3 --to swiftui
go run ./cmd/loom patterns:transfer MainWindow.xaml --from winui3 --to swiftui --format json
```

The transfer report matches each node to its canonical Pattern, checks the
target mapping, includes an ASCII Pattern, and separates clean layout transfer
from policy decisions, native behavior contracts, lossy semantics, and missing
target support.

Analyze a WinUI XAML view:

```sh
go run ./cmd/loom inspect:xaml MainWindow.xaml
go run ./cmd/loom inspect:xaml MainWindow.xaml --format json
```

Generate a SwiftUI scaffold from WinUI XAML:

```sh
swift run loom generate:swiftui MainWindow.xaml --view-name MainWindowScaffold
swift run loom generate:swiftui MainWindow.xaml \
  --view-name MainWindowScaffold \
  --output Generated/MainWindowScaffold.swift
```

Analyze a computed subview property:

```sh
swift run loom inspect:source MyView.swift \
  --root-view ContentView \
  --component operatorTopBar
```

Discover reachable SwiftUI layout components:

```sh
swift run loom graph:components Sources/App --root-view ContentView
swift run loom graph:components Sources/App \
  --root-view ContentView \
  --format dot \
  --output Generated/component-graph.dot
```

`graph:components` scans Swift files under the supplied path, starts from
`RootView.body` by default, resolves lower-case references that match computed
view properties on the current view, resolves upper-case references that match
custom SwiftUI `View` structs, detects cycles, and reports unresolved custom
views. Use repeated `--include` and `--exclude` globs to constrain large source
sets deterministically.

Generate a reviewable WinUI 3 XAML fragment:

```sh
swift run loom generate:xaml MyView.swift \
  --root-view ContentView \
  --output Generated/ContentView.xaml
```

Generate WinUI target contracts for behavior Loom intentionally leaves native:

```sh
swift run loom generate:contracts MyView.swift --root-view ContentView
swift run loom generate:contracts MyView.swift \
  --root-view ContentView \
  --theme-prefix Voci \
  --format json \
  --output Generated/ContentView.contracts.json
```

Contracts are the companion checklist for native Windows implementation:
bindings, button actions, lifecycle handlers, resource tokens, visibility
rules, collections, accessibility metadata, component boundaries, and
unsupported expressions.

Update a Loom-owned region inside an existing XAML file:

```xml
<!-- LOOM-BEGIN shell.main -->
<!-- generated content lives here -->
<!-- LOOM-END shell.main -->
```

```sh
swift run loom generate:xaml MyView.swift \
  --root-view ContentView \
  --replace-region MainWindow.xaml \
  --region-id shell.main
```

Loom refuses to write unless exactly one matching begin/end marker pair exists.
Handwritten XAML outside the marked region is left untouched.

For first-time generated host files only, opt into guarded initialization:

```sh
swift run loom generate:xaml MyView.swift \
  --root-view ContentView \
  --replace-region Generated/Shell.xaml \
  --region-id shell.main \
  --init-region
```

`--init-region` creates a missing XAML file with one Loom-owned region. It does
not add markers to an existing unmarked file.

Trace which OS-agnostic patterns drove emitted XAML nodes:

```sh
swift run loom generate:xaml MyView.swift \
  --root-view ContentView \
  --patterns-dir Patterns \
  --pattern-comments
```

Projects with a consistent XAML resource prefix can retain their theme tokens:

```sh
swift run loom generate:xaml MyView.swift \
  --root-view ContentView \
  --theme-prefix Voci \
  --output Generated/ContentView.xaml
```

Compare SwiftUI-derived structure with an existing XAML file:

```sh
swift run loom inspect:parity MyView.swift \
  --root-view ContentView \
  --xaml MainWindow.xaml
```

Run the Voci example from this repository:

```sh
./Examples/Voci/analyze-voci.sh /private/SDF/Voci
```

Validate and build any configured project directly:

```sh
swift run loom config:validate path/to/loom.json --project-root path/to/project
swift run loom project:build path/to/loom.json \
  --project-root path/to/project \
  --output-dir Generated/Loom
```

`project:build` analyzes every component named in the manifest before writing
anything. A successful run emits per-component analysis JSON and XAML, an
optional parity report, and `project.summary.json`.

A minimal versioned manifest looks like:

```json
{
  "schema_version": "1",
  "project": "MyApp",
  "source": "platform/macos/MyApp.swift",
  "rootView": "ContentView",
  "target": "winui3",
  "themeResourcePrefix": "MyApp",
  "existingXaml": "platform/windows/MainWindow.xaml",
  "components": ["body", "sidebar", "contentPanel"]
}
```

Relative paths resolve from `--project-root`, or from the manifest directory
when no project root is supplied. Run `loom config:schema` for the complete
machine-readable contract.

## Translation policy

Loom translates intent, not spelling. The same Pattern vocabulary is intended
to support SwiftUI-to-XAML and XAML-to-SwiftUI workflows; each platform is a
mapping of a shared semantic pattern, not the canonical definition. For
example, `HStack` does not always
become `StackPanel`: a stack containing `Spacer`, flexible frames, or competing
minimum widths needs a `Grid` with star-sized columns. `ZStack` naturally maps
to layered Grid content. `ForEach` normally becomes a collection control and
data template rather than repeated literal elements.

Every loss of semantics must result in a diagnostic or an explanatory XAML
comment. SwiftUI scaffolds generated from XAML preserve structure first:
platform behavior, commands, bindings, styles, and unsupported controls are
left as reviewable placeholders or comments. Silent approximation is treated
as a bug.

## Repository structure

- `cmd/loom`: Go command-line entrypoint.
- `internal/loom`: Go runtime, Pattern catalog, XAML ingestion, audit,
  transfer, suggestions, and CLI dispatch.
- `Sources/LoomCore`: Swift reference parser, layout model, emitters, and
  parity checker.
- `Sources/LoomCLI`: Swift reference command-line interface.
- `Patterns`: OS-agnostic semantic UI definitions and their normative schema.
- `Tests/LoomCoreTests`: unit and integration-style fixtures.
- `Examples/Voci`: configuration and commands for the sibling Voci project.
- `docs`: architecture and translation support matrix.

See the [command organization](docs/COMMANDS.md) and the phased
[roadmap](docs/ROADMAP.md) for the current action plan.

## Roadmap

1. Port SwiftUI parsing and source diagnostics to Go.
2. Port XAML and SwiftUI emitters to Go.
3. Port project manifests, owned-region generation, and self-healing guards.
4. Expand Pattern-driven emission from trace comments into full mapping policy.
5. Generate target contract stubs for events and bindings.
6. Build Windows-hosted XAML compilation and screenshot comparison adapters.
