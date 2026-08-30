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
5. **Target lowering**: WinUI mapping decides whether a SwiftUI stack needs a
   `StackPanel`, `Grid`, collection control, layered container, or placeholder.
6. **XAML emission**: Loom emits a valid, reviewable fragment with comments at
   every semantic boundary that still needs a human decision.
7. **Parity checking**: source-derived constraints and component references are
   compared with an existing XAML surface. Later adapters can compile and render
   the result on Windows for image and accessibility-tree comparison.
8. **Project workflow**: a validated `loom.json` manifest selects components,
   target resources, existing XAML, and deterministic output artifacts.

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

## Component Graphs

`graph:components` builds a deterministic dependency graph over reachable
SwiftUI layout components. Nodes are identified as `View.component`, where the
component is usually `body` or a computed view property. Lower-case component
references only become graph edges when they match a computed property on the
current view, which prevents action calls from becoming visual dependencies.
Upper-case references become edges when they match a custom SwiftUI `View` with
an extractable `body`.

## Trust boundaries

- SwiftSyntax validates and structures Swift declarations. The extracted
  SwiftUI DSL remains intentionally narrower than the complete Swift grammar.
- Generated XAML is an input to review and Windows compilation, not proof of
  visual or behavioral parity.
- Existing platform code is never overwritten by `analyze` or `parity`.
- `generate` writes only when an explicit output path is supplied by the user.

## Incremental generation direction

Whole-file generation is unsuitable for mature native applications. Loom will
use named generated regions or generated UserControls, leaving event routing,
native menus, media surfaces, and platform services in handwritten files. A
future merge engine must parse XAML and refuse to alter unowned nodes.
