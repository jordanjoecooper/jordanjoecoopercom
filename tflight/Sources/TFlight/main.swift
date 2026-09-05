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
            CommandGroup(after: .saveItem) { Button("Save working copy") { library.saveCurrent() }.keyboardShortcut("s"); Button("Export to Astro…") { library.exportCurrent() }.keyboardShortcut("e", modifiers: [.command, .shift]) }
        }
    }
}

struct Note: Identifiable, Hashable, Codable {
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
        repositoryURL = root
        let loaded = (try? FileManager.default.contentsOfDirectory(at: posts, includingPropertiesForKeys: nil).filter { $0.pathExtension == "md" }.compactMap(read))?.sorted { $0.date > $1.date } ?? []
        notes = restoreWorkingCopies(over: loaded)
        selectedID = notes.first?.id; message = "Ready · \(notes.count) notes"
    }

    func newNote() { let note = Note(id: "note-\(Int(Date().timeIntervalSince1970))"); notes.insert(note, at: 0); selectedID = note.id; persistWorkingCopies() }

    func update(_ note: Note) { selected = note; persistWorkingCopies() }

    func discardLocalDraft() {
        guard let selected, selected.id.hasPrefix("note-") else { message = "Repository posts are protected"; return }
        notes.removeAll { $0.id == selected.id }; selectedID = notes.first?.id; persistWorkingCopies(); message = "Local draft discarded"
    }

    func saveCurrent() { guard let selected else { return }; if write(selected) { message = "Saved just now" } }

    func exportCurrent() {
        guard let selected else { return }
        let title = selected.title.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !title.isEmpty else { message = "Add a title before exporting"; return }
        guard !selected.body.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { message = "Write something before exporting"; return }
        guard selected.heroImage.isEmpty || !selected.heroAlt.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else { message = "Add hero image alt text before exporting"; return }
        if write(selected) { message = "Exported to Astro · ready for Git" }
    }

    func attachImage(asHero: Bool = false) {
        guard var note = selected, let root = repositoryURL else { return }
        if note.id.hasPrefix("note-"), !note.title.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty { write(note); note = selected ?? note }
        let panel = NSOpenPanel(); panel.allowedContentTypes = [.image, .audio, .movie]; panel.allowsMultipleSelection = false; panel.prompt = asHero ? "Choose hero image" : "Add to note"
        guard panel.runModal() == .OK, let source = panel.url else { return }
        guard !asHero || UTType(filenameExtension: source.pathExtension)?.conforms(to: .image) == true else { message = "Hero images must be image files"; return }
        let slug = note.id.hasPrefix("note-") ? note.title.lowercased().replacingOccurrences(of: "[^a-z0-9]+", with: "-", options: .regularExpression).trimmingCharacters(in: CharacterSet(charactersIn: "-")) : note.id
        let finalSlug = slug.isEmpty ? note.id : slug
        let folder = root.appendingPathComponent("src/assets/posts/\(finalSlug)"); try? FileManager.default.createDirectory(at: folder, withIntermediateDirectories: true)
        let filename = source.lastPathComponent.lowercased().replacingOccurrences(of: "[^a-z0-9._-]+", with: "-", options: .regularExpression)
        let destination = folder.appendingPathComponent(filename)
        do { try FileManager.default.copyItem(at: source, to: destination) }
        catch { message = "Could not copy media: \(error.localizedDescription)"; return }
        var updated = note
        let relativePath = "../../assets/posts/\(finalSlug)/\(filename)"
        if asHero { updated.heroImage = relativePath; updated.heroAlt = source.deletingPathExtension().lastPathComponent.replacingOccurrences(of: "-", with: " ") }
        else if UTType(filenameExtension: source.pathExtension)?.conforms(to: .image) == true { updated.body += "\n\n![\(source.deletingPathExtension().lastPathComponent)](\(relativePath))\n" }
        else { updated.body += "\n\n[\(filename)](\(relativePath))\n" }
        update(updated); message = asHero ? "Hero image set · remember to export" : "Image added · remember to export"
    }

    private var workingCopyKey: String { "tflight.working-copies.\(repositoryURL?.path ?? "unconnected")" }

    private func persistWorkingCopies() {
        guard let data = try? JSONEncoder().encode(notes) else { return }
        UserDefaults.standard.set(data, forKey: workingCopyKey)
    }

    private func restoreWorkingCopies(over loaded: [Note]) -> [Note] {
        guard let data = UserDefaults.standard.data(forKey: workingCopyKey), let cached = try? JSONDecoder().decode([Note].self, from: data) else { return loaded }
        let byID = Dictionary(uniqueKeysWithValues: cached.map { ($0.id, $0) })
        var restored = loaded.map { note -> Note in var value = byID[note.id] ?? note; value.body = normaliseBody(value.body); return value }
        restored.insert(contentsOf: cached.filter { $0.id.hasPrefix("note-") }, at: 0)
        return restored.sorted { $0.date > $1.date }
    }

    private func read(_ url: URL) -> Note? {
        guard let source = try? String(contentsOf: url, encoding: .utf8) else { return nil }
        let parts = source.components(separatedBy: "\n---\n"); let front = parts.first?.replacingOccurrences(of: "---\n", with: "") ?? ""; let body = parts.dropFirst().joined(separator: "\n---\n")
        func value(_ key: String) -> String { front.split(separator: "\n").first(where: { $0.hasPrefix(key + ":") }).map { String($0.dropFirst(key.count + 1)).trimmingCharacters(in: .whitespaces).trimmingCharacters(in: CharacterSet(charactersIn: "\"")) } ?? "" }
        let formatter = ISO8601DateFormatter(); formatter.formatOptions = [.withFullDate]
        return Note(id: url.deletingPathExtension().lastPathComponent, title: value("title"), description: value("description"), date: formatter.date(from: value("pubDate")) ?? Date(), draft: value("draft") != "false", pinned: value("pinned") == "true", keywords: value("keywords"), body: normaliseBody(body), heroImage: value("heroImage"), heroAlt: value("heroAlt"))
    }

    private func normaliseBody(_ source: String) -> String {
        guard source.range(of: "<[^>]+>", options: .regularExpression) != nil else { return source.trimmingCharacters(in: .whitespacesAndNewlines) }
        var value = source
        for level in 1...6 { value = value.replacingOccurrences(of: "(?is)<h\(level)[^>]*>(.*?)</h\(level)>", with: String(repeating: "#", count: level) + " $1\n\n", options: .regularExpression) }
        value = value.replacingOccurrences(of: "(?is)<img[^>]*alt=[\"']([^\"']*)[\"'][^>]*src=[\"']([^\"']+)[\"'][^>]*>", with: "![$1]($2)\n\n", options: .regularExpression)
        value = value.replacingOccurrences(of: "(?is)<img[^>]*src=[\"']([^\"']+)[\"'][^>]*>", with: "![]($1)\n\n", options: .regularExpression)
        value = value.replacingOccurrences(of: "(?is)<a[^>]*href=[\"']([^\"']+)[\"'][^>]*>(.*?)</a>", with: "[$2]($1)", options: .regularExpression)
        value = value.replacingOccurrences(of: "(?is)<(strong|b)[^>]*>", with: "**", options: .regularExpression).replacingOccurrences(of: "(?is)</(strong|b)>", with: "**", options: .regularExpression)
        value = value.replacingOccurrences(of: "(?is)<(em|i)[^>]*>", with: "*", options: .regularExpression).replacingOccurrences(of: "(?is)</(em|i)>", with: "*", options: .regularExpression)
        value = value.replacingOccurrences(of: "(?is)<li[^>]*>", with: "- ", options: .regularExpression).replacingOccurrences(of: "(?is)</li>", with: "\n", options: .regularExpression)
        value = value.replacingOccurrences(of: "(?is)<blockquote[^>]*>", with: "> ", options: .regularExpression).replacingOccurrences(of: "(?is)</blockquote>", with: "\n\n", options: .regularExpression)
        value = value.replacingOccurrences(of: "(?is)<br\\s*/?>", with: "\n", options: .regularExpression)
        value = value.replacingOccurrences(of: "(?is)</?(p|div|ul|ol)[^>]*>", with: "\n\n", options: .regularExpression)
        value = value.replacingOccurrences(of: "<[^>]+>", with: "", options: .regularExpression)
        value = value.replacingOccurrences(of: "&nbsp;", with: " ").replacingOccurrences(of: "&amp;", with: "&").replacingOccurrences(of: "&quot;", with: "\"").replacingOccurrences(of: "&#39;", with: "'").replacingOccurrences(of: "&lt;", with: "<").replacingOccurrences(of: "&gt;", with: ">")
        value = value.replacingOccurrences(of: "(?m)^\\*{1,3}\\s*$", with: "", options: .regularExpression)
        return value.replacingOccurrences(of: "\n{3,}", with: "\n\n", options: .regularExpression).trimmingCharacters(in: .whitespacesAndNewlines)
    }

    @discardableResult private func write(_ note: Note) -> Bool {
        guard let root = repositoryURL else { message = "Connect an Astro site first"; return false }
        let slug = note.id.hasPrefix("note-") ? note.title.lowercased().replacingOccurrences(of: "[^a-z0-9]+", with: "-", options: .regularExpression).trimmingCharacters(in: CharacterSet(charactersIn: "-")) : note.id
        let finalSlug = slug.isEmpty ? note.id : slug
        let posts = root.appendingPathComponent("src/content/posts"); try? FileManager.default.createDirectory(at: posts, withIntermediateDirectories: true)
        let formatter = ISO8601DateFormatter(); formatter.formatOptions = [.withFullDate]
        let imageLine = note.heroImage.isEmpty ? "" : "heroImage: \(note.heroImage)\nheroAlt: \"\(note.heroAlt.replacingOccurrences(of: "\"", with: "\\\""))\"\n"
        let yaml = "---\ntitle: \"\(note.title.replacingOccurrences(of: "\"", with: "\\\""))\"\ndescription: \"\(note.description.replacingOccurrences(of: "\"", with: "\\\""))\"\npubDate: \(formatter.string(from: note.date))\ndraft: \(note.draft)\npinned: \(note.pinned)\nkeywords: \"\(note.keywords)\"\n\(imageLine)---\n\n\(note.body.trimmingCharacters(in: .whitespacesAndNewlines))\n"
        do { try yaml.write(to: posts.appendingPathComponent(finalSlug + ".md"), atomically: true, encoding: .utf8) }
        catch { message = "Could not write Astro post: \(error.localizedDescription)"; return false }
        if finalSlug != note.id { if let index = notes.firstIndex(where: { $0.id == note.id }) { notes[index].id = finalSlug; selectedID = finalSlug } }
        return true
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
        Spacer(); if library.selected?.id.hasPrefix("note-") == true { Button { library.discardLocalDraft() } label: { Label("Discard local draft", systemImage: "trash") }.buttonStyle(.plain).foregroundStyle(.secondary) }; Button { library.chooseRepository() } label: { Label(library.repositoryURL == nil ? "Connect Astro site" : "Change site", systemImage: "folder") }.buttonStyle(.plain).foregroundStyle(.secondary); Text(library.message).font(.caption2).foregroundStyle(.secondary).lineLimit(2)
    }.padding(18) }
}

