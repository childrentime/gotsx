import type { PagedCards } from "host:catalog";
import ProductCard from "../ui/ProductCard.server";
import Pager from "../ui/Pager.server";

interface SortTab {
  k: string;
  label: string;
}

const sorts: SortTab[] = [
  { k: "rec", label: "综合" },
  { k: "sales", label: "销量" },
  { k: "price", label: "价格 ↑" },
  { k: "priceDesc", label: "价格 ↓" },
];

export default function Listing({ title, base, sort, page, wished }: { title: string; base: string; sort: string; page: PagedCards; wished: string[] }) {
  return (
    <div>
      <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-baseline gap-2">
          <h1 class="text-xl font-extrabold tracking-tight">{title}</h1>
          <span class="text-sm text-ink-400">{page.total} 件商品</span>
        </div>
        <div class="flex items-center gap-1 rounded-full border border-ink-200 bg-white p-1 text-[13px]">
          {sorts.map((t) => (
            <a href={`${base}sort=${t.k}`} class={t.k === sort ? "rounded-full bg-brand-500 px-3.5 py-1.5 font-semibold text-white" : "rounded-full px-3.5 py-1.5 text-ink-600 transition hover:bg-ink-100"}>{t.label}</a>
          ))}
        </div>
      </div>
      {page.total === 0 && (
        <div class="rounded-xl2 border border-ink-100 bg-white py-24 text-center">
          <div class="text-6xl">🔍</div>
          <p class="mt-4 text-ink-500">没有找到相关商品</p>
          <p class="mt-1 text-sm text-ink-400">换个关键词, 或看看别的分类</p>
          <a href="/" class="mt-5 inline-block rounded-full bg-brand-500 px-6 py-2.5 text-sm font-bold text-white transition hover:bg-brand-600">回首页逛逛</a>
        </div>
      )}
      <div class="grid grid-cols-2 gap-3 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
        {page.items.map((c) => <ProductCard card={c} wished={wished.includes(c.id)} />)}
      </div>
      {page.pages > 1 && <Pager base={`${base}sort=${sort}&`} page={page.page} list={page.pageList} />}
    </div>
  );
}
