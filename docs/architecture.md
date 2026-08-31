# loom architecture

loom keeps a platform-neutral layout model at the center and maps it through
OS-agnostic `patterns` and command outputs.

## Runtime direction

loom's active runtime is Go. It is designed to run on macOS, Windows, and
Linux as a native CLI (`cmd/loom`).

## Current Go-owned surface

- `inspect:xaml`: parse WinUI XAML into loom's shared tree.
- `inspect:ascii`: render that tree as a plain text structure.
- `inspect:errors`: classify and report source, XAML, manifest, and pattern
  issues.
- `patterns:list|show|validate|lint|export|transfer`: canonical pattern
  validation, transfer scoring, and interoperability exports.
- `accessibility:audit`: identify layout design and accessibility risks.
- `suggestions:os-errors`: return curated fix recommendations.
- `status`, `verify`, `checks:command-catalog`, `guards:summary`,
  `config:validate`, `config:schema`, `self-heal:plan`.

## Catalog compatibility placeholders

These are intentionally present in the command catalog but not yet implemented in
Go:

- `inspect:source`
- `inspect:parity`
- `graph:components`
- `generate:xaml`
- `generate:swiftui`
- `generate:contracts`
- `project:build`

Invocations return a deterministic
`is reserved for catalog parity only and is not yet available in the Go runtime`
diagnostic.

## Core pipeline (implemented in Go)

1. **Input normalization**: parse WinUI XAML and normalize to loom IR.
2. **pattern matching**: map normalized nodes to canonical patterns where
   possible.
3. **Transfer planning**: score each node as `direct`, `needs-policy`,
   `needs-native-contract`, `lossy`, or `unsupported`.
4. **ASCII rendering**: produce deterministic text tree output for review and logs.
5. **Diagnostics**: run command-scoped validation with structured severities and
   optional JSON output.
6. **Accessibility/risk audit**: report missing names, weak interaction targets,
   malformed/redundant structures, unsupported native boundaries, and other
   transfer hazards.

## Trust boundaries

- Layout analysis is intentionally conservative and does not imply behavioral
  parity.
- Unsupported constructs are surfaced as diagnostics, not silently flattened.
- Generated or derived output should always be reviewed before running in another UI
  runtime.

## pattern model

`patterns/*.pattern.json` remain canonical. They define:

- semantic intent, child policy, sizing/order, constraints, and stable identity;
- typed attributes with explicit value domains;
- optional variant policy metadata;
- optional accessibility metadata and target mappings.

The catalog is OS-agnostic by design. Platform mappings are treated as platform
realization guidance, not the definition of meaning.

## Reporting

Most commands support:

- `--json` for machine-readable output.
- `--output` for writing report artifacts.
- `--quiet` to suppress success chatter.
- `--verbose` to include write diagnostics.

Suggested fixes are attached to relevant findings to support both human review and
AI-agent actioning.
