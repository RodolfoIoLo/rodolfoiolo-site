import type { Metadata } from 'next';
import { MarkdownContent } from '@/components/markdown/markdown-content';
import { Container } from '@/components/ui/container';
import { api } from '@/lib/api';
export const metadata: Metadata = { title: '关于', description: '关于我、学习方向和联系方式。' };
export default async function AboutPage() {
  const settings = await api.settings();
  return (
    <main className="py-16">
      <Container className="max-w-4xl">
        <h1 className="font-display text-4xl font-semibold">关于</h1>
        <div className="mt-8">
          <MarkdownContent source={settings.about_content} />
        </div>
        <ul className="mt-10 flex flex-wrap gap-5">
          {settings.social_links.map((link) => (
            <li key={link.url}>
              <a href={link.url} rel="me noreferrer" target="_blank">
                {link.name}
              </a>
            </li>
          ))}
        </ul>
      </Container>
    </main>
  );
}
