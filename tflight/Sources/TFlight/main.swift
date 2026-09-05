import SwiftUI
import AppKit
import UniformTypeIdentifiers

@main
struct TFlightApp: App {
    @StateObject private var library = Library()

    var body: some Scene {
        WindowGroup("TFlight") {
            ContentView(library: library)
                .frame(minWidth: 1060, minHeight: 680)
        }
        .commands {
            CommandGroup(replacing: .newItem) { Button("New Note") { library.newNote() }.keyboardShortcut("n") }
            CommandGroup(after: .saveItem) { Button("Export to Astro…") { library.exportCurrent() }.keyboardShortcut("e", modifiers: [.command, .shift]) }
        }
    }
}

struct Note: Identifiable, Hashable {
    var id: String
    var title = "Untitled note"
    var description = ""
    var date = Date()
    var draft = true
    var pinned = false
    var keywords = ""
    var body = ""
    var heroImage = ""
    var heroAlt = ""
    var dirty = false
}

@MainActor final class Library: ObservableObject {
    @Published var notes: [Note] = []
    @Published var selectedID: String?
    @Published var repositoryURL: URL?
    @Published var message = "Choose your Astro site to begin"

    var selected: Note? {
        get { notes.first { $0.id == selectedID } }
        set { guard let newValue, let index = notes.firstIndex(where: { $0.id == newValue.id }) else { return }; notes[index] = newValue }
    }

    init() { if let path = UserDefaults.standard.url(forKey: "tflight.repository") { open(path) } }

    func chooseRepository() {
        let panel = NSOpenPanel(); panel.canChooseDirectories = true; panel.canChooseFiles = false; panel.prompt = "Use Site"
        if panel.runModal() == .OK, let url = panel.url { open(url); UserDefaults.standard.set(url, forKey: "tflight.repository") }
    }

    func open(_ root: URL) {
        let posts = root.appendingPathComponent("src/content/posts")
        guard FileManager.default.fileExists(atPath: posts.path) else { message = "That folder is not an Astro site with src/content/posts"; return }
        repositoryURL = root; notes = (try? FileManager.default.contentsOfDirectory(at: posts, includingPropertiesForKeys: nil).filter { $0.pathExtension == "md" }.compactMap(read))?.sorted { $0.date > $1.date } ?? []
        selectedID = notes.first?.id; message = "Ready · \(notes.count) notes"
    }

    func newNote() { let note = Note(id: "note-\(Int(Date().timeIntervalSince1970))"); notes.insert(note, at: 0); selectedID = note.id }

    func update(_ note: Note) { selected = note }

    func saveCurrent() { guard let selected else { return }; write(selected); message = "Saved just now" }

    func exportCurrent() { guard let selected else { return }; write(selected); message = "Exported to Astro · ready for Git" }

    func attachImage(asHero: Bool = false) {
        guard var note = selected, let root = repositoryURL else { return }
        if note.id.hasPrefix("note-"), !note.title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty { write(note); note = selected ?? note }
        let panel = NSOpenPanel(); panel.allowedContentTypes = [.image]; panel.allowsMultipleSelection = false; panel.prompt = "Add to note"
        guard panel.runModal() == .OK, let source = panel.url else { return }
        let slug = note.id.hasPrefix("note-") ? note.title.lowercased().replacingOccurrences(of: "[^a-z0-9]+", with: "-", options: .regularExpression).trimmingCharacters(in: CharacterSet(charactersIn: "-")) : note.id
        let finalSlug = slug.isEmpty ? note.id : slug
        let folder = root.appendingPathComponent("src/assets/posts/\(finalSlug)"); try? FileManager.default.createDirectory(at: folder, withIntermediateDirectories: true)
        let destination = folder.appendingPathComponent(source.lastPathComponent)
        try? FileManager.default.copyItem(at: source, to: destination)
        var updated = note
        let relativePath = "../../assets/posts/\(finalSlug)/\(source.lastPathComponent)"
        if asHero { updated.heroImage = relativePath; updated.heroAlt = source.deletingPathExtension().lastPathComponent.replacingOccurrences(of: "-", with: " ") }
        else { updated.body += "\n\n![\(source.deletingPathExtension().lastPathComponent)](\(relativePath))\n" }
        update(updated); message = asHero ? "Hero image set · remember to export" : "Image added · remember to export"
    }

