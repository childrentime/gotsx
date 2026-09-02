import { useState, useEffect, on } from "gotsx";

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
        <div class={t.kind === "err"
          ? "toast-in flex items-center gap-2 rounded-lg border border-rose-200 bg-white px-4 py-3 text-sm text-rose-700 shadow-lg"
          : "toast-in flex items-center gap-2 rounded-lg border border-emerald-200 bg-white px-4 py-3 text-sm text-emerald-700 shadow-lg"}>
          <span>{t.kind === "err" ? "⚠️" : "✅"}</span>{t.msg}
        </div>
      ))}
    </div>
  );
}
