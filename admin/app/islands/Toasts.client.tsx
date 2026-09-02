import { useState, useEffect, on } from "gotsx";
import Icon from "../ui/Icon";

interface Toast { id: number; msg: string; kind: string; }

export default function Toasts() {
  const [items, setItems] = useState<Toast[]>([]);
  useEffect(() => {
    on("admin:toast", (d: any) => {
      const id = Date.now() + Math.floor(Math.random() * 1000);
      setItems((cur) => [...cur, { id, msg: d.msg, kind: d.kind ?? "ok" }]);
      setTimeout(() => setItems((cur) => cur.filter((t) => t.id !== id)), 2600);
    });
  }, []);
  return (
    <div class="fixed right-4 top-4 z-50 flex w-72 flex-col gap-2">
      {items.map((t) => (
        <div key={t.id} class="card pop-in flex items-center gap-2.5 px-4 py-3 text-sm shadow-lg">
          <span class={t.kind === "err" ? "text-destructive" : "text-success"}><Icon name={t.kind === "err" ? "alert" : "check"} /></span>
          <span class="font-medium">{t.msg}</span>
        </div>
      ))}
    </div>
  );
}
