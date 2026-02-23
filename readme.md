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
├── post-template.html  # Template for new posts
├── posts/             # Individual post HTML files
│   ├── aphorisms.html
│   └── about-this-site.html
├── styles.css         # All CSS styles
├── feed.xml           # RSS feed
├── scripts/
│   └── update-post-links.js  # Updates homepage + feed when you add a post
└── images/           # Static images
    ├── companies/     # Company logos for timeline
    ├── dp.jpg        # Profile image
    └── enso.jpg      # Footer image
```

## How to Add a New Post

**Quick way (CLI)**
1. Run **`node scripts/new-post.js`** and answer the prompts (title, description, keywords, date, slug). This creates the post file with all meta fields filled and adds the new post to the top of the homepage and RSS feed.
2. Edit the post body in `posts/{slug}.html`.

**Manual way**
1. Copy `post-template.html` to `posts/your-post-name.html` and replace all placeholders.
2. Run **`node scripts/update-post-links.js posts/your-post-name.html`**.

No build step, no npm install (scripts use Node only).

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
