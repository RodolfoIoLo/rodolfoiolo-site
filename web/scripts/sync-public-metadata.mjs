import { mkdir, writeFile } from 'node:fs/promises';

const apiBaseURL = process.env.CONTENT_API_BASE_URL ?? 'http://localhost:8080/api/v1';
const siteURL = process.env.NEXT_PUBLIC_SITE_URL ?? 'http://localhost:3000';

await mkdir(new URL('../public/', import.meta.url), { recursive: true });

for (const filename of ['rss.xml', 'sitemap.xml']) {
  const response = await fetch(`${apiBaseURL}/${filename}`);
  if (!response.ok) throw new Error(`${filename} fetch failed with ${response.status}`);
  await writeFile(new URL(`../public/${filename}`, import.meta.url), await response.text(), 'utf8');
}

const robots = `User-agent: *\nAllow: /\nDisallow: /admin/\nSitemap: ${siteURL}/sitemap.xml\n`;
await writeFile(new URL('../public/robots.txt', import.meta.url), robots, 'utf8');
