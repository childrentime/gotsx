import type { PageProps, Meta } from "gotsx";
import { categories } from "host:catalog";
import Layout from "../components/Layout.server";
import Icon from "../ui/Icon";

export function meta(props: PageProps): Meta {
  const lc = props.locale !== "" ? props.locale : "en";
  return { title: t(lc, "error.notFound"), noIndex: true };
}

// pages/_404.server.tsx → gen.NotFound: any unknown path, wrapped by _layout like every page
export default function NotFound({ locale }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  return (
    <Layout locale={lc} wide>
      <div class="card flex flex-col items-center py-24 text-center">
        <span class="flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground"><Icon name="search" className="h-5 w-5" /></span>
        <p class="mt-4 font-medium">{t(lc, "error.notFound")}</p>
        <a href={lpath(lc, "/")} class="btn btn-primary mt-6">{t(lc, "error.home")}</a>
        <div class="mt-8 flex flex-wrap justify-center gap-2">
          {categories().map((c) => (
            <a href={lpath(lc, `/c/${c.key}`)} class="badge badge-outline transition-colors hover:bg-accent">{c.label}</a>
          ))}
        </div>
      </div>
    </Layout>
  );
}
