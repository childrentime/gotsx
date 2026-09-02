import type { PageProps } from "gotsx";
import LoginForm from "../islands/LoginForm.client";

export default function Login({}: PageProps) {
  return (
    <html lang="zh-CN">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>登录 · gotsx admin</title>
        <link rel="stylesheet" href="/public/tailwind.css" />
      </head>
      <body class="flex min-h-screen items-center justify-center bg-gradient-to-br from-slate-100 to-slate-200 font-sans text-slate-900 antialiased">
        <div id="gotsx-bar"></div>
        <div class="w-full max-w-sm rounded-2xl border border-slate-200 bg-white p-8 shadow-xl">
          <div class="mb-6 flex flex-col items-center">
            <span class="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-500 to-brand-700 text-2xl font-black text-white">g</span>
            <h1 class="mt-3 text-xl font-black">登录后台</h1>
            <p class="mt-1 text-xs text-slate-400">gotsx 企业管理示例</p>
          </div>
          <LoginForm />
          <div class="mt-5 rounded-lg bg-slate-50 p-3 text-center text-xs text-slate-500">
            演示账号:<b>admin</b> / <b>admin123</b> &nbsp;或&nbsp; <b>demo</b> / <b>demo</b>
          </div>
        </div>
      </body>
    </html>
  );
}
