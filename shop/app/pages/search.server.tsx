import type { PageProps, Meta } from "gotsx";
import { listCards } from "host:catalog";
import { list as wishList } from "host:wish";
import Layout from "../components/Layout.server";
import Listing from "../components/Listing.server";

export function meta(props: PageProps): Meta {
  const lc = props.locale !== "" ? props.locale : "en";
  return { title: tv(lc, "search.results", { q: props.query.q ?? "" }), noIndex: true };
}

export default function SearchPage({ query, cookies, locale, path }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  const sid = cookies.sid ?? "";
  const q = query.q ?? "";
  const sort = query.sort ?? "rec";
  const page = listCards("", q, sort, Number(query.p ?? "1"));
  const title = tv(lc, "search.results", { q });
  return (
    <Layout sid={sid} locale={lc} q={q} wide path={path}>
      <Listing title={title} base={`/search?q=${encodeURIComponent(q)}&`} sort={sort} page={page} wished={wishList(sid)} locale={lc} />
    </Layout>
  );
}
