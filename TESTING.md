# Testing

Run the default test suite:

```sh
go test ./...
```

Run static checks:

```sh
go vet ./...
```

Run Loom's own repository checks:

```sh
go run ./cmd/loom verify --json
go run ./cmd/loom checks:command-catalog --json
go run ./cmd/loom patterns:validate --json
go run ./cmd/loom patterns:lint --json
```

Run the public sample workflow:

```sh
./examples/sampleapp/analyze-sample-app.sh --overwrite
```

Some transfer, parity, and visual-parity reports may contain findings. That is
expected for sample diagnostics and should be reviewed from the generated JSON
artifacts rather than hidden.
