export function GET() {
  return new Response(JSON.stringify({ name: 'Jordan Joe Cooper', short_name: 'Jordan Joe Cooper', icons: [{ src: '/android-chrome-192x192.png', sizes: '192x192', type: 'image/png' }, { src: '/android-chrome-512x512.png', sizes: '512x512', type: 'image/png' }], theme_color: '#f5f0eb', background_color: '#f5f0eb', display: 'standalone' }), { headers: { 'Content-Type': 'application/manifest+json' } });
}
