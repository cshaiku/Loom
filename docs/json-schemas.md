# JSON Schema Contracts

Loom JSON reports use `schema_version: "1"` for the v1 automation contract.
Additive fields may appear in minor releases. Existing documented fields should
not be removed or have their meaning changed without a new schema version or
deprecation notice.

## Common Fields

All command reports include:

| Field | Type | Meaning |
| --- | --- | --- |
| `schema_version` | string | Report schema version. Current value is `"1"`. |
| `status` | string | Command-specific result. `ok` means no blocking findings. |

Commands that inspect an input usually also include `sourcePath`, `rootView`,
`component`, `diagnostics`, and a command-specific body such as `layout`,
`items`, `findings`, `contracts`, or `artifacts`.

## Stable Report Surfaces

| Command | Primary report type | Stable body fields |
| --- | --- | --- |
| `inspect:source`, `inspect:swiftui`, `inspect:xaml`, `inspect:qt` | `Analysis` | `sourcePath`, `rootView`, `component`, `syntaxNodeCount`, `layout`, `diagnostics` |
| `inspect:ascii` | text only | Not a JSON surface. |
| `inspect:errors` | `LoomErrorInspectionReport` | `inspectedKind`, `source`, `findings` |
| `inspect:font` | `FontInspectionReport` | `source`, `family`, `faces`, `diagnostics` |
| `inspect:parity` | `ParityReport` | `sourcePath`, `targetPath`, `sourceDialect`, `targetDialect`, `sourceCount`, `targetCount`, `findings` |
| `inspect:visual-parity` | `VisualParityReport` | `sourcePath`, `targetPath`, `sourceDialect`, `targetDialect`, `profile`, `summary`, `findings`, `diagnostics` |
| `accessibility:audit` | `AuditReport` | `sourcePath`, `rootView`, `component`, `summary`, `findings`, `diagnostics` |
| `patterns:validate`, `patterns:lint` | `PatternValidationReport` | `directory`, `patternCount`, `issues` |
| `patterns:transfer` | `TransferReport` | `sourcePath`, `from`, `to`, `rootView`, `component`, `asciiPattern`, `summary`, `items`, `diagnostics` |
| `graph:components` | `ComponentGraphReport` | `source`, `root`, `components`, `edges`, `diagnostics` |
| `generate:xaml`, `generate:swiftui` | `GeneratedArtifactReport` | `sourcePath`, `from`, `to`, `rootView`, `component`, `outputKind`, `text`, `diagnostics` |
| `generate:contracts` | `ContractReport` | `sourcePath`, `target`, `rootView`, `component`, `contracts`, `diagnostics` |
| `project:build` | `ProjectBuildReport` | `project`, `manifestPath`, `projectRoot`, `outputDir`, `artifacts`, `diagnostics` |
| `status` | `LoomStatusReport` | `version`, `workingDirectory`, `commands`, `patternDirectory`, `patternStatus`, `patternCount`, `issues` |
| `verify` | `LoomVerifyReport` | `commandCatalog`, `patterns`, `patternLint` |
| `checks:command-catalog` | `LoomCommandCatalogCheckReport` | `commands`, `aliases`, `issues` |
| `guards:summary` | `LoomGuardsReport` | `entries` |
| `self-heal:plan` | `LoomSelfHealPlan` | `entries` |
| `suggestions:os-errors` | `OSErrorSuggestionReport` | `platform`, `query`, `suggestions` |

## Status Values

Common status values:

- `ok`: no blocking diagnostics or findings.
- `warning`: non-blocking findings need review.
- `review`: generated scaffold or contract output needs human review before
  production use.
- `partial`: transfer planning found lossy, policy-dependent, native-contract,
  or unsupported work.
- `error`: blocking diagnostics or invalid input.
- `source-invalid`: source analysis produced error-level diagnostics.

Callers should treat unknown non-`ok` statuses as requiring review.

## Stability Rules

- Empty lists are emitted as `[]`.
- Paths are reported as the command received or resolved them; callers should
  not assume platform separators.
- Text generation reports put generated source in `text`.
- File-writing commands require explicit output flags and refuse existing files
  unless `--overwrite` is provided.
- `project:build` emits a summary plus individual artifacts so automation can
  archive or inspect each report independently.
