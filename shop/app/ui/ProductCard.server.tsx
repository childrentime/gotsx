import type { Card } from "host:catalog";
import CardShell from "./CardShell";
import WishButton from "../islands/WishButton.client";

/** Server-side listing card: CardShell + the wishlist island */
export default function ProductCard({ card, wished, locale = "en" }: { card: Card; wished: boolean; locale?: string }) {
  return (
    <CardShell card={card} locale={locale}>
      <WishButton id={card.id} wished={wished} label={t(locale, "wish.label")} />
    </CardShell>
  );
}
