import type { Metadata } from 'next';
import { ArticleCard } from '@/components/cards/article-card';
import { Container } from '@/components/ui/container';
import { api } from '@/lib/api';

export const metadata: Metadata = {
  title: '文章',
  description: '软件工程、计算机科学与项目复盘文章。',
};
export default async function ArticlesPage() {
  const page = await api.articles(100);
  return (
    <main className="py-16">
      <Container>
        <h1 className="font-display text-4xl font-semibold">文章</h1>
        <p className="mt-3 text-[var(--text-muted)]">按时间记录问题、取舍和验证过程。</p>
        <div className="mt-10 grid gap-5 md:grid-cols-2">
          {page.items.length ? (
            page.items.map((article) => <ArticleCard key={article.id} article={article} />)
          ) : (
            <p>还没有已发布文章。</p>
          )}
        </div>
      </Container>
    </main>
  );
}
