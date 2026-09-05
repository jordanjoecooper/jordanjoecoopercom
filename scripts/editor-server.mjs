import { createServer } from 'node:http';
import { mkdir, readFile, readdir, stat, writeFile } from 'node:fs/promises';
import { extname, join, normalize, basename } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = fileURLToPath(new URL('../', import.meta.url));
const postsDirectory = join(root, 'src/content/posts');
const assetsDirectory = join(root, 'src/assets/posts');
const editorDirectory = join(root, 'editor');
const host = '127.0.0.1';
const port = Number(process.env.EDITOR_PORT || 4322);

const send = (res, status, body, type = 'application/json') => {
  res.writeHead(status, { 'Content-Type': `${type}; charset=utf-8`, 'Cache-Control': 'no-store' });
  res.end(type === 'application/json' ? JSON.stringify(body) : body);
};
const safeSlug = (value) => value && /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(value);
const safeFilename = (value) => value && basename(value) === value && /^[a-zA-Z0-9][a-zA-Z0-9._-]*$/.test(value);
const readJson = async (req) => {
  let data = '';
  for await (const chunk of req) data += chunk;
  if (Buffer.byteLength(data) > 3_000_000) throw new Error('Request is too large.');
  return JSON.parse(data || '{}');
};
const yamlValue = (frontmatter, key) => frontmatter.match(new RegExp(`^${key}:\\s*(.*)$`, 'm'))?.[1]?.trim() || '';
const unquote = (value) => value.replace(/^['"]|['"]$/g, '').replaceAll('\\"', '"');
const parsePost = (source, slug) => {
  const match = source.match(/^---\s*\n([\s\S]*?)\n---\s*\n?([\s\S]*)$/);
  if (!match) return { slug, title: slug, description: '', pubDate: '', draft: true, pinned: false, keywords: '', heroImage: '', heroAlt: '', body: source };
  const [, frontmatter, body] = match;
  return { slug, title: unquote(yamlValue(frontmatter, 'title')), description: unquote(yamlValue(frontmatter, 'description')), pubDate: unquote(yamlValue(frontmatter, 'pubDate')), draft: yamlValue(frontmatter, 'draft') !== 'false', pinned: yamlValue(frontmatter, 'pinned') === 'true', keywords: unquote(yamlValue(frontmatter, 'keywords')), heroImage: unquote(yamlValue(frontmatter, 'heroImage')), heroAlt: unquote(yamlValue(frontmatter, 'heroAlt')), body: body.trimStart() };
};
const quote = (value) => `"${String(value || '').replaceAll('\\', '\\\\').replaceAll('"', '\\"').replaceAll('\n', ' ')}"`;
const serializePost = (post) => `---\ntitle: ${quote(post.title)}\ndescription: ${quote(post.description)}\npubDate: ${post.pubDate || new Date().toISOString().slice(0, 10)}\ndraft: ${post.draft ? 'true' : 'false'}\npinned: ${post.pinned ? 'true' : 'false'}\nkeywords: ${quote(post.keywords)}\n${post.heroImage ? `heroImage: ${post.heroImage}\nheroAlt: ${quote(post.heroAlt)}\n` : ''}---\n\n${String(post.body || '').trim()}\n`;
const listPosts = async () => Promise.all((await readdir(postsDirectory)).filter((name) => name.endsWith('.md')).map(async (name) => { const slug = name.replace(/\.md$/, ''); const post = parsePost(await readFile(join(postsDirectory, name), 'utf8'), slug); return { slug, title: post.title || slug, draft: post.draft, pubDate: post.pubDate }; }));
const routeParts = (pathname) => pathname.split('/').filter(Boolean).map(decodeURIComponent);

const server = createServer(async (req, res) => {
  try {
    const url = new URL(req.url, `http://${host}:${port}`);
    if (req.method === 'GET' && url.pathname === '/api/posts') return send(res, 200, await listPosts());
    if (url.pathname.startsWith('/api/posts/')) {
      const [, , slug] = routeParts(url.pathname);
      if (!safeSlug(slug)) return send(res, 400, { error: 'Invalid post slug.' });
      const postPath = join(postsDirectory, `${slug}.md`);
      if (req.method === 'GET') return send(res, 200, parsePost(await readFile(postPath, 'utf8'), slug));
      if (req.method === 'PUT') { const post = await readJson(req); if (!post.title?.trim()) return send(res, 400, { error: 'A title is required.' }); await writeFile(postPath, serializePost({ ...post, slug }), 'utf8'); return send(res, 200, { ok: true }); }
    }
    if (req.method === 'POST' && url.pathname === '/api/images') {
      const image = await readJson(req);
      if (!safeSlug(image.slug) || !safeFilename(image.filename) || !String(image.data).startsWith('data:image/')) return send(res, 400, { error: 'Invalid image upload.' });
      const extension = extname(image.filename).toLowerCase();
      if (!['.png', '.jpg', '.jpeg', '.gif', '.webp', '.avif'].includes(extension)) return send(res, 400, { error: 'Use PNG, JPG, GIF, WebP, or AVIF images.' });
      await mkdir(join(assetsDirectory, image.slug), { recursive: true });
      await writeFile(join(assetsDirectory, image.slug, image.filename), Buffer.from(image.data.split(',')[1], 'base64'));
      return send(res, 200, { path: `../../assets/posts/${image.slug}/${image.filename}` });
    }
    const fileName = url.pathname === '/' ? 'index.html' : normalize(url.pathname).replace(/^\//, '');
    if (fileName.includes('..')) return send(res, 404, 'Not found', 'text/plain');
    const filePath = join(editorDirectory, fileName); const fileStat = await stat(filePath); if (!fileStat.isFile()) throw new Error('Not found');
    const type = filePath.endsWith('.html') ? 'text/html' : filePath.endsWith('.css') ? 'text/css' : 'text/javascript';
    return send(res, 200, await readFile(filePath, 'utf8'), type);
  } catch (error) { send(res, error.code === 'ENOENT' ? 404 : 500, { error: error.code === 'ENOENT' ? 'Not found.' : error.message }); }
});
server.listen(port, host, () => console.log(`Local editor running at http://${host}:${port}`));
