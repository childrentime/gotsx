import type { PageProps } from "gotsx";
import { listCards } from "host:catalog";
import { list as wishList } from "host:wish";
import Layout from "../components/Layout.server";
import Listing from "../components/Listing.server";

export default function SearchPage({ query, cookies, locale }: PageProps) {
  const sid = cookies.sid ?? "";
  const q = query.q ?? "";
  const sort = query.sort ?? "rec";
  const page = listCards("", q, sort, Number(query.p ?? "1"));
  return (
    <Layout title={`“${q}” 的搜索结果`} sid={sid} locale={locale} q={q} wide>
      <Listing title={`“${q}” 的搜索结果`} base={`/search?q=${encodeURIComponent(q)}&`} sort={sort} page={page} wished={wishList(sid)} />
    </Layout>
  );
}
