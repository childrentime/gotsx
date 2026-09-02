import type { PageProps } from "gotsx";
import { view } from "host:cart";
import Layout from "../components/Layout.server";
import CartPage from "../islands/CartPage.client";

export default function Cart({ cookies, locale }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  const sid = cookies.sid ?? "";
  return (
    <Layout title={t(lc, "cart.title")} sid={sid} locale={lc} wide>
      <h1 class="mb-5 text-xl font-semibold tracking-tight">{t(lc, "cart.title")}</h1>
      <CartPage cart={view(sid)} locale={lc} />
    </Layout>
  );
}
