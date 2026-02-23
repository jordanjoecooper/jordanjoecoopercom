#!/usr/bin/env node
/**
 * Update homepage (index.html) and RSS feed (feed.xml) with a new post.
 * Run after adding a new HTML file to posts/ and filling in its meta tags.
 *
 * Usage: node scripts/update-post-links.js posts/your-post.html
 *
 * Reads title, description, and date from the post's meta tags and inserts
 * the new entry at the top of the Writing list and the feed.
 */

const fs = require('fs');
const path = require('path');

const MONTHS = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];

function extractMeta(html, attr, value) {
  const re = new RegExp(`<meta\\s+[^>]*${attr}="${value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}"[^>]*content="([^"]*)"`, 'i');
  const m = html.match(re);
  return m ? m[1].trim() : null;
}

function extractTitle(html) {
  const fromOg = extractMeta(html, 'property', 'og:title') || extractMeta(html, 'name', 'twitter:title');
  if (fromOg) {
    return fromOg.replace(/\s*-\s*Jordan Joe Cooper\s*$/i, '').trim();
  }
  const fromTitle = html.match(/<title>([^<]+)<\/title>/i);
  return fromTitle ? fromTitle[1].replace(/\s*-\s*Jordan Joe Cooper\s*$/i, '').trim() : null;
}

function formatDisplayDate(isoDate) {
  const [y, m, d] = isoDate.split('-').map(Number);
  const month = MONTHS[m - 1];
  return `${month} ${d}, ${y}`;
}

function formatRssDate(isoDate) {
  const date = new Date(isoDate + 'T00:00:00Z');
  return date.toUTCString().replace(' GMT', ' +0000');
}

function main() {
  const postPath = process.argv[2];
  if (!postPath || !postPath.startsWith('posts/') || !postPath.endsWith('.html')) {
    console.error('Usage: node scripts/update-post-links.js posts/your-post.html');
    process.exit(1);
  }

  const root = path.resolve(__dirname, '..');
  const postFullPath = path.join(root, postPath);

  if (!fs.existsSync(postFullPath)) {
    console.error('Post file not found:', postFullPath);
    process.exit(1);
  }

  const html = fs.readFileSync(postFullPath, 'utf8');
  const title = extractTitle(html);
  const description = extractMeta(html, 'name', 'description') || extractMeta(html, 'property', 'og:description');
  const dateRaw = extractMeta(html, 'name', 'article:published_time') || extractMeta(html, 'property', 'article:published_time');

  if (!title) {
    console.error('Could not extract title from post. Check og:title or <title>.');
    process.exit(1);
  }
  if (!description) {
    console.error('Could not extract description. Check meta name="description".');
    process.exit(1);
  }
  if (!dateRaw || !/^\d{4}-\d{2}-\d{2}$/.test(dateRaw)) {
    console.error('Could not extract date. Check meta name="article:published_time" (YYYY-MM-DD).');
    process.exit(1);
  }

  const slug = path.basename(postPath, '.html');
  const displayDate = formatDisplayDate(dateRaw);
  const rssDate = formatRssDate(dateRaw);
  const url = `https://jordanjoecooper.com/posts/${slug}.html`;

  // Escape for XML (description in RSS)
  const escapeXml = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  const descriptionEscaped = escapeXml(description);

  // Update index.html: insert new <li> as first item in .post-list
  const indexPath = path.join(root, 'index.html');
  let indexContent = fs.readFileSync(indexPath, 'utf8');
  const newListItem = `      <li>
        <h2 class="post-title">
          <span class="post-date">${displayDate}</span>
          <a href="posts/${slug}.html">${title.replace(/&/g, '&amp;').replace(/</g, '&lt;')}</a>
        </h2>
      </li>
      `;
  const listMarker = '<ul class="post-list">';
  const idx = indexContent.indexOf(listMarker);
  if (idx === -1) {
    console.error('Could not find Writing list in index.html.');
    process.exit(1);
  }
  indexContent = indexContent.slice(0, idx + listMarker.length) + '\n' + newListItem + indexContent.slice(idx + listMarker.length);
  fs.writeFileSync(indexPath, indexContent);
  console.log('Updated index.html (Writing section).');

  // Update feed.xml: insert new <item> before first existing <item>
  const feedPath = path.join(root, 'feed.xml');
  let feedContent = fs.readFileSync(feedPath, 'utf8');
  const newItem = `    <item>
      <title>${escapeXml(title)}</title>
      <link>${url}</link>
      <guid>${url}</guid>
      <pubDate>${rssDate}</pubDate>
      <description>${descriptionEscaped}</description>
    </item>
    `;
  const itemMarker = '<item>';
  const feedIdx = feedContent.indexOf(itemMarker);
  if (feedIdx === -1) {
    console.error('Could not find <item> in feed.xml.');
    process.exit(1);
  }
  feedContent = feedContent.slice(0, feedIdx) + newItem + feedContent.slice(feedIdx);
  fs.writeFileSync(feedPath, feedContent);
  console.log('Updated feed.xml.');

  console.log('Done. New post added to homepage and RSS:', title);
}

main();
