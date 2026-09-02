import { useState } from "gotsx";

export default function CopyButton({ text }: { text: string }) {
  const [done, setDone] = useState(false);
  const copy = async () => {
    await navigator.clipboard.writeText(text);
    setDone(true);
    setTimeout(() => setDone(false), 1500);
  };
  return (
    <button class="rounded px-2 py-0.5 text-xs text-zinc-500 hover:bg-zinc-200 hover:text-zinc-900 dark:hover:bg-zinc-700 dark:hover:text-zinc-100" onClick={copy}>
      {done ? "已复制 ✓" : "复制"}
    </button>
  );
}
