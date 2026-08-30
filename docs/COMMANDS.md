# Command organization

Starting in 0.17.0, Loom is migrating to Go as the forward cross-platform CLI.
Commands already available in the Go runtime are marked `Go`. Commands still
available only in the Swift 0.16 reference are marked `Swift ref`.

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
| `accessibility` | Audit accessible semantics and layout design quality. |
| `inspection` | Read SwiftUI structure and compare existing XAML. |
| `generation` | Lower one component to a reviewable XAML fragment. |
| `projects` | Run manifest-driven multi-component translation workflows. |
| `diagnostics` | Report local readiness, command metadata, write guards, and self-healing actions. |
| `patterns` | Inspect and validate OS-agnostic UI semantics. |
| `suggestions` | Show user/agent fixes for OS and framework errors. |
| `setup` | Describe and validate Loom project configuration. |

## Canonical commands and compatibility aliases

| Canonical command | Runtime | Access | Alias |
| --- | --- | --- | --- |
| `accessibility:audit` | Go | `r/w` through `--output` | `a11y` |
| `inspect:source` | Swift ref | `r/w` through `--output` | `analyze` |
| `inspect:xaml` | Go | `r/w` through `--output` | `xaml` |
| `inspect:ascii` | Go | `r/w` through `--output` | `ascii` |
| `inspect:errors` | Swift ref | `r/w` through `--output` | `errors` |
| `inspect:parity` | Swift ref | `r/w` through `--output` | `parity` |
| `graph:components` | Swift ref | `r/w` through `--output` | `graph` |
| `generate:swiftui` | Swift ref | `r/w` through `--output` | — |
| `generate:xaml` | Swift ref | `r/w` through `--output`, `--replace-region` | `generate` |
| `generate:contracts` | Swift ref | `r/w` through `--output` | `contracts` |
| `project:build` | Swift ref | `w` | `project` |
| `status` | Go | `r` | — |
| `verify` | Go | `r` | — |
| `checks:command-catalog` | Swift ref | `r` | — |
| `guards:summary` | Swift ref | `r` | — |
| `self-heal:plan` | Swift ref | `r` | — |
| `patterns:list` | Go | `r/w` through `--output` | — |
| `patterns:show` | Go | `r/w` through `--output` | — |
| `patterns:validate` | Go | `r/w` through `--output` | — |
| `patterns:lint` | Go | `r/w` through `--output` | — |
| `patterns:export` | Go | `r/w` through `--output` | — |
| `patterns:transfer` | Go | `r/w` through `--output` | `transfer` |
| `suggestions:os-errors` | Go | `r/w` through `--output` | `os-errors` |
| `config:validate` | Swift ref | `r` | — |
| `config:schema` | Swift ref | `r` | — |

Aliases preserve the 0.1 command surface, but documentation and automation
should use canonical names.

## Diagnostics commands

Loom adopts a small subset of Vigil's operational command shape for local
readiness and policy visibility:

```sh
loom status
loom status --patterns-dir Patterns --json
loom verify
```

`status` reports version, command count, and Pattern catalog health. `verify`
runs Pattern validation/lint and command availability checks.

The Swift reference also retains:

```sh
loom checks:command-catalog --json
loom guards:summary
loom self-heal:plan
```

## Accessibility audit command

`accessibility:audit` checks the normalized layout tree for accessibility and
design-transfer problems:

```sh
loom accessibility:audit ContentView.swift --root-view ContentView
loom accessibility:audit MainWindow.xaml --format json
loom accessibility:audit ContentView.swift --fail-on warning
```

It reports missing button names, unlabeled images, inputs relying only on
placeholders, undersized interactive targets, color-only semantic risks,
unsupported or malformed nodes, empty containers, redundant wrappers, repeated
nested layout wrappers, nested scroll regions, geometry-dependent layouts, and
unsupported native WinUI component boundaries.

Use `--fail-on error` or `--fail-on warning` in automation. By default the
command reports findings but exits successfully.

