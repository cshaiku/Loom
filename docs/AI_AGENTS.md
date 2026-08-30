# AI Agents

This document is the operating contract for AI agents using Loom in a repository.
It explains which commands are safe, which commands write, and how to inspect
errors before generating or updating target files.

## Default agent workflow

Use this order when entering an unfamiliar Loom workspace:

```sh
loom status --json
loom verify --json
loom guards:summary
loom self-heal:plan
loom list --json
```

Then inspect the specific source surface before generating:

```sh
loom inspect:errors ContentView.swift --root-view ContentView --json
loom accessibility:audit ContentView.swift --root-view ContentView --json
loom suggestions:os-errors --platform winui3 --format json
loom inspect:source ContentView.swift --root-view ContentView --json
loom patterns:transfer ContentView.swift --from swiftui --to winui3 --format json
loom inspect:ascii ContentView.swift --root-view ContentView
loom generate:contracts ContentView.swift --root-view ContentView --json
loom generate:xaml ContentView.swift --root-view ContentView --output Generated/ContentView.xaml
```

For mature native targets, prefer owned-region updates over whole-file
replacement:

```sh
loom generate:xaml ContentView.swift \
  --root-view ContentView \
  --replace-region MainWindow.xaml \
  --region-id shell.main
```

## Read/write safety

Loom command access is exposed through `loom list` and `loom guards:summary`.

- `r` commands are read-only.
- `w` commands write by design.
- `r/w` commands read by default and write only when a write flag is supplied.

Current write-authorizing flags include `--output` and `--replace-region`.
`project:build` always writes generated artifacts.

Agents should not infer permission to overwrite handwritten platform files.
Use `--replace-region` only when the file contains explicit Loom ownership
markers:

```xml
<!-- LOOM-BEGIN shell.main -->
<!-- generated content lives here -->
<!-- LOOM-END shell.main -->
```

`--init-region` is the only self-healing write mode. It creates a missing XAML
host file with one owned region. It intentionally refuses to modify existing
unmarked files.

## Error inspection

Use `inspect:errors` before generation when source quality is unknown:

```sh
loom inspect:errors ContentView.swift --kind swift --root-view ContentView
loom inspect:errors MainWindow.xaml --kind xaml
loom inspect:errors loom.json --kind manifest
loom inspect:errors Patterns --kind patterns
```

The command reports Swift parser diagnostics, Loom extraction diagnostics, XAML
parse failures, manifest validation issues, and Pattern validation/lint issues.

For automation, opt into failure semantics:

```sh
loom inspect:errors ContentView.swift --root-view ContentView --fail-on error
loom inspect:errors ContentView.swift --root-view ContentView --fail-on warning
```

Without `--fail-on`, error inspection reports findings but exits successfully.
This makes it useful in exploratory agent workflows.

`inspect:errors` findings may include `suggested_fixes`. Use those before
editing. The same catalog is available directly:

```sh
loom suggestions:os-errors --platform swiftui --format json
loom suggestions:os-errors --platform winui3 --message StaticResource
loom suggestions:os-errors --platform xaml --message XamlParseException
```

The suggestion catalog is curated. If no suggestion matches, use the generic
fallback and verify against platform documentation before making broad changes.

## Accessibility and layout audit

Use `accessibility:audit` before generation when reviewing UI quality or
transfer readiness:

```sh
loom accessibility:audit ContentView.swift --root-view ContentView --json
loom accessibility:audit MainWindow.xaml --format json
loom accessibility:audit ContentView.swift --fail-on warning
```

Treat `error` findings as blockers. Warnings usually mean a Pattern, layout
policy, or native implementation contract needs an explicit decision. The audit
covers missing accessible names, unlabeled images and inputs, small targets,
color-only semantic risks, unsupported/malformed nodes, empty or redundant
containers, nested scroll regions, geometry-dependent layouts, and unsupported
native WinUI component boundaries.

Read `suggested_fixes` in JSON output before editing. `user` fixes identify
decisions that need human product/design input. `agent` fixes identify safe
implementation actions or follow-up Loom commands. Do not apply a user-decision
fix as an agent edit unless the user has already supplied the decision.

## Output discipline

Prefer JSON for machine use:

```sh
loom status --json
loom verify --json
loom inspect:errors ContentView.swift --json
loom generate:contracts ContentView.swift --json
loom patterns:export --format dtcg
```

Use global runtime flags for clean automation logs:

```sh
loom --quiet generate:xaml ContentView.swift --output Generated/ContentView.xaml
loom --verbose generate:xaml ContentView.swift --output Generated/ContentView.xaml
```

`--quiet` suppresses successful write chatter. Fatal errors still print to
stderr. `--verbose` adds write details on stderr.

## Pattern export

Use `patterns:export` when another tool needs Loom's semantic vocabulary:

```sh
loom patterns:export --format loom
loom patterns:export --format dtcg
loom patterns:export --format open-ui
loom patterns:export --format aria
loom patterns:export --format style-dictionary
```

Exports are derived from the canonical `Patterns/*.pattern.json` files. Do not
edit exported files and treat them as disposable build artifacts unless a
project explicitly commits generated metadata.

## Pattern transfer and ASCII Patterns

Use `patterns:transfer` before generation when the goal is to move interface
design between platforms:

```sh
loom patterns:transfer ContentView.swift --from swiftui --to winui3 --format json
loom patterns:transfer MainWindow.xaml --from winui3 --to swiftui --format json
```

The report classifies each layout element as `direct`, `needs-policy`,
`needs-native-contract`, `lossy`, or `unsupported`. Treat `unsupported` and
`lossy` items as review blockers. Treat `needs-native-contract` items as
required native implementation work beside generated layout.

Use `inspect:ascii` when a compact plaintext layout sketch is more useful than
full JSON:

```sh
loom inspect:ascii ContentView.swift --root-view ContentView
loom inspect:ascii MainWindow.xaml
```

## Translation boundaries

Loom translates layout and semantic structure. It does not own application
behavior, platform services, gestures, media surfaces, native menus, or view
model state. Agents should treat `generate:contracts` as the required companion
to `generate:xaml`: layout output shows what can be scaffolded; contracts show
what must be implemented natively.
