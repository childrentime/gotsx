import type { PageProps } from "gotsx";
import { listCards, catLabel } from "host:catalog";
import { list as wishList } from "host:wish";
import Layout from "../../components/Layout.server";
import Listing from "../../components/Listing.server";

export default function CategoryPage({ params, query, cookies, locale }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  const sid = cookies.sid ?? "";
  const sort = query.sort ?? "rec";
  const page = listCards(params.cat, "", sort, Number(query.p ?? "1"));
  const label = catLabel(params.cat);
  return (
    <Layout title={label} sid={sid} locale={lc} active={params.cat} wide path={`/c/${params.cat}`} desc={tv(lc, "category.meta", { label, n: String(page.total) })}>
      <Listing title={label} base={`/c/${params.cat}?`} sort={sort} page={page} wished={wishList(sid)} locale={lc} />
    </Layout>
  );
}
