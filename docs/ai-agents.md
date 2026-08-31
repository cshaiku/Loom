# AI Agents Guide for loom

Use this guide when working with the Go-only loom runtime.

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
- `loom patterns:transfer <swift-file> --from swiftui --to windows --json`

If these return errors, stop and either fix input or escalate to a human decision.

## Read/write controls

- `r` commands are read-only by catalog metadata.
- `r/w` commands can write only with write flags (commonly `--output`).
- Use `--quiet` for CI/automation and `--verbose` for extra write diagnostics.

examples:

```sh
loom --quiet inspect:errors mainwindow.xaml --kind xaml --json --fail-on error
loom --verbose patterns:validate --json
```

## Current supported generation path

Generation and project commands are currently catalog-compatible only and are not yet available in Go:

- `generate:xaml`
- `generate:swiftui`
- `generate:contracts`
- `project:build`
- `inspect:parity`
- `graph:components`

When these are available, treat them as write-capable and run with explicit output
targets.

## Diagnostics and suggestions

- `inspect:source` supports `.swift`, `.xaml`, and `.xml` inputs.
- `inspect:errors` supports `xaml`, `manifest`, `patterns`, and `swift` kinds.
- `config:validate` validates manifest shape and referenced files.
- `suggestions:os-errors` returns user/agent fix guidance and command suggestions
  when patterns match.

Use JSON output for downstream automations and preserve `platform`, `query`, and
`status` fields.

## Function-call contract

- Prefer `--json` for all machine calls.
- Treat `schema_version`, `status`, command-specific `summary` objects, and
  `suggested_fixes` as stable coordination fields.
- Empty lists are emitted as `[]` for implemented report surfaces.
- Use `--line-ending lf` for deterministic cross-OS artifacts, `--line-ending crlf`
  for Windows-facing text files, and `--line-ending native` only when matching
  the current host is explicitly useful.
- Reserved commands return a deterministic unavailable-command error; do not
  retry them as implementation steps.
- Use `--quiet` when also passing `--output` so successful writes do not add
  extra chatter around file artifacts.

## Command availability policy

- If a command is cataloged as a future-go feature (for example, any generation
  or parity command), stop execution and route the workflow through supported
  commands instead.
- Keep `--quiet` and `--json` for automation, and always preserve command output
  fields required for machine interpretation.

## pattern workflows

- Use `patterns:validate` and `patterns:lint` for semantic and operational checks.
- Use `patterns:transfer` to inspect portability risk before editing source layouts.
- Use `patterns:export` for downstream tooling formats (`dtcg`, `open-ui`, `aria`,
  `style-dictionary`).
