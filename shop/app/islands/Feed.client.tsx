import { useState, useEffect } from "gotsx";
import type { Card } from "host:catalog";
import CardShell from "../ui/CardShell";
import SkeletonCard from "../ui/SkeletonCard";
import Icon from "../ui/Icon";

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
    <button class={beat ? "heart-pop flex h-8 w-8 items-center justify-center rounded-full border border-border bg-background/90 text-foreground shadow-xs" : "flex h-8 w-8 items-center justify-center rounded-full border border-border bg-background/90 text-muted-foreground shadow-xs transition-colors hover:text-foreground"} onClick={toggle} aria-label="心愿单" aria-pressed={w}>
      {w ? <Icon name="heart" fill="currentColor" className="text-foreground" /> : <Icon name="heart" />}
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
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
        {cards.map((c) => <CardShell card={c}><Heart id={c.id} /></CardShell>)}
        {loading && [0, 1, 2, 3, 4, 5, 6, 7, 8, 9].map(() => <SkeletonCard />)}
      </div>
      <div class="mt-8 text-center">
        {!hasMore && cards.length > 0 && <p class="text-sm text-muted-foreground">已经到底了</p>}
        {hasMore && !loading && (
          <button class="btn btn-outline" onClick={load}>加载更多</button>
        )}
        {loading && cards.length > 0 && <span class="inline-flex items-center gap-2 text-sm text-muted-foreground"><Icon name="loader" className="animate-spin" /> 加载中…</span>}
      </div>
    </div>
  );
}
