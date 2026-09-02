import { readFileSync } from "node:fs";
import { join } from "node:path";
import Counter from "./Counter";
export const dynamic = "force-dynamic"; // render on every request, like the other contenders
type Product = { id: string; title: string; brand: string; desc: string; price: number; tags: string[]; rating: number };
const items: Product[] = JSON.parse(readFileSync(join(process.cwd(), "products.json"), "utf8"));
const money = (c: number) => `$${Math.floor(c / 100)}.${String(c % 100).padStart(2, "0")}`;
export default function Home() {
  return (
    <>
      <header><strong>bench · next</strong><Counter start={0} /></header>
      <main>
        <h1>Products</h1>
        <p className="muted">{items.length} products</p>
        <div className="grid">
          {items.map((p) => (
            <article className="card" key={p.id}>
              <h2>{p.title}</h2>
              <p className="muted">{p.brand} · {p.desc}</p>
              <div className="row">
                {p.tags.map((t) => <span className="chip" key={t}>{t}</span>)}
                <span className="price">{money(p.price)}</span>
              </div>
            </article>
          ))}
        </div>
      </main>
      <footer className="muted">rendered by next.js</footer>
    </>
  );
}
