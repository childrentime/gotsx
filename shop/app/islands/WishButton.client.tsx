import { useState } from "gotsx";
import { toggle } from "host:wish";              // typed action: WishModule.Toggle → Promise<boolean>
import Icon from "../ui/Icon";

export default function WishButton({ id, wished, label = "Add to wishlist" }: { id: string; wished: boolean; label?: string }) {
  const [w, setW] = useState(wished);
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
