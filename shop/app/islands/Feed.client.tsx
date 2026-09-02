import { useState, useEffect } from "gotsx";
import { feed } from "host:catalog";             // typed action: CatalogModule.Feed(page) → Promise<FeedResult>
import { toggle } from "host:wish";
import type { Card } from "host:catalog";
import CardShell from "../ui/CardShell";
import SkeletonCard from "../ui/SkeletonCard";
import Icon from "../ui/Icon";

/** Mini wishlist button inside feed cards (local state) */
function Heart({ id, label }: { id: string; label: string }) {
  const [w, setW] = useState(false);
  const [beat, setBeat] = useState(false);
  const onToggle = async () => {
    setW(await toggle(id));
    setBeat(true);
    setTimeout(() => setBeat(false), 400);
  };
  return (
    <button class={beat ? "heart-pop flex h-8 w-8 items-center justify-center rounded-full border border-border bg-background/90 text-foreground shadow-xs" : "flex h-8 w-8 items-center justify-center rounded-full border border-border bg-background/90 text-muted-foreground shadow-xs transition-colors hover:text-foreground"} onClick={onToggle} aria-label={label} aria-pressed={w}>
      {w ? <Icon name="heart" fill="currentColor" className="text-foreground" /> : <Icon name="heart" />}
    </button>
  );
}

/** Home page "for you" feed: fetches after mount, skeleton first, load more */
export default function Feed({ locale = "" }: { locale?: string }) {
  const [cards, setCards] = useState<Card[]>([]);
  const [page, setPage] = useState(0);
  const [loading, setLoading] = useState(true);
  const [hasMore, setHasMore] = useState(true);
  const load = async () => {
    setLoading(true);
    const next = page + 1;
    const d = await feed(next);
    setCards([...cards, ...d.cards]);
    setPage(next);
    setHasMore(d.hasMore);
    setLoading(false);
  };
  useEffect(() => { load(); }, []);
  return (
    <div>
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
        {cards.map((c) => <CardShell card={c} locale={locale}><Heart id={c.id} label={t(locale, "wish.label")} /></CardShell>)}
        {loading && [0, 1, 2, 3, 4, 5, 6, 7, 8, 9].map(() => <SkeletonCard />)}
      </div>
      <div class="mt-8 text-center">
        {!hasMore && cards.length > 0 && <p class="text-sm text-muted-foreground">{t(locale, "feed.end")}</p>}
        {hasMore && !loading && (
          <button class="btn btn-outline" onClick={load}>{t(locale, "feed.more")}</button>
        )}
        {loading && cards.length > 0 && <span class="inline-flex items-center gap-2 text-sm text-muted-foreground"><Icon name="loader" className="animate-spin" /> {t(locale, "feed.loading")}</span>}
      </div>
    </div>
  );
}
