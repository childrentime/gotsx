/** 语法高亮 token → class(颜色变量在 app/tailwind.css 里定义)。共享模块: 服务端 CodeBlock 和客户端 CodeTabs 都用 */
export function tokenCls(kind: string): string {
  if (kind === "kw") return "tok-kw";
  if (kind === "str") return "tok-str";
  if (kind === "cmt") return "tok-cmt";
  if (kind === "num") return "tok-num";
  if (kind === "tag") return "tok-tag";
  if (kind === "attr") return "tok-attr";
  if (kind === "type") return "tok-type";
  if (kind === "punct") return "tok-punct";
  return "";
}
