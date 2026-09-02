import { useState, useEffect } from "gotsx";
import type { Card } from "host:catalog";
import CardShell from "../ui/CardShell";
import SkeletonCard from "../ui/SkeletonCard";

/** 心愿单迷你按钮(信息流卡片内, 局部状态) */
function Heart({ id }: { id: string }) {
  const [w, setW] = useState(false);
  const [beat, setBeat] = useState(false);
  const toggle = async () => {
    const r = await fetch("/actions/wish", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ id }) });
    const d = await r.json();
    setW(d.wished);
    setBeat(true);
    setTimeout(() => setBeat(false), 400);
  };
  return (
    <button class={beat ? "heart-pop flex h-8 w-8 items-center justify-center rounded-full bg-white/95 text-[15px] shadow-md" : "flex h-8 w-8 items-center justify-center rounded-full bg-white/95 text-[15px] shadow-md transition hover:scale-110"} onClick={toggle} aria-label="心愿单">
      {w ? "❤️" : "🤍"}
    </button>
  );
}

/** 首页"为你推荐"信息流: 挂载后 fetch, 先骨架后内容, 支持加载更多 */
export default function Feed() {
  const [cards, setCards] = useState<Card[]>([]);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(true);
  const [hasMore, setHasMore] = useState(true);
  const load = async () => {
    setLoading(true);
    const next = page + 1;
    const r = await fetch(`/api/feed?page=${next}`);
    const d = await r.json();
    setCards([...cards, ...(d.cards as Card[])]);
    setPage(next);
    setHasMore(d.hasMore);
    setLoading(false);
  };
  useEffect(() => { load(); }, []);
  return (
    <div>
      <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
        {cards.map((c) => <CardShell card={c}><Heart id={c.id} /></CardShell>)}
        {loading && [0, 1, 2, 3, 4, 5, 6, 7, 8, 9].map(() => <SkeletonCard />)}
      </div>
      <div class="mt-8 text-center">
        {!hasMore && cards.length > 0 && <p class="text-sm text-ink-300">— 已经到底了 —</p>}
        {hasMore && !loading && (
          <button class="rounded-full border border-ink-200 bg-white px-8 py-2.5 text-sm font-semibold text-ink-700 transition hover:border-brand-400 hover:text-brand-600" onClick={load}>加载更多好物</button>
        )}
        {loading && cards.length > 0 && <span class="inline-flex items-center gap-2 text-sm text-ink-400"><span class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-ink-200 border-t-brand-500"></span> 加载中…</span>}
      </div>
    </div>
  );
}
