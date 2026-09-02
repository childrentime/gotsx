import { tokens } from "host:hl";          // Go 写的 tokenizer
import { tokenCls } from "./hl";
import CopyButton from "../islands/CopyButton.client";

/** 服务端代码块: Go 切 token, 方言渲染 span, 右上角是一个复制岛 */
export default function CodeBlock({ code, lang, title = "" }: { code: string; lang: string; title?: string }) {
  const toks = tokens(code, lang);
  return (
    <div class="group my-4 overflow-hidden rounded-xl border border-zinc-200 bg-zinc-50 text-[13px] dark:border-zinc-800 dark:bg-zinc-900/70">
      <div class="flex items-center justify-between border-b border-zinc-200 px-4 py-1.5 text-xs text-zinc-500 dark:border-zinc-800">
        <span class="font-mono">{title !== "" ? title : lang}</span>
        <CopyButton text={code} />
      </div>
      <pre class="overflow-x-auto p-4 leading-6"><code class="font-mono">{toks.map((t) => <span class={tokenCls(t.kind)}>{t.text}</span>)}</code></pre>
    </div>
  );
}
