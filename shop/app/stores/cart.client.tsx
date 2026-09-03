import { createStore } from "gotsx";
import type { CartView } from "host:cart";

// The cart every island shares: the document layout seeds it once per request (seed(cart, view(sid)) in
// pages/_layout.server.tsx), so the badge, the cart page and the checkout form render the same value on the
// server and start from it in the browser. Actions return a fresh CartView; islands hand it to cart.set(...)
// and every subscriber updates — no event bus, no copies.
export const cart = createStore<CartView>({ items: [], empty: true });
