export const site = {
  name: 'Rodolfo Iolo',
  description: 'Computer science, software projects, and engineering notes.',
  url: process.env.NEXT_PUBLIC_SITE_URL ?? 'http://localhost:3000',
  apiBaseURL: process.env.CONTENT_API_BASE_URL ?? 'http://localhost:8080/api/v1',
} as const;
