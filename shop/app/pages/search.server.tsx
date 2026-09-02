import type { PageProps } from "gotsx";
import { listCards } from "host:catalog";
import { list as wishList } from "host:wish";
import Layout from "../components/Layout.server";
import Listing from "../components/Listing.server";

export default function SearchPage({ query, cookies, locale }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  const sid = cookies.sid ?? "";
  const q = query.q ?? "";
  const sort = query.sort ?? "rec";
  const page = listCards("", q, sort, Number(query.p ?? "1"));
  const title = tv(lc, "search.results", { q });
  return (
    <Layout title={title} sid={sid} locale={lc} q={q} wide>
      <Listing title={title} base={`/search?q=${encodeURIComponent(q)}&`} sort={sort} page={page} wished={wishList(sid)} locale={lc} />
    </Layout>
  );
}
