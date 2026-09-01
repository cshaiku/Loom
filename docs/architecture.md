# loom architecture

loom keeps a platform-neutral layout model at the center and maps it through
OS-agnostic `patterns`, generator contracts, transfer reports, and command
outputs.

## Runtime direction

loom's active runtime is Go. It is designed to run on macOS, Windows, and
Linux as a native CLI (`cmd/loom`).

## Product Direction

The `v1.0.0` product is an analyzer, generator, and translator:

- **Analyzer**: parse SwiftUI, WinUI XAML, Qt, fonts, and pattern catalogs into
  deterministic reports.
- **Generator**: emit reviewable target UI scaffolds and native contracts.
- **Translator**: run manifest-directed source-to-target workflows with transfer,
  parity, accessibility, and generated artifact outputs.

`0.24.0` implements the analyzer, transfer-planning, component-graph, and
analysis-only project build foundation. Generator commands are cataloged but
remain release-blocking.

## Current Go-Owned Surface

- `inspect:xaml`: parse WinUI XAML into loom's shared tree.
- `inspect:swiftui`: parse common SwiftUI layout/control constructs into
  loom's shared tree.
- `inspect:qt`: parse Qt QML, Qt Designer UI, and common Qt C++ layout/control
  constructs into loom's shared tree.
- `inspect:font`: extract intrinsic font material properties from supplied font
  files or installed family names.
- `inspect:source`: auto-detect SwiftUI, WinUI XAML, or Qt source and inspect it.
- `inspect:parity`: compare normalized layout structure across supported
  dialects.
- `inspect:visual-parity`: compare profile-normalized visual metrics across
  supported dialects as pre-render visual regression infrastructure.
- `graph:components`: discover component boundaries and source-tree dependency
  edges.
- `inspect:ascii`: render the shared tree as a plain text structure.
- `inspect:errors`: classify and report source, XAML, manifest, and pattern
  issues.
- `patterns:list|show|validate|lint|export|transfer`: canonical pattern
  validation, transfer scoring, and interoperability exports.
- `accessibility:audit`: identify layout design and accessibility risks.
- `suggestions:os-errors`: return curated fix recommendations.
- `status`, `verify`, `checks:command-catalog`, `guards:summary`,
  `config:validate`, `config:schema`, `project:build`, `self-heal:plan`.

## Catalog Compatibility Placeholders

These are intentionally present in the command catalog but not yet implemented in
Go:

- `generate:xaml`
- `generate:swiftui`
- `generate:contracts`

Invocations return a deterministic unavailable-command diagnostic so automation
fails closed.

## Current Pipeline

1. **Input normalization**: parse WinUI XAML, common SwiftUI constructs, or Qt
   layout sources and normalize to loom IR.
2. **pattern matching**: map normalized nodes to canonical patterns where
   possible.
3. **Transfer planning**: score each node as `direct`, `needs-policy`,
   `needs-native-contract`, `lossy`, or `unsupported`.
4. **ASCII rendering**: produce deterministic text tree output for review and logs.
5. **Font material inspection**: extract family names, fallback candidates,
   ascender/descender metrics, line gaps, cap height, x-height, weight/width,
   italic/fixed-pitch data, kerning presence, and normalized ratios from real
   font files or installed family names.
6. **Visual parity profiles**: normalize platform defaults for fonts, fallback
   stacks, kerning, line height, baseline offset, spacing, padding, margins,
   control minimums, and dimensional tolerances before render-backed comparison.
7. **Provenance and trust**: attach source, font-material, resolved-resource,
   style-setter, explicit-style-setter, profile, resource-reference,
   default-profile, or unknown origins plus confidence to visual parity metrics
   and findings. XAML visual parity can resolve document-local resources, local
   merged dictionaries, implicit styles, explicit styles, and simple BasedOn
   chains before comparing material values. Common object-valued resources and
   setters are normalized into comparable scalar material values where possible.
8. **Diagnostics**: run command-scoped validation with structured severities and
   optional JSON output.
9. **Accessibility/risk audit**: report missing names, weak interaction targets,
   malformed/redundant structures, unsupported native boundaries, and other
   transfer hazards.
10. **Project build**: run manifest-directed analysis-only bundles that write
    validation, source analysis, component graph, transfer, parity, and summary
    artifacts.

## v1.0.0 Pipeline

1. **Component graphing**: discover reachable layout components and dependency
   boundaries.
2. **Target generation**: emit reviewable SwiftUI or WinUI XAML from Loom IR.
3. **Contract generation**: report behavior, state, action, accessibility,
   resource, and native implementation requirements.
4. **Project build**: run manifest-directed analyzer, generator, transfer,
   parity, and audit outputs into a stable artifact directory.
5. **Guarded replacement**: replace owned regions only when the target file,
   region ID, and overwrite intent are explicit.

## Trust boundaries

- Layout analysis is intentionally conservative and does not imply behavioral
  parity.
- Unsupported constructs are surfaced as diagnostics, not silently flattened.
- Generated or derived output should always be reviewed before running in another UI
  runtime.
- Commands that write files must require explicit write targets and preserve
  overwrite guards.

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
