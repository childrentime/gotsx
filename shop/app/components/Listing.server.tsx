import type { PagedCards } from "host:catalog";
import ProductCard from "../ui/ProductCard.server";
import Pager from "../ui/Pager.server";
import Icon from "../ui/Icon";

const sortKeys = ["rec", "sales", "price", "priceDesc"];

export default function Listing({ title, base, sort, page, wished, locale = "en" }: { title: string; base: string; sort: string; page: PagedCards; wished: string[]; locale?: string }) {
  return (
    <div>
      <div class="mb-5 flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-baseline gap-2">
          <h1 class="text-xl font-semibold tracking-tight">{title}</h1>
          <span class="text-sm text-muted-foreground tabular-nums">{plural(locale, "listing.count", page.total)}</span>
        </div>
        <div class="inline-flex items-center gap-0.5 rounded-md border border-border bg-background p-0.5 text-xs">
          {sortKeys.map((k) => (
            <a href={`${base}sort=${k}`} class={k === sort ? "rounded-sm bg-primary px-3 py-1.5 font-medium text-primary-foreground" : "rounded-sm px-3 py-1.5 text-muted-foreground transition-colors hover:text-foreground"}>{t(locale, "sort." + k)}</a>
          ))}
        </div>
      </div>
      {page.total === 0 && (
        <div class="card flex flex-col items-center py-24 text-center">
          <span class="flex h-12 w-12 items-center justify-center rounded-full bg-muted text-muted-foreground"><Icon name="search" className="h-5 w-5" /></span>
          <p class="mt-4 font-medium">{t(locale, "listing.empty")}</p>
          <p class="mt-1 text-sm text-muted-foreground">{t(locale, "listing.emptySub")}</p>
          <a href="/" class="btn btn-outline mt-6">{t(locale, "listing.back")}</a>
        </div>
      )}
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
        {page.items.map((c) => <ProductCard card={c} wished={wished.includes(c.id)} locale={locale} />)}
      </div>
      {page.pages > 1 && <Pager base={`${base}sort=${sort}&`} page={page.page} list={page.pageList} />}
    </div>
  );
}
