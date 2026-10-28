import Link from 'next/link';
import type { ReactNode } from 'react';
export function ButtonLink({
  href,
  children,
  secondary = false,
}: {
  href: string;
  children: ReactNode;
  secondary?: boolean;
}) {
  const color = secondary
    ? 'border-[var(--border)] bg-[var(--panel-bg)]'
    : 'border-moss bg-moss text-white';
  return (
    <Link
      className={`inline-flex min-h-11 items-center justify-center border px-4 py-2 font-medium no-underline transition hover:-translate-y-0.5 ${color}`}
      href={href}
    >
      {children}
    </Link>
  );
}
