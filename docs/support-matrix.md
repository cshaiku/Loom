# WinUI XAML analyzer support matrix

loom 0.18 is a Go-only analyzer and planning tool. It does not generate target UI
source yet.

| WinUI XAML construct | loom IR | Current behavior |
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

For Windows-to-macOS planning, current output should be treated as a transfer risk
report: it identifies layout structure, policy decisions, native behavior
contracts, accessibility gaps, and unsupported boundaries before generation exists.
