import type { Metadata } from 'next';
import { notFound } from 'next/navigation';
import { MarkdownContent } from '@/components/markdown/markdown-content';
import { Container } from '@/components/ui/container';
import { Tag } from '@/components/ui/tag';
import { api } from '@/lib/api';
import { site } from '@/lib/site';

type Props = { params: Promise<{ slug: string }> };
export async function generateStaticParams() {
  const page = await api.articles(100);
  return page.items.map(({ slug }) => ({ slug }));
}
export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  try {
    const article = await api.article(slug);
    return {
      title: article.title,
      description: article.summary,
      alternates: { canonical: `/articles/${article.slug}/` },
      openGraph: {
        type: 'article',
        title: article.title,
        description: article.summary,
        publishedTime: article.published_at ?? undefined,
        modifiedTime: article.updated_at,
      },
    };
  } catch {
    return { title: '文章不存在' };
  }
}
export default async function ArticlePage({ params }: Props) {
  const { slug } = await params;
  let article;
  try {
    article = await api.article(slug);
  } catch {
    notFound();
  }
  const structured = {
    '@context': 'https://schema.org',
    '@type': 'Article',
    headline: article.title,
    description: article.summary,
    datePublished: article.published_at,
    dateModified: article.updated_at,
    author: { '@type': 'Person', name: article.author.nickname },
    mainEntityOfPage: `${site.url}/articles/${article.slug}/`,
  };
  return (
    <main className="py-16">
      <Container className="max-w-4xl">
        <article>
          <header className="border-b border-[var(--border)] pb-8">
            <p className="text-sm text-[var(--text-muted)]">
              {article.published_at
                ? new Intl.DateTimeFormat('zh-CN', { dateStyle: 'long' }).format(
                    new Date(article.published_at),
                  )
                : ''}
            </p>
            <h1 className="mt-3 font-display text-4xl font-semibold leading-tight sm:text-5xl">
              {article.title}
            </h1>
            <p className="mt-5 text-lg text-[var(--text-muted)]">{article.summary}</p>
            <ul className="mt-5 flex flex-wrap gap-2">
              {article.tags.map((tag) => (
                <li key={tag.id}>
                  <Tag>{tag.name}</Tag>
                </li>
              ))}
            </ul>
          </header>
          <div className="mt-10">
            <MarkdownContent source={article.content} />
          </div>
        </article>
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{
            __html: JSON.stringify(structured).replaceAll('<', '\\u003c'),
          }}
        />
      </Container>
    </main>
  );
}
