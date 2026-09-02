import { useState } from "gotsx";
import Badge from "../ui/Badge";

export interface Row {
  cat: string;
  syntax: string;
  note: string;
  ok: boolean;
}

/** 语法参考表: 数据由服务端当 props 传入(JSON), 过滤是 memo, 表格是响应式列表 */
export default function SyntaxSearch({ rows, locale = "en" }: { rows: Row[]; locale?: string }) {
  const zh = locale === "zh";
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
          class="input max-w-sm"
          placeholder={zh ? "搜索语法, 比如 map / useState / class" : "Search syntax, e.g. map / useState / class"}
          value={q}
          onInput={(e) => setQ(e.target.value)}
        />
        <label class="flex items-center gap-2 text-sm text-muted-foreground">
          <input type="checkbox" class="checkbox" checked={onlyNo} onChange={() => setOnlyNo(!onlyNo)} />
          {zh ? "只看不支持的" : "Unsupported only"}
        </label>
        <span class="text-sm tabular-nums text-muted-foreground">{shown.length} / {rows.length}</span>
      </div>
      <div class="overflow-x-auto rounded-lg border border-border">
        <table class="table">
          <thead>
            <tr><th>{zh ? "类别" : "Category"}</th><th>{zh ? "语法" : "Syntax"}</th><th>{zh ? "说明" : "Notes"}</th><th>{zh ? "状态" : "Status"}</th></tr>
          </thead>
          <tbody>
            {shown.map((r) => (
              <tr>
                <td class="whitespace-nowrap text-muted-foreground">{r.cat}</td>
                <td class="font-mono text-[13px]">{r.syntax}</td>
                <td class="text-muted-foreground">{r.note}</td>
                <td>{r.ok ? <Badge color="green">{zh ? "支持" : "Supported"}</Badge> : <Badge color="red">{zh ? "不支持" : "Unsupported"}</Badge>}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
