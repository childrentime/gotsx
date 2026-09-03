import type { PageProps, Meta } from "gotsx";
import Layout from "../components/Layout.server";
import CartPage from "../islands/CartPage.client";

export function meta(props: PageProps): Meta {
  const lc = props.locale !== "" ? props.locale : "en";
  return { title: t(lc, "cart.title"), noIndex: true };
}

export default function Cart({ locale, path }: PageProps) {
  const lc = locale !== "" ? locale : "en";
  return (
    <Layout locale={lc} wide path={path}>
      <h1 class="mb-5 text-xl font-semibold tracking-tight">{t(lc, "cart.title")}</h1>
      <CartPage locale={lc} />
    </Layout>
  );
}
