import type { Metadata } from 'next';
import { ProjectCard } from '@/components/cards/project-card';
import { Container } from '@/components/ui/container';
import { api } from '@/lib/api';
export const metadata: Metadata = {
  title: '项目',
  description: '从需求、架构、实现到部署的项目实践。',
};
export default async function ProjectsPage() {
  const page = await api.projects(100);
  return (
    <main className="py-16">
      <Container>
        <h1 className="font-display text-4xl font-semibold">项目</h1>
        <p className="mt-3 text-[var(--text-muted)]">不仅展示结果，也记录工程过程与关键取舍。</p>
        <div className="mt-10 grid gap-5 md:grid-cols-2">
          {page.items.length ? (
            page.items.map((project) => <ProjectCard key={project.id} project={project} />)
          ) : (
            <p>还没有公开项目。</p>
          )}
        </div>
      </Container>
    </main>
  );
}
