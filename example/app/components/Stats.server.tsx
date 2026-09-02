import { stats, trending } from "host:data";
import { fmtNumber } from "host:intl";

/** 放在 <Suspense> 里: 这里的宿主调用在外壳发出后才执行, 两个慢查询各自在 goroutine 里跑 */
export default function Stats() {
  const s = stats();
  const top = trending();
  return (
    <div class="card grid grid-3">
      <div class="stat"><span class="muted">Models</span><span class="value">{fmtNumber(s.total)}</span></div>
      <div class="stat"><span class="muted">Likes</span><span class="value">{fmtNumber(s.likes)}</span></div>
      <div class="stat"><span class="muted">Trending</span><span class="value" style="font-size:15px">{top.map((m) => m.title).join(" · ")}</span></div>
    </div>
  );
}
