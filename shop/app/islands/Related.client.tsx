import { useState, useEffect } from "gotsx";
import { related } from "host:catalog";          // typed action: CatalogModule.Related(id) → Promise<Card[]>
import type { Card } from "host:catalog";
import CardShell from "../ui/CardShell";
import SkeletonCard from "../ui/SkeletonCard";

export default function Related({ id, locale = "" }: { id: string; locale?: string }) {
  const [cards, setCards] = useState<Card[]>([]);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    const run = async () => {
      setCards(await related(id));
      setLoading(false);
    };
    run();
  }, []);
  return (
    <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
      {loading && [0, 1, 2, 3, 4, 5].map(() => <SkeletonCard />)}
      {cards.map((c) => <CardShell card={c} locale={locale}></CardShell>)}
    </div>
  );
}
