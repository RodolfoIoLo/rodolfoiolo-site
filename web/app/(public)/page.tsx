import Link from 'next/link';
import { ArticleCard } from '@/components/cards/article-card';
import { ProjectCard } from '@/components/cards/project-card';
import { ButtonLink } from '@/components/ui/button-link';
import { Container } from '@/components/ui/container';
import { api } from '@/lib/api';

export default async function HomePage() {
  const [articles, projects, stats] = await Promise.all([
    api.articles(3),
    api.projects(3),
    api.stats(),
  ]);
  return (
    <main>
      <section className="flex min-h-[72vh] items-center border-b border-[var(--border)] py-20">
        <Container>
          <p className="font-medium text-moss">计算机专业学生 · 独立开发者</p>
          <h1 className="mt-4 max-w-4xl font-display text-5xl font-semibold leading-tight sm:text-6xl">
            Rodolfo Iolo
          </h1>
          <p className="mt-6 max-w-2xl text-lg text-[var(--text-muted)]">
            记录软件工程实践、项目复盘，以及把想法从空仓库推进到生产环境的全过程。
          </p>
          <div className="mt-8 flex flex-wrap gap-3">
            <ButtonLink href="/projects/">查看项目</ButtonLink>
            <ButtonLink href="/articles/" secondary>
              阅读文章
            </ButtonLink>
          </div>
          <dl className="mt-12 grid max-w-2xl grid-cols-2 gap-4 sm:grid-cols-4">
            {[
              ['文章', stats.total_articles],
              ['项目', stats.total_projects],
              ['分类', stats.total_categories],
              ['标签', stats.total_tags],
            ].map(([label, value]) => (
              <div key={label} className="border-l-2 border-moss pl-3">
                <dt className="text-sm text-[var(--text-muted)]">{label}</dt>
                <dd className="m-0 font-display text-2xl font-semibold">{value}</dd>
              </div>
            ))}
          </dl>
        </Container>
      </section>
      <section className="py-20">
        <Container>
          <div className="flex items-end justify-between gap-4">
            <div>
              <p className="text-sm text-coral">Writing</p>
              <h2 className="font-display text-3xl font-semibold">最新文章</h2>
            </div>
            <Link href="/articles/">全部文章</Link>
          </div>
          <div className="mt-8 grid gap-5 lg:grid-cols-3">
            {articles.items.length ? (
              articles.items.map((article) => <ArticleCard key={article.id} article={article} />)
            ) : (
              <p>文章正在准备中。</p>
            )}
          </div>
        </Container>
      </section>
      <section className="border-t border-[var(--border)] py-20">
        <Container>
          <div>
            <p className="text-sm text-sky">Selected work</p>
            <h2 className="font-display text-3xl font-semibold">项目实践</h2>
          </div>
          <div className="mt-8 grid gap-5 md:grid-cols-2 lg:grid-cols-3">
            {projects.items.length ? (
              projects.items.map((project) => <ProjectCard key={project.id} project={project} />)
            ) : (
              <p>项目正在整理中。</p>
            )}
          </div>
        </Container>
      </section>
    </main>
  );
}
