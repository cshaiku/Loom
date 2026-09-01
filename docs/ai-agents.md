# AI Agents Guide For Loom

Use this guide when working with the Go-only Loom runtime.

Loom's `v1.0.0` target is an analyzer, generator, and translator. The current
`0.24.0` runtime implements the analyzer, transfer-planning, component graph,
and analysis-only project build foundation.

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
- `loom inspect:parity <swift-file> --target <qt-file> --from swiftui --to qt --json`
- `loom graph:components <source-dir> --json`

If these return errors, stop and either fix input or escalate to a human decision.

## Read/write controls

- `r` commands are read-only by catalog metadata.
- `r/w` commands can write only with write flags (commonly `--output`).
- Existing outputs require `--overwrite`; never assume replacement is allowed.
- Use `--quiet` for CI/automation and `--verbose` for extra write diagnostics.

examples:

```sh
loom --quiet inspect:errors mainwindow.xaml --kind xaml --json --fail-on error
loom --verbose patterns:validate --json
```

## Current Generation Path

Generation commands are catalog-compatible only and are not yet available in Go:

- `generate:xaml`
- `generate:swiftui`
- `generate:contracts`

Use `project:build` for manifest-directed analysis bundles. It writes outputs to
`--output-dir`; use `--overwrite` when replacing an existing bundle.

## Diagnostics and suggestions

- `inspect:source` supports `.swift`, `.xaml`, `.xml`, `.qml`, `.ui`, and common
  Qt C++ file extensions.
- `inspect:errors` supports `xaml`, `qt`, `manifest`, `patterns`, and `swift`
  kinds.
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
- Reserved generator commands return a deterministic unavailable-command error;
  do not retry them as implementation steps.
- Use `--quiet` when also passing `--output` so successful writes do not add
  extra chatter around file artifacts.

## Command Availability Policy

- If a generator command is cataloged as a future-Go feature, stop execution and
  route the workflow through supported analysis, transfer, audit, graph, project
  build, and parity commands instead.
- Keep `--quiet` and `--json` for automation, and always preserve command output
  fields required for machine interpretation.

## pattern workflows

- Use `patterns:validate` and `patterns:lint` for semantic and operational checks.
- Use `patterns:transfer` to inspect portability risk before editing source layouts.
- Use `patterns:export` for downstream tooling formats (`dtcg`, `open-ui`, `aria`,
  `style-dictionary`).
