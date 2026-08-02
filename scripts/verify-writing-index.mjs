import { readFile, readdir } from 'node:fs/promises';

const postsDirectory = new URL('../src/content/posts/', import.meta.url);
const writingIndex = await readFile(new URL('../dist/writing/index.html', import.meta.url), 'utf8');
const publishedSlugs = [];

for (const name of await readdir(postsDirectory)) {
  if (!name.endsWith('.md')) continue;

  const source = await readFile(new URL(name, postsDirectory), 'utf8');
  const isDraft = /^draft:\s*true\s*$/m.test(source);
  if (!isDraft) publishedSlugs.push(name.replace(/\.md$/, ''));
}

if (!publishedSlugs.length || publishedSlugs.some((slug) => !writingIndex.includes(`/posts/${slug}/`))) {
  throw new Error('Writing index is missing one or more published posts.');
}

console.log(`Writing index includes ${publishedSlugs.length} published posts.`);
