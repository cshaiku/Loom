# Loom Patterns

Patterns are Loom’s canonical, operating-system-independent definitions of layout and
control intent.

Each `*.pattern.json` file is a metadata contract for how a UI construct should be
understood before any target mapping. Patterns define:

- stable ID and kind
- intent, lifecycle context, and category
- child structure, sizing, ordering, and constraints
- typed attributes with defaults and ranges
- optional accessibility metadata and variant policy profiles
- optional platform mappings (`winui3`, `swiftui`/`macos`, etc.)

The `.pattern.json` files are the source of truth. Export formats are derived.

## Canonical operations

- `loom patterns:list [--directory Patterns] [--json]`
- `loom patterns:show <id> [--directory Patterns] [--output path]`
- `loom patterns:validate [--directory Patterns] [--json]`
- `loom patterns:lint [--directory Patterns] [--json]`
- `loom patterns:export [--directory Patterns] [--format loom|dtcg|open-ui|aria|style-dictionary]`

## Transfer and audit entry points

- `loom patterns:transfer MainWindow.xaml --from winui3 --to macos`
- `loom accessibility:audit MainWindow.xaml --json`
- `loom inspect:errors MainWindow.xaml --kind xaml --json`

## Export formats

- `loom patterns:export --format dtcg`
- `loom patterns:export --format open-ui`
- `loom patterns:export --format aria`
- `loom patterns:export --format style-dictionary`

Use exported formats for downstream tooling. Do not mutate canonical
`*.pattern.json` files from exported output.
