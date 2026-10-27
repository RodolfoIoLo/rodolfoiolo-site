import Link from 'next/link';
export default function NotFound() {
  return (
    <main className="grid min-h-screen place-items-center px-5 text-center">
      <div>
        <p className="font-display text-6xl font-semibold">404</p>
        <h1 className="mt-4 text-2xl font-semibold">没有找到这个页面</h1>
        <p className="mt-3 text-[var(--text-muted)]">地址可能已变化，或者内容尚未发布。</p>
        <Link className="mt-6 inline-block" href="/">
          返回首页
        </Link>
      </div>
    </main>
  );
}
