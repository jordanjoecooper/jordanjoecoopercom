# jordanjoecooper.com

Personal site for Jordan Joe Cooper, built as a static Astro site and deployed to GitHub Pages.

## Architecture

- Astro 5 static output; no server runtime
- GitHub Actions runs `npm ci`, then `npm run build`, then deploys `dist/`
- Legacy HTML, images, Go post CLI, and historical feeds live in `archive/`
- The Astro app is the repository root; do not add new site work to `archive/`

## Key files

- `src/pages/index.astro` — Homepage
- `src/pages/writing.astro` — Writing index
- `src/pages/posts/[slug].astro` — Post route
- `src/content/posts/` — Markdown posts and front matter
- `src/content.config.ts` — Post schema
- `src/layouts/BaseLayout.astro` — Shared SEO, canonical, social, and structured data
- `src/styles/global.css` — Astro styling layer; it imports legacy CSS from `archive/styles.css` while transitional pages remain
- `docs/MIGRATION.md` — Cutover and legacy-site migration notes

## Adding a post

From the repository root:

- `npm run new:post` — create a guided draft and optional image folder
- Write the post in `src/content/posts/<slug>.md`
- Put post images in `src/assets/posts/<slug>/`
- Set `draft: false` only when the article is ready
- Run `npm run build` before committing; it validates content and generates RSS, sitemap, SEO metadata, and the static site

Posts can use `heroImage` and `heroAlt` front matter. Hero images are optimized and used in social metadata. See `README.md` for the exact front matter and inline-image syntax.

## Important reminders

- Do not manually edit generated `dist/` output.
- Keep legacy `.html` post compatibility endpoints in `src/pages/posts/[slug].html.ts` until existing links and search indexing have migrated.
- Do not remove files from `archive/` until the corresponding wrapped legacy pages have been converted.
- The GitHub Actions workflow is `.github/workflows/deploy-pages.yml`; it assumes the Astro app lives at the repository root.
