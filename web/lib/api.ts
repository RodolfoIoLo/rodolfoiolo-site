import { site } from '@/lib/site';
import type {
  ArchiveGroup,
  ArticleDetail,
  ArticleListResponse,
  Envelope,
  ProjectDetail,
  ProjectListResponse,
  PublicStats,
  SiteSettings,
} from '@/lib/types';

async function get<T>(path: string): Promise<T> {
  const response = await fetch(`${site.apiBaseURL}${path}`, {
    headers: { Accept: 'application/json' },
  });
  if (!response.ok) {
    throw new Error(`Content API ${path} failed with ${response.status}`);
  }
  const envelope = (await response.json()) as Envelope<T>;
  if (envelope.code !== 0) {
    throw new Error(`Content API ${path} returned code ${envelope.code}`);
  }
  return envelope.data;
}

export const api = {
  articles: (pageSize = 20) => get<ArticleListResponse>(`/articles?page=1&page_size=${pageSize}`),
  article: (slug: string) => get<ArticleDetail>(`/articles/${encodeURIComponent(slug)}`),
  archive: () => get<ArchiveGroup[]>('/articles/archive'),
  projects: (pageSize = 20) => get<ProjectListResponse>(`/projects?page=1&page_size=${pageSize}`),
  project: (slug: string) => get<ProjectDetail>(`/projects/${encodeURIComponent(slug)}`),
  settings: () => get<SiteSettings>('/site/settings'),
  stats: () => get<PublicStats>('/site/stats'),
};
