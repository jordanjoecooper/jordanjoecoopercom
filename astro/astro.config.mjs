import { defineConfig } from 'astro/config';
import sitemap from '@astrojs/sitemap';

export default defineConfig({
  site: 'https://jordanjoecooper.com',
  output: 'static',
  // Reuse the existing image directory while this migration lives alongside the legacy site.
  publicDir: '../images',
  integrations: [sitemap()],
});
