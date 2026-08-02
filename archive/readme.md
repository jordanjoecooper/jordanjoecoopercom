# jordanjoecooper.com

A fully static personal website — pure HTML and CSS, no build step.

## Post editor

A local web editor for creating and editing blog posts. Saving a post automatically rebuilds the homepage writing list and RSS feed.

**Run from the repo root:**

```
go run ./cmd/postcli
```

This starts a server at `http://localhost:3000` and opens it in your browser.

- **List** — all posts, newest first
- **New Post** — title, description, date, slug (auto-generated from title), contenteditable body
- **Edit** — open any existing post, edit content and metadata inline
- **Save** — writes the HTML file and rebuilds `index.html` and `feed.xml` automatically
- **Cmd+S / Ctrl+S** — keyboard shortcut to save

Requires Go 1.21+. No external dependencies.

## Structure

```
├── index.html          # Homepage
├── experience.html     # Work experience
├── about.html          # About page
├── advisory.html       # Advisory (unlisted)
├── cv.html             # CV (unlisted)
├── lite.html           # Data-light version
├── 404.html            # Custom 404
├── post-template.html  # Template for new posts
├── posts/              # Individual post HTML files
├── cmd/postcli/        # Local post editor (Go)
├── styles.css          # All styles
├── feed.xml            # RSS feed
└── images/             # Static images
```

## Deployment

Pushes to `main` deploy automatically via GitHub Pages (`.github/workflows/static.yml`).
