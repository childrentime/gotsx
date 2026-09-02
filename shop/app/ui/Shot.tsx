/** 商品棚拍图: 近白影棚背景 + 落地投影, size 控制 emoji 大小 */
export default function Shot({ emoji, hue, size = "text-7xl", rounded = "rounded-t-xl2" }: { emoji: string; hue: number; size?: string; rounded?: string }) {
  return (
    <div class={`shot flex aspect-square items-center justify-center ${rounded}`} style={`background:radial-gradient(120% 120% at 50% 18%, #ffffff 0%, hsl(${hue} 46% 96%) 62%, hsl(${hue} 40% 92%) 100%)`}>
      <span class={`emoji ${size}`}>{emoji}</span>
    </div>
  );
}
