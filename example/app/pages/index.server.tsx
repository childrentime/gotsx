import type { PageProps } from "gotsx";
import { models } from "host:data";        // Go 实现: 编译后是直接的 Go 调用, 零编组
import { fmtNumber } from "host:intl";
import Layout from "../components/Layout.server";
import ModelCard from "../components/ModelCard.server";

export default function Home({ query }: PageProps) {
  const q = query.q ?? "";
  const list = models.search(q);           // 同步: 并发由 goroutine 提供, 语言里没有 async
  return (
    <Layout title="模型">
      <form method="get" action="/" class="search">
        <input name="q" value={q} placeholder="搜索模型…(回车提交, 走 SPA 跳转)" />
      </form>
      <p class="muted">
        {fmtNumber(list.length)} 个模型{q !== "" && <span> · 匹配 “{q}”</span>}
      </p>
      <div class="grid">{list.map((m) => <ModelCard model={m} />)}</div>
      {list.length === 0 && <p class="empty">没有匹配的模型</p>}
    </Layout>
  );
}
