import fs from 'fs-extra';
import path from 'path';
import matter from 'gray-matter';
import { marked } from 'marked';
import { minify } from 'html-minifier-terser';

import RSS from 'rss';
import Handlebars from 'handlebars';
import { fileURLToPath } from 'url';
import dotenv from 'dotenv';

dotenv.config();

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const CONTENT_DIR = path.join(__dirname, '../content/posts');
const TEMPLATE_DIR = path.join(__dirname, '../templates');
const PARTIALS_DIR = path.join(__dirname, '../templates/partials');
const DIST_DIR = path.join(__dirname, '../dist');
const STATIC_FILES = ['../styles.css', '../images'];

const siteMeta = {
  title: process.env.SITE_TITLE || "Your Blog",
  description: process.env.SITE_DESCRIPTION || 'Personal blog about technology, thoughts, and experiences',
  url: process.env.SITE_URL || 'https://yourdomain.com',
  author: process.env.SITE_AUTHOR || 'Your Name',
  logoText: process.env.SITE_LOGO_TEXT || 'Y'
};

function formatDate(date) {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric'
  });
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

    // Copy CSS
    console.log('📝 Processing CSS...');
    const cssPath = path.join(__dirname, STATIC_FILES[0]);
    if (await fs.pathExists(cssPath)) {
      const css = await fs.readFile(cssPath, 'utf8');
      // Simple CSS minification - remove comments and extra whitespace
      const minifiedCss = css
        .replace(/\/\*[\s\S]*?\*\//g, '') // Remove comments
        .replace(/\s+/g, ' ') // Collapse whitespace
        .replace(/;\s*}/g, '}') // Remove trailing semicolons
        .replace(/:\s+/g, ':') // Remove space after colons
        .replace(/\s*{\s*/g, '{') // Remove space around braces
        .replace(/\s*}\s*/g, '}') // Remove space around braces
        .trim();
      await fs.writeFile(path.join(DIST_DIR, 'styles.css'), minifiedCss);
    }

    // Copy images (if any)
    console.log('🖼️  Copying images...');
    const imagesPath = path.join(__dirname, STATIC_FILES[1]);
    if (await fs.pathExists(imagesPath)) {
      await fs.copy(imagesPath, path.join(DIST_DIR, 'images'));
    }

    // Copy site.webmanifest
    console.log('📄 Copying site manifest...');
    const manifestPath = path.join(__dirname, '../site.webmanifest');
    if (await fs.pathExists(manifestPath)) {
      await fs.copy(manifestPath, path.join(DIST_DIR, 'site.webmanifest'));
    }

    // Copy favicon files
    console.log('🎨 Copying favicon files...');
    const faviconFiles = [
      'favicon.ico',
      'favicon-16x16.png',
      'favicon-32x32.png'
    ];
    
    for (const file of faviconFiles) {
      const sourcePath = path.join(__dirname, '../images', file);
      if (await fs.pathExists(sourcePath)) {
        await fs.copy(sourcePath, path.join(DIST_DIR, 'images', file));
      }
    }

    // Copy robots.txt
    console.log('🤖 Copying robots.txt...');
    const robotsPath = path.join(__dirname, '../robots.txt');
    if (await fs.pathExists(robotsPath)) {
      await fs.copy(robotsPath, path.join(DIST_DIR, 'robots.txt'));
    }

    // Read templates and partials
    console.log('📋 Reading templates...');
    const postTemplate = await fs.readFile(path.join(TEMPLATE_DIR, 'post.hbs'), 'utf8');
    const indexTemplate = await fs.readFile(path.join(TEMPLATE_DIR, 'index.hbs'), 'utf8');
    const headerPartial = await fs.readFile(path.join(PARTIALS_DIR, 'header.hbs'), 'utf8');
    const footerPartial = await fs.readFile(path.join(PARTIALS_DIR, 'footer.hbs'), 'utf8');

    // Register Handlebars partials
    Handlebars.registerPartial('header', headerPartial);
    Handlebars.registerPartial('footer', footerPartial);

    // Register Handlebars helpers
    Handlebars.registerHelper('concat', function() {
      return Array.from(arguments).slice(0, -1).join('');
    });
    Handlebars.registerHelper('eq', function(a, b) {
      return a === b;
    });

    // Compile templates
    const postTemplateCompiled = Handlebars.compile(postTemplate);
    const indexTemplateCompiled = Handlebars.compile(indexTemplate);

    // Read and process posts
    console.log('📄 Processing posts...');
    const files = (await fs.readdir(CONTENT_DIR)).filter(f => f.endsWith('.md'));
    let posts = [];
    let allPosts = [];
    
    for (const file of files) {
      const filePath = path.join(CONTENT_DIR, file);
      const { data, content } = matter(await fs.readFile(filePath, 'utf8'));
      const slug = file.replace(/\.md$/, '');
      const htmlContent = marked.parse(content);
      
      const post = {
        ...data,
        slug,
        content: htmlContent,
        date: data.date || new Date().toISOString().split('T')[0],
        formattedDate: formatDate(data.date || new Date()),
        keywords: data.keywords || '',
        categories: data.categories || []
      };

      // Generate post URL and image URL
      const postUrl = `${siteMeta.url}/posts/${slug}.html`;
      const imageUrl = data.img ? `${siteMeta.url}/${data.img.replace('../', '')}` : '';

      // Render post HTML using Handlebars
      const postData = {
        ...post,
        url: postUrl,
        image: imageUrl,
        baseUrl: siteMeta.url,
        author: siteMeta.author,
        logoText: siteMeta.logoText,
        cssPath: '../styles.css',
        faviconPath: '../favicon.ico',
        homePath: '../index.html',
        aboutPath: '../about.html'
      };

      let html = postTemplateCompiled(postData);

      html = await minify(html, { 
        collapseWhitespace: true, 
        removeComments: true,
        removeAttributeQuotes: true,
        removeEmptyAttributes: true,
        removeOptionalTags: true,
        removeRedundantAttributes: true,
        removeScriptTypeAttributes: true,
        removeStyleLinkTypeAttributes: true,
        useShortDoctype: true
      });
      
      await fs.writeFile(path.join(DIST_DIR, 'posts', `${slug}.html`), html);
      allPosts.push(post);
      
      // Only add published posts to the public posts array
      if (data.published !== false) {
        posts.push(post);
      }
    }

    // Sort posts by date (newest first)
    posts.sort((a, b) => new Date(b.date) - new Date(a.date));

    // Generate index.html
    console.log('🏠 Generating index page...');
    
    const indexData = {
      posts: posts,
      title: siteMeta.title,
      description: siteMeta.description,
      baseUrl: siteMeta.url,
      author: siteMeta.author,
      logoText: siteMeta.logoText,
      cssPath: 'styles.css',
      faviconPath: 'favicon.ico',
      homePath: 'index.html',
      aboutPath: 'about.html'
    };

    let indexHtml = indexTemplateCompiled(indexData);

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

    // Generate about page
    console.log('📄 Generating about page...');
    const aboutTemplate = await fs.readFile(path.join(TEMPLATE_DIR, 'about.hbs'), 'utf8');
    const aboutTemplateCompiled = Handlebars.compile(aboutTemplate);
    
    const aboutData = {
      title: siteMeta.title,
      description: siteMeta.description,
      baseUrl: siteMeta.url,
      author: siteMeta.author,
      logoText: siteMeta.logoText,
      cssPath: 'styles.css',
      faviconPath: 'favicon.ico',
      homePath: 'index.html',
      aboutPath: 'about.html'
    };
    
    let aboutHtml = aboutTemplateCompiled(aboutData);
    aboutHtml = await minify(aboutHtml, { 
      collapseWhitespace: true, 
      minifyCSS: true, 
      removeComments: true,
      minifyJS: true
    });
    
    await fs.writeFile(path.join(DIST_DIR, 'about.html'), aboutHtml);

    // Read and compile writing.hbs
    const writingTemplate = await fs.readFile(path.join(TEMPLATE_DIR, 'writing.hbs'), 'utf8');
    const writingTemplateCompiled = Handlebars.compile(writingTemplate);

    // Generate writing.html
    console.log('📝 Generating writing page...');
    const writingData = {
      posts: posts,
      title: siteMeta.title,
      description: siteMeta.description,
      baseUrl: siteMeta.url,
      author: siteMeta.author,
      logoText: siteMeta.logoText,
      cssPath: 'styles.css',
      faviconPath: 'favicon.ico',
      homePath: 'index.html',
      aboutPath: 'about.html'
    };
    let writingHtml = writingTemplateCompiled(writingData);
    writingHtml = await minify(writingHtml, { 
      collapseWhitespace: true, 
      minifyCSS: true, 
      removeComments: true,
      minifyJS: true
    });
    await fs.writeFile(path.join(DIST_DIR, 'writing.html'), writingHtml);

    console.log('✅ Build completed successfully!');
    console.log(`📊 Generated ${allPosts.length} posts (${posts.length} published)`);
    console.log(`📁 Output directory: ${DIST_DIR}`);
    
  } catch (error) {
    console.error('❌ Build failed:', error);
    process.exit(1);
  }
}

build();