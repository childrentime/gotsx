import { useState, useEffect, on } from "gotsx";
import type { Flash } from "gotsx";
import Icon from "../ui/Icon";

interface Toast { id: number; msg: string; kind: string; }

/** Toasts: the session's flash messages (server-rendered by the layout, then hydrated) plus client events
 *  (emit("admin:toast", { msg, kind })). Kinds: ok | error | info (the alert-* classes). */
export default function Toasts({ initial }: { initial: Flash[] }) {
  const [items, setItems] = useState<Toast[]>(initial.map((f, i) => ({ id: i + 1, msg: f.text, kind: f.kind })));
  const dismiss = (id: number) => { setTimeout(() => setItems((cur) => cur.filter((t) => t.id !== id)), 3200); };
  useEffect(() => {
    for (let i = 0; i < initial.length; i++) dismiss(i + 1);
    on("admin:toast", (d: any) => {
      const id = Date.now() + Math.floor(Math.random() * 1000);
      setItems((cur) => [...cur, { id, msg: d.msg, kind: d.kind ?? "ok" }]);
      dismiss(id);
    });
  }, []);
  const kind = (k: string) => (k === "err" ? "error" : k);
  return (
    <div class="fixed right-4 top-4 z-50 flex w-72 flex-col gap-2">
      {items.map((t) => (
        <div key={t.id} class={"alert alert-" + kind(t.kind) + " pop-in flex items-center gap-2.5 shadow-lg"} role="status">
          <Icon name={kind(t.kind) === "ok" ? "check" : "alert"} />
          <span class="font-medium">{t.msg}</span>
        </div>
      ))}
    </div>
  );
}
