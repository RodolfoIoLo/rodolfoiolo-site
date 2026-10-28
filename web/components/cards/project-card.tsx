import Link from 'next/link';
import { Tag } from '@/components/ui/tag';
import type { Project } from '@/lib/types';
export function ProjectCard({ project }: { project: Project }) {
  return (
    <article className="glass grid min-h-64 content-between gap-5 p-5 transition hover:-translate-y-1">
      <div>
        <p className="text-sm font-medium text-coral">{project.featured ? '精选项目' : '项目'}</p>
        <h2 className="mt-2 font-display text-2xl font-semibold">
          <Link href={`/projects/${project.slug}/`}>{project.name}</Link>
        </h2>
        <p className="mt-3 text-[var(--text-muted)]">{project.summary}</p>
      </div>
      <ul className="flex flex-wrap gap-2" aria-label="技术栈">
        {project.technologies.map((technology) => (
          <li key={technology}>
            <Tag>{technology}</Tag>
          </li>
        ))}
      </ul>
    </article>
  );
}
