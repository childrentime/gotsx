import type { PageProps } from "gotsx";
import { view } from "host:cart";
import Layout from "../components/Layout.server";
import CartPage from "../islands/CartPage.client";

export default function Cart({ cookies, locale }: PageProps) {
  const sid = cookies.sid ?? "";
  return (
    <Layout title="购物车" sid={sid} locale={locale} wide>
      <h1 class="mb-4 text-xl font-extrabold tracking-tight">购物车</h1>
      <CartPage cart={view(sid)} />
    </Layout>
  );
}
