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
    <div class="overflow-hidden rounded-lg border border-border bg-muted/40 text-[13px]">
      <div class="flex gap-4 border-b border-border px-4" role="tablist">
        {tabs.map((tab, i) => (
          <button
            role="tab"
            aria-selected={cur === i}
            class={cur === i ? "-mb-px border-b-2 border-foreground py-2 font-mono text-xs font-medium text-foreground" : "-mb-px border-b-2 border-transparent py-2 font-mono text-xs text-muted-foreground transition-colors hover:text-foreground"}
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
