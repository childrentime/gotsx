import { useState, useEffect } from "gotsx";

export default function Countdown({ left0 }: { left0: number }) {
  const [left, setLeft] = useState(left0);
  useEffect(() => {
    setInterval(() => setLeft((l) => Math.max(0, l - 1000)), 1000);
  });
  const s = Math.floor(left / 1000);
  const pad = (n: number) => (n < 10 ? "0" : "") + String(n);
  const cell = "flex h-6 min-w-6 items-center justify-center rounded-sm bg-primary px-1 font-mono text-[12px] font-medium text-primary-foreground tabular-nums";
  return (
    <span class="flex items-center gap-1">
      <span class={cell}>{pad(Math.floor(s / 3600))}</span>
      <span class="text-muted-foreground">:</span>
      <span class={cell}>{pad(Math.floor(s % 3600 / 60))}</span>
      <span class="text-muted-foreground">:</span>
      <span class={cell}>{pad(s % 60)}</span>
    </span>
  );
}
