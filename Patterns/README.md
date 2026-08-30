# Loom Patterns

Patterns are Loom's canonical, operating-system-independent definitions of UI
layout and control intent. A pattern describes what an element means before a
SwiftUI parser or WinUI emitter decides how to express it.

Every `*.pattern.json` file is independently versioned and declares:

- stable identity, semantic kind, lifecycle status, and category;
- intended and inappropriate uses;
- child, sizing, and ordering semantics;
- typed attributes with defaults, ranges, units, and enumerated values;
- cross-attribute constraints;
- accessibility behavior;
- non-canonical source and target mappings for each platform.
- optional named variants for compact, dense, accessibility, or adaptive
  layout policies.

Platform mappings are evidence about how an OS framework can represent the
pattern. They do not define the pattern itself.

Validate the complete catalog with:

```sh
swift run loom patterns:validate
```

Export derived integration views with:

```sh
swift run loom patterns:export --format dtcg
swift run loom patterns:export --format open-ui
swift run loom patterns:export --format aria
swift run loom patterns:export --format style-dictionary
```

The export formats are downstream compatibility shapes for design-token,
component-inventory, accessibility-review, and design-system tooling. The
`*.pattern.json` files remain canonical.

Plan OS-to-OS layout transfer with:

```sh
swift run loom patterns:transfer ContentView.swift --from swiftui --to winui3
swift run loom patterns:transfer MainWindow.xaml --from winui3 --to swiftui
```

Render the same normalized layout as an ASCII Pattern with:

```sh
swift run loom inspect:ascii ContentView.swift --root-view ContentView
```

The normative metadata contract is `pattern.schema.json`.
