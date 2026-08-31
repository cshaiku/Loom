import SwiftUI

struct ContentView: View {
    @State private var name = "example item"
    @State private var enabled = true

    var body: some View {
        VStack(spacing: 12) {
            Text("sample workspace")
            Text("a neutral SwiftUI layout used to demonstrate loom analysis.")

            HStack(spacing: 16) {
                List {
                    Text("first item")
                    Text("second item")
                }

                VStack(spacing: 8) {
                    Text("details")
                    TextField("name", text: $name)
                    Toggle("enabled", isOn: $enabled)
                    Button("save changes") {
                    }
                }
            }

            Text("ready")
        }
        .padding(24)
    }
}
