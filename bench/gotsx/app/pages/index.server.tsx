import type { PageProps } from "gotsx";
import { products, money } from "host:data";
import Counter from "../islands/Counter.client";

export default function Home({ query }: PageProps) {
  const items = products();
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <title>Bench · gotsx</title>
        <link rel="stylesheet" href="/public/shared.css" />
      </head>
      <body>
        <header><strong>bench · gotsx</strong><Counter start={0} /></header>
        <main>
          <h1>Products</h1>
          <p class="muted">{items.length} products</p>
          <div class="grid">
            {items.map((p) => (
              <article class="card">
                <h2>{p.title}</h2>
                <p class="muted">{p.brand} · {p.desc}</p>
                <div class="row">
                  {p.tags.map((t) => <span class="chip">{t}</span>)}
                  <span class="price">{money(p.price)}</span>
                </div>
              </article>
            ))}
          </div>
        </main>
        <footer class="muted">rendered by gotsx</footer>
      </body>
    </html>
  );
}
