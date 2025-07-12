# Jordan's Static Blog

A lightweight, SEO-optimized static blog built with JavaScript, Markdown, and HTML templates. This system provides all the essential features of a modern blog while maintaining simplicity and performance.

## Features

- **Markdown Support**: Write content in Markdown with YAML frontmatter
- **SEO Optimization**: Built-in meta tags, structured data, and sitemaps
- **RSS Feed**: Automatic RSS feed generation for subscribers
- **Responsive Design**: Mobile-first design that works on all devices
- **Performance**: Static HTML pages with minified assets
- **Easy Deployment**: Deploy to GitHub Pages, Netlify, or any static hosting

## Project Structure

```
jordan/
├── content/
│   └── posts/          # Markdown blog posts
├── templates/          # HTML templates
├── scripts/
│   └── build.js       # Build script
├── dist/              # Generated static files
├── images/            # Static images
├── styles.css         # Main stylesheet
└── package.json       # Dependencies and scripts
```

## Getting Started

### Prerequisites

- Node.js (v14 or higher)
- npm

### Installation

1. Clone the repository:
```bash
git clone <your-repo-url>
cd jordan
```

2. Install dependencies:
```bash
npm install
```

3. Configure your site settings:
   - Copy `env.example` to `.env`
   - Update the values in `.env` with your information:
   ```bash
   SITE_URL=https://yourdomain.com
   SITE_TITLE=Your Blog Title
   SITE_DESCRIPTION=Personal blog about technology, thoughts, and experiences
   SITE_AUTHOR=Your Name
   SITE_LOGO_TEXT=Y
   ```

### Building the Site

Run the build script to generate the static site:

```bash
npm run build
```

This will:
- Process all Markdown files in `content/posts/`
- Generate HTML pages using templates
- Create an index page with all posts
- Generate RSS feed and sitemap
- Minify CSS and HTML
- Copy static assets

### Development

To build and serve the site locally:

```bash
npm run dev
```

This will build the site and start a local server at `http://localhost:8000`.

## Local Development & Preview

To build the site:

```sh
npm run build
```

To preview the site locally with working navigation, run:

```sh
npm run serve
```

This will start a local server (using [serve](https://www.npmjs.com/package/serve)) and open your site at http://localhost:3000 (or another port if 3000 is in use).

**Note:**
- Navigation links use root-relative paths (e.g., `/writing.html`) for best compatibility in production and on GitHub Pages.
- Opening HTML files directly in your browser (using the `file://` protocol) will break navigation. Always use a local server for previewing the site.

## Writing Posts

### Creating a New Post

1. Create a new `.md` file in `content/posts/`
2. Add YAML frontmatter with metadata:

```markdown
---
title: "Your Post Title"
description: "A brief description of your post"
date: "2025-01-15"
img: "../images/your-image.jpg"
keywords: "keyword1, keyword2, keyword3"
categories: ["Category1", "Category2"]
---

# Your Post Content

Write your post content in Markdown here...
```

### Frontmatter Options

- `title`: Post title (required)
- `description`: Post description for SEO (required)
- `date`: Publication date in YYYY-MM-DD format
- `img`: Path to featured image (optional)
- `keywords`: Comma-separated keywords for SEO
- `categories`: Array of categories

### Markdown Features

The system supports standard Markdown plus:

- **Code blocks** with syntax highlighting
- **Images** with lazy loading
- **Links** with proper styling
- **Lists** and **blockquotes**
- **Tables** and **footnotes**

## Customization

### Templates

Edit the HTML templates in `templates/` to customize the layout:

- `post.html`: Template for individual blog posts
- `index.html`: Template for the homepage

### Styling

Modify `styles.css` to customize the appearance. The CSS includes:

- Responsive grid layout
- Modern typography
- Hover effects and animations
- Mobile-first design

### Build Script

The build script (`scripts/build.js`) can be customized to:

- Add new content types
- Modify the build process
- Add custom plugins
- Change output format

## Deployment

### GitHub Pages

1. Push your code to GitHub
2. Enable GitHub Pages in repository settings
3. Set source to `/docs` or `/` branch
4. Update `siteMeta.url` in build script

### Netlify

1. Connect your GitHub repository to Netlify
2. Set build command: `npm run build`
3. Set publish directory: `dist`
4. Deploy automatically on push

### Other Platforms

The generated `dist/` folder contains static files that can be deployed to:

- Vercel
- AWS S3
- Cloudflare Pages
- Any static hosting service

## SEO Features

### Meta Tags

Each page includes:
- Open Graph tags for social sharing
- Twitter Card support
- Proper meta descriptions
- Canonical URLs

### Structured Data

JSON-LD schema markup for:
- BlogPosting schema
- Person schema
- Organization schema

### Technical SEO

- XML sitemap generation
- RSS feed for content syndication
- Proper heading structure
- Image alt tags
- Lazy loading

## Performance

### Optimizations

- **Static Generation**: All pages pre-built as HTML
- **Asset Minification**: CSS and HTML compressed
- **Lazy Loading**: Images load only when needed
- **No JavaScript**: Pure HTML/CSS for maximum speed

### Performance Tips

1. Optimize images before adding to `images/`
2. Use descriptive alt tags for accessibility
3. Keep posts focused and well-structured
4. Regularly update dependencies

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test the build process
5. Submit a pull request

## License

This project is open source and available under the [MIT License](LICENSE).

## Support

If you have questions or need help:

1. Check the documentation
2. Look at existing issues
3. Create a new issue with details

---

Built with ❤️ using JavaScript, Markdown, and HTML.

## Automation

### Local Git Hooks

The project includes a `post-commit` git hook that automatically:

1. Rebuilds the site after each commit
2. Commits any changes to the `dist/` directory
3. Provides feedback on build success/failure

The hook is located at `.git/hooks/post-commit` and runs automatically when you commit changes.

### GitHub Actions

A GitHub Actions workflow (`.github/workflows/deploy.yml`) automatically:

1. Builds the site on every push to `main`
2. Deploys to GitHub Pages using the `gh-pages` branch
3. Runs on both pushes and pull requests

To enable GitHub Pages deployment:

1. Go to your repository settings
2. Navigate to "Pages" section
3. Set source to "Deploy from a branch"
4. Select `gh-pages` branch and `/ (root)` folder
5. Save the settings

The workflow will automatically deploy your site to `https://yourusername.github.io/your-repo-name/`

### Manual Build

You can still build manually anytime:

```bash
npm run build
```
