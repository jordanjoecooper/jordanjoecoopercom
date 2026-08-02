import { readdir, readFile } from 'node:fs/promises';
import { join } from 'node:path';

const postsDirectory = new URL('../src/content/posts/', import.meta.url);
const failures = [];

for (const name of await readdir(postsDirectory)) {
  if (!name.endsWith('.md')) continue;

  const path = join(postsDirectory.pathname, name);
  const source = await readFile(path, 'utf8');
  const match = source.match(/^---\s*\n([\s\S]*?)\n---\s*\n?([\s\S]*)$/);

  if (!match) {
    failures.push(`${name}: missing or invalid frontmatter`);
    continue;
  }

  const [, frontmatter, body] = match;
  const get = (field) => frontmatter.match(new RegExp(`^${field}:\\s*["']?(.+?)["']?\\s*$`, 'm'))?.[1]?.trim();
  const isDraft = get('draft') === 'true';
  const readableBody = body.replace(/<[^>]*>/g, '').replace(/&nbsp;/g, ' ').trim();

  if (!get('title')) failures.push(`${name}: missing title`);
  if (!get('description')) failures.push(`${name}: missing description`);
  if (!get('pubDate')) failures.push(`${name}: missing pubDate`);
  if (!isDraft && !readableBody) failures.push(`${name}: published posts need body content (set draft: true until ready)`);
}

if (failures.length) {
  console.error('Content validation failed:\n' + failures.map((failure) => `- ${failure}`).join('\n'));
  process.exit(1);
}

console.log('Content validation passed.');
