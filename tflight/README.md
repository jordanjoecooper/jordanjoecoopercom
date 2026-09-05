# TFlight

TFlight is a native macOS writing studio for this Astro site. It keeps the writing experience local and focused while making the publishing boundary explicit: an export writes a normal Markdown post and its metadata into the repository for review in Git.

## Run

From the repository root:

```sh
swift run --package-path tflight
```

The first launch asks you to choose the site folder. It expects `src/content/posts/` and writes exported posts there. Keep images in `src/assets/posts/<slug>/` and reference them from the Markdown body or front matter using the relative paths documented in the site README.

## Current workflow

- Connect an Astro repository from the sidebar.
- Create or select a note.
- Choose a writing font, edit title/description/date/draft/pinned state, and write in the Markdown-compatible canvas.
- Use Preview to read the finished piece, then Export or `⌘⇧E`.
- Run `npm run build` and review the generated Markdown/assets in Git before publishing.

This is intentionally a separate desktop target; the Astro app remains static and deployment-safe.
