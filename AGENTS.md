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
- `feed.xml` — RSS feed (updated via script when adding posts)
- `scripts/new-post.js` — Interactive CLI to create a new post (prompts for title, description, keywords, date, slug) and writes `posts/{slug}.html`
- `scripts/update-post-links.js` — Updates index.html and feed.xml from a new post’s meta tags

## Adding a new post

**Option A — CLI (recommended)**
1. Run **`node scripts/new-post.js`** and answer the prompts (title, description, keywords, date, slug). This creates `posts/{slug}.html` from the template, fills all meta fields, and adds the new post to the top of the homepage Writing section and to `feed.xml`.
2. Edit the post body in the new file.

**Option B — Manual**
1. Copy `post-template.html` to `posts/your-post-name.html` and fill placeholders.
2. Run **`node scripts/update-post-links.js posts/your-post-name.html`** to add to the homepage and feed.

## Important reminders

- **Run the script when publishing a new post** so the homepage and RSS stay in sync. If you add or edit a post manually in index.html or feed.xml, the script does not remove duplicates — run it only once per new post.
- Styles use CSS custom properties (`:root` variables) — update these for theme changes rather than editing individual colour values throughout the file.
- The site supports automatic dark mode via `prefers-color-scheme` media query.
- The accent colour is a warm terracotta (`#b05a3a` light / `#d4775a` dark). To change it, update the `--accent` and `--accent-hover` values in both the `:root` and dark mode blocks in `styles.css`.
- Every HTML page should include `<link rel="alternate" type="application/rss+xml" ...>` in the `<head>` for RSS discovery.
