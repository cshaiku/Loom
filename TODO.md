# TODO

## v1.0.0 Release Blockers

- Implement `generate:xaml` for reviewable WinUI XAML output from normalized
  source layouts.
- Implement `generate:swiftui` for reviewable SwiftUI output from normalized
  WinUI/Qt layouts.
- Implement `generate:contracts` for target native behavior, state, action, and
  accessibility contracts.
- Extend `project:build` beyond analysis-only bundles once generator commands
  exist, including generated fragments and contracts.
- Freeze and document stable JSON schemas for all public automation surfaces.
- Add cross-platform CI and release artifacts for macOS, Linux, and Windows.
- Add generator-specific tests for owned-region replacement, overwrite guards,
  malformed input, and generated code reviewability.
- Add native smoke evidence for supported release archives.

## Open-Source Follow-Ups

- Decide the public GitHub repository location before adding permanent badges.
- Decide maintainer aliases for `CODEOWNERS`.
- Decide whether releases should publish Homebrew, direct archives, `go install`,
  or all three.
- Replace generic security contact if the project gets a dedicated address.
