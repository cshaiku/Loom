# Support Policy

This policy separates code Loom can compile from platforms and behaviors Loom
has actually exercised.

## Release Lines

- The latest stable minor release receives correctness and security fixes.
- The previous stable minor receives critical correctness and security fixes for
  90 days after its successor is published.
- Prerelease and `0.x` channels are evaluation channels and may receive fixes
  only on the current line.
- Security advisories may require an immediate upgrade when preserving old
  behavior would retain material risk.

## Current Support State

`0.24.x` is supported as the current analyzer, transfer-planning,
component-graph, and analysis-only project build line. Generator scaffolds on
main are pre-release until a tagged release includes schema and platform
evidence.

No stable `v1.0.0` line exists yet. `v1.0.0` is blocked until generation,
translation, release, documentation, and CI evidence are complete.

## Platform Evidence

The Go runtime is intended for:

| Operating system | Architecture | Status |
| --- | --- | --- |
| macOS | amd64, arm64 | Local development target; release smoke required. |
| Linux | amd64, arm64 | CI and release smoke required. |
| Windows | amd64, arm64 | CI and native release smoke required. |

Cross-compilation alone is not enough to call a platform supported. A supported
platform needs source tests, generated artifact checks, and native smoke evidence.

## Toolchain Window

The Go version in `go.mod` is the authoritative build toolchain. Release
workflows should use that exact version. Older toolchains are unsupported.

## Stable Contract

For a stable release, Loom should preserve:

- command names and documented aliases;
- JSON `schema_version`, `status`, `summary`, `findings`, and command-specific
  fields documented for automation;
- default LF output behavior;
- explicit write guards for commands that create or replace files.

Breaking contract changes require a new schema version or the deprecation policy.
