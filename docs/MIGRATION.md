# Migrating jordanjoecooper.com to Astro

This guide covers the move from the legacy static site, retained in `archive/`, to the Astro site at the repository root. The target is a static GitHub Pages deployment with the same custom domain, content URLs that are easy to author, and SEO generated as part of every build.

## What the new site already does

The Astro app has one content source for blog posts: `src/content/posts/*.md`.

- Published posts create `/posts/<slug>/` pages, appear on `/writing/`, enter `feed.xml`, and enter the sitemap.
- Drafts create none of those public outputs.
- Every page receives a canonical URL, description, Open Graph metadata, Twitter metadata, RSS discovery, and JSON-LD structured data.
- A post hero image is optimized for the article page and becomes the sharing image.
- `npm run build`, which GitHub Actions runs before deploying, validates published content and builds the sitemap and RSS feed.

## Legacy-site audit

The root site is an HTML/CSS site. It has a good small footprint, but content and metadata are duplicated across individual files.

| Area | Current root site | Astro target | Migration action |
| --- | --- | --- | --- |
| Home, writing, feed | Hand-maintained HTML/XML | Content collection, RSS route, writing route | Use Astro as the source of truth after cutover. |
| Blog posts | `posts/*.html` | `src/content/posts/*.md` | Migrate body copy and front matter; publish only completed posts. |
| Existing post URLs | `/posts/<slug>.html` | `/posts/<slug>/` | Keep the generated compatibility pages during cutover. |
| Non-blog pages | Root HTML | Astro wrapper imports the legacy body | Leave the wrapper in place initially; convert a page only when changing it. |
| Images | `archive/images/` | Existing directory is reused; new post images go in `src/assets/posts/` | Do not delete archived images until non-blog pages are fully converted. |
| CSS | `archive/styles.css` | Imported by `src/styles/global.css`, then extended | Keep shared CSS until the last legacy page is converted. |

### Findings to carry into the cutover

- **Content drift is already visible.** The legacy `writing.html`, `feed.xml`, and `sitemap.xml` list four posts, while the `posts/` directory contains newer article files such as `career-blindness.html` and `moving-fast-vs-looking-fast.html`. Astro derives its writing list, feed, and sitemap from the same collection, so they cannot get out of sync.
- **Metadata is duplicated per post.** Each legacy post carries its own title, canonical URL, social fields, and RSS link. The Astro layout owns those fields once, with per-post front matter supplying only the content that changes.
- **The current deployment has two URL conventions.** Legacy links end in `.html`; Astro uses trailing slashes. The generated compatibility endpoints preserve existing shared links during the move.
- **The remaining static pages are transitional.** `about`, `experience`, both CV pages, and `advisory` are wrapped rather than rewritten. This is safe for cutover, but their archived HTML and images remain deployment dependencies.

## How to publish a post

1. Run `npm run new:post` from the repository root.
2. Write the draft in `src/content/posts/<slug>.md`.
3. For a hero image, copy the image to `src/assets/posts/<slug>/`, add `heroImage` and meaningful `heroAlt` front matter, then use normal Markdown for inline images.
4. Run `npm run build`. It refuses empty published posts and produces the exact static site GitHub Pages will serve.
5. When the post is ready, change `draft: true` to `draft: false`, run the build again, then commit and push.

## Cutover checklist

1. Commit the repository-root Astro files, `archive/`, and `.github/workflows/deploy-pages.yml`. An uncommitted migration cannot reach GitHub Pages.
2. In GitHub repository settings, set Pages source to **GitHub Actions**.
3. Add `GITHUB_CONTRIBUTIONS_TOKEN` as an Actions secret if the activity calendar should be published. The build succeeds without it and simply omits that section.
4. Push to `main` and confirm the Actions deployment completes.
5. Check these deployed URLs:

   - `/`
   - `/writing/`
   - `/posts/avoiding-persian-messenger-syndrome/`
   - `/posts/avoiding-persian-messenger-syndrome.html` (compatibility page)
   - `/feed.xml`
   - `/sitemap-index.xml`

6. In Google Search Console, submit the new sitemap. Keep the legacy `.html` compatibility pages for at least several months so old links can move to the canonical URLs.
7. Only after traffic and search indexing have settled, migrate the remaining static pages from their root HTML files and then remove the legacy files and shared asset dependencies in a separate change.

## Why the compatibility pages matter

GitHub Pages serves static files and cannot create server-side HTTP 301 redirects. Astro now generates a small `.html` compatibility page for every published article. It immediately sends visitors to the new trailing-slash URL and tells crawlers that the new URL is canonical. This protects existing shared links while moving the site to cleaner URLs.

## Deployment SEO reference

| Output | Source | Included on deployment |
| --- | --- | --- |
| Page title and description | Page/post front matter | Yes |
| Canonical URL | `Astro.site` and current route | Yes |
| Open Graph and Twitter image | Hero image, otherwise profile image | Yes |
| BlogPosting JSON-LD | Article title, date, description, canonical URL, image | Yes |
| RSS | Published content collection | Yes |
| Sitemap | `@astrojs/sitemap` | Yes |
| Robots file | `src/pages/robots.txt.ts` | Yes |

## Known follow-up work

The non-blog pages are deliberately transitional: the Astro route reads their legacy HTML body. It avoids a risky all-at-once redesign, but it also means the old markup and image paths still need to exist. Convert those pages one by one when you next edit them, starting with `about`, then `experience`, then the CV pages.
