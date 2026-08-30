# AI Agents Guide for Loom

Use this guide when working with the Go-only Loom runtime.

## Preferred workflow

1. `loom status --json`
2. `loom verify --json`
3. `loom checks:command-catalog --json`
4. `loom guards:summary --json`
5. `loom list --json`
6. Run command-specific checks for the artifact in scope.

## Safe error gate

Before touching files:

- `loom inspect:errors <path> --json --fail-on error`
- `loom accessibility:audit <xaml-file> --json --fail-on warning`

If these return errors, stop and either fix input or escalate to a human decision.

## Read/write controls

- `r` commands are read-only by catalog metadata.
- `r/w` commands can write only with write flags (commonly `--output`).
- Use `--quiet` for CI/automation and `--verbose` for extra write diagnostics.

Examples:

```sh
loom --quiet inspect:errors MainWindow.xaml --kind xaml --json --fail-on error
loom --verbose patterns:validate --json
```

## Current supported generation path

Generation and project commands are currently catalog-compatible only and are not yet available in Go:

- `generate:xaml`
- `generate:swiftui`
- `generate:contracts`
- `project:build`
- `inspect:source`
- `inspect:parity`
- `graph:components`

When these are available, treat them as write-capable and run with explicit output
targets.

## Diagnostics and suggestions

- `inspect:errors` supports `xaml`, `manifest`, `patterns`, and `swift` kinds.
- `config:validate` validates manifest shape and referenced files.
- `suggestions:os-errors` returns user/agent fix guidance and command suggestions
  when patterns match.

Use JSON output for downstream automations and preserve `platform`, `query`, and
`Status` fields.

## Command availability policy

- If a command is cataloged as a future-go feature (for example, any generation
  or parity command), stop execution and route the workflow through supported
  commands instead.
- Keep `--quiet` and `--json` for automation, and always preserve command output
  fields required for machine interpretation.

## Pattern workflows

- Use `patterns:validate` and `patterns:lint` for semantic and operational checks.
- Use `patterns:transfer` to inspect portability risk before editing source layouts.
- Use `patterns:export` for downstream tooling formats (`dtcg`, `open-ui`, `aria`,
  `style-dictionary`).
