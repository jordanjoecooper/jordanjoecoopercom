import fs from 'fs-extra';
import path from 'path';
import matter from 'gray-matter';
import { marked } from 'marked';
import { minify } from 'html-minifier-terser';
import CleanCSS from 'clean-css';
import RSS from 'rss';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const CONTENT_DIR = path.join(__dirname, '../content/posts');
const TEMPLATE_DIR = path.join(__dirname, '../templates');
const DIST_DIR = path.join(__dirname, '../dist');
const STATIC_FILES = ['../styles.css', '../images'];

const siteMeta = {
  title: "Jordan's Blog",
  description: 'Personal blog about technology, thoughts, and experiences',
  url: 'https://yourdomain.com', // Change to your domain!
  author: 'Jordan'
};

function formatDate(date) {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
}

function calculateReadTime(content) {
  const wordsPerMinute = 200;
  const wordCount = content.split(/\s+/).length;
  return Math.ceil(wordCount / wordsPerMinute);
}

function generateSitemap(posts) {
  const baseUrl = siteMeta.url;
  let sitemap = '<?xml version="1.0" encoding="UTF-8"?>\n';
  sitemap += '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n';
  
  // Add homepage
  sitemap += `  <url>
    <loc>${baseUrl}</loc>
    <lastmod>${new Date().toISOString().split('T')[0]}</lastmod>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>\n`;
  
  // Add about page
  sitemap += `  <url>
    <loc>${baseUrl}/about.html</loc>
    <lastmod>${new Date().toISOString().split('T')[0]}</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.8</priority>
  </url>\n`;
  
  // Add blog posts
  posts.forEach(post => {
    sitemap += `  <url>
    <loc>${baseUrl}/posts/${post.slug}.html</loc>
    <lastmod>${new Date(post.date).toISOString().split('T')[0]}</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.7</priority>
  </url>\n`;
  });
  
  sitemap += '</urlset>';
  return sitemap;
}

function generateRSSFeed(posts) {
  const feed = new RSS({
    title: siteMeta.title,
    description: siteMeta.description,
    feed_url: `${siteMeta.url}/rss.xml`,
    site_url: siteMeta.url,
    image_url: `${siteMeta.url}/images/logo.png`,
    managingEditor: siteMeta.author,
    webMaster: siteMeta.author,
    copyright: `2025 ${siteMeta.author}`,
    language: 'en',
    pubDate: new Date().toUTCString(),
    ttl: '60'
  });

  posts.forEach(post => {
    feed.item({
      title: post.title,
      description: post.description,
      url: `${siteMeta.url}/posts/${post.slug}.html`,
      guid: post.slug,
      categories: post.categories || [],
      author: siteMeta.author,
      date: post.date
    });
  });

  return feed.xml({ indent: true });
}

