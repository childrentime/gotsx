import { useState } from "gotsx";
import type { Token } from "host:hl";       // 客户端只能 import type
import { tokenCls } from "../ui/hl";

export interface Tab {
  label: string;
  tokens: Token[];
}

/** 代码标签页: 服务端先用 Go 切好 token 当 props 传进来, 两端都只负责渲染 */
export default function CodeTabs({ tabs }: { tabs: Tab[] }) {
  const [cur, setCur] = useState(0);
  return (
    <div class="overflow-hidden rounded-xl border border-zinc-200 bg-zinc-50 text-[13px] dark:border-zinc-800 dark:bg-zinc-900/70">
      <div class="flex gap-1 border-b border-zinc-200 px-2 pt-2 dark:border-zinc-800" role="tablist">
        {tabs.map((tab, i) => (
          <button
            role="tab"
            aria-selected={cur === i}
            class={cur === i ? "rounded-t-md bg-white px-3 py-1.5 font-mono text-xs font-semibold text-brand-700 shadow-sm dark:bg-zinc-800 dark:text-brand-300" : "rounded-t-md px-3 py-1.5 font-mono text-xs text-zinc-500 hover:text-zinc-900 dark:hover:text-zinc-100"}
            onClick={() => setCur(i)}
          >
            {tab.label}
          </button>
        ))}
      </div>
      <pre class="overflow-x-auto p-4 leading-6"><code class="font-mono">{tabs[cur].tokens.map((t) => <span class={tokenCls(t.kind)}>{t.text}</span>)}</code></pre>
    </div>
  );
}
