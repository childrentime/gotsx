import { useState } from "gotsx";

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
    <button class={beat ? "heart-pop flex h-8 w-8 items-center justify-center rounded-full bg-white/95 text-[15px] shadow-md" : "flex h-8 w-8 items-center justify-center rounded-full bg-white/95 text-[15px] shadow-md transition hover:scale-110"} onClick={toggle} aria-label="加入心愿单">
      {w ? "❤️" : "🤍"}
    </button>
  );
}
