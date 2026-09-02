export default function Pager({ base, page, list }: { base: string; page: number; list: number[] }) {
  return (
    <nav class="mt-8 flex items-center justify-center gap-1.5" aria-label="pagination">
      {list.map((n) => (
        n === page
          ? <span class="btn btn-primary btn-sm min-w-8 px-2 tabular-nums" aria-current="page">{n}</span>
          : <a href={`${base}p=${n}`} class="btn btn-outline btn-sm min-w-8 px-2 tabular-nums">{n}</a>
      ))}
    </nav>
  );
}
