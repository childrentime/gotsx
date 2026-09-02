/** 中性头像: 首字 + muted 圆底 */
export default function Avatar({ name, size = "sm" }: { name: string; size?: string }) {
  const ch = name.slice(0, 1);
  const s = size === "lg" ? "h-10 w-10 text-sm" : "h-8 w-8 text-xs";
  return <span class={`flex ${s} shrink-0 items-center justify-center rounded-full bg-muted font-medium text-foreground`}>{ch}</span>;
}
