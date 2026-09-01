# Release Evidence

Use this file to record the evidence for each stable Loom release.

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

## v1.0.0 Release Evidence

Release date: 2026-09-01

- Commit SHA: `10279a80c97b0540d02d909f107fcedf21908fe2`
- CI run: https://github.com/cshaiku/Loom/actions/runs/33476210041
- Release workflow run: https://github.com/cshaiku/Loom/actions/runs/33476210005
- Release: https://github.com/cshaiku/Loom/releases/tag/v1.0.0

Native release workflow smoke passed for:

- Linux amd64
- Linux arm64
- macOS amd64
- macOS arm64
- Windows amd64
- Windows arm64

Release archive checksums:

| Archive | SHA-256 |
| --- | --- |
| `loom-v1.0.0-linux-amd64.tar.gz` | `60b0a344ee30702247b7ce8aad30ee2cda5bcce9aefa0aaa035cc6f5e9f2548b` |
| `loom-v1.0.0-linux-arm64.tar.gz` | `9210360892d0754faba0487a1a94a7e36621407d2c40a5a58d5ebd9a78edef49` |
| `loom-v1.0.0-macos-amd64.tar.gz` | `4c3d91abd5260ba989b937a78542eedf8862ee066e64aaee617aae0cb812597a` |
| `loom-v1.0.0-macos-arm64.tar.gz` | `5c96cd8f8fb471456dad251869c39f0b2431cd00eb3386864c27d9e2804c15bb` |
| `loom-v1.0.0-windows-amd64.zip` | `c3fac983972737540a2fa3bf67f67b79796b57e3d1e74bcd0391a5b7f52cea86` |
| `loom-v1.0.0-windows-arm64.zip` | `514873332c07d9e532089498a18e45972948c36dccf1f70c6b9f95fb4c3cb489` |

Accepted limitation: generated code is conservative scaffold output. Loom
preserves component boundaries and emits contract/transfer reports for
target-platform review instead of inventing native behavior silently.

## Future Release Notes

For future stable releases, record:

- commit SHA;
- CI run URL;
- release workflow run URL;
- archive names and checksums;
- native smoke results per platform;
- any accepted limitations for generated scaffold output.
