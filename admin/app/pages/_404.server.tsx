import type { PageProps, Meta } from "gotsx";
import Shell from "../components/Shell.server";
import Icon from "../ui/Icon";

export function meta(): Meta {
  return { title: "Page not found", noIndex: true };
}

// Signed-in visitors get the dashboard chrome around the message; anonymous ones only the message
// (the sidebar is not exposed on unknown paths).
export default function NotFound({ path, session }: PageProps) {
  const body = (
    <div class="flex flex-col items-center justify-center py-32 text-center">
      <span class="text-muted-foreground"><Icon name="search" cls="h-8 w-8" /></span>
      <p class="mt-4 text-sm text-muted-foreground">There is nothing at {path}.</p>
      <a href={session.user === "" ? "/login" : "/"} class="btn btn-outline mt-5">{session.user === "" ? "Sign in" : "Back to dashboard"}</a>
    </div>
  );
  if (session.user === "") return <main class="container-page">{body}</main>;
  return (
    <Shell title="Not found" name={session.name} role={session.role}>
      {body}
    </Shell>
  );
}
