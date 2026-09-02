import type { Card } from "host:catalog";
import CardShell from "./CardShell";
import WishButton from "../islands/WishButton.client";

/** 服务端列表卡: CardShell + 心愿单岛 */
export default function ProductCard({ card, wished }: { card: Card; wished: boolean }) {
  return (
    <CardShell card={card}>
      <WishButton id={card.id} wished={wished} />
    </CardShell>
  );
}
