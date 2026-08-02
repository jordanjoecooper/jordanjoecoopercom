import { mkdir, writeFile } from 'node:fs/promises';
import { createInterface } from 'node:readline/promises';
import { stdin as input, stdout as output } from 'node:process';

const postsDirectory = new URL('../src/content/posts/', import.meta.url);
const assetsDirectory = new URL('../src/assets/posts/', import.meta.url);
const prompt = createInterface({ input, output });
const answer = async (label, fallback = '') => (await prompt.question(`${label}${fallback ? ` [${fallback}]` : ''}: `)).trim() || fallback;
const slugify = (value) => value.toLowerCase().trim().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '');

try {
  const title = await answer('Title');
  if (!title) throw new Error('A title is required.');

  const description = await answer('Search/social description');
  if (!description) throw new Error('A description is required.');

  const slug = slugify(await answer('Slug', slugify(title)));
  if (!slug) throw new Error('A valid slug is required.');

  const pubDate = await answer('Publication date (YYYY-MM-DD)', new Date().toISOString().slice(0, 10));
  const keywords = await answer('Keywords (comma-separated, optional)');
  const hasHero = (await answer('Add a hero image? (y/N)', 'N')).toLowerCase() === 'y';
  const heroFile = hasHero ? await answer('Image filename (place it in the created asset folder)') : '';
  const heroAlt = heroFile ? await answer('Image alt text') : '';

  if (hasHero && (!heroFile || !heroAlt)) throw new Error('Hero images need both a filename and meaningful alt text.');

  const assetFolder = new URL(`${slug}/`, assetsDirectory);
  await mkdir(assetFolder, { recursive: true });
  const hero = heroFile ? `heroImage: ../../assets/posts/${slug}/${heroFile}\nheroAlt: "${heroAlt.replaceAll('"', '\\"')}"\n` : '';
  const post = `---\ntitle: "${title.replaceAll('"', '\\"')}"\ndescription: "${description.replaceAll('"', '\\"')}"\npubDate: ${pubDate}\ndraft: true\nkeywords: "${keywords.replaceAll('"', '\\"')}"\n${hero}---\n\nWrite your post here.\n`;

  await writeFile(new URL(`${slug}.md`, postsDirectory), post, { flag: 'wx' });
  console.log(`Created src/content/posts/${slug}.md`);
  console.log(`Image folder: src/assets/posts/${slug}/`);
  console.log('The post is a draft. Add content, set draft: false, then run npm run build.');
} catch (error) {
  console.error(error.message);
  process.exitCode = 1;
} finally {
  prompt.close();
}
