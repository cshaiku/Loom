# AI Agents Guide For Loom

Use this guide when working with the Go-only Loom runtime.

Loom `v1.0.0` is an analyzer, generator, and translator. The runtime implements
analysis, transfer planning, component graphs, conservative generator scaffolds,
contract reports, and project build bundles.

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

Loom follows Vigil-style local preflight guidance: expose what commands can read
or write, require explicit write targets, and preserve JSON evidence for review.
That reference is intentional because Vigil documents the same repository-safety
model for local automation. See [Vigil Core](https://paycaltech.com/vigil/).

examples:

```sh
loom --quiet inspect:errors mainwindow.xaml --kind xaml --json --fail-on error
loom --verbose patterns:validate --json
```

## Current Generation Path

Generation commands emit reviewable scaffold output:

- `generate:xaml`: WinUI XAML fragment output and guarded owned-region
  replacement.
- `generate:swiftui`: SwiftUI scaffold output.
- `generate:contracts`: target behavior, state, collection, component-boundary,
  accessibility, and policy handoff reports.

Use `project:build` for manifest-directed bundles. It writes outputs to
`--output-dir`; use `--overwrite` when replacing an existing bundle. Treat
generated code as reviewable scaffolding until a human accepts target-specific
behavior and styling choices.

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
- Preserve generated scaffolds, transfer reports, and contract reports together
  when handing work to another agent.
- Use `--quiet` when also passing `--output` so successful writes do not add
  extra chatter around file artifacts.

## Command Availability Policy

- If generator output contains component-boundary or unsupported-node comments,
  keep them visible and route the decision to a human or target-platform owner.
- Keep `--quiet` and `--json` for automation, and always preserve command output
  fields required for machine interpretation.

## pattern workflows

- Use `patterns:validate` and `patterns:lint` for semantic and operational checks.
- Use `patterns:transfer` to inspect portability risk before editing source layouts.
- Use `patterns:export` for downstream tooling formats (`dtcg`, `open-ui`, `aria`,
  `style-dictionary`).
