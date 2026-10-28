import type { components } from '@/types/openapi';

export type Article = components['schemas']['Article'];
export type ArticleDetail = components['schemas']['ArticleDetail'];
export type ArticleListResponse = components['schemas']['ArticleListResponse'];
export type Project = components['schemas']['Project'];
export type ProjectDetail = components['schemas']['ProjectDetail'];
export type ProjectListResponse = components['schemas']['ProjectListResponse'];
export type ArchiveGroup = components['schemas']['ArchiveGroup'];
export type SiteSettings = components['schemas']['SiteSettings'];
export type PublicStats = components['schemas']['PublicStats'];

export type Envelope<T> = { code: number; message: string; data: T };
