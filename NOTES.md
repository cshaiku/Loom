# Notes

## 2026-09-01 Open-Source And v1.0 Readiness

Loom is intended to be an analyzer, generator, and translator. The current
`0.24.0` codebase has a healthy analyzer, transfer-planning, component graph,
and analysis-only project build foundation, but generator commands remain
placeholders. Public docs should therefore describe `0.24.x` as pre-1.0 and
reserve `v1.0.0` for the complete analyze-generate-translate product.

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
