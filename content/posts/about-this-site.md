---
title: "How This Site is Built"
description: "A behind-the-scenes look at the technology, structure, and workflow powering jordanjoecooper.com."
date: "2025-07-12"
keywords: "static site, blog, JavaScript, Handlebars, markdown, GitHub Pages, SEO"
categories: ["Tech", "Meta"]
---

# How This Site is Built

This site is a lightweight, SEO-optimized static blog built with JavaScript, Markdown, and HTML templates. The goal is to provide all the essential features of a modern blog while keeping things simple, fast, and easy to maintain.

## Key Features

- **Markdown Content**: All posts are written in Markdown with YAML frontmatter for metadata (title, date, description, etc.).
- **Static Generation**: A custom Node.js build script processes Markdown files, applies Handlebars templates, and outputs minified static HTML.
- **SEO & Social**: Every page includes meta tags, Open Graph, Twitter Card, and JSON-LD structured data for rich previews and discoverability.
- **RSS & Sitemap**: The build process generates an RSS feed and XML sitemap automatically.
- **Responsive & Fast**: The design is mobile-first, with no client-side JavaScript, and all assets are minified for speed.
- **Easy Deployment**: The site is deployed to GitHub Pages using GitHub Actions, but can be hosted anywhere that serves static files.

## Project Structure

```
jordanjoecoopercom/
├── content/
│   └── posts/          # Markdown blog posts
├── templates/          # Handlebars HTML templates
├── scripts/
│   └── build.js        # Build script
├── dist/               # Generated static files
├── images/             # Static images
├── styles.css          # Main stylesheet
└── package.json        # Dependencies and scripts
```

## How It Works

1. **Write in Markdown**: Each post lives in `content/posts/` as a `.md` file with YAML frontmatter.
2. **Build**: Run `npm run build` to process all posts, apply templates, generate the homepage, RSS, sitemap, and minify everything into `dist/`.
3. **Templates**: Layout and structure are handled by Handlebars templates in `templates/`. You can customize the look and feel by editing these files and `styles.css`.
4. **Deploy**: The `dist/` folder is deployed to GitHub Pages via a workflow, but you can use Netlify, Vercel, or any static host.

## Customization & Extensibility

- **Add new templates** or partials for different page types.
- **Edit the build script** (`scripts/build.js`) to add features or change the workflow.
- **Tweak SEO** by editing meta tags in the header partial.
- **Style** the site by editing `styles.css`.

## Why This Approach?

- **Performance**: Pure static HTML is as fast as it gets.
- **Simplicity**: No databases, no frameworks, no client-side JS.
- **Portability**: The output works anywhere you can serve static files.
- **Control**: Full control over markup, SEO, and build process.

## Want to Build Your Own?

Check out the [README](https://github.com/jordanjoecooper/jordanjoecoopercom/blob/main/README.md) for setup instructions, or fork the repo and start customizing! 