# Security Policy

Please report suspected security issues privately. Loom parses local UI source
files, reads pattern catalogs, emits generated artifacts, and is intended to be
used by people and coding agents, so command execution boundaries, generated
code safety, path handling, and supply-chain integrity are in scope.

## Reporting

Email: `info@paycaltech.com`

Include:

- affected Loom version or commit;
- operating system and shell;
- exact command and input shape;
- reproduction steps;
- expected impact;
- relevant logs, with secrets and private source removed.

Do not include live secrets, private keys, access tokens, proprietary source, or
private customer data. If a proof of concept requires sensitive material,
describe the data shape and wait for maintainer guidance.

## Scope

In scope:

- unintended file overwrite, path traversal, or mutation without an explicit
  write flag;
- parser crashes or resource exhaustion caused by malformed SwiftUI, XAML, Qt,
  manifest, font, or pattern inputs;
- generated code that introduces unsafe runtime behavior contrary to documented
  contracts;
- command catalog, release, dependency, or packaging issues that can mislead
  automation;
- secret leakage through logs, generated artifacts, diagnostics, or support
  material.

Out of scope:

- social engineering maintainers or users;
- denial-of-service volume testing against public services;
- issues requiring access to systems, repositories, or data you do not own or
  have explicit permission to test;
- unsupported forks or local modifications unless the issue also affects Loom.

## Safe Testing Rules

- Use only local test files or source you are authorized to inspect.
- Do not publish private UI source, customer data, credentials, or generated
  artifacts containing sensitive material.
- Stop testing and report immediately if you encounter secrets or private data.

## Supported Versions

| Version | Security support |
| --- | --- |
| `main` | Supported for coordinated reports against unreleased code. |
| `0.22.x` | Supported as the current analyzer and transfer-planning line. |
| Earlier releases | Fixes target the current line unless immediate user risk requires otherwise. |

No stable `v1.0.0` line exists yet. Stable support windows are defined in
[docs/support-policy.md](docs/support-policy.md).

## Response Targets

- Initial human acknowledgement: within 3 business days.
- Initial severity assessment: within 7 business days after enough information
  is available.
- Status updates: at least every 14 days while a validated report remains open.

These are targets, not guarantees.

## Coordinated Disclosure

Please give maintainers a reasonable opportunity to investigate and release a
fix before public disclosure. The default disclosure target is 90 days after
validation, or sooner by mutual agreement when a fix is available.
