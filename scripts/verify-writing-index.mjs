import { readFile, readdir } from 'node:fs/promises';

const postsDirectory = new URL('../src/content/posts/', import.meta.url);
const publicOutputs = await Promise.all([
  'index.html',
  'writing/index.html',
  'feed.xml',
  'sitemap-0.xml',
].map((path) => readFile(new URL(`../dist/${path}`, import.meta.url), 'utf8')));
const publishedSlugs = [];
const draftSlugs = [];

for (const name of await readdir(postsDirectory)) {
  if (!name.endsWith('.md')) continue;

  const source = await readFile(new URL(name, postsDirectory), 'utf8');
  const isDraft = /^draft:\s*true\s*$/m.test(source);
  const slug = name.replace(/\.md$/, '');
  if (isDraft) draftSlugs.push(slug);
  else publishedSlugs.push(slug);
}

const [homePage, writingIndex, feed, sitemap] = publicOutputs;
const publicationOutputs = [homePage, writingIndex, feed, sitemap];

if (!publishedSlugs.length || publishedSlugs.some((slug) => !writingIndex.includes(`/posts/${slug}/`))) {
  throw new Error('Writing index is missing one or more published posts.');
}

if (draftSlugs.some((slug) => publicationOutputs.some((output) => output.includes(`/posts/${slug}`)))) {
  throw new Error('A draft post appears in a public listing.');
}

for (const slug of draftSlugs) {
  try {
    await readFile(new URL(`../dist/posts/${slug}/index.html`, import.meta.url));
    throw new Error(`A draft post route was generated: ${slug}`);
  } catch (error) {
    if (error.code !== 'ENOENT') throw error;
  }
}

console.log(`Writing index includes ${publishedSlugs.length} published posts.`);