struct Welcome: View { @ObservedObject var library: Library; var body: some View { VStack(spacing: 14) { Text("A quieter place to write.").font(.system(size: 38, design: .serif)); Text("Notes become beautiful, deployable Astro articles.").foregroundStyle(.secondary); Button("Connect your Astro site") { library.chooseRepository() }.buttonStyle(.borderedProminent).tint(.orange) } } }

struct Editor: View {
    @ObservedObject var library: Library
    @AppStorage("tflight.editorFont") private var font = "New York"
    @AppStorage("tflight.showPreview") private var showPreview = true
    @State private var focusMode = false
    var note: Note { library.selected ?? Note(id: "empty") }
    var body: some View { VStack(spacing: 0) {
        HStack { TextField("Untitled note", text: Binding(get: { note.title }, set: { var n = note; n.title = $0; n.dirty = true; library.update(n) })).textFieldStyle(.plain).font(.system(size: 26, design: .serif)); Spacer(); Picker("Font", selection: $font) { Text("New York").tag("New York"); Text("Avenir").tag("Avenir"); Text("Mono").tag("Menlo") }.frame(width: 130); Button("Add media") { library.attachImage() }; Button("Set hero") { library.attachImage(asHero: true) }; Button(focusMode ? "Exit focus" : "Focus") { focusMode.toggle() }; if !focusMode { Button(showPreview ? "Hide preview" : "Preview") { showPreview.toggle() } }; Button("Export") { library.exportCurrent() }.buttonStyle(.borderedProminent).tint(.orange) }.padding(.horizontal, 28).padding(.vertical, 16)
        Divider()
        HStack(spacing: 0) { if focusMode { RichMarkdownCanvas(text: Binding(get: { note.body }, set: { var n = note; n.body = $0; n.dirty = true; library.update(n) }), fontName: font).frame(minHeight: 500).padding(36) } else { FormPane(note: note, library: library, font: font); if showPreview { Divider(); PreviewPane(note: note) } } }
    } }
}

