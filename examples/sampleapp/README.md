# Loom Sample App

This sample contains equivalent small layouts in SwiftUI, WinUI XAML, and Qt
QML. It is intentionally neutral so public contributors can run Loom without
private app code.

## Files

- `contentview.swift`: SwiftUI source.
- `mainwindow.xaml`: WinUI XAML source.
- `mainwindow.qml`: Qt QML source.
- `loom.json`: sample translation manifest.
- `visual-profile.json`: profile-normalized visual parity settings.
- `analyze-sample-app.sh`: repeatable analysis workflow.

## Run

```sh
./examples/sampleapp/analyze-sample-app.sh --overwrite
```

Outputs are written to `examples/sampleapp/generated/` and are ignored by git.

The workflow validates the manifest, inspects each source, audits the XAML,
plans SwiftUI/WinUI/Qt transfers, compares structural parity, and compares
profile-normalized visual metrics.
