import type { Metadata } from 'next';
import { notFound } from 'next/navigation';
import { ExternalLink, Github } from 'lucide-react';
import { MarkdownContent } from '@/components/markdown/markdown-content';
import { Container } from '@/components/ui/container';
import { Tag } from '@/components/ui/tag';
import { api } from '@/lib/api';
import { site } from '@/lib/site';

type Props = { params: Promise<{ slug: string }> };
export async function generateStaticParams() {
  const page = await api.projects(100);
  const params = page.items.map(({ slug }) => ({ slug }));
  return params.length ? params : [{ slug: '__empty__' }];
}
export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  try {
    const project = await api.project(slug);
    return {
      title: project.name,
      description: project.summary,
      alternates: { canonical: `/projects/${project.slug}/` },
    };
  } catch {
    return { title: '项目不存在' };
  }
}
export default async function ProjectPage({ params }: Props) {
  const { slug } = await params;
  let project;
  try {
    project = await api.project(slug);
  } catch {
    notFound();
  }
  const structured = {
    '@context': 'https://schema.org',
    '@type': 'SoftwareSourceCode',
    name: project.name,
    description: project.summary,
    codeRepository: project.repository_url,
    url: `${site.url}/projects/${project.slug}/`,
    programmingLanguage: project.technologies,
  };
  return (
    <main className="py-16">
      <Container className="max-w-4xl">
        <article>
          <header className="border-b border-[var(--border)] pb-8">
            <p className="text-sm text-coral">{project.featured ? '精选项目' : '项目'}</p>
            <h1 className="mt-3 font-display text-4xl font-semibold sm:text-5xl">{project.name}</h1>
            <p className="mt-5 text-lg text-[var(--text-muted)]">{project.summary}</p>
            <ul className="mt-5 flex flex-wrap gap-2">
              {project.technologies.map((technology) => (
                <li key={technology}>
                  <Tag>{technology}</Tag>
                </li>
              ))}
            </ul>
            <div className="mt-6 flex gap-4">
              {project.repository_url ? (
                <a
                  className="inline-flex items-center gap-2"
                  href={project.repository_url}
                  rel="noreferrer"
                  target="_blank"
                >
                  <Github size={18} aria-hidden />
                  源码
                </a>
              ) : null}
              {project.demo_url ? (
                <a
                  className="inline-flex items-center gap-2"
                  href={project.demo_url}
                  rel="noreferrer"
                  target="_blank"
                >
                  <ExternalLink size={18} aria-hidden />
                  演示
                </a>
              ) : null}
            </div>
          </header>
          <div className="mt-10">
            <MarkdownContent source={project.content} />
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