struct FormPane: View { let note: Note; @ObservedObject var library: Library; let font: String; var body: some View { VStack(alignment: .leading, spacing: 16) { TextField("A short description for search and sharing", text: Binding(get: { note.description }, set: { var n = note; n.description = $0; library.update(n) })).textFieldStyle(.plain).foregroundStyle(.secondary); TextField("Keywords · comma separated", text: Binding(get: { note.keywords }, set: { var n = note; n.keywords = $0; library.update(n) })).textFieldStyle(.plain).font(.caption).foregroundStyle(.secondary); if !note.heroImage.isEmpty { HStack(spacing: 8) { Image(systemName: "photo").foregroundStyle(.orange); TextField("Hero image alt text", text: Binding(get: { note.heroAlt }, set: { var n = note; n.heroAlt = $0; library.update(n) })).textFieldStyle(.plain); Image(systemName: note.heroAlt.isEmpty ? "exclamationmark.triangle" : "checkmark.circle").foregroundStyle(note.heroAlt.isEmpty ? .red : .green) } }; Divider(); HStack { DatePicker("", selection: Binding(get: { note.date }, set: { var n = note; n.date = $0; library.update(n) }), displayedComponents: .date).labelsHidden(); Toggle("Draft", isOn: Binding(get: { note.draft }, set: { var n = note; n.draft = $0; library.update(n) })); Toggle("Pinned", isOn: Binding(get: { note.pinned }, set: { var n = note; n.pinned = $0; library.update(n) })) }; VStack(alignment: .leading, spacing: 10) { FormatBar(); StructureBar(note: note, library: library) }; RichMarkdownCanvas(text: Binding(get: { note.body }, set: { var n = note; n.body = $0; n.dirty = true; library.update(n) }), fontName: font).frame(minHeight: 500); HStack { Text("\(note.body.split(whereSeparator: \.isWhitespace).count) words"); Spacer(); Text("Markdown export · ⌘⇧E"); }.font(.caption).foregroundStyle(.secondary) }.padding(28).frame(maxWidth: .infinity, alignment: .leading) }
}

