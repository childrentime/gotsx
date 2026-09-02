import type { Node } from "gotsx";
import { now } from "host:intl";
import Counter from "../islands/Counter.client";

export interface LayoutProps {
  title: string;
  children?: Node;
}

/** 整个文档由方言渲染; 头部的 Counter 是岛, SPA 跳转时状态存活 */
export default function Layout({ title, children }: LayoutProps) {
  return (
    <html lang="zh-CN">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>{title} · gotsx</title>
        <link rel="stylesheet" href="/public/app.css" />
      </head>
      <body>
        <header class="bar">
          <a href="/" class="brand">gotsx</a>
          <Counter start={0} />
        </header>
        <main class="wrap">{children}</main>
        <footer class="foot">Go 原生渲染, 编译自 TSX 方言 · 服务端时间 {now()}</footer>
      </body>
    </html>
  );
}
