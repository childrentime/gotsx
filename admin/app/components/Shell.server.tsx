import type { Node } from "gotsx";
import Avatar from "../ui/Avatar";
import Icon from "../ui/Icon";
import Logout from "../islands/Logout.client";

interface Item { href: string; label: string; icon: string; key: string; }
const nav: Item[] = [
  { href: "/", label: "仪表盘", icon: "dashboard", key: "home" },
  { href: "/users", label: "用户管理", icon: "users", key: "users" },
];

function navCls(active: string, me: string): string {
  return active === me
    ? "nav-link-active flex items-center gap-2.5 px-2.5 py-2"
    : "nav-link flex items-center gap-2.5 px-2.5 py-2";
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
      <body class="min-h-screen bg-background text-foreground">
        <div id="gotsx-bar"></div>
        <div class="flex min-h-screen">
          <aside class="fixed inset-y-0 left-0 hidden w-60 flex-col border-r border-border bg-background p-4 md:flex">
            <a href="/" class="mb-6 flex items-center gap-2.5 px-2">
              <span class="flex h-8 w-8 items-center justify-center rounded-md bg-primary font-mono text-sm font-semibold text-primary-foreground">g</span>
              <span class="flex flex-col leading-tight">
                <span class="text-sm font-semibold tracking-tight">gotsx admin</span>
                <span class="text-[11px] text-muted-foreground">后台管理</span>
              </span>
            </a>
            <nav class="flex-1 space-y-1">
              {nav.map((it) => <a href={it.href} class={navCls(active, it.key)}><Icon name={it.icon} />{it.label}</a>)}
            </nav>
            <div class="separator mb-3"></div>
            <p class="px-2 text-[11px] leading-5 text-muted-foreground">gotsx 企业示例 · 页面由 Go 渲染 · 交互编译成 signals · 单二进制部署</p>
          </aside>
          <div class="flex min-w-0 flex-1 flex-col md:pl-60">
            <header class="page-header flex h-14 items-center justify-between gap-3 px-6">
              <div class="flex min-w-0 items-center gap-3">
                <a href="/" class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary font-mono text-xs font-semibold text-primary-foreground md:hidden" aria-label="gotsx admin">g</a>
                <h1 class="truncate text-base font-semibold tracking-tight">{title}</h1>
                <nav class="flex items-center gap-0.5 md:hidden">
                  {nav.map((it) => <a href={it.href} class={active === it.key ? "nav-link-active px-2 py-1 text-xs" : "nav-link px-2 py-1 text-xs"}>{it.label}</a>)}
                </nav>
              </div>
              {name !== "" && (
                <div class="flex items-center gap-3">
                  <div class="hidden text-right sm:block">
                    <div class="text-sm font-medium leading-tight">{name}</div>
                    <div class="text-[11px] text-muted-foreground">{role === "admin" ? "管理员" : role === "editor" ? "编辑" : "只读"}</div>
                  </div>
                  <Avatar name={name} />
                  <Logout />
                </div>
              )}
            </header>
            <main class="fade-up flex-1 p-6">{children}</main>
          </div>
        </div>
      </body>
    </html>
  );
}