struct RichMarkdownCanvas: NSViewRepresentable {
    @Binding var text: String
    let fontName: String

    func makeCoordinator() -> Coordinator { Coordinator(self) }
    func makeNSView(context: Context) -> NSScrollView {
        let view = NSTextView()
        view.delegate = context.coordinator; view.string = text; view.font = .init(name: fontName, size: 17) ?? .systemFont(ofSize: 17)
        view.isRichText = false; view.isEditable = true; view.isSelectable = true; view.usesFontPanel = false; view.drawsBackground = false; view.textContainerInset = NSSize(width: 4, height: 8); view.autoresizingMask = [.width, .height]
        style(view, fontName: fontName)
        let scroll = NSScrollView(); scroll.drawsBackground = false; scroll.hasVerticalScroller = true; scroll.hasHorizontalScroller = false; scroll.borderType = .noBorder; scroll.documentView = view; return scroll
    }
    func updateNSView(_ scroll: NSScrollView, context: Context) { guard let view = scroll.documentView as? NSTextView else { return }; if view.string != text { view.string = text }; view.font = .init(name: fontName, size: 17) ?? .systemFont(ofSize: 17); style(view, fontName: fontName) }
    private func style(_ view: NSTextView, fontName: String) {
        guard let storage = view.textStorage else { return }; let full = NSRange(location: 0, length: (view.string as NSString).length); let base = NSFont(name: fontName, size: 17) ?? .systemFont(ofSize: 17)
        storage.beginEditing(); storage.setAttributes([.font: base, .foregroundColor: NSColor.textColor], range: full)
        for pattern in ["(?m)^#{1,6}\\s+", "(?m)^>\\s?", "(?m)^[-*]\\s+", "\\*{1,3}", "`{1,3}"] { let expression = try? NSRegularExpression(pattern: pattern); expression?.enumerateMatches(in: view.string, range: full) { match, _, _ in if let match { storage.addAttribute(.foregroundColor, value: NSColor.clear, range: match.range) } } }
        for pattern in ["!\\[[^\\]]*\\]\\([^)]*\\)", "\\[[^]]+\\]\\([^)]*\\)"] {
            guard let expression = try? NSRegularExpression(pattern: pattern) else { continue }
            expression.enumerateMatches(in: view.string, range: full) { match, _, _ in
                guard let match else { return }; let token = (view.string as NSString).substring(with: match.range)
                guard let open = token.firstIndex(of: "["), let close = token.lastIndex(of: ")") else { return }
                let start = token.distance(from: token.startIndex, to: open); let end = token.distance(from: token.startIndex, to: close) + 1
                storage.addAttribute(.foregroundColor, value: NSColor.clear, range: NSRange(location: match.range.location, length: start)); storage.addAttribute(.foregroundColor, value: NSColor.clear, range: NSRange(location: match.range.location + end, length: match.range.length - end))
            }
        }
        storage.endEditing()
    }
    final class Coordinator: NSObject, NSTextViewDelegate { var parent: RichMarkdownCanvas; init(_ parent: RichMarkdownCanvas) { self.parent = parent }; func textDidChange(_ notification: Notification) { guard let view = notification.object as? NSTextView else { return }; parent.text = view.string } }
}

