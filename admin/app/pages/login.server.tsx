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
      <body class="flex min-h-screen items-center justify-center bg-background p-6 text-foreground">
        <div id="gotsx-bar"></div>
        <div class="card fade-up w-full max-w-sm p-8">
          <div class="mb-6 flex flex-col items-center text-center">
            <span class="flex h-10 w-10 items-center justify-center rounded-md bg-primary font-mono text-base font-semibold text-primary-foreground">g</span>
            <h1 class="mt-4 text-lg font-semibold tracking-tight">登录后台</h1>
            <p class="mt-1 text-sm text-muted-foreground">gotsx 企业管理示例</p>
          </div>
          <LoginForm />
          <div class="separator my-5"></div>
          <p class="text-center text-xs text-muted-foreground">
            演示账号 <span class="kbd">admin</span> / <span class="kbd">admin123</span> 或 <span class="kbd">demo</span> / <span class="kbd">demo</span>
          </p>
        </div>
      </body>
    </html>
  );
}
