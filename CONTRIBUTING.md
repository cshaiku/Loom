# Contributing To Loom

Loom is an open-source CLI for analyzing, generating, and translating UI layout
intent across SwiftUI, WinUI XAML, and Qt.

Contributions are welcome when they keep Loom deterministic, local-first,
reviewable, and useful for both people and automation.

## Good First Contributions

- Documentation fixes that make setup, commands, generated artifacts, or release
  expectations easier to verify.
- Reproducible bug reports with exact input files, command arguments, expected
  output, actual output, operating system, and Loom version.
- Parser coverage for common SwiftUI, WinUI XAML, Qt QML, Qt Designer UI, or Qt
  C++ layout constructs.
- Generator and translator work that preserves source intent and emits reviewable
  target code.
- Accessibility, visual parity, and malformed-input tests.

## Before Opening A Pull Request

1. Search existing issues and pull requests to avoid duplicates.
2. Keep changes scoped to one behavior or documentation concern.
3. Do not include private source code, customer data, credentials, API keys,
   screenshots with sensitive data, or copied proprietary UI assets.
4. Add or update tests when behavior changes.
5. Update documentation when commands, schemas, generated output, or setup
   expectations change.

## Local Setup

```sh
git clone https://github.com/cshaiku/loom.git loom
cd loom
go test ./...
go vet ./...
go run ./cmd/loom verify --json
```

Run the sample workflow:

```sh
./examples/sampleapp/analyze-sample-app.sh --overwrite
```

## Pull Request Expectations

- Explain the user-facing impact.
- Identify affected commands, schemas, patterns, or generated artifacts.
- Include the exact checks you ran.
- Call out parser, generator, translation, accessibility, security, and
  cross-platform impact explicitly.
- Keep generated artifacts and dependency churn out of unrelated changes.

## Mutation Policy

Commands that write files must require an explicit output or mutation flag such
as `--output`, `--overwrite`, or a documented replacement flag. Read-only command
metadata must stay accurate in the command catalog.

## Security Reports

Do not open public issues for suspected vulnerabilities. Follow
[SECURITY.md](SECURITY.md).
