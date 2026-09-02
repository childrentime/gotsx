import { useState } from "gotsx";
import Badge from "../ui/Badge";

export interface Row {
  cat: string;
  syntax: string;
  note: string;
  ok: boolean;
}

/** 语法参考表: 数据由服务端当 props 传入(JSON), 过滤是 memo, 表格是响应式列表 */
export default function SyntaxSearch({ rows }: { rows: Row[] }) {
  const [q, setQ] = useState("");
  const [onlyNo, setOnlyNo] = useState(false);
  const shown = rows.filter((r) => {
    const k = q.toLowerCase();
    const hit = k === "" || r.syntax.toLowerCase().includes(k) || r.note.toLowerCase().includes(k) || r.cat.includes(q);
    return hit && (!onlyNo || !r.ok);
  });
  return (
    <div>
      <div class="mb-3 flex flex-wrap items-center gap-3">
        <input
          class="h-10 w-full max-w-sm rounded-lg border border-zinc-300 bg-white px-3 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200 dark:border-zinc-700 dark:bg-zinc-900 dark:focus:ring-brand-900"
          placeholder="搜索语法, 比如 map / useState / class"
          value={q}
          onInput={(e) => setQ(e.target.value)}
        />
        <label class="flex items-center gap-2 text-sm text-zinc-600 dark:text-zinc-300">
          <input type="checkbox" checked={onlyNo} onChange={() => setOnlyNo(!onlyNo)} />
          只看不支持的
        </label>
        <span class="text-sm text-zinc-500">{shown.length} / {rows.length}</span>
      </div>
      <div class="overflow-x-auto rounded-xl border border-zinc-200 dark:border-zinc-800">
        <table class="w-full text-left text-sm">
          <thead class="bg-zinc-50 text-xs uppercase text-zinc-500 dark:bg-zinc-900">
            <tr><th class="px-4 py-2">类别</th><th class="px-4 py-2">语法</th><th class="px-4 py-2">说明</th><th class="px-4 py-2">状态</th></tr>
          </thead>
          <tbody class="divide-y divide-zinc-200 dark:divide-zinc-800">
            {shown.map((r) => (
              <tr class="bg-white dark:bg-zinc-950">
                <td class="whitespace-nowrap px-4 py-2 text-zinc-500">{r.cat}</td>
                <td class="px-4 py-2 font-mono text-[13px]">{r.syntax}</td>
                <td class="px-4 py-2 text-zinc-600 dark:text-zinc-300">{r.note}</td>
                <td class="px-4 py-2">{r.ok ? <Badge color="green">支持</Badge> : <Badge color="red">不支持</Badge>}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