    private func read(_ url: URL) -> Note? {
        guard let source = try? String(contentsOf: url, encoding: .utf8) else { return nil }
        let parts = source.components(separatedBy: "\n---\n"); let front = parts.first?.replacingOccurrences(of: "---\n", with: "") ?? ""; let body = parts.dropFirst().joined(separator: "\n---\n")
        func value(_ key: String) -> String { front.split(separator: "\n").first(where: { $0.hasPrefix(key + ":") }).map { String($0.dropFirst(key.count + 1)).trimmingCharacters(in: .whitespaces).trimmingCharacters(in: CharacterSet(charactersIn: "\"")) } ?? "" }
        let formatter = ISO8601DateFormatter(); formatter.formatOptions = [.withFullDate]
        return Note(id: url.deletingPathExtension().lastPathComponent, title: value("title"), description: value("description"), date: formatter.date(from: value("pubDate")) ?? Date(), draft: value("draft") != "false", pinned: value("pinned") == "true", keywords: value("keywords"), body: body.trimmingCharacters(in: .whitespacesAndNewlines), heroImage: value("heroImage"), heroAlt: value("heroAlt"))
    }

    private func write(_ note: Note) {
        guard let root = repositoryURL else { return }
        let slug = note.id.hasPrefix("note-") ? note.title.lowercased().replacingOccurrences(of: "[^a-z0-9]+", with: "-", options: .regularExpression).trimmingCharacters(in: CharacterSet(charactersIn: "-")) : note.id
        let finalSlug = slug.isEmpty ? note.id : slug
        let posts = root.appendingPathComponent("src/content/posts"); try? FileManager.default.createDirectory(at: posts, withIntermediateDirectories: true)
        let formatter = ISO8601DateFormatter(); formatter.formatOptions = [.withFullDate]
        let imageLine = note.heroImage.isEmpty ? "" : "heroImage: \(note.heroImage)\nheroAlt: \"\(note.heroAlt.replacingOccurrences(of: "\"", with: "\\\""))\"\n"
        let yaml = "---\ntitle: \"\(note.title.replacingOccurrences(of: "\"", with: "\\\""))\"\ndescription: \"\(note.description.replacingOccurrences(of: "\"", with: "\\\""))\"\npubDate: \(formatter.string(from: note.date))\ndraft: \(note.draft)\npinned: \(note.pinned)\nkeywords: \"\(note.keywords)\"\n\(imageLine)---\n\n\(note.body.trimmingCharacters(in: .whitespacesAndNewlines))\n"
        try? yaml.write(to: posts.appendingPathComponent(finalSlug + ".md"), atomically: true, encoding: .utf8)
        if finalSlug != note.id { if let index = notes.firstIndex(where: { $0.id == note.id }) { notes[index].id = finalSlug; selectedID = finalSlug } }
    }
}

struct ContentView: View {
    @ObservedObject var library: Library
    var body: some View { NavigationSplitView { Sidebar(library: library) } detail: { if library.selected != nil { Editor(library: library) } else { Welcome(library: library) } }.navigationSplitViewStyle(.balanced) }
}

struct Sidebar: View {
    @ObservedObject var library: Library
    var body: some View { VStack(alignment: .leading, spacing: 18) {
        HStack { Text("✦").font(.title).foregroundStyle(.orange); VStack(alignment: .leading) { Text("TFlight").font(.headline); Text("Writing studio").font(.caption).foregroundStyle(.secondary) } }.padding(.horizontal, 8)
        Button { library.newNote() } label: { Label("New note", systemImage: "plus") }.buttonStyle(.borderedProminent).tint(.orange).keyboardShortcut("n")
        HStack { Text("NOTES").font(.caption).foregroundStyle(.secondary); Spacer(); Text("\(library.notes.count)").font(.caption).foregroundStyle(.secondary) }
        List(selection: $library.selectedID) { ForEach(library.notes) { note in VStack(alignment: .leading, spacing: 4) { Text(note.title.isEmpty ? "Untitled note" : note.title).lineLimit(1); Text(note.draft ? "Draft" : "Published") .font(.caption2).foregroundStyle(.secondary) }.tag(note.id) } }.listStyle(.sidebar)
        Spacer(); Button { library.chooseRepository() } label: { Label(library.repositoryURL == nil ? "Connect Astro site" : "Change site", systemImage: "folder") }.buttonStyle(.plain).foregroundStyle(.secondary); Text(library.message).font(.caption2).foregroundStyle(.secondary).lineLimit(2)
    }.padding(18) }
}

