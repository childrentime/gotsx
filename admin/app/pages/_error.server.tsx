import type { ErrorProps, Meta } from "gotsx";
import Shell from "../components/Shell.server";
import Icon from "../ui/Icon";

export function meta(): Meta {
  return { title: "Something went wrong", noIndex: true };
}

// pages/_error.server.tsx → gen.ErrorPage: a host method returned an error or a page panicked (message only in dev)
export default function ErrorPage({ message, session }: ErrorProps) {
  return (
    <Shell title="Error" name={session.name} role={session.role}>
      <div class="flex flex-col items-center justify-center py-32 text-center">
        <span class="text-destructive"><Icon name="alert" cls="h-8 w-8" /></span>
        <p class="mt-4 text-sm text-muted-foreground">Something went wrong, please try again later.</p>
        {message !== "" && <p class="mt-2 font-mono text-xs text-muted-foreground">{message}</p>}
        <a href="/" class="btn btn-outline mt-5">Back to dashboard</a>
      </div>
    </Shell>
  );
}
