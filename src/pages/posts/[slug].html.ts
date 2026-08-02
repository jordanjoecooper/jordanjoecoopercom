import { getCollection } from 'astro:content';
import type { APIRoute } from 'astro';

export async function getStaticPaths() {
  return (await getCollection('posts', ({ data }) => !data.draft)).map((post) => ({
    params: { slug: post.id.replace(/\.md$/, '') },
  }));
}

export const GET: APIRoute = ({ params, site }) => {
  const destination = `/posts/${params.slug}/`;
  const canonical = new URL(destination, site).href;
  const escapedDestination = destination.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/"/g, '&quot;');

  return new Response(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="robots" content="noindex">
    <link rel="canonical" href="${canonical}">
    <meta http-equiv="refresh" content="0;url=${escapedDestination}">
    <title>Post moved</title>
  </head>
  <body><p>This post has moved to <a href="${escapedDestination}">${escapedDestination}</a>.</p></body>
</html>`, { headers: { 'Content-Type': 'text/html; charset=utf-8' } });
};
