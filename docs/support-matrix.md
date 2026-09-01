# Loom Support Matrix

Loom `0.22.0` is a pre-1.0 analyzer and transfer-planning release. The `v1.0.0`
target is analyzer, generator, and translator support across SwiftUI, WinUI XAML,
and Qt.

## Current Analyzer Support

| Source dialect | Status | Current behavior |
| --- | --- | --- |
| WinUI XAML | Supported analyzer baseline | Parses common layout/control constructs, captures Grid track metadata, resolves selected visual resources and styles, audits transfer/accessibility risks. |
| SwiftUI | Supported analyzer baseline | Parses common layout/control constructs and visual modifiers into Loom's shared layout model. |
| Qt QML | Supported analyzer baseline | Parses common Qt Quick and Controls layout/control constructs into Loom's shared layout model. |
| Qt Designer UI | Conservative analyzer support | Parses common XML widget/layout constructs. |
| Qt C++ | Conservative analyzer support | Parses common layout/control construction heuristics. |
| Fonts | Supported inspection | Extracts intrinsic OpenType/TrueType/TTC/WOFF metrics or installed family metrics for visual parity profiles. |

## Current WinUI XAML Mapping

| WinUI XAML construct | Loom IR | Current behavior |
| --- | --- | --- |
| `Grid` | `grid` | Parsed as a container. `Grid.RowDefinitions` and `Grid.ColumnDefinitions` are captured as metadata for transfer policy. |
| `StackPanel Orientation="Vertical"` | `verticalStack` | Parsed as a linear vertical layout. |
| `StackPanel Orientation="Horizontal"` | `horizontalStack` | Parsed as a linear horizontal layout. |
| `TextBlock` | `text` | Captures text content. |
| `Button`, `AppBarButton`, `HyperlinkButton` | `button` | Captures visible label when available and reports native action contract needs. |
| `TextBox`, `PasswordBox` | `textField` | Captures header or placeholder text and reports state-binding contract needs. |
| `Image` | `image` | Captures source and audits missing accessibility intent. |
| `ScrollViewer` | `scrollView` | Captures scroll container and audits nested scroll risk. |
| `ListView`, `GridView`, `ItemsRepeater` | `list` | Captures collection surface and reports collection/template contract needs. |
| `Slider`, `ProgressBar` | `slider` | Captures range/progress surface and reports state contract needs. |
| `ToggleSwitch`, `CheckBox` | `toggle` | Captures binary state surface and reports state contract needs. |
| `Rectangle Height="1"` / `Width="1"` | `divider` | Treated as a separator. |
| `Border Background=...` | `color` | Treated as a surface/color token candidate. |
| Other native WinUI controls | `component` | Preserved as unsupported native component boundaries with warnings and suggested fixes. |

## Generator And Translator Support

| Command | Current status | v1.0.0 expectation |
| --- | --- | --- |
| `patterns:transfer` | Implemented | Stable transfer planning across supported source/target pairs. |
| `inspect:parity` | Implemented | Stable structural parity report. |
| `inspect:visual-parity` | Implemented foundation | Stable profile-normalized visual parity report. |
| `graph:components` | Placeholder | Discover reachable layout components and custom dependencies. |
| `generate:xaml` | Placeholder | Emit reviewable WinUI XAML and support guarded owned-region replacement. |
| `generate:swiftui` | Placeholder | Emit reviewable SwiftUI scaffolds. |
| `generate:contracts` | Placeholder | Emit target native contracts for behavior, state, action, and accessibility. |
| `project:build` | Placeholder | Run manifest-directed analyzer/generator/translator workflows. |

For current releases, generated output should not be advertised as available.
Transfer reports identify structure, policy decisions, native behavior contracts,
accessibility gaps, and unsupported boundaries before generation exists.