struct Welcome: View { @ObservedObject var library: Library; var body: some View { VStack(spacing: 14) { Text("A quieter place to write.").font(.system(size: 38, design: .serif)); Text("Notes become beautiful, deployable Astro articles.").foregroundStyle(.secondary); Button("Connect your Astro site") { library.chooseRepository() }.buttonStyle(.borderedProminent).tint(.orange) } } }

struct Editor: View {
    @ObservedObject var library: Library
    @State private var font = "New York"
    @State private var showPreview = true
    var note: Note { library.selected ?? Note(id: "empty") }
    var body: some View { VStack(spacing: 0) {
        HStack { TextField("Untitled note", text: Binding(get: { note.title }, set: { var n = note; n.title = $0; n.dirty = true; library.update(n) })).textFieldStyle(.plain).font(.system(size: 26, design: .serif)); Spacer(); Picker("Font", selection: $font) { Text("New York").tag("New York"); Text("Avenir").tag("Avenir"); Text("Mono").tag("Menlo") }.frame(width: 130); Button("Add image") { library.attachImage() }; Button("Set hero") { library.attachImage(asHero: true) }; Button(showPreview ? "Hide preview" : "Preview") { showPreview.toggle() }; Button("Export") { library.exportCurrent() }.buttonStyle(.borderedProminent).tint(.orange) }.padding(.horizontal, 28).padding(.vertical, 16)
        Divider()
        HStack(spacing: 0) { FormPane(note: note, library: library, font: font); if showPreview { Divider(); PreviewPane(note: note) } }
    } }
}

struct FormPane: View { let note: Note; @ObservedObject var library: Library; let font: String; var body: some View { ScrollView { VStack(alignment: .leading, spacing: 16) { TextField("A short description for search and sharing", text: Binding(get: { note.description }, set: { var n = note; n.description = $0; library.update(n) })).textFieldStyle(.plain).foregroundStyle(.secondary); Divider(); HStack { DatePicker("", selection: Binding(get: { note.date }, set: { var n = note; n.date = $0; library.update(n) }), displayedComponents: .date).labelsHidden(); Toggle("Draft", isOn: Binding(get: { note.draft }, set: { var n = note; n.draft = $0; library.update(n) })); Toggle("Pinned", isOn: Binding(get: { note.pinned }, set: { var n = note; n.pinned = $0; library.update(n) })) }; StructureBar(note: note, library: library); TextEditor(text: Binding(get: { note.body }, set: { var n = note; n.body = $0; n.dirty = true; library.update(n) })).font(.custom(font, size: 17)).scrollContentBackground(.hidden).frame(minHeight: 500); HStack { Text("\(note.body.split(whereSeparator: \.isWhitespace).count) words"); Spacer(); Text("Markdown export · ⌘⇧E"); }.font(.caption).foregroundStyle(.secondary) }.padding(28).frame(maxWidth: .infinity, alignment: .leading) } }
}

struct StructureBar: View { let note: Note; @ObservedObject var library: Library
    var body: some View { HStack(spacing: 6) { Text("STRUCTURE").font(.caption2).foregroundStyle(.secondary); Button("Heading") { insert("## ") }; Button("Quote") { insert("> ") }; Button("List") { insert("- ") }; Button("Divider") { insert("\n---\n") } }.buttonStyle(.bordered).controlSize(.small) }
    private func insert(_ value: String) { var updated = note; updated.body += (updated.body.isEmpty ? "" : "\n\n") + value; updated.dirty = true; library.update(updated) }
}

struct PreviewPane: View { let note: Note; var body: some View { ScrollView { VStack(alignment: .leading, spacing: 18) { Text(note.title.isEmpty ? "Untitled note" : note.title).font(.system(size: 40, design: .serif)); Text(note.description).foregroundStyle(.secondary); Divider(); if let rendered = try? AttributedString(markdown: note.body) { Text(rendered).font(.system(size: 17, design: .serif)).lineSpacing(7) } else { Text(note.body).font(.system(size: 17, design: .serif)).lineSpacing(7) } }.padding(42).frame(maxWidth: .infinity, alignment: .leading) }.background(Color(nsColor: .textBackgroundColor)) } }
