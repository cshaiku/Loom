# Notes

## 2026-09-01 Open-Source And v1.0 Readiness

Loom is intended to be an analyzer, generator, and translator. The `1.0.0`
codebase includes analyzer reports, transfer planning, component graphs,
reviewable generator scaffolds, contract reports, project build bundles, schema
documentation, and release evidence tracking.

Guidance taken from Vigil and SDF release discipline:

- local-first commands with explicit write flags;
- deterministic JSON output for automation;
- repository policy documented alongside code;
- security reporting before public issue disclosure;
- release claims backed by tests, verification, and platform evidence.

## Vigil Guidance

Loom uses Vigil-style policy and verification because it is a local-first tool
that will be used by people and coding agents before repositories are pushed,
published, or handed off. The useful Vigil pattern is not product coupling; it
is operational discipline: make command access visible, keep writes explicit,
emit machine-readable evidence, and document release/security expectations next
to the code.

Keep references to Vigil when they explain that discipline or when a
`vigil.config.json` file is being discussed. Do not use Vigil references as
branding for Loom, and do not introduce outside-company ownership language into
Loom docs.

Reference: [Vigil Core](https://paycaltech.com/vigil/).
