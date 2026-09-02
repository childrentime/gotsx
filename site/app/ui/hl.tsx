/** 语法高亮 token → Tailwind class。共享模块: 服务端 CodeBlock 和客户端 CodeTabs 都用 */
export function tokenCls(kind: string): string {
  if (kind === "kw") return "text-purple-700 dark:text-purple-300";
  if (kind === "str") return "text-emerald-700 dark:text-emerald-300";
  if (kind === "cmt") return "text-zinc-500 italic";
  if (kind === "num") return "text-amber-600 dark:text-amber-300";
  if (kind === "tag") return "text-sky-700 dark:text-sky-300";
  if (kind === "attr") return "text-amber-700 dark:text-amber-200";
  if (kind === "type") return "text-teal-700 dark:text-teal-300";
  if (kind === "punct") return "text-zinc-500 dark:text-zinc-400";
  return "";
}
