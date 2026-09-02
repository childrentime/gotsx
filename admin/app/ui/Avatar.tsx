export default function Avatar({ name, size = "sm" }: { name: string; size?: string }) {
  const ch = name.slice(0, 1);
  const hue = (name.length * 47) % 360;
  const s = size === "lg" ? "h-10 w-10 text-base" : "h-8 w-8 text-sm";
  return <span class={`flex ${s} shrink-0 items-center justify-center rounded-full font-semibold text-white`} style={`background:hsl(${hue} 60% 55%)`}>{ch}</span>;
}
