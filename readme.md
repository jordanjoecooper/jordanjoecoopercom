# Jordan Joe Cooper - Personal Site

A fully static personal website. No build process, no dependencies, just pure HTML and CSS.

## Features

- **Fully Static**: Pure HTML files, no build step required
- **Minimal Design**: Clean, focused design that loads instantly
- **Easy to Edit**: Just edit HTML files directly
- **Fast**: No JavaScript, no frameworks, minimal CSS

## How to Add a New Post

1. Copy `post-template.html` to `posts/your-post-name.html`
2. Replace the placeholders:
   - `POST_TITLE` → Your post title
   - `POST_DESCRIPTION` → Brief description
   - `YYYY-MM-DD` → Date in ISO format
   - `Month Day, Year` → Human-readable date
3. Write your content in HTML (or copy from markdown and convert)
4. Add a link to your new post in:
   - `writing.html` (add to the list)
   - `index.html` (add to recent posts if desired)

That's it! No build step, no dependencies, just edit and deploy.

## Project Structure

```
├── index.html          # Homepage
├── writing.html        # All posts listing
├── about.html          # About page
├── post-template.html  # Template for new posts
├── posts/             # Individual post HTML files
│   ├── aphorisms.html
│   └── about-this-site.html
├── styles.css         # Minimal CSS
└── images/            # Static images (if any)
```

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
