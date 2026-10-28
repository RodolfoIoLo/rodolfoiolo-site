import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import '@/app/globals.css';
import { site } from '@/lib/site';

export const metadata: Metadata = {
  metadataBase: new URL(site.url),
  title: { default: site.name, template: `%s | ${site.name}` },
  description: site.description,
  openGraph: {
    type: 'website',
    siteName: site.name,
    title: site.name,
    description: site.description,
    url: site.url,
  },
  twitter: { card: 'summary_large_image', title: site.name, description: site.description },
  manifest: '/manifest.webmanifest',
  alternates: { types: { 'application/rss+xml': `${site.url}/rss.xml` } },
};

export default function RootLayout({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <html lang="zh-CN" data-scroll-behavior="smooth">
      <body>{children}</body>
    </html>
  );
}
