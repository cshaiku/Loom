# Deprecation Policy

Stable public contracts are not removed without a documented transition.

## Required Sequence

1. Publish a scoped compatibility decision in release notes or an accepted RFC.
2. Mark the command, flag, schema field, or pattern contract as deprecated in
   metadata or documentation.
3. Document the replacement, migration, rollback, and last supported release.
4. Emit a stable machine-readable warning code where practical.
5. Preserve the old contract for at least two minor releases and 90 calendar
   days, whichever is longer.
6. Remove only in the announced release, with tests and release notes updated in
   the same change.

Security issues may shorten the window when continued availability creates
material risk. The advisory must explain the exception and provide the safest
available migration.

## Current Deprecations

None.