struct FormatBar: View {
    var body: some View { HStack(spacing: 4) { Text("FORMAT").font(.caption2).foregroundStyle(.secondary); Button("Bold") { wrap("**", "**") }; Button("Italic") { wrap("*", "*") }; Button("Code") { wrap("`", "`") }; Button("Link") { wrap("[", "](https://)") } }.buttonStyle(.bordered).controlSize(.small) }
    private func wrap(_ prefix: String, _ suffix: String) {
        guard let root = NSApp.keyWindow?.contentView, let view = findTextView(in: root) else { return }
        let range = view.selectedRange(); let selected = (view.string as NSString).substring(with: range)
        view.insertText(prefix + (selected.isEmpty ? "text" : selected) + suffix, replacementRange: range); view.window?.makeFirstResponder(view)
    }
    private func findTextView(in view: NSView) -> NSTextView? { if let textView = view as? NSTextView { return textView }; for child in view.subviews { if let found = findTextView(in: child) { return found } }; return nil }
}

struct StructureBar: View { let note: Note; @ObservedObject var library: Library
    var body: some View { HStack(spacing: 6) { Text("STRUCTURE").font(.caption2).foregroundStyle(.secondary); Button("Heading") { insert("## ") }; Button("Quote") { insert("> ") }; Button("List") { insert("- ") }; Button("Callout") { insert("> **Note:** ") }; Button("Code block") { insert("```\n\n```") }; Button("Divider") { insert("\n---\n") } }.buttonStyle(.bordered).controlSize(.small) }
    private func insert(_ value: String) { var updated = note; updated.body += (updated.body.isEmpty ? "" : "\n\n") + value; updated.dirty = true; library.update(updated) }
}

struct PreviewPane: View { let note: Note; var body: some View { ScrollView { VStack(alignment: .leading, spacing: 18) { Text(note.title.isEmpty ? "Untitled note" : note.title).font(.system(size: 40, design: .serif)); Text(note.description).foregroundStyle(.secondary); Divider(); if let rendered = try? AttributedString(markdown: note.body) { Text(rendered).font(.system(size: 17, design: .serif)).lineSpacing(7) } else { Text(note.body).font(.system(size: 17, design: .serif)).lineSpacing(7) } }.padding(42).frame(maxWidth: .infinity, alignment: .leading) }.background(Color(nsColor: .textBackgroundColor)) } }
