import type { ReactNode } from 'react';
export function Tag({ children }: { children: ReactNode }) {
  return (
    <span className="inline-flex items-center border border-[var(--border)] bg-[var(--panel-bg)] px-2 py-1 text-xs text-[var(--text-muted)]">
      {children}
    </span>
  );
}
