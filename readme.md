# Jordan Joe Cooper - Personal Site

A fully static personal website. No build process, no dependencies, just pure HTML and CSS.

## Features

- **Fully Static**: Pure HTML files, no build step required
- **Minimal Design**: Clean, focused design that loads instantly
- **Easy to Edit**: Just edit HTML files directly
- **Fast**: No JavaScript, no frameworks, minimal CSS

## Project Structure

```
├── index.html          # Homepage
├── experience.html     # Work experience timeline
├── projects.html       # Projects page (currently hidden)
├── post-template.html  # Template for new posts
├── posts/             # Individual post HTML files
│   ├── aphorisms.html
│   └── ...
├── cmd/
│   └── postcli/        # Go CLI for managing posts and links
├── styles.css         # All CSS styles
├── feed.xml           # RSS feed
└── images/           # Static images
    ├── companies/     # Company logos for timeline
    ├── dp.jpg        # Profile image
    └── enso.jpg      # Footer image
```

## How to Add a New Post

**Recommended way (Go CLI)**

From the repo root:

- Build once (optional):
  `go build -o postcli ./cmd/postcli`

- Or run directly with Go:
  `go run ./cmd/postcli`

The CLI supports:

- `./postcli` or `go run ./cmd/postcli` — interactive menu (New / Edit / List / Quit)
- `./postcli new` or `go run ./cmd/postcli new` — create a new post (prompts for title, description, keywords, date, slug).
  This creates `posts/{slug}.html` from `post-template.html` and updates **`index.html`** and **`feed.xml`**.
- `./postcli edit` or `go run ./cmd/postcli edit` — pick a post to edit in `$EDITOR` (or `nano` by default).
- `./postcli edit <slug>` — open that post directly.
- `./postcli list` — list all posts.
- `./postcli update-links posts/your-post.html` — add a manually created post to the homepage and RSS feed.

After running `new`, edit the post body in `posts/{slug}.html`.

**Manual way**
1. Copy `post-template.html` to `posts/your-post-name.html` and replace all placeholders.
2. Run **`go run ./cmd/postcli update-links posts/your-post-name.html`** once to update the homepage and RSS feed.

No build step or npm install required for the site itself; the Go CLI uses only the standard library.

## Deployment

Just upload all the HTML files and `styles.css` to any static hosting:
- GitHub Pages
- Netlify
- Vercel
- Any web server

No build step needed!

## Converting Markdown to HTML

If you prefer writing in markdown, you can:
1. Use any online markdown-to-HTML converter
2. Use a simple tool like `pandoc`: `pandoc post.md -o post.html`
3. Or just write directly in HTML (it's simpler than you think!)

## Design Philosophy

- **Simplicity**: No build tools, no complexity
- **Speed**: Pure HTML/CSS, loads instantly
- **Focus**: Minimal design that doesn't distract
- **Maintainability**: Easy to understand and modify
