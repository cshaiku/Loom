# OS error suggestions

Loom provides a curated suggestion catalog for common OS/framework errors that
affect interface transfer. It is intentionally practical, not exhaustive.

Use:

```sh
loom suggestions:os-errors
loom suggestions:os-errors --platform swiftui --format json
loom suggestions:os-errors --platform winui3 --message StaticResource
loom inspect:errors MainWindow.xaml --kind xaml --json
```

`inspect:errors` automatically attaches `suggested_fixes` to findings. Each
fix is split by audience:

- `user`: product/design decision or review action;
- `agent`: implementation action or follow-up Loom command.

## Covered areas

- Swift parser and SwiftUI result-builder failures.
- SwiftUI accessibility labels and decorative-hidden content.
- WinUI/XAML parse and resource lookup failures.
- Unsupported native WinUI component boundaries.
- WinUI AutomationProperties naming and accessibility tree exposure.
- Custom WinUI automation peer work.
- Windows binding/data-context contract failures.

## Source basis

The catalog is based on current platform guidance:

- Apple SwiftUI accessibility label/hidden guidance.
- Apple Swift and SwiftSyntax parser diagnostics behavior.
- Microsoft WinUI/XAML parse exception guidance.
- Microsoft XAML resource dictionary and StaticResource behavior.
- Microsoft AutomationProperties.Name and AccessibilityView guidance.
- Microsoft custom automation peer guidance for WinUI controls.

When adding new suggestions, prefer official platform documentation and make
the fix actionable for both a human reviewer and an AI agent.
