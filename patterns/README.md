# loom patterns

patterns are loom's canonical, operating-system-independent definitions of layout and
control intent.

Each `*.pattern.json` file is a metadata contract for how a UI construct should be
understood before any target mapping. patterns define:

- stable ID and kind
- intent, lifecycle context, and category
- child structure, sizing, ordering, and constraints
- typed attributes with defaults and ranges
- optional accessibility metadata and variant policy profiles
- optional platform mappings (`winui3`, `swiftui`/`macos`, etc.)
- Qt mappings for Linux/Qt transfer planning.

The `.pattern.json` files are the source of truth. Export formats are derived.

## canonical operations

- `loom patterns:list [--directory patterns] [--json]`
- `loom patterns:show <id> [--directory patterns] [--output path]`
- `loom patterns:validate [--directory patterns] [--json]`
- `loom patterns:lint [--directory patterns] [--json]`
- `loom patterns:export [--directory patterns] [--format loom|dtcg|open-ui|aria|style-dictionary]`

## transfer and audit entry points

- `loom patterns:transfer mainwindow.xaml --from winui3 --to macos`
- `loom patterns:transfer contentview.swift --from swiftui --to windows`
- `loom patterns:transfer mainwindow.qml --from qt --to windows`
- `loom inspect:parity contentview.swift --target mainwindow.qml --from swiftui --to qt --json`
- `loom accessibility:audit mainwindow.xaml --json`
- `loom inspect:errors mainwindow.xaml --kind xaml --json`

## export formats

- `loom patterns:export --format dtcg`
- `loom patterns:export --format open-ui`
- `loom patterns:export --format aria`
- `loom patterns:export --format style-dictionary`

Use exported formats for downstream tooling. Do not mutate canonical
`*.pattern.json` files from exported output.
