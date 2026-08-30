# Command organization

Loom follows the command ergonomics established by Vigil without copying its
repository-policy domain. The shared conventions are:

- Canonical commands use `type:operation` names.
- A central registry owns command names, descriptions, categories, access
  modes, write flags, aliases, synopsis text, and examples.
- `loom list` groups commands by purpose and places setup last.
- `loom list --json` exposes the same registry as a stable machine contract.
- `loom help <command>` and `loom explain <command>` resolve through the same
  metadata used by dispatch.
- Access markers make filesystem behavior visible: `r` is read-only, `w`
  writes, and `r/w` writes only when a declared flag such as `--output` is used.
- Global runtime flags can be placed before any command: `--quiet`/`-q`
  suppresses successful write chatter, while `--verbose`/`-v` reports write
  details on stderr. Fatal errors still print under quiet mode.

## Categories

| Category | Purpose |
| --- | --- |
| `inspection` | Read SwiftUI structure and compare existing XAML. |
| `generation` | Lower one component to a reviewable XAML fragment. |
| `projects` | Run manifest-driven multi-component translation workflows. |
| `patterns` | Inspect and validate OS-agnostic UI semantics. |
| `setup` | Describe and validate Loom project configuration. |

## Canonical commands and compatibility aliases

| Canonical command | Access | Alias |
| --- | --- | --- |
| `inspect:source` | `r/w` through `--output` | `analyze` |
| `inspect:xaml` | `r/w` through `--output` | `xaml` |
| `inspect:parity` | `r/w` through `--output` | `parity` |
| `graph:components` | `r/w` through `--output` | `graph` |
| `generate:xaml` | `r/w` through `--output`, `--replace-region` | `generate` |
| `generate:swiftui` | `r/w` through `--output` | — |
| `project:build` | `w` | `project` |
| `patterns:list` | `r` | — |
| `patterns:show` | `r` | — |
| `patterns:validate` | `r` | — |
| `patterns:lint` | `r` | — |
| `config:validate` | `r` | — |
| `config:schema` | `r` | — |

Aliases preserve the 0.1 command surface, but documentation and automation
should use canonical names.

## XAML inspection command

`inspect:xaml` parses WinUI XAML and normalizes supported elements into Loom's
shared layout tree. It is the ingestion half of the future Windows-to-macOS
translation path.

```sh
loom inspect:xaml MainWindow.xaml
loom inspect:xaml MainWindow.xaml --format json --output mainwindow.analysis.json
```

## Component graph command

`graph:components` discovers the reachable SwiftUI layout surface from a root
view without requiring every computed subview in a manifest. It accepts a Swift
file or source directory, resolves same-view computed properties and custom
`View` structs, and emits text, JSON, or DOT.

```sh
loom graph:components Sources/App --root-view ContentView
loom graph:components Sources/App --root-view ContentView --format dot --output graph.dot
```

Use repeated `--include glob` and `--exclude glob` options to constrain large
source trees. Hidden directories, `.git`, `.build`, `Build`, and `DerivedData`
are skipped by default.

## Pattern commands

`patterns:validate` checks structural correctness and semantic-kind coverage.
`patterns:lint` adds operational rules for mappings that can support both
SwiftUI-to-XAML and XAML-to-SwiftUI translation paths.

`generate:xaml --patterns-dir Patterns --pattern-comments` loads the Pattern
registry and annotates emitted nodes with the pattern id and WinUI mapping that
drove the output.

## Owned XAML regions

`generate:xaml --replace-region <xaml-file> --region-id <id>` updates only the
content between matching Loom markers:

```xml
<!-- LOOM-BEGIN shell.main -->
<!-- generated content lives here -->
<!-- LOOM-END shell.main -->
```

The command refuses to write if the marker pair is missing, duplicated, or out
of order. `--replace-region` cannot be combined with `--output`.

`--init-region` is an explicit self-healing mode for first-time generated host
files:

```sh
loom generate:xaml ContentView.swift \
  --replace-region Generated/Shell.xaml \
  --region-id shell.main \
  --init-region
```

It creates a missing XAML file containing one Loom-owned region. It refuses to
add markers to an existing unmarked file, because that would require guessing
where generated ownership should begin.

## SwiftUI generation command

`generate:swiftui` parses WinUI XAML through `inspect:xaml` and emits a
reviewable SwiftUI scaffold. It preserves supported layout structure and leaves
behavior, bindings, styles, and unsupported XAML controls as placeholders.

```sh
loom generate:swiftui MainWindow.xaml --view-name MainWindowScaffold
loom generate:swiftui MainWindow.xaml --view-name MainWindowScaffold --output MainWindowScaffold.swift
```
