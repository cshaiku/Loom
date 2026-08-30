# Loom architecture

## Design goal

Loom converts the layout semantics that SwiftUI and WinUI 3 can reasonably
share. It never hides platform behavior behind an apparently successful but
incorrect conversion.

## Pipeline

1. **Swift frontend**: SwiftParser produces a source-accurate SwiftSyntax tree.
   Loom uses it to identify real `View` declarations and their source regions.
2. **View expression extraction**: a conservative result-builder parser reads a
   selected computed view property and recognizes SwiftUI construction syntax.
3. **Layout IR**: a Codable tree records containers, controls, conditions,
   component references, raw arguments, modifiers, and diagnostics.
4. **Component graph**: recursive discovery starts from a root `View`
   component, resolves same-view computed properties and custom SwiftUI `View`
   structs across source files, and reports cycles or missing custom views.
5. **XAML frontend**: WinUI XAML is parsed as XML and normalized into the same
   `LoomNode` IR used by SwiftUI extraction.
6. **Target lowering**: WinUI mapping decides whether a SwiftUI stack needs a
   `StackPanel`, `Grid`, collection control, layered container, or placeholder.
7. **Target emission**: Loom emits valid, reviewable target scaffolds with
   comments at every semantic boundary that still needs a human decision.
   Current targets are WinUI XAML fragments and SwiftUI scaffolds from XAML.
8. **Target contracts**: behavior and state that cannot be safely translated as
   layout is emitted as a companion contract report for native WinUI wiring.
9. **Pattern transfer planning**: each normalized layout node is matched to a
   canonical Pattern and classified for OS-to-OS transfer readiness.
10. **ASCII Pattern rendering**: the normalized tree can be emitted as a compact
    plaintext structural sketch for reviews, logs, and diffs.
11. **Parity checking**: source-derived constraints and component references are
   compared with an existing XAML surface. Later adapters can compile and render
   the result on Windows for image and accessibility-tree comparison.
12. **Project workflow**: a validated `loom.json` manifest selects components,
   target resources, existing XAML, and deterministic output artifacts.
13. **Owned output**: generated XAML can be written back only inside explicit
    Loom marker pairs, preserving surrounding handwritten platform code.
14. **Runtime controls**: command success chatter is controlled globally by
    `--quiet` and `--verbose`, while fatal diagnostics always remain visible.

## Command architecture

Loom mirrors Vigil's command ergonomics at a smaller scale. A central catalog
defines canonical colon-namespaced commands, categories, access modes, write
flags, aliases, synopsis text, and examples. Help, list, JSON catalog, manuals,
and dispatch all resolve through that same metadata so the advertised surface
cannot drift independently from the executable.

## Semantic Patterns

`Patterns/*.pattern.json` is the canonical vocabulary between source parsing
and target emission. A pattern defines OS-agnostic intent, structure, sizing,
attributes, constraints, accessibility behavior, and stable identity. SwiftUI
and WinUI entries are explicitly mappings, not definitions of meaning. This is
direction-neutral by design: SwiftUI-to-XAML and XAML-to-SwiftUI should both
normalize through the same shared IR and Pattern vocabulary.

The catalog is governed by `Patterns/pattern.schema.json` and a typed runtime
validator. Every meaningful `LoomNodeKind` must have exactly one pattern;
synthetic roots and unsupported-source placeholders are intentionally excluded.
`patterns:lint` adds operational mapping rules, and the XAML emitter can load a
Pattern registry to trace which pattern drives an emitted WinUI node.

`patterns:export` derives interoperability views from the canonical Pattern
files. DTCG-style tokens, Open UI-style component metadata, ARIA summaries, and
Style Dictionary-compatible packages are export shapes for external tools;
they are not alternate sources of truth. Changes should be made in
`*.pattern.json` first and then re-exported.

Patterns may also declare optional variants. A variant describes a named
layout policy under conditions such as compact width, dense information mode,
large text, or accessibility-first rendering. Variants are intentionally
policy metadata, not platform-specific code.

## Pattern Transfer

`patterns:transfer` is the transferability layer between inspection and
generation. It answers whether the interface design can move from one platform
to another before code is emitted. Each visual node is classified:

- `direct`: target mapping exists and no extra policy or native contract was
  detected.
- `needs-policy`: target mapping exists, but spacing, sizing, typography,
  adaptive breakpoints, or tokens require a project decision.
- `needs-native-contract`: layout transfers, but state, actions, lifecycle,
  collections, accessibility metadata, or component boundaries need native
  implementation.
- `lossy`: target representation exists but platform-specific behavior may
  degrade without review.
- `unsupported`: no target mapping exists or the node is explicitly
  unsupported.

This makes Patterns operational: they are not only definitions of UI meaning,
but also evidence for safe interface transfer.

## ASCII Patterns

`inspect:ascii` renders the normalized layout tree as a plain text structure:

```text
= ContentView.body
\-- vertical-stack / VStack
    |-- text / Text
    \-- button / Button
```

The renderer uses simple ASCII connectors so the output works in terminals,
plain logs, code reviews, and agent messages without relying on screenshots.

## Component Graphs

`graph:components` builds a deterministic dependency graph over reachable
SwiftUI layout components. Nodes are identified as `View.component`, where the
component is usually `body` or a computed view property. Lower-case component
references only become graph edges when they match a computed property on the
current view, which prevents action calls from becoming visual dependencies.
Upper-case references become edges when they match a custom SwiftUI `View` with
an extractable `body`.

## XAML Ingestion

`inspect:xaml` converts WinUI controls into the same platform-neutral IR used
by `inspect:source`. The first supported set covers `Grid`, `StackPanel`,
`TextBlock`, `Button`, `TextBox`, `Image`, `ScrollViewer`, `ListView`, sliders,
toggles, dividers, and background borders. Original attributes are retained
under `xaml.*` properties so later emitters can preserve platform detail when
generating SwiftUI scaffolds.

## SwiftUI Scaffolds

`generate:swiftui` converts XAML-normalized IR into a SwiftUI source scaffold.
It maps supported containers and controls to SwiftUI equivalents, carries simple
frame and accessibility metadata, and preserves unsupported XAML controls as
comments plus `EmptyView()` placeholders. The output is intended for review and
native completion, not as proof that Windows behavior or styling has been
ported.

## Target Contracts

`generate:contracts` converts SwiftUI-normalized IR into a WinUI implementation
contract. The report intentionally separates layout from platform behavior:
dynamic text becomes a binding requirement, buttons become command/click
requirements, lifecycle modifiers become page/control lifecycle hooks,
conditionals become visibility or visual-state requirements, collections become
ItemsSource/DataTemplate requirements, and component references become
generated-region/UserControl decisions. This keeps generated XAML reviewable
without hiding missing native work.

## Trust boundaries

- SwiftSyntax validates and structures Swift declarations. The extracted
  SwiftUI DSL remains intentionally narrower than the complete Swift grammar.
- Generated XAML is an input to review and Windows compilation, not proof of
  visual or behavioral parity.
- Existing platform code is never overwritten by `analyze` or `parity`.
- `generate` writes only when an explicit output path is supplied by the user.
- `generate:xaml --replace-region` writes only between a single matching
  `LOOM-BEGIN` / `LOOM-END` marker pair and refuses ambiguous marker state.
- `--init-region` only creates a missing generated host file; existing files
  still require explicit markers before Loom will write into them.

## Incremental generation direction

Whole-file generation is unsuitable for mature native applications. Loom will
use named generated regions or generated UserControls, leaving event routing,
native menus, media surfaces, and platform services in handwritten files. A
future merge engine must parse XAML and refuse to alter unowned nodes.
