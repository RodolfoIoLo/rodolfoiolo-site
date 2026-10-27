import Link from 'next/link';
import { Container } from '@/components/ui/container';
const links = [
  ['文章', '/articles/'],
  ['项目', '/projects/'],
  ['归档', '/archive/'],
  ['关于', '/about/'],
] as const;
export function Header() {
  return (
    <header className="sticky top-0 z-40 border-b border-[var(--border)] bg-[color-mix(in_srgb,var(--page-bg)_86%,transparent)] backdrop-blur-xl">
      <Container className="flex min-h-16 items-center justify-between gap-6">
        <Link className="font-display text-lg font-semibold no-underline" href="/">
          Rodolfo Iolo
        </Link>
        <nav aria-label="主导航">
          <ul className="flex flex-wrap items-center justify-end gap-x-5 gap-y-2 text-sm">
            {links.map(([label, href]) => (
              <li key={href}>
                <Link href={href}>{label}</Link>
              </li>
            ))}
          </ul>
        </nav>
      </Container>
    </header>
  );
}
