# loom

Version: **0.22.0**

loom is a cross-platform Go CLI for UI layout analysis, pattern catalog
validation, transfer planning, and workflow diagnostics.

## what loom does today

- Parse WinUI XAML and normalize it into loom's shared layout model (`inspect:xaml`).
- Parse common SwiftUI layout/control constructs into the same shared model
  (`inspect:swiftui`, `inspect:source`).
- Parse common Qt QML, Qt Designer UI, and Qt C++ layout constructs into the
  same shared model (`inspect:qt`, `inspect:source`).
- Preserve WinUI Grid row/column definitions as layout metadata for transfer planning.
- Render parsed layouts as a compact plaintext ASCII tree (`inspect:ascii`).
- Validate and lint the `patterns` catalog (`patterns:validate`, `patterns:lint`).
- Transfer-plan layout compatibility in both WinUI → macOS and macOS/SwiftUI →
  Windows directions (`patterns:transfer`).
- Compare layout parity across supported source dialects (`inspect:parity`).
- Audit accessibility/layout quality for unsupported boundaries, small targets,
  malformed or redundant structures, and scan-friendly risks (`accessibility:audit`).
- Run manifest validation (`config:validate` / `config:schema`).
- Report errors for source analysis, XAML parsing, manifests, and patterns
  (`inspect:errors`).
- Provide curated cross-platform error guidance and suggested fixes (`suggestions:os-errors`).
- Provide CLI self-diagnostics and guard reporting (`status`, `verify`,
  `checks:command-catalog`, `guards:summary`, `self-heal:plan`).
- Write deterministic LF output by default, with `--line-ending crlf` for Windows
  artifacts and `--line-ending native` for host-native text output.

## patterns catalog

loom ships with its own `patterns` catalog. These files are public, neutral,
OS-agnostic examples that describe common interface elements such as buttons,
text, grids, lists, scroll regions, split views, toggles, and text input.

When you run loom from this repository, it uses `./patterns` by default. When
installed with `make build`, the same catalog is copied to
`/opt/homebrew/share/loom/patterns` so the installed `loom` command can validate,
lint, list, export, and transfer-plan against loom's own pattern definitions.

The following commands are retained in the catalog for parity but are intentionally
not yet implemented in the Go runtime. They return a clear message when invoked:

- `graph:components`
- `generate:xaml`
- `generate:swiftui`
- `generate:contracts`
- `project:build`

## build and test

```sh
make build      # installs /opt/homebrew/bin/loom
make test       # or: go test ./...
```

## command quickstart

```sh
loom list
loom help <command>
loom status --json
loom verify --json
loom checks:command-catalog --json
loom inspect:source contentview.swift --json
loom inspect:swiftui contentview.swift --format json
loom inspect:qt mainwindow.qml --format json
loom inspect:xaml mainwindow.xaml --format json
loom inspect:parity contentview.swift --target mainwindow.qml --from swiftui --to qt --json
loom inspect:ascii mainwindow.xaml --output layout.txt
loom inspect:ascii mainwindow.xaml --output layout.txt --line-ending crlf
loom accessibility:audit mainwindow.xaml --format json --fail-on warning
loom patterns:lint
loom patterns:transfer mainwindow.xaml --from winui3 --to macos
loom patterns:transfer contentview.swift --from swiftui --to windows
loom patterns:transfer mainwindow.qml --from qt --to windows
```

## repository layout

- `cmd/loom`: Go CLI entrypoint
- `internal/loom`: CLI runtime, parsers, catalog logic, and reports
- `patterns`: OS-agnostic canonical layout and control pattern metadata
- `examples/sampleapp`: neutral public sample XAML and analysis script
- `docs`: operating docs, commands, and guides
- `README`: project context and command surface

## documentation

- [docs/commands.md](docs/commands.md)
- [docs/ai-agents.md](docs/ai-agents.md)
- [docs/architecture.md](docs/architecture.md)
- [docs/roadmap.md](docs/roadmap.md)

## release checklist

1. Ensure `VERSION` and `internal/loom/catalog.go` version constants match.
2. Run `make test`.
3. Run `git add`, commit, and push.
4. Tag the release, e.g.:
   `git tag -a v0.22.0 -m "qt inspection and cross-platform parity"; git push --tags`.
