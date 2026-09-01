# Release Evidence

Use this file to record the evidence for each release candidate before tagging a
stable Loom release.

## Required Gates

| Gate | Evidence |
| --- | --- |
| Unit tests | `go test ./...` passes. |
| Static check | `go vet ./...` passes. |
| Loom verification | `go run ./cmd/loom verify --json` returns `status: ok`. |
| Command catalog | `go run ./cmd/loom checks:command-catalog --json` returns `status: ok`. |
| Sample workflow | `./examples/sampleapp/analyze-sample-app.sh --overwrite` completes and writes analysis, graph, generated scaffold, contract, transfer, parity, and visual-parity artifacts. |
| Release archives | Tag workflow uploads macOS, Linux, and Windows archives. |
| Native smoke | Each release archive runs `loom version` and `loom verify --json` on its target OS/architecture or an approved equivalent runner. |

## Current Local Evidence

As of the current `main` work:

- `go test ./...` passed.
- `go vet ./...` passed.
- `go run ./cmd/loom verify --json` returned `status: ok`.
- `./examples/sampleapp/analyze-sample-app.sh --overwrite` completed.
- Generator tests cover reviewable XAML output, SwiftUI scaffold output,
  contract reports, owned-region replacement, overwrite guards, malformed
  source reporting, and nested output path creation.

## v1.0 Release Candidate Notes

Before tagging `v1.0.0`, record:

- commit SHA;
- CI run URL;
- release workflow run URL;
- archive names and checksums;
- native smoke results per platform;
- any accepted limitations for generated scaffold output.
