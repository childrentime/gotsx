import type { ErrorProps, Meta } from "gotsx";
import Layout from "../components/Layout.server";
import Icon from "../ui/Icon";

export function meta(): Meta {
  return { title: "Error", noIndex: true };
}

// pages/_error.server.tsx → gen.ErrorPage: a host method returned an error or a page panicked (the message is only
// shown to the user in dev; production shows the generic text)
export default function ErrorPage({ cookies, locale, message }: ErrorProps) {
  const lc = locale !== "" ? locale : "en";
  return (
    <Layout sid={cookies.sid ?? ""} locale={lc} wide>
      <div class="card flex flex-col items-center py-24 text-center">
        <span class="flex h-12 w-12 items-center justify-center rounded-full bg-destructive/10 text-destructive"><Icon name="alert" className="h-5 w-5" /></span>
        <p class="mt-4 font-medium">{t(lc, "error.server")}</p>
        {message !== "" && <p class="mt-2 max-w-md font-mono text-xs text-muted-foreground">{message}</p>}
        <a href={lpath(lc, "/")} class="btn btn-primary mt-6">{t(lc, "error.home")}</a>
      </div>
    </Layout>
  );
}
