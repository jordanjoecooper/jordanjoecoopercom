import { defineConfig } from 'astro/config';
import sitemap from '@astrojs/sitemap';

export default defineConfig({
  site: 'https://jordanjoecooper.com',
  output: 'static',
  // Legacy assets stay available while the remaining static pages are migrated.
  publicDir: 'archive/images',
  integrations: [sitemap()],
});
