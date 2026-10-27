import type { Metadata } from 'next';
import Link from 'next/link';
import { Container } from '@/components/ui/container';
import { api } from '@/lib/api';
export const metadata: Metadata = { title: '归档', description: '按年份和月份浏览所有文章。' };
export default async function ArchivePage() {
  const groups = await api.archive();
  return (
    <main className="py-16">
      <Container className="max-w-4xl">
        <h1 className="font-display text-4xl font-semibold">归档</h1>
        <div className="mt-10 space-y-12">
          {groups.length ? (
            groups.map((year) => (
              <section key={year.year}>
                <h2 className="font-display text-3xl font-semibold">{year.year}</h2>
                {year.months.map((month) => (
                  <div className="mt-6 grid gap-3 sm:grid-cols-[5rem_1fr]" key={month.month}>
                    <h3 className="text-[var(--text-muted)]">{month.month} 月</h3>
                    <ul className="space-y-3">
                      {month.articles.map((article) => (
                        <li
                          className="flex flex-wrap justify-between gap-3 border-b border-[var(--border)] pb-3"
                          key={article.slug}
                        >
                          <Link href={`/articles/${article.slug}/`}>{article.title}</Link>
                          <time className="text-sm text-[var(--text-muted)]">
                            {new Intl.DateTimeFormat('zh-CN', {
                              month: '2-digit',
                              day: '2-digit',
                            }).format(new Date(article.published_at))}
                          </time>
                        </li>
                      ))}
                    </ul>
                  </div>
                ))}
              </section>
            ))
          ) : (
            <p>还没有归档文章。</p>
          )}
        </div>
      </Container>
    </main>
  );
}
