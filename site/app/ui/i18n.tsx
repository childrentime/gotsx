/** 双语内联助手: 英文为主, 中文为可选。loc(locale, english, chinese) → 当前语言的字符串。
 *  服务端(Go)与客户端(岛)都编译, 行为一致。 */
export function loc(locale: string, en: string, zh: string): string {
  return locale === "zh" ? zh : en;
}
