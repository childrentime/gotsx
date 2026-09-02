import type { PageProps, Meta } from "gotsx";
import LoginForm from "../islands/LoginForm.client";

export function meta(): Meta {
  return { title: "Sign in", description: "Sign in to the gotsx back-office demo." };
}

export default function Login({ session }: PageProps) {
  if (session.user !== "") redirect("/");
  return (
    <div class="flex min-h-screen items-center justify-center p-6">
      <div class="card fade-up w-full max-w-sm p-8">
        <div class="mb-6 flex flex-col items-center text-center">
          <span class="flex h-10 w-10 items-center justify-center rounded-md bg-primary font-mono text-base font-semibold text-primary-foreground">g</span>
          <h1 class="mt-4 text-lg font-semibold tracking-tight">Sign in</h1>
          <p class="mt-1 text-sm text-muted-foreground">gotsx back-office demo</p>
        </div>
        <LoginForm />
        <div class="separator my-5"></div>
        <p class="text-center text-xs text-muted-foreground">
          Demo accounts: <span class="kbd">admin</span> / <span class="kbd">admin123</span> or <span class="kbd">demo</span> / <span class="kbd">demo</span>
        </p>
      </div>
    </div>
  );
}
