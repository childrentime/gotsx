import { useState, useEffect } from "gotsx";
import type { Card } from "host:catalog";
import CardShell from "../ui/CardShell";
import SkeletonCard from "../ui/SkeletonCard";

export default function Related({ id }: { id: string }) {
  const [cards, setCards] = useState<Card[]>([]);
  const [loading, setLoading] = useState(true);
  useEffect(() => {
    const run = async () => {
      const r = await fetch(`/api/related?id=${encodeURIComponent(id)}`);
      const d = await r.json();
      setCards(d.cards as Card[]);
      setLoading(false);
    };
    run();
  }, []);
  return (
    <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-6">
      {loading && [0, 1, 2, 3, 4, 5].map(() => <SkeletonCard />)}
      {cards.map((c) => <CardShell card={c}></CardShell>)}
    </div>
  );
}
