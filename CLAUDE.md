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
- `feed.xml` — RSS feed (hand-maintained)

## Adding a new post

1. Copy `post-template.html` to `posts/your-post-name.html`
2. Update the title, description, date, meta tags, and content
3. Add a link in the "Writing" section of `index.html`
4. **Update `feed.xml`** — add a new `<item>` entry with the post's title, link, publication date, and description

## Important reminders

- **Always update `feed.xml` when publishing new content.** The RSS feed is hand-maintained and must be updated manually whenever a new post is added or an existing post is modified.
- Styles use CSS custom properties (`:root` variables) — update these for theme changes rather than editing individual colour values throughout the file.
- The site supports automatic dark mode via `prefers-color-scheme` media query.
- The accent colour is a warm terracotta (`#b05a3a` light / `#d4775a` dark). To change it, update the `--accent` and `--accent-hover` values in both the `:root` and dark mode blocks in `styles.css`.
- Every HTML page should include `<link rel="alternate" type="application/rss+xml" ...>` in the `<head>` for RSS discovery.
