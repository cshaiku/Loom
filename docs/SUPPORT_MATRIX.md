# SwiftUI ↔ WinUI 3 support matrix (for transfer planning)

| SwiftUI construct | WinUI 3 strategy | Current status |
| --- | --- | --- |
| `VStack`, `HStack` | Row/column `Grid`; star cell for `Spacer` | Generated |
| `ZStack` | Layered `Grid` | Generated |
| `HSplitView`, `VSplitView` | Constraint-aware `Grid` plus splitter integration point | Structural |
| `GeometryReader` | `Grid` plus explicit runtime sizing diagnostic | Structural |
| `ScrollView` | `ScrollViewer` | Generated |
| `List`, `Table` | `ListView` with ItemsSource/DataTemplate integration point | Structural |
| `ForEach` | ItemsSource/DataTemplate integration point | Structural |
| `Text`, `TextField` | `TextBlock`, `TextBox` | Generated |
| `Button` | `Button` plus event/binding integration point | Structural |
| `Image` | `Image` | Generated |
| `Slider`, `ProgressView` | `Slider`, `ProgressBar` | Generated |
| `Toggle`, `Picker` | `ToggleSwitch` pending configurable control policy | Structural |
| `Spacer`, `Divider` | Star-sized cell, themed `Rectangle` | Generated |
| `if`/`else` | Visibility binding or visual-state integration point | Structural |
| `.frame` | Width/height constraints when statically numeric | Partial |
| `.padding` | XAML margin translation for static edge/value forms | Partial |
| Accessibility identifier | Automation ID | Generated |
| Gestures and animations | Explicit platform implementation | Diagnostic only |
| `NSViewRepresentable` | Native Windows control or interop surface | Diagnostic only |
| Custom `Layout` | Project mapping or handwritten target control | Diagnostic only |
| Arbitrary runtime geometry | View model or `SizeChanged` adapter | Diagnostic only |

Rows describe planned transfer outcomes in Loom’s `patterns:transfer` model; many
entries are still design-policy decisions and require human review before release.
