# Astro site

This is a static Astro migration of jordanjoecooper.com. It uses Astro's directory-style URLs (for example, `/posts/my-post/`), ships no client-side JavaScript, and reuses the existing `../images` folder while both versions live in this repository.

## Commands

```sh
npm install
npm run dev
npm run build
```

Deploy `dist/` to GitHub Pages. The build emits `CNAME`, so the existing custom domain is retained.

## GitHub contribution calendar

The homepage calendar is a static SVG generated from GitHub's GraphQL API during the production build, so no account token is ever sent to site visitors. Add a fine-grained personal access token as the repository Actions secret `GITHUB_CONTRIBUTIONS_TOKEN`; every push to `main` will refresh the calendar before deployment.

If the secret is unavailable or GitHub cannot provide the calendar, the activity section is omitted entirely. No placeholder is published.

To generate it locally:

```sh
GITHUB_CONTRIBUTIONS_TOKEN=github_pat_… npm run generate:github-activity
```

It writes `src/generated/github-contributions.ts`. The export includes whatever contributions GitHub shows to the token holder, including private contributions when that is enabled in the account's contribution settings. The SVG only publishes per-day activity levels and totals—not repository names or commit details.

For another profile or output location:

```sh
GITHUB_CONTRIBUTIONS_TOKEN=github_pat_… GITHUB_USERNAME=octocat npm run generate:github-activity
```

The homepage uses the light theme by default. A dark GitHub-style export is also available with `GITHUB_CONTRIBUTIONS_THEME=dark` or `--theme dark`.

## Homepage writing

The homepage shows four posts. They are the four newest published posts by default. To keep an older post on the homepage, add `pinned: true` to its frontmatter; pinned posts take priority, and any remaining slots are filled by the newest unpinned posts.

## Writing a post

Start an interactive draft with:

```sh
npm run new:post
```

The command creates a draft in `src/content/posts/` and an image folder in `src/assets/posts/<slug>/`. It asks for the title, description, slug, date, keywords, and optional hero image.

Use Markdown for post bodies. Set `draft: true` to keep a post out of production routes, the writing page, RSS, sitemap, and legacy compatibility routes. A deployment build rejects an empty published post.

```md
---
title: "Post title"
description: "A concise description for search and sharing."
pubDate: 2026-07-21
draft: false
keywords: "optional, comma-separated, keywords"
heroImage: ../../assets/posts/your-slug/hero.jpg # optional
heroAlt: "Describe the image for people who cannot see it" # required when heroImage is set
---

Write the post here.
```

Put post images in `src/assets/posts/your-slug/`. A `heroImage` is automatically resized for the article page and used for Open Graph, Twitter, and Article structured data. Inline images work with standard Markdown:

```md
![Alt text](../../assets/posts/your-slug/diagram.png)
```

Before changing `draft` to `false`, run:

```sh
npm run validate:content
npm run build
```

The deployment workflow already runs `npm run build`, so content validation, sitemap generation, RSS generation, canonical tags, social metadata, and structured data are all included in every production deployment.

## Migration

See [the migration guide](docs/MIGRATION.md) for the legacy-site audit, compatibility plan, and cutover checklist.