async function build() {
  try {
    console.log('🚀 Starting build process...');
    
    // Clean and create dist directory
    await fs.remove(DIST_DIR);
    await fs.ensureDir(DIST_DIR);
    await fs.ensureDir(path.join(DIST_DIR, 'posts'));

    // Copy and minify CSS
    console.log('📝 Processing CSS...');
    const cssPath = path.join(__dirname, STATIC_FILES[0]);
    if (await fs.pathExists(cssPath)) {
      const css = await fs.readFile(cssPath, 'utf8');
      const minifiedCss = new CleanCSS().minify(css).styles;
      await fs.writeFile(path.join(DIST_DIR, 'styles.css'), minifiedCss);
    }

    // Copy images (if any)
    console.log('🖼️  Copying images...');
    const imagesPath = path.join(__dirname, STATIC_FILES[1]);
    if (await fs.pathExists(imagesPath)) {
      await fs.copy(imagesPath, path.join(DIST_DIR, 'images'));
    }

    // Read templates
    console.log('📋 Reading templates...');
    const postTemplate = await fs.readFile(path.join(TEMPLATE_DIR, 'post.html'), 'utf8');
    const indexTemplate = await fs.readFile(path.join(TEMPLATE_DIR, 'index.html'), 'utf8');

    // Read and process posts
    console.log('📄 Processing posts...');
    const files = (await fs.readdir(CONTENT_DIR)).filter(f => f.endsWith('.md'));
    let posts = [];
    
    for (const file of files) {
      const filePath = path.join(CONTENT_DIR, file);
      const { data, content } = matter(await fs.readFile(filePath, 'utf8'));
      const slug = file.replace(/\.md$/, '');
      const htmlContent = marked.parse(content);
      const readTime = calculateReadTime(content);
      
      const post = {
        ...data,
        slug,
        content: htmlContent,
        date: data.date || new Date().toISOString().split('T')[0],
        readTime,
        formattedDate: formatDate(data.date || new Date()),
        keywords: data.keywords || '',
        categories: data.categories || []
      };

      // Generate post URL and image URL
      const postUrl = `${siteMeta.url}/posts/${slug}.html`;
      const imageUrl = data.img ? `${siteMeta.url}/${data.img.replace('../', '')}` : '';

      // Render and minify post HTML
      let html = postTemplate
        .replace(/{{title}}/g, post.title)
        .replace(/{{description}}/g, post.description)
        .replace(/{{date}}/g, post.date)
        .replace(/{{formattedDate}}/g, post.formattedDate)
        .replace(/{{img}}/g, data.img || '')
        .replace(/{{content}}/g, post.content)
        .replace(/{{url}}/g, postUrl)
        .replace(/{{image}}/g, imageUrl)
        .replace(/{{keywords}}/g, post.keywords)
        .replace(/{{baseUrl}}/g, siteMeta.url)
        .replace(/{{readTime}}/g, readTime);

      html = await minify(html, { 
        collapseWhitespace: true, 
        minifyCSS: true, 
        removeComments: true,
        minifyJS: true
      });
      
      await fs.writeFile(path.join(DIST_DIR, 'posts', `${slug}.html`), html);
      posts.push(post);
    }

    // Sort posts by date (newest first)
    posts.sort((a, b) => new Date(b.date) - new Date(a.date));

    // Generate index.html
    console.log('🏠 Generating index page...');
    const postsHtml = posts.map(post => `
      <article class="blog-card">
        ${post.img ? `<div class="blog-thumbnail">
          <img src="${post.img}" alt="${post.title}" loading="lazy">
        </div>` : ''}
        <div class="blog-content">
          <h3 class="blog-title">
            <a href="posts/${post.slug}.html">${post.title}</a>
          </h3>
          <div class="blog-meta">
            <time datetime="${post.date}" class="blog-date">${post.formattedDate}</time>
            ${post.readTime ? `<span class="blog-read-time">${post.readTime} min read</span>` : ''}
          </div>
          <div class="blog-excerpt">
            <p>${post.description}</p>
          </div>
          <a href="posts/${post.slug}.html" class="button">Read More</a>
        </div>
      </article>
    `).join('\n');

    let indexHtml = indexTemplate
      .replace(/{{baseUrl}}/g, siteMeta.url)
      .replace(/{{#each posts}}([\s\S]*?){{\/each}}/g, (match, template) => {
        return posts.map(post => {
          return template
            .replace(/{{slug}}/g, post.slug)
            .replace(/{{title}}/g, post.title)
            .replace(/{{description}}/g, post.description)
            .replace(/{{date}}/g, post.date)
            .replace(/{{formattedDate}}/g, post.formattedDate)
            .replace(/{{img}}/g, post.img || '')
            .replace(/{{readTime}}/g, post.readTime || '');
        }).join('\n');
      });

    indexHtml = await minify(indexHtml, { 
      collapseWhitespace: true, 
      minifyCSS: true, 
      removeComments: true,
      minifyJS: true
    });
    
    await fs.writeFile(path.join(DIST_DIR, 'index.html'), indexHtml);

    // Generate sitemap
    console.log('🗺️  Generating sitemap...');
    const sitemap = generateSitemap(posts);
    await fs.writeFile(path.join(DIST_DIR, 'sitemap.xml'), sitemap);

    // Generate RSS feed
    console.log('📡 Generating RSS feed...');
    const rssFeed = generateRSSFeed(posts);
    await fs.writeFile(path.join(DIST_DIR, 'rss.xml'), rssFeed);

    // Copy about page if it exists
    const aboutPath = path.join(__dirname, '../about.html');
    if (await fs.pathExists(aboutPath)) {
      await fs.copy(aboutPath, path.join(DIST_DIR, 'about.html'));
    }

    console.log('✅ Build completed successfully!');
    console.log(`📊 Generated ${posts.length} posts`);
    console.log(`📁 Output directory: ${DIST_DIR}`);
    
  } catch (error) {
    console.error('❌ Build failed:', error);
    process.exit(1);
  }
}

// Run the build
build();