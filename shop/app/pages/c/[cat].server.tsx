import type { PageProps, Meta } from "gotsx";
import { listCards, catLabel, count } from "host:catalog";
import { list as wishList } from "host:wish";
import Layout from "../../components/Layout.server";
import Listing from "../../components/Listing.server";

export function meta(props: PageProps): Meta {
  const lc = props.locale !== "" ? props.locale : "en";
  const label = catLabel(props.params.cat);
  return { title: label, description: tv(lc, "category.meta", { label, n: String(count(props.params.cat)) }) };
}

export default function CategoryPage({ params, query, cookies, locale, path }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  const sid = cookies.sid ?? "";
  const sort = query.sort ?? "rec";
  const page = listCards(params.cat, "", sort, Number(query.p ?? "1"));
  const label = catLabel(params.cat);
  return (
    <Layout sid={sid} locale={lc} active={params.cat} wide path={path}>
      <Listing title={label} base={`/c/${params.cat}?`} sort={sort} page={page} wished={wishList(sid)} locale={lc} />
    </Layout>
  );
}
