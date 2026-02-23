# jordanjoecooper.com

Personal website for Jordan Joe Cooper. Fully static HTML — no build process, no dependencies, no frameworks.

## Architecture

- Pure static HTML + CSS
- No JavaScript
- Hosted on GitHub Pages
- No build step — edit files directly

## Key files

- `index.html` — Homepage
- `experience.html` — Work experience timeline
- `posts/` — Blog posts (each is a standalone HTML file)
- `post-template.html` — Template for new posts
- `styles.css` — All styles, uses CSS custom properties for theming and dark mode
- `feed.xml` — RSS feed (updated by post CLI when adding posts)
- `cmd/postcli/` — Go CLI for posts (no external deps). Build: **`go build -o postcli ./cmd/postcli`**. Run from repo root.

## Adding a new post

**Using the CLI (recommended)** — From repo root:
- **`./postcli`** or **`go run ./cmd/postcli`** — interactive menu (New / Edit / List / Quit).
- **`./postcli new`** — create a new post (prompts for title, description, keywords, date, slug). Creates the file and updates index.html + feed.xml.
- **`./postcli edit`** — list posts, pick one to open in **$EDITOR** (default `nano`).
- **`./postcli edit <slug>`** — open that post directly.
- **`./postcli list`** — list all posts.
- **`./postcli update-links posts/your-post.html`** — add a manually created post to the homepage and feed.

## Important reminders

- **Run the CLI when publishing a new post** so the homepage and RSS stay in sync. `postcli new` updates index and feed automatically. If you add a post file manually, run **`postcli update-links posts/your-post.html`** once.
- Styles use CSS custom properties (`:root` variables) — update these for theme changes rather than editing individual colour values throughout the file.
- The site supports automatic dark mode via `prefers-color-scheme` media query.
- The accent colour is a warm terracotta (`#b05a3a` light / `#d4775a` dark). To change it, update the `--accent` and `--accent-hover` values in both the `:root` and dark mode blocks in `styles.css`.
- Every HTML page should include `<link rel="alternate" type="application/rss+xml" ...>` in the `<head>` for RSS discovery.
