import { useState } from "gotsx";
import Icon from "../ui/Icon";

export default function CopyButton({ text }: { text: string }) {
  const [done, setDone] = useState(false);
  const copy = async () => {
    await navigator.clipboard.writeText(text);
    setDone(true);
    setTimeout(() => setDone(false), 1500);
  };
  return (
    <button class="btn btn-ghost h-6 gap-1.5 px-2 text-xs text-muted-foreground hover:text-foreground" onClick={copy} aria-label="copy">
      {done ? <Icon name="check" className="icon h-3.5 w-3.5 text-success" /> : <Icon name="copy" className="icon h-3.5 w-3.5" />}
      {done ? "已复制" : "复制"}
    </button>
  );
}
