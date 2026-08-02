import type { APIRoute } from 'astro';
import { githubContributionCalendarSvg } from '../generated/github-contributions';

export const GET: APIRoute = async () => new Response(githubContributionCalendarSvg, {
  headers: { 'Content-Type': 'image/svg+xml; charset=utf-8' },
});
