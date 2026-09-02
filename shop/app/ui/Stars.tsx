/** 评分星: 半星用叠加实现 */
export default function Stars({ rating, size = "sm" }: { rating: number; size?: string }) {
  const pct = Math.round(rating / 5 * 100);
  const cls = size === "md" ? "text-[15px]" : "text-[11px]";
  return (
    <span class={`relative inline-block ${cls} leading-none`} aria-label={`${rating} 星`}>
      <span class="text-ink-200">★★★★★</span>
      <span class="absolute inset-0 overflow-hidden text-gold-500" style={`width:${pct}%`}>★★★★★</span>
    </span>
  );
}
