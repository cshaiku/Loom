# Release Checklist

Use this checklist before tagging a public Loom release.

## Version

- `VERSION` matches `internal/loom/catalog.go`.
- `README.md` and `CHANGELOG.md` name the same release.
- Release notes distinguish analyzer, generator, and translator support.

## Quality Gates

```sh
go test ./...
go vet ./...
go run ./cmd/loom verify --json
go run ./cmd/loom checks:command-catalog --json
go run ./cmd/loom patterns:validate --json
go run ./cmd/loom patterns:lint --json
./examples/sampleapp/analyze-sample-app.sh --overwrite
```

## Public Source

- No credentials, private source, customer data, local artifacts, or generated
  sample outputs are committed unintentionally.
- `LICENSE`, `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, and
  `SUPPORT.md` are present.
- Issue and pull request templates are present.
- CI is green on the supported platform matrix.

## v1.0.0 Gate

`v1.0.0` additionally requires:

- implemented `generate:xaml`;
- implemented `generate:swiftui`;
- implemented `generate:contracts`;
- project builds that include generated fragments and contracts, not only
  analysis bundles;
- stable JSON schemas for analysis, transfer, audit, parity, visual parity, and
  generation reports;
- documented support and deprecation policies;
- release archives or install instructions that work outside a Homebrew-only
  local machine.
