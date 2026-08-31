# loom roadmap

## 0.17 - Go runtime foundation

- Forward Go CLI introduced under `cmd/loom`.
- Command catalog, status/verification diagnostics, pattern operations, and command
  guard outputs.
- `inspect:xaml`, `inspect:ascii`, `inspect:errors`, `accessibility:audit`,
  `patterns:*`, and OS error suggestions.
- CLI output contracts for automation (`--json`, `--quiet`, `--verbose`).

## 0.18 - Go analyzer baseline

- `inspect:errors` and `inspect:source` path planning moved toward Go ownership.
- `graph:components` command retained in catalog and prepared for re-implementation.
- Cross-command consistency checks expanded (`checks:command-catalog`).

## 0.19 - Go emitter and manifest parity work (next)

- Implement `generate:xaml`, `generate:swiftui`, `generate:contracts`,
  and `project:build` in Go.
- Add first-class owned-region replacement and merge safety in Go.
- Stabilize project workflow outputs (summary/build manifests + parity reports).

## Current compatibility layer

`inspect:source`, `inspect:parity`, `graph:components`, `generate:xaml`,
`generate:swiftui`, `generate:contracts`, and `project:build` remain catalog
placeholders until they are implemented in the Go runtime.

## Near-term priorities

- Strengthen pattern linting for malformed / redundant structures and platform
  boundary mapping completeness.
- Add explicit malformed-input and memory/guard regression tests.
- Improve malformed layout diagnostics for nested scroll and geometry-dependent
  constructs.
- Extend suggestions to cover additional WinUI, XAML, SwiftUI, and Windows
  parity failure classes.
