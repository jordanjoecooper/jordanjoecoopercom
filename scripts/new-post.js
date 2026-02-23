#!/usr/bin/env node
/**
 * Interactive CLI to create a new post from the template.
 * Prompts for title, description, keywords, date, and slug; copies post-template.html
 * to posts/{slug}.html, fills in all placeholders, then runs update-post-links.js so
 * the new post is added at the top of the homepage Writing section and the RSS feed.
 *
 * Usage: node scripts/new-post.js
 * Optional: node scripts/new-post.js "Post title" "Short description" "keywords" "2025-02-23" "post-slug"
 *   (any number of args; missing ones are prompted)
 */

const fs = require('fs');
const path = require('path');
const readline = require('readline');
const { spawnSync } = require('child_process');

const MONTHS = ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'];

const root = path.resolve(__dirname, '..');
const templatePath = path.join(root, 'post-template.html');
const postsDir = path.join(root, 'posts');

function slugify(s) {
  return s
    .toLowerCase()
    .trim()
    .replace(/\s+/g, '-')
    .replace(/[^a-z0-9-]/g, '');
}

function formatDisplayDate(isoDate) {
  const [y, m, d] = isoDate.split('-').map(Number);
  return `${MONTHS[m - 1]} ${d}, ${y}`;
}

function ask(rl, question, defaultVal = '') {
  const suffix = defaultVal ? ` (${defaultVal})` : '';
  return new Promise((resolve) => {
    rl.question(`${question}${suffix}: `, (answer) => {
      resolve((answer && answer.trim()) || defaultVal || '');
    });
  });
}

function validateDate(s) {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(s)) return false;
  const d = new Date(s + 'T12:00:00Z');
  return d.getUTCFullYear() === parseInt(s.slice(0, 4), 10) &&
    (d.getUTCMonth() + 1) === parseInt(s.slice(5, 7), 10) &&
    d.getUTCDate() === parseInt(s.slice(8, 10), 10);
}

async function promptFields() {
  const argv = process.argv.slice(2);
  const hasArgs = argv.length > 0;

  const rl = readline.createInterface({ input: process.stdin, output: process.stdout });

  const title = hasArgs && argv[0] ? argv[0] : await ask(rl, 'Title');
  if (!title) {
    console.error('Title is required.');
    rl.close();
    process.exit(1);
  }

  const defaultSlug = slugify(title);
  const description = hasArgs && argv[1] ? argv[1] : await ask(rl, 'Description (meta, one line)');
  const keywords = hasArgs && argv[2] ? argv[2] : await ask(rl, 'Keywords (comma-separated)');
  const dateArg = hasArgs && argv[3] ? argv[3] : await ask(rl, 'Date (YYYY-MM-DD)', new Date().toISOString().slice(0, 10));
  const slug = hasArgs && argv[4] ? argv[4] : await ask(rl, 'Slug (filename)', defaultSlug);

  rl.close();

  const date = dateArg.trim();
  if (!validateDate(date)) {
    console.error('Invalid date. Use YYYY-MM-DD.');
    process.exit(1);
  }

  const finalSlug = (slug || defaultSlug).replace(/\.html$/, '').replace(/[^a-z0-9-]/gi, '-').replace(/-+/g, '-').replace(/^-|-$/g, '') || defaultSlug;
  if (!finalSlug) {
    console.error('Slug is required.');
    process.exit(1);
  }

  return {
    title: title.trim(),
    description: (description || '').trim(),
    keywords: (keywords || '').trim(),
    date,
    slug: finalSlug,
    displayDate: formatDisplayDate(date),
  };
}

function escapeForHtml(s) {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

function main() {
  if (!fs.existsSync(templatePath)) {
    console.error('Post template not found:', templatePath);
    process.exit(1);
  }
  if (!fs.existsSync(postsDir)) {
    fs.mkdirSync(postsDir, { recursive: true });
  }

  promptFields().then((fields) => {
    const outPath = path.join(postsDir, `${fields.slug}.html`);
    if (fs.existsSync(outPath)) {
      console.error('File already exists:', outPath);
      process.exit(1);
    }

    let content = fs.readFileSync(templatePath, 'utf8');

    content = content
      .replace(/POST_TITLE/g, escapeForHtml(fields.title))
      .replace(/POST_DESCRIPTION/g, escapeForHtml(fields.description))
      .replace(/POST_KEYWORDS/g, escapeForHtml(fields.keywords))
      .replace(/POST_SLUG/g, fields.slug)
      .replace(/YYYY-MM-DD/g, fields.date)
      .replace(/Month Day, Year/g, fields.displayDate);

    // Ensure RSS link is present (template may not have it)
    if (!content.includes('application/rss+xml')) {
      content = content.replace(
        '  <link rel="manifest" href="../site.webmanifest">\n  \n  <link rel="stylesheet"',
        '  <link rel="manifest" href="../site.webmanifest">\n  <link rel="alternate" type="application/rss+xml" title="Jordan Joe Cooper" href="../feed.xml">\n\n  <link rel="stylesheet"'
      );
    }

    fs.writeFileSync(outPath, content, 'utf8');
    console.log('Created:', outPath);

    const postRelative = `posts/${fields.slug}.html`;
    const updateResult = spawnSync(process.execPath, [path.join(__dirname, 'update-post-links.js'), postRelative], {
      cwd: root,
      stdio: 'inherit',
    });
    if (updateResult.status !== 0) {
      console.error('Post created but failed to update homepage/feed. Run manually: node scripts/update-post-links.js', postRelative);
      process.exit(1);
    }

    console.log('');
    console.log('Next: edit the post body in', outPath);
  }).catch((err) => {
    console.error(err);
    process.exit(1);
  });
}

main();
