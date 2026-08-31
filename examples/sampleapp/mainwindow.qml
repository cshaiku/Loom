import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

ColumnLayout {
    spacing: 12

    Text {
        text: "sample workspace"
    }

    Text {
        text: "a neutral Qt layout used to demonstrate loom analysis."
    }

    RowLayout {
        spacing: 16

        ListView {
            Text { text: "first item" }
            Text { text: "second item" }
        }

        ColumnLayout {
            spacing: 8
            Text { text: "details" }
            TextField { placeholderText: "name" }
            Switch { text: "enabled" }
            Button { text: "save changes" }
        }
    }

    Text {
        text: "ready"
    }
}
