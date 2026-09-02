import { useState } from "gotsx";
import Icon from "../ui/Icon";

export default function WishButton({ id, wished }: { id: string; wished: boolean }) {
  const [w, setW] = useState(wished);
  const [beat, setBeat] = useState(false);
  const toggle = async () => {
    const r = await fetch("/actions/wish", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ id }) });
    const d = await r.json();
    setW(d.wished);
    setBeat(true);
    setTimeout(() => setBeat(false), 400);
  };
  return (
    <button class={beat ? "heart-pop flex h-8 w-8 items-center justify-center rounded-full border border-border bg-background/90 text-foreground shadow-xs" : "flex h-8 w-8 items-center justify-center rounded-full border border-border bg-background/90 text-muted-foreground shadow-xs transition-colors hover:text-foreground"} onClick={toggle} aria-label="加入心愿单" aria-pressed={w}>
      {w ? <Icon name="heart" fill="currentColor" className="text-foreground" /> : <Icon name="heart" />}
    </button>
  );
}
