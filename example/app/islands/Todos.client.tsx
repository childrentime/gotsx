import { useState } from "gotsx";

interface Item {
  id: number;
  text: string;
}

/** keyed 列表: <li key={x.id}> 让客户端按 key 复用 DOM —— 每行里的 input 内容在增删 / 上移后都还在 */
export default function Todos({ initial }: { initial: string[] }) {
  const [items, setItems] = useState<Item[]>(initial.map((text, i) => ({ id: i, text })));
  const [draft, setDraft] = useState("");
  const [seq, setSeq] = useState(initial.length);
  const [showLen, setShowLen] = useState(false);
  const add = () => {
    if (draft.trim() === "") return;
    setItems([...items, { id: seq, text: draft.trim() }]);
    setSeq(seq + 1);
    setDraft("");
  };
  const remove = (id: number) => setItems(items.filter((x) => x.id !== id));
  const up = (id: number) => {
    const i = items.findIndex((x) => x.id === id);
    if (i <= 0) return;
    const copy = items.slice();
    const [moved] = copy.splice(i, 1);
    copy.splice(i - 1, 0, moved);
    setItems(copy);
  };
  return (
    <div class="todos">
      <div class="row">
        <input class="input" style="flex:1" value={draft} placeholder="Add an item" onInput={(e) => setDraft(e.target.value)} onKeyDown={(e) => { if (e.key === "Enter") add(); }} />
        <button class="btn" onClick={add}>Add</button>
        <button class="btn btn-outline toggle-len" onClick={() => setShowLen(!showLen)}>{showLen ? "hide" : "show"} lengths</button>
      </div>
      <ul>
        {items.map((x) => (
          <li key={x.id} data-id={x.id}>
            <input class="input note" placeholder="notes (kept on reorder)" />
            <span>{x.text}</span>
            {showLen && <em class="len badge badge-outline">{x.text.length}</em>}
            <button class="btn btn-ghost btn-sm up" aria-label="move up" onClick={() => up(x.id)}>↑</button>
            <button class="btn btn-ghost btn-sm del" aria-label="remove" onClick={() => remove(x.id)}>✕</button>
          </li>
        ))}
      </ul>
      <p class="muted">{items.length} items · keyed list: DOM nodes are reused on reorder</p>
    </div>
  );
}
