// hono + bun: JSX rendered on the server, no hydration story (the counter is an inline script).
import { Hono } from "hono";
import { html, raw } from "hono/html";
import { readFileSync } from "node:fs";
type Product = { id: string; title: string; brand: string; desc: string; price: number; tags: string[]; rating: number };
const items: Product[] = JSON.parse(readFileSync(new URL("./products.json", import.meta.url), "utf8"));
const css = readFileSync(new URL("./shared.css", import.meta.url), "utf8");
const money = (c: number) => `$${Math.floor(c / 100)}.${String(c % 100).padStart(2, "0")}`;
const app = new Hono();
app.get("/healthz", (c) => c.text("ok"));
app.get("/", (c) =>
  c.html(
    <html lang="en">
      <head><meta charset="utf-8" /><meta name="viewport" content="width=device-width, initial-scale=1" /><title>Bench · hono</title><style>{raw(css)}</style></head>
      <body>
        <header><strong>bench · hono</strong><button id="c">count: <span id="n">0</span></button></header>
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
        <footer class="muted">rendered by hono (bun)</footer>
        {html`<script>var n=0,el=document.getElementById('n');document.getElementById('c').onclick=function(){el.textContent=++n}</script>`}
      </body>
    </html>,
  ),
);
export default { port: Number(process.env.PORT || 3006), fetch: app.fetch };
