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
4. **Target lowering**: WinUI mapping decides whether a SwiftUI stack needs a
   `StackPanel`, `Grid`, collection control, layered container, or placeholder.
5. **XAML emission**: Loom emits a valid, reviewable fragment with comments at
   every semantic boundary that still needs a human decision.
6. **Parity checking**: source-derived constraints and component references are
   compared with an existing XAML surface. Later adapters can compile and render
   the result on Windows for image and accessibility-tree comparison.

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
