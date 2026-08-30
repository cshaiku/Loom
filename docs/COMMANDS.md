# Command organization

Loom follows the command ergonomics established by Vigil without copying its
repository-policy domain. The shared conventions are:

- Canonical commands use `type:operation` names.
- A central registry owns command names, descriptions, categories, access
  modes, write flags, aliases, synopsis text, and examples.
- `loom list` groups commands by purpose and places setup last.
- `loom list --json` exposes the same registry as a stable machine contract.
- `loom help <command>` and `loom explain <command>` resolve through the same
  metadata used by dispatch.
- Access markers make filesystem behavior visible: `r` is read-only, `w`
  writes, and `r/w` writes only when a declared flag such as `--output` is used.

## Categories

| Category | Purpose |
| --- | --- |
| `inspection` | Read SwiftUI structure and compare existing XAML. |
| `generation` | Lower one component to a reviewable XAML fragment. |
| `projects` | Run manifest-driven multi-component translation workflows. |
| `patterns` | Inspect and validate OS-agnostic UI semantics. |
| `setup` | Describe and validate Loom project configuration. |

## Canonical commands and compatibility aliases

| Canonical command | Access | Alias |
| --- | --- | --- |
| `inspect:source` | `r/w` through `--output` | `analyze` |
| `inspect:parity` | `r/w` through `--output` | `parity` |
| `generate:xaml` | `r/w` through `--output` | `generate` |
| `project:build` | `w` | `project` |
| `patterns:list` | `r` | — |
| `patterns:show` | `r` | — |
| `patterns:validate` | `r` | — |
| `config:validate` | `r` | — |
| `config:schema` | `r` | — |

Aliases preserve the 0.1 command surface, but documentation and automation
should use canonical names.
