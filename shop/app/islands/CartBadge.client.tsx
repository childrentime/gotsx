import { useState, useEffect } from "gotsx";
import { cart } from "../stores/cart";
import Icon from "../ui/Icon";

// The header badge: `count` is a signal of the shared cart store — seeded by the layout on the server, updated by
// whichever island hands an action's CartView to cart.set(...). The pop animation runs on every change after mount.
export default function CartBadge({ label = "" }: { label?: string }) {
  const { count } = cart;
  const [beat, setBeat] = useState(false);
  let seen = -1;
  useEffect(() => {
    const n = count;
    if (seen >= 0 && n !== seen) {
      setBeat(true);
      setTimeout(() => setBeat(false), 400);
    }
    seen = n;
  });
  return (
    <a href="/cart" class="btn btn-outline relative shrink-0">
      <Icon name="cart" />
      <span class="hidden sm:inline">{label}</span>
      {count > 0 && <span class={beat ? "heart-pop absolute -right-2 -top-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1.5 text-[11px] font-medium text-primary-foreground tabular-nums" : "absolute -right-2 -top-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-primary px-1.5 text-[11px] font-medium text-primary-foreground tabular-nums"}>{count}</span>}
    </a>
  );
}
