import Link from 'next/link';
import { Tag } from '@/components/ui/tag';
import type { Article } from '@/lib/types';
export function ArticleCard({ article }: { article: Article }) {
  return (
    <article className="glass grid min-h-52 content-between gap-5 p-5 transition hover:-translate-y-1">
      <div>
        <p className="text-sm text-[var(--text-muted)]">
          {article.published_at
            ? new Intl.DateTimeFormat('zh-CN', { dateStyle: 'medium' }).format(
                new Date(article.published_at),
              )
            : '未发布'}
        </p>
        <h2 className="mt-2 font-display text-2xl font-semibold">
          <Link href={`/articles/${article.slug}/`}>{article.title}</Link>
        </h2>
        <p className="mt-3 text-[var(--text-muted)]">{article.summary}</p>
      </div>
      <ul className="flex flex-wrap gap-2" aria-label="文章标签">
        {article.tags.map((tag) => (
          <li key={tag.id}>
            <Tag>{tag.name}</Tag>
          </li>
        ))}
      </ul>
    </article>
  );
}
