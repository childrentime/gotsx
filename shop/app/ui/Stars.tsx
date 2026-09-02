/** Rating stars: half stars via overlay — solid foreground stars layered over border-colored empty ones */
export default function Stars({ rating, size = "sm" }: { rating: number; size?: string }) {
  const pct = Math.round(rating / 5 * 100);
  const cls = size === "md" ? "text-[14px]" : "text-[11px]";
  return (
    <span class={`relative inline-block ${cls} leading-none tracking-[.05em]`} aria-label={`${rating} out of 5`}>
      <span class="text-border">★★★★★</span>
      <span class="absolute inset-0 overflow-hidden text-foreground" style={`width:${pct}%`}>★★★★★</span>
    </span>
  );
}
