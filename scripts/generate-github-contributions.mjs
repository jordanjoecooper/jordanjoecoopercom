#!/usr/bin/env node

import { mkdir, writeFile } from 'node:fs/promises';
import { dirname, resolve } from 'node:path';

const DEFAULT_USERNAME = 'jordanjoecooper';
const DEFAULT_OUTPUT = 'src/generated/github-contributions.ts';
const DAY_LABELS = new Map([[1, 'Mon'], [3, 'Wed'], [5, 'Fri']]);
const THEMES = {
  light: {
    background: '#fffdf9', border: '#ddd5ca', heading: '#2a2a2a', label: '#7a7268',
    colors: { NONE: '#e8e2d9', FIRST_QUARTILE: '#9be9a8', SECOND_QUARTILE: '#40c463', THIRD_QUARTILE: '#30a14e', FOURTH_QUARTILE: '#216e39' },
  },
  dark: {
    background: '#0d1117', border: '#30363d', heading: '#f0f6fc', label: '#8b949e',
    colors: { NONE: '#161b22', FIRST_QUARTILE: '#0e4429', SECOND_QUARTILE: '#006d32', THIRD_QUARTILE: '#26a641', FOURTH_QUARTILE: '#39d353' },
  },
};

const args = process.argv.slice(2);
const option = (name, fallback) => {
  const index = args.indexOf(name);
  return index === -1 ? fallback : args[index + 1] ?? fallback;
};

const username = option('--username', process.env.GITHUB_USERNAME || DEFAULT_USERNAME);
const output = resolve(option('--output', process.env.GITHUB_CONTRIBUTIONS_OUTPUT || DEFAULT_OUTPUT));
const themeName = option('--theme', process.env.GITHUB_CONTRIBUTIONS_THEME || 'light');
const token = process.env.GITHUB_CONTRIBUTIONS_TOKEN || process.env.GITHUB_TOKEN;

if (!THEMES[themeName]) {
  throw new Error(`Unknown theme \"${themeName}\". Use \"light\" or \"dark\".`);
}
const theme = THEMES[themeName];

const writeCalendarModule = async (hasCalendar, svg = '') => {
  const source = `// This file is regenerated before every production build. Do not edit manually.\nexport const hasGithubContributionCalendar = ${hasCalendar};\nexport const githubContributionCalendarSvg = ${JSON.stringify(svg)};\n`;
  await mkdir(dirname(output), { recursive: true });
  await writeFile(output, source);
};

if (!token) {
  await writeCalendarModule(false);
  console.warn('Skipping GitHub contribution export: no contribution token is configured.');
  process.exit(0);
}

const end = new Date();
const start = new Date(end);
start.setUTCFullYear(start.getUTCFullYear() - 1);
const query = `
  query Contributions($login: String!, $from: DateTime!, $to: DateTime!) {
    user(login: $login) {
      contributionsCollection(from: $from, to: $to) {
        contributionCalendar {
          totalContributions
          weeks {
            contributionDays {
              date
              contributionLevel
            }
          }
        }
      }
    }
  }
`;

const response = await fetch('https://api.github.com/graphql', {
  method: 'POST',
  headers: {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    query,
    variables: { login: username, from: start.toISOString(), to: end.toISOString() },
  }),
});

const result = await response.json();
if (!response.ok || result.errors || !result.data?.user) {
  throw new Error(result.errors?.map((error) => error.message).join('; ') || `GitHub request failed (${response.status}).`);
}

const calendar = result.data.user.contributionsCollection.contributionCalendar;
const cell = 11;
const gap = 3;
const step = cell + gap;
const chartX = 48;
const chartY = 37;
const width = chartX + calendar.weeks.length * step + 10;
const height = 151;
const escapeXml = (value) => String(value).replace(/[&<>"']/g, (character) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&apos;' })[character]);

let previousMonth = -1;
const months = [];
const cells = [];
calendar.weeks.forEach((week, weekIndex) => {
  week.contributionDays.forEach((day) => {
    const date = new Date(`${day.date}T00:00:00Z`);
    const dayOfWeek = date.getUTCDay();
    const month = date.getUTCMonth();
    if (dayOfWeek === 0 && month !== previousMonth) {
      months.push(`<text x="${chartX + weekIndex * step}" y="22" class="month">${date.toLocaleString('en-GB', { month: 'short', timeZone: 'UTC' })}</text>`);
      previousMonth = month;
    }
    const x = chartX + weekIndex * step;
    const y = chartY + dayOfWeek * step;
    const fill = theme.colors[day.contributionLevel] || theme.colors.NONE;
    const label = `${day.date}: ${day.contributionLevel.replaceAll('_', ' ').toLowerCase()}`;
    cells.push(`<rect x="${x}" y="${y}" width="${cell}" height="${cell}" rx="2" fill="${fill}"><title>${escapeXml(label)}</title></rect>`);
  });
});

const labels = [...DAY_LABELS].map(([day, label]) => `<text x="0" y="${chartY + day * step + 9}" class="day">${label}</text>`).join('');
const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${width} ${height}" role="img" aria-labelledby="title description">
  <title id="title">${escapeXml(`${calendar.totalContributions.toLocaleString('en-GB')} GitHub contributions in the last year`)}</title>
  <desc id="description">A GitHub contribution calendar for ${escapeXml(username)}.</desc>
  <style>.background{fill:${theme.background};stroke:${theme.border};stroke-width:1}.heading{fill:${theme.heading};font:600 16px -apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}.month,.day{fill:${theme.label};font:12px -apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}.day{font-weight:600}</style>
  <rect class="background" x=".5" y=".5" width="${width - 1}" height="${height - 1}" rx="8"/>
  <text x="0" y="16" class="heading">${escapeXml(`${calendar.totalContributions.toLocaleString('en-GB')} contributions in the last year`)}</text>
  ${months.join('')}
  ${labels}
  ${cells.join('')}
</svg>`;

await writeCalendarModule(true, svg);
console.log(`Wrote ${output} for @${username}.`);
