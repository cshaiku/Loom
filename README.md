# loom

Version: **0.24.0**

loom is a cross-platform Go CLI for UI layout analysis, generation planning,
translation, pattern catalog validation, transfer planning, and workflow
diagnostics.

The product goal for `v1.0.0` is an analyzer, generator, and translator for
moving UI layout intent between SwiftUI, WinUI XAML, and Qt. The current
`0.24.0` release is the analyzer, transfer-planning, component-graph, and
analysis-only project build foundation. Generator commands are visible in the
catalog as stable targets, but they remain release-blocking work for `v1.0.0`.

## What Loom Does Today

- Parse WinUI XAML and normalize it into loom's shared layout model (`inspect:xaml`).
- Parse common SwiftUI layout/control constructs into the same shared model
  (`inspect:swiftui`, `inspect:source`).
- Parse common Qt QML, Qt Designer UI, and Qt C++ layout constructs into the
  same shared model (`inspect:qt`, `inspect:source`).
- Extract intrinsic font material properties from supplied TrueType, OpenType,
  TrueType Collection, and WOFF font files or installed family names
  (`inspect:font`).
- Preserve WinUI Grid row/column definitions as layout metadata for transfer planning.
- Render parsed layouts as a compact plaintext ASCII tree (`inspect:ascii`).
- Validate and lint the `patterns` catalog (`patterns:validate`, `patterns:lint`).
- Transfer-plan layout compatibility in both WinUI → macOS and macOS/SwiftUI →
  Windows directions (`patterns:transfer`).
- Compare layout parity across supported source dialects (`inspect:parity`).
- Compare profile-normalized visual metrics such as typography, spacing, control
  minimums, padding, margins, and sizing (`inspect:visual-parity`).
- Audit accessibility/layout quality for unsupported boundaries, small targets,
  malformed or redundant structures, and scan-friendly risks (`accessibility:audit`).
- Run manifest validation (`config:validate` / `config:schema`).
- Discover source-tree component dependencies (`graph:components`).
- Run manifest-directed analysis builds that write validation, analysis, graph,
  transfer, parity, and summary artifacts (`project:build`).
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

The following generator commands define the remaining planned generation surface
for `v1.0.0`. They are retained in the catalog for parity but are intentionally
not yet implemented in the Go runtime. They return a clear message when invoked:

- `generate:xaml`
- `generate:swiftui`
- `generate:contracts`

## Build And Test

```sh
make build      # installs /opt/homebrew/bin/loom on local Homebrew systems
make test       # runs go test ./...
go vet ./...    # static correctness check
```

## Command Quickstart

```sh
loom list
loom help <command>
loom status --json
loom verify --json
loom checks:command-catalog --json
loom inspect:source contentview.swift --json
loom inspect:swiftui contentview.swift --format json
loom inspect:qt mainwindow.qml --format json
loom inspect:font Inter.ttf --json
loom inspect:font --family "Segoe UI" --json
loom inspect:xaml mainwindow.xaml --format json
loom inspect:parity contentview.swift --target mainwindow.qml --from swiftui --to qt --json
loom inspect:visual-parity contentview.swift --target mainwindow.xaml --from swiftui --to winui3 --profile visual-profile.json --json
loom inspect:visual-parity contentview.swift --target mainwindow.xaml --source-font Inter.ttf --target-font-family "Segoe UI" --json
loom inspect:ascii mainwindow.xaml --output layout.txt
loom inspect:ascii mainwindow.xaml --output layout.txt --line-ending crlf
loom accessibility:audit mainwindow.xaml --format json --fail-on warning
loom graph:components examples/sampleapp --format dot --output component-graph.dot
loom patterns:lint
loom patterns:transfer mainwindow.xaml --from winui3 --to macos
loom patterns:transfer contentview.swift --from swiftui --to windows
loom patterns:transfer mainwindow.qml --from qt --to windows
loom project:build examples/sampleapp/loom.json --output-dir examples/sampleapp/generated/project-build --overwrite --json
```

## Example Workflow

```sh
./examples/sampleapp/analyze-sample-app.sh --overwrite
```

The sample produces analysis, audit, transfer, component graph, project build,
parity, and visual-parity reports under `examples/sampleapp/generated/`.

## Repository Layout

- `cmd/loom`: Go CLI entrypoint
- `internal/loom`: CLI runtime, parsers, catalog logic, and reports
- `patterns`: OS-agnostic canonical layout and control pattern metadata
- `examples/sampleapp`: neutral public sample XAML and analysis script
- `docs`: operating docs, commands, and guides
- `README`: project context and command surface

## Documentation

- [docs/commands.md](docs/commands.md)
- [docs/ai-agents.md](docs/ai-agents.md)
- [docs/architecture.md](docs/architecture.md)
- [docs/roadmap.md](docs/roadmap.md)
- [docs/support-policy.md](docs/support-policy.md)
- [docs/deprecations.md](docs/deprecations.md)
- [docs/release-checklist.md](docs/release-checklist.md)
- [CONTRIBUTING.md](CONTRIBUTING.md)
- [SECURITY.md](SECURITY.md)
- [TESTING.md](TESTING.md)

## License

Loom is open source under the 0BSD license. See [LICENSE](LICENSE).

## Release Checklist

1. Ensure `VERSION` and `internal/loom/catalog.go` version constants match.
2. Run `make test`, `go vet ./...`, and `go run ./cmd/loom verify --json`.
3. Run the sample workflow with `--overwrite`.
4. Review [TODO.md](TODO.md) before declaring `v1.0.0`.
5. Run `git add`, commit, and push.
6. Tag the release, e.g.:
   `git tag -a v0.24.0 -m "pre-1.0 analyzer project build release"; git push --tags`.
