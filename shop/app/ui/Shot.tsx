/** 商品棚拍图: 中性浅灰影棚面 + 单一柔和投影, size 控制 emoji 大小。hue 保留在数据里但视觉上不再着色 */
export default function Shot({ emoji, hue, size = "text-7xl", rounded = "rounded-t-lg" }: { emoji: string; hue: number; size?: string; rounded?: string }) {
  return (
    <div class={`shot flex aspect-square items-center justify-center ${rounded}`} data-hue={hue}>
      <span class={`emoji ${size}`}>{emoji}</span>
    </div>
  );
}
