/** Product studio shot: neutral light-gray surface + one soft shadow; size sets the emoji size. hue stays in the data but no longer tints anything */
export default function Shot({ emoji, hue, size = "text-7xl", rounded = "rounded-t-lg" }: { emoji: string; hue: number; size?: string; rounded?: string }) {
  return (
    <div class={`shot flex aspect-square items-center justify-center ${rounded}`} data-hue={hue}>
      <span class={`emoji ${size}`}>{emoji}</span>
    </div>
  );
}
