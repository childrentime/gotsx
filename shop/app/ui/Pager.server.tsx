export default function Pager({ base, page, list }: { base: string; page: number; list: number[] }) {
  return (
    <nav class="mt-8 flex items-center justify-center gap-1.5">
      {list.map((n) => (
        n === page
          ? <span class="flex h-9 min-w-9 items-center justify-center rounded-lg bg-brand-500 px-2.5 text-sm font-bold text-white shadow-sm">{n}</span>
          : <a href={`${base}p=${n}`} class="flex h-9 min-w-9 items-center justify-center rounded-lg border border-ink-200 bg-white px-2.5 text-sm text-ink-600 transition hover:border-brand-400 hover:text-brand-600">{n}</a>
      ))}
    </nav>
  );
}
