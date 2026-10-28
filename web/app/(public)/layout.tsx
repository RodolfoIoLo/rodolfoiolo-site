import type { ReactNode } from 'react';
import { Footer } from '@/components/layout/footer';
import { Header } from '@/components/layout/header';
import { api } from '@/lib/api';
export default async function PublicLayout({ children }: { children: ReactNode }) {
  const settings = await api.settings();
  return (
    <>
      <Header />
      {children}
      <Footer settings={settings} />
    </>
  );
}
