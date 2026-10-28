import { Container } from '@/components/ui/container';
import type { SiteSettings } from '@/lib/types';
export function Footer({ settings }: { settings: SiteSettings }) {
  return (
    <footer className="border-t border-[var(--border)] py-10 text-sm text-[var(--text-muted)]">
      <Container className="flex flex-wrap items-center justify-between gap-4">
        <p>{settings.footer_text}</p>
        <p>
          © {new Date().getUTCFullYear()} {settings.site_title}
        </p>
      </Container>
    </footer>
  );
}
