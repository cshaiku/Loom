# Notes

## 2026-09-01 Open-Source And v1.0 Readiness

Loom is intended to be an analyzer, generator, and translator. The current
`0.23.0` codebase has a healthy analyzer and transfer-planning foundation, but
the generator and project build commands remain placeholders. Public docs should
therefore describe `0.22.x` as pre-1.0 and reserve `v1.0.0` for the complete
analyze-generate-translate product.

Guidance taken from Vigil and PayCal-style release discipline:

- local-first commands with explicit write flags;
- deterministic JSON output for automation;
- repository policy documented alongside code;
- security reporting before public issue disclosure;
- release claims backed by tests, verification, and platform evidence.
