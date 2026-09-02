import type { Node } from "gotsx";
import Avatar from "../ui/Avatar";
import Logout from "../islands/Logout.client";

interface Item { href: string; label: string; icon: string; key: string; }
const nav: Item[] = [
  { href: "/", label: "仪表盘", icon: "📊", key: "home" },
  { href: "/users", label: "用户管理", icon: "👥", key: "users" },
];

function navCls(active: string, me: string): string {
  return active === me
    ? "flex items-center gap-3 rounded-lg bg-brand-500 px-3 py-2 text-sm font-semibold text-white"
    : "flex items-center gap-3 rounded-lg px-3 py-2 text-sm text-slate-600 hover:bg-slate-100";
}

export default function Shell({ title, active = "", name = "", role = "", children }: { title: string; active?: string; name?: string; role?: string; children?: Node }) {
  return (
    <html lang="zh-CN">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>{title} · gotsx admin</title>
        <link rel="stylesheet" href="/public/tailwind.css" />
      </head>
      <body class="min-h-screen bg-slate-50 font-sans text-slate-900 antialiased">
        <div id="gotsx-bar"></div>
        <div class="flex min-h-screen">
          <aside class="fixed inset-y-0 left-0 hidden w-60 flex-col border-r border-slate-200 bg-white p-4 md:flex">
            <div class="mb-6 flex items-center gap-2 px-2">
              <span class="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-brand-500 to-brand-700 text-lg font-black text-white">g</span>
              <div><div class="text-sm font-black leading-tight">gotsx admin</div><div class="text-[11px] text-slate-400">后台管理</div></div>
            </div>
            <nav class="flex-1 space-y-1">
              {nav.map((it) => <a href={it.href} class={navCls(active, it.key)}><span>{it.icon}</span>{it.label}</a>)}
            </nav>
            <div class="rounded-lg bg-slate-50 p-3 text-[11px] leading-5 text-slate-400">
              gotsx 企业示例 · 页面由 Go 渲染 · 交互编译成 signals · 单二进制部署
            </div>
          </aside>
          <div class="flex min-w-0 flex-1 flex-col md:pl-60">
            <header class="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-slate-200 bg-white/85 px-5 backdrop-blur">
              <h1 class="text-base font-bold">{title}</h1>
              {name !== "" && (
                <div class="flex items-center gap-3">
                  <div class="text-right">
                    <div class="text-sm font-semibold leading-tight">{name}</div>
                    <div class="text-[11px] text-slate-400">{role === "admin" ? "管理员" : role === "editor" ? "编辑" : "只读"}</div>
                  </div>
                  <Avatar name={name} />
                  <Logout />
                </div>
              )}
            </header>
            <main class="flex-1 p-5">{children}</main>
          </div>
        </div>
      </body>
    </html>
  );
}
