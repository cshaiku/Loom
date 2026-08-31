# loom commands

## Go runtime command surface

- `status`
- `verify`
- `checks:command-catalog`
- `guards:summary`
- `self-heal:plan`
- `config:validate`
- `config:schema`
- `accessibility:audit`
- `inspect:xaml`
- `inspect:ascii`
- `inspect:errors`
- `patterns:list`
- `patterns:show`
- `patterns:validate`
- `patterns:lint`
- `patterns:export`
- `patterns:transfer`
- `suggestions:os-errors`

## Placeholder commands retained in catalog

The following are not implemented in the Go runtime yet and return a clear
`not yet available` message:

- `inspect:source`
- `inspect:parity`
- `graph:components`
- `generate:xaml`
- `generate:swiftui`
- `generate:contracts`
- `project:build`

## Notes

- `--quiet` suppresses successful write confirmations.
- `--verbose` prints write diagnostics to stderr.
- `--line-ending lf|crlf|native` controls text output line endings. loom uses
  `lf` by default for deterministic artifacts.
- Most read/write report commands accept `--json`.
- Most commands accept `--help` for catalog-driven synopsis.

## examples

```sh
loom status --json
loom verify --json
loom checks:command-catalog --json
loom guards:summary
loom self-heal:plan
loom inspect:errors mainwindow.xaml --kind xaml --json --fail-on error
loom patterns:transfer mainwindow.xaml --from winui3 --to macos --format json
loom accessibility:audit mainwindow.xaml --fail-on warning
loom inspect:ascii mainwindow.xaml --output layout.txt --line-ending crlf
```