Each finding includes `recommendation` plus `suggested_fixes`. Suggested fixes
are split by audience:

- `user`: product/design decision or review action;
- `agent`: concrete implementation or follow-up command guidance.

## XAML inspection command

`inspect:xaml` parses WinUI XAML and normalizes supported elements into Loom's
shared layout tree. It is the ingestion half of the future Windows-to-macOS
translation path.

```sh
loom inspect:xaml MainWindow.xaml
loom inspect:xaml MainWindow.xaml --format json --output mainwindow.analysis.json
```

## ASCII Pattern command

`inspect:ascii` renders SwiftUI or WinUI XAML as a plaintext layout tree. It is
intended for reviews, terminal output, agent logs, and diffs.

```sh
loom inspect:ascii ContentView.swift --root-view ContentView
loom inspect:ascii ContentView.swift --root-view ContentView --output layout.ascii.txt
loom inspect:ascii MainWindow.xaml
```

The structural tree uses plain ASCII markers such as `=`, `|--`, and `\--`.

## Error inspection command

`inspect:errors` reports syntax and Loom-specific diagnostics without requiring
generation:

```sh
loom inspect:errors ContentView.swift --root-view ContentView
loom inspect:errors MainWindow.xaml --kind xaml --json
loom inspect:errors loom.json --kind manifest
loom inspect:errors Patterns --kind patterns --fail-on error
```

Supported kinds are `swift`, `xaml`, `manifest`, and `patterns`. If `--kind` is
omitted, Loom infers it from the path. `--fail-on none|error|warning` controls
automation exit behavior; default inspection does not fail just because it
found errors.

Findings include `suggested_fixes` when Loom recognizes a relevant OS/framework
error pattern.

## OS error suggestions command

`suggestions:os-errors` lists curated fix guidance for common SwiftUI, WinUI,
XAML, macOS, and Windows errors:

```sh
loom suggestions:os-errors
loom suggestions:os-errors --platform winui3 --message StaticResource
loom suggestions:os-errors --platform swiftui --format json
```

The catalog is intended for user and AI-agent workflows. It is based on common
platform failure modes such as Swift parser/result-builder errors, XAML parse
exceptions, unresolved XAML resources, WinUI AutomationProperties naming,
AccessibilityView tree exposure, custom automation peers, native component
boundaries, and binding/data-context gaps.

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

`patterns:export` converts the canonical `.pattern.json` files into
integration-oriented JSON shapes:

```sh
loom patterns:export --format loom
loom patterns:export --format dtcg --output loom.tokens.json
loom patterns:export --format open-ui
loom patterns:export --format aria
loom patterns:export --format style-dictionary
```

Supported formats are `loom`, `dtcg`, `open-ui`, `aria`, and
`style-dictionary`. Pattern commands print to stdout by default and write only
when `--output` is supplied.

`patterns:transfer` plans the movement of an interface layout between
platforms:

```sh
loom patterns:transfer ContentView.swift --from swiftui --to winui3
loom patterns:transfer MainWindow.xaml --from winui3 --to swiftui --format json
```

It parses the source into Loom IR, matches each node to a canonical Pattern,
checks whether the target platform has a mapping, embeds an ASCII Pattern, and
classifies each item as `direct`, `needs-policy`, `needs-native-contract`,
`lossy`, or `unsupported`.

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

## Target contract command

`generate:contracts` emits the native WinUI contract surface that must be wired
around generated layout. It reports bindings, button actions, lifecycle
handlers, visibility rules, collection sources/templates, accessibility
metadata, theme resources, component boundaries, and unsupported behavior.

```sh
loom generate:contracts ContentView.swift --root-view ContentView
loom generate:contracts ContentView.swift --root-view ContentView --format json --output ContentView.contracts.json
```

Use this beside `generate:xaml`: XAML provides the reviewable layout fragment;
contracts provide the target-side implementation checklist.
