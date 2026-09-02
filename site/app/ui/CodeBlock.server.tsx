import { tokens } from "host:hl";          // Go 写的 tokenizer
import { tokenCls } from "./hl";
import CopyButton from "../islands/CopyButton.client";

/** 服务端代码块: Go 切 token, 方言渲染 span, 右上角是一个复制岛 */
export default function CodeBlock({ code, lang, title = "" }: { code: string; lang: string; title?: string }) {
  const toks = tokens(code, lang);
  return (
    <div class="my-4 overflow-hidden rounded-lg border border-border bg-muted/40 text-[13px]">
      <div class="flex h-9 items-center justify-between border-b border-border px-4 text-xs text-muted-foreground">
        <span class="truncate font-mono">{title !== "" ? title : lang}</span>
        <CopyButton text={code} />
      </div>
      <pre class="overflow-x-auto p-4 leading-6"><code class="font-mono">{toks.map((t) => <span class={tokenCls(t.kind)}>{t.text}</span>)}</code></pre>
    </div>
  );
}
