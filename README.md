# Loom

Version: **0.19.0**

Loom is a cross-platform Go CLI for UI layout analysis, Pattern catalog
validation, transfer planning, and workflow diagnostics.

## What Loom does today

- Parse WinUI XAML and normalize it into Loom's shared layout model (`inspect:xaml`).
- Preserve WinUI Grid row/column definitions as layout metadata for transfer planning.
- Render parsed layouts as a compact plaintext ASCII tree (`inspect:ascii`).
- Validate and lint the `Patterns` catalog (`patterns:validate`, `patterns:lint`).
- Transfer-plan layout compatibility from WinUI to macOS UI patterns (`patterns:transfer`).
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

The following commands are retained in the catalog for parity but are intentionally
not yet implemented in the Go runtime. They return a clear message when invoked:

- `inspect:source`
- `inspect:parity`
- `graph:components`
- `generate:xaml`
- `generate:swiftui`
- `generate:contracts`
- `project:build`

## Build and test

```sh
make build      # installs /opt/homebrew/bin/loom
make test       # or: go test ./...
```

## Command quickstart

```sh
loom list
loom help <command>
loom status --json
loom verify --json
loom checks:command-catalog --json
loom inspect:xaml MainWindow.xaml --format json
loom inspect:ascii MainWindow.xaml --output Layout.txt
loom inspect:ascii MainWindow.xaml --output Layout.txt --line-ending crlf
loom accessibility:audit MainWindow.xaml --format json --fail-on warning
loom patterns:lint
loom patterns:transfer MainWindow.xaml --from winui3 --to macos
```

## Repository layout

- `cmd/loom`: Go CLI entrypoint
- `internal/loom`: CLI runtime, parsers, catalog logic, and reports
- `Patterns`: OS-agnostic canonical layout and control pattern metadata
- `docs`: operating docs, commands, and guides
- `README`: project context and command surface

## Documentation

- [docs/COMMANDS.md](docs/COMMANDS.md)
- [docs/AI_AGENTS.md](docs/AI_AGENTS.md)
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)
- [docs/ROADMAP.md](docs/ROADMAP.md)

## Release checklist

1. Ensure `VERSION` and `internal/loom/catalog.go` version constants match.
2. Run `make test`.
3. Run `git add`, commit, and push.
4. Tag the release, e.g.:
   `git tag -a v0.19.0 -m "Line-ending controls and function output polish"; git push --tags`.
