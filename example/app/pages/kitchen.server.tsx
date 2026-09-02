import type { PageProps } from "gotsx";
import { models } from "host:data";
import Todos from "../islands/Todos.client";

/** 语言"厨房水槽": 0.5 新增的循环 / switch / 原地修改 / 可选对象真值 / redirect / notFound 全在这一页, 也是集成测试 */

interface Named {
  name: string;
}
interface Scored extends Named {
  score: number;
}

function grade(score: number): string {
  switch (true) {
    case score >= 90:
      return "A";
    case score >= 60:
      return "B";
    default:
      return "C";
  }
}

function label(kind: string): string {
  let out = "";
  switch (kind) {
    case "a":
    case "b":
      out = "ab";
      break;
    case "c":
      out = "c";
    default:
      out += "!";
  }
  return out;
}

export default function Kitchen({ query }: PageProps) {
  if (query.go !== "") return redirect(query.go);
  if (query.missing !== "") notFound();

  // 经典循环 + break / continue
  let total = 0;
  for (let i = 1; i <= 10; i++) {
    if (i % 2 === 0) continue;
    total += i;
  }
  let n = 0;
  while (n < 5) {
    n++;
    if (n === 3) break;
  }

  // 原地修改的数组方法
  let names: string[] = [];
  for (const m of models.list()) names.push(m.title);
  names.splice(0, 1);
  const first = names.shift() ?? "";
  names.unshift(first);
  const sorted = names.sort((a, b) => a.localeCompare(b));
  const idx = sorted.findIndex((s) => s === first);

  // 可选对象: find 没找到就是"undefined"(Go 零值)
  const people: Scored[] = [{ name: "ann", score: 91 }, { name: "bob", score: 42 }];
  const found = people.find((p) => p.name === "bob");
  const missing = people.find((p) => p.name === "zed");

  let counter = 7;
  counter %= 4;
  counter++;
  const arr = [1, 2, 3];
  arr[0] = 10;

  // 正则(RE2 子集, 编译期校验)与 Date
  const titles = models.list().map((m) => m.title).join(" | ");
  const nums = titles.match(/\d+/g);
  const slug = first.toLowerCase().replace(/[^a-z0-9]+/g, "-");
  const hasTPU = /tpu/i.test(titles);
  const year = isoDate(Date.parse("2026-08-02T10:00:00Z")).slice(0, 4);
  const fresh = Date.now() > 0;

  // Record: 缺席的键读出零值, Object.hasOwn 判断存在
  const counts: Record<string, number> = { a: 1 };
  counts.b++;
  delete counts.a;
  const hasA = Object.hasOwn(counts, "a");

  return (
    <div class="stack">
      <h1>Language kitchen sink</h1>
      <ul class="kitchen">
        <li>odd sum 1..10 = {total}</li>
        <li>while / break → {n}</li>
        <li>sorted: {sorted.join(", ")} (first at {idx})</li>
        <li>found: {found ? found.name : "-"} · missing: {missing ? "yes" : "no"} · {missing === undefined ? "undefined" : "defined"}</li>
        <li>grades: {people.map((p) => `${p.name}=${grade(p.score)}`).join(" ")} · labels: {label("a")} {label("c")} {label("z")}</li>
        <li>counter {counter.toString()} · arr[0] {arr[0]}</li>
        <li>regex: {nums.length} numbers in titles · slug "{slug}" · tpu? {hasTPU ? "yes" : "no"}</li>
        <li>date: parsed year {year} · now {fresh ? "ok" : "?"}</li>
        <li>record: b={counts.b} · hasOwn a? {hasA ? "yes" : "no"} · a reads as {counts.a}</li>
      </ul>
      <p class="muted">Try <a class="link" href="/kitchen?go=/">?go=/</a> (redirect) and <a class="link" href="/kitchen?missing=1">?missing=1</a> (404), or <a class="link" href="/docs/a/b/c">/docs/a/b/c</a> (catch-all route).</p>
      <Todos initial={sorted} />
    </div>
  );
}
