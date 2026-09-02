/* gotsx 岛加载器 + SPA 跳转。页面本身零 JS; 岛按需 import(); 跳转 = 拉 HTML → idiomorph → 岛按 DOM 同一性存活 */
import { mount } from "./runtime.js";
/* idiomorph(~11 KB gz)只在第一次 SPA 跳转时才需要: 懒加载, 首屏 JS 只有 runtime + loader; 悬停预取时提前拉 */
let morphMod = null;
const loadMorph = () => morphMod || (morphMod = import("./idiomorph.esm.js").then((m) => m.Idiomorph));

const manifest = () => (window.__GOTSX && window.__GOTSX.islands) || {};

/* i18n 前缀模式: 把内链自动补上当前语言前缀(默认语言不补), 保持同语言导航。base-aware(静态导出到子路径时) */
function localize(url) {
  const g = window.__GOTSX || {};
  const i = g.i18n || {};
  if (!i.prefix || !i.locale || i.locale === i.default) return url;
  let u;
  try { u = new URL(url, location.origin); } catch { return url; }
  if (u.origin !== location.origin) return url;
  const base = g.base || "";
  let p = u.pathname;
  const underBase = base !== "" && (p === base || p.startsWith(base + "/"));
  if (underBase) p = p.slice(base.length) || "/";
  if (p === "/" + i.locale || p.startsWith("/" + i.locale + "/")) return url;   // 已带前缀
  for (const skip of ["/_gotsx", "/api", "/actions", "/public", "/img", "/robots", "/sitemap", "/manifest", "/icon", "/healthz", "/readyz"])
    if (p.startsWith(skip)) return url;
  p = "/" + i.locale + (p === "/" ? "" : p);
  u.pathname = (underBase ? base : "") + p;
  return u.href;
}
const mods = new Map();
function load(name) {
  if (!mods.has(name)) mods.set(name, import(manifest()[name]).then((m) => m.default));
  return mods.get(name);
}

async function mountIsland(el) {
  const name = el.getAttribute("name");
  const raw = el.getAttribute("props") || "{}";
  // 错误边界: 单个岛失败不应白屏, 也不应影响其它岛
  try {
    const Comp = await load(name);
    const props = JSON.parse(raw);
    const fresh = !el.__gotsx;
    try {
      mount(Comp, props, el, fresh);          // 第一次 hydrate; props 变了则重建
    } catch (e) {
      console.error("[gotsx] 岛 " + name + " hydrate 失败, 回退到客户端渲染:", e);
      mount(Comp, props, el, false);          // hydrate 不匹配的兜底: 从头渲染
    }
    el.__gotsx = { name, props: raw };
  } catch (e) {
    console.error("[gotsx] 岛 " + name + " 挂载失败, 保留服务端内容:", e);
    el.__gotsx = { name, props: raw, failed: true };  // 保留 SSR HTML, 标记失败不再重试
  }
}

function reconcile() {
  document.querySelectorAll("gotsx-island").forEach((el) => {
    const s = el.__gotsx;
    if (!s || s.name !== el.getAttribute("name") || s.props !== (el.getAttribute("props") || "{}")) mountIsland(el);
  });
}

/* ---------- SPA 跳转 ---------- */
function mergeHead(doc) {
  document.title = doc.title;
  const m = doc.querySelector('meta[name="gotsx-render-us"]');
  const mine = document.querySelector('meta[name="gotsx-render-us"]');
  if (m && mine) mine.content = m.content;
}
function keepIslands(oldNode, newNode) {
  if (oldNode.nodeType !== 1 || oldNode.tagName !== "GOTSX-ISLAND" || !oldNode.__gotsx || newNode.nodeType !== 1) return;
  oldNode.setAttribute("name", newNode.getAttribute("name"));
  oldNode.setAttribute("props", newNode.getAttribute("props") || "{}");
  return false;     // 岛的子树归运行时管, morph 不碰
}

/* 顶部导航进度条 */
const bar = () => document.getElementById("gotsx-bar");
let barTimer = null;
function barStart() {
  const b = bar(); if (!b) return;
  clearTimeout(barTimer); b.style.transition = "none"; b.style.width = "0"; b.style.opacity = "1";
  requestAnimationFrame(() => { b.style.transition = ""; b.style.width = "75%"; });
}
function barDone() {
  const b = bar(); if (!b) return;
  b.style.width = "100%";
  barTimer = setTimeout(() => { b.style.opacity = "0"; b.style.width = "0"; }, 250);
}

/* 预取: hover / touchstart 时把目标页 HTML 拉进缓存, 点击即秒开(Core Web Vitals) */
const prefetchCache = new Map();
function prefetch(url) {
  if (prefetchCache.has(url)) return;
  const pr = fetch(url, { headers: { "X-Gotsx-Nav": "1", "Purpose": "prefetch" } })
    .then(async (r) => ({ ok: r.ok && (r.headers.get("content-type") || "").includes("text/html"), html: await r.text() }))
    .catch(() => ({ ok: false }));
  prefetchCache.set(url, pr);
  if (prefetchCache.size > 40) prefetchCache.delete(prefetchCache.keys().next().value); // 上限, 防内存膨胀
}
function prefetchable(a) {
  return a && a.origin === location.origin && !a.target && !a.hasAttribute("download") &&
    a.getAttribute("href") && !a.href.includes("#") && a.dataset.noPrefetch === undefined && a.href !== location.href;
}
let hoverTimer;
document.addEventListener("pointerover", (e) => {
  const a = e.target.closest && e.target.closest("a[href]");
  if (!prefetchable(a)) return;
  clearTimeout(hoverTimer);
  loadMorph();
  hoverTimer = setTimeout(() => prefetch(a.dataset.raw ? a.href : localize(a.href)), 60);   // 悬停 60ms 才预取
});
document.addEventListener("pointerout", () => clearTimeout(hoverTimer));
document.addEventListener("touchstart", (e) => {
  const a = e.target.closest && e.target.closest("a[href]");
  if (prefetchable(a)) prefetch(a.href);
}, { passive: true });

let navSeq = 0, navAbort = null;
async function navigate(url, push = true) {
  const seq = ++navSeq;
  if (navAbort) navAbort.abort();
  navAbort = new AbortController();
  barStart();
  const morphReady = loadMorph();
  let html;
  const cached = prefetchCache.get(url);
  try {
    if (cached) {                       // 命中预取: 秒开
      const res = await cached;
      prefetchCache.delete(url);
      if (!res.ok) { barDone(); location.href = url; return; }
      html = res.html;
    } else {
      const res = await fetch(url, { headers: { "X-Gotsx-Nav": "1" }, signal: navAbort.signal });
      if (!res.ok || !(res.headers.get("content-type") || "").includes("text/html")) { barDone(); location.href = url; return; }
      html = await res.text();
    }
  } catch (e) { if (e.name !== "AbortError") { barDone(); location.href = url; } return; }
  if (seq !== navSeq) return;
  let Idiomorph;
  try { Idiomorph = await morphReady; } catch { barDone(); location.href = url; return; }
  if (seq !== navSeq) return;
  barDone();
  const doc = new DOMParser().parseFromString(html, "text/html");
  if (push) history.pushState({ gotsx: true }, "", url);
  const swap = () => {
    mergeHead(doc);
    Idiomorph.morph(document.body, doc.body, { morphStyle: "outerHTML", callbacks: { beforeNodeMorphed: keepIslands } });
    reconcile();
  };
  const applyFills = () => document.querySelectorAll("template[data-gotsx-fill]").forEach((t) => window.__gotsxFill && window.__gotsxFill(t.getAttribute("data-gotsx-fill")));
  if (document.startViewTransition) {
    const vt = document.startViewTransition(swap);
    vt.ready.catch(() => {});                       // 过渡被跳过(如紧接着又一次跳转)不算错
    try { await vt.finished; } catch { /* 同上 */ }
  } else swap();
  applyFills();                                  // 流式 Suspense 的填充随 HTML 一起到达, 跳转后手动应用
  if (push) window.scrollTo(0, 0);
  document.dispatchEvent(new CustomEvent("gotsx:navigated", { detail: { url } }));
}

document.addEventListener("click", (e) => {
  const a = e.target.closest && e.target.closest("a[href]");
  if (!a || e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
  if (a.origin !== location.origin || a.target || a.hasAttribute("download")) return;
  e.preventDefault();
  navigate(a.dataset.raw ? a.href : localize(a.href));
});
document.addEventListener("submit", (e) => {
  const f = e.target;
  if ((f.method || "get").toLowerCase() !== "get" || f.target) return;
  e.preventDefault();
  const u = new URL(f.action || location.href);
  u.search = new URLSearchParams(new FormData(f)).toString();
  navigate(u.href);
});
addEventListener("popstate", () => navigate(location.href, false));

/* 客户端遥测(仅当服务端开启 OnClientEvent 时): JS 错误 / 未处理 rejection / 页面浏览 */
const logURL = () => window.__GOTSX && window.__GOTSX.log;
let sent = 0;
function report(ev) {
  const url = logURL();
  if (!url || sent > 50) return;                  // 上限, 防风暴
  sent++;
  const body = JSON.stringify(ev);
  if (navigator.sendBeacon) navigator.sendBeacon(url, new Blob([body], { type: "application/json" }));
  else fetch(url, { method: "POST", headers: { "Content-Type": "application/json" }, body, keepalive: true }).catch(() => {});
}
addEventListener("error", (e) => report({ type: "error", message: String(e.message || e.error), stack: e.error && e.error.stack ? String(e.error.stack).slice(0, 2000) : "", url: location.href }));
addEventListener("unhandledrejection", (e) => report({ type: "rejection", message: String(e.reason && e.reason.message || e.reason), stack: e.reason && e.reason.stack ? String(e.reason.stack).slice(0, 2000) : "", url: location.href }));
if (logURL()) report({ type: "pageview", url: location.href, ref: document.referrer });
document.addEventListener("gotsx:navigated", (e) => report({ type: "pageview", url: e.detail.url }));

/* dev 模式: 监听 /_gotsx/dev 的 SSE。gotsx dev 重启进程后 bootID 变了 → 整页刷新(编译失败时旧进程还在, id 不变, 不刷新) */
if (window.__GOTSX && window.__GOTSX.dev && typeof EventSource !== "undefined") {
  let boot = null;
  const es = new EventSource("/_gotsx/dev");
  es.onmessage = (e) => {
    if (boot === null) boot = e.data;
    else if (e.data !== boot) { es.close(); location.reload(); }
  };
  let overlay = null;
  /* Compile-error overlay: gotsx dev writes .gotsx/diagnostics.json → the server sends an SSE `diag` event → drawn here; an empty object clears it */
  es.addEventListener("diag", (e) => {
    let d = {};
    try { d = JSON.parse(e.data); } catch { /* ignore */ }
    const old = document.getElementById("gotsx-overlay");
    if (!d.errors || !d.errors.length) { if (old) old.remove(); return; }
    const box = old || document.createElement("div");
    box.id = "gotsx-overlay";
    box.setAttribute("style", "position:fixed;inset:0;z-index:2147483647;background:rgba(9,9,11,.94);color:#fafafa;font:13px/1.5 ui-monospace,SFMono-Regular,Menlo,monospace;padding:32px;overflow:auto;box-sizing:border-box");
    const esc = (s) => String(s).replace(/[&<>]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;" }[c]));
    box.innerHTML = '<div style="max-width:960px;margin:0 auto"><div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px"><strong style="font-size:15px;color:#f87171">' + esc(d.title || "gotsx: build failed") + '</strong><button id="gotsx-overlay-x" style="background:none;border:1px solid #3f3f46;color:#a1a1aa;border-radius:6px;padding:4px 10px;cursor:pointer">dismiss</button></div>' +
      d.errors.map((x) => '<div style="margin:0 0 14px;padding:12px 14px;background:#18181b;border:1px solid #27272a;border-radius:8px"><div style="color:#a1a1aa;margin-bottom:4px">' + esc(x.file ? x.file + ":" + x.line + ":" + x.col : "") + '</div><div style="white-space:pre-wrap">' + esc(x.msg) + "</div></div>").join("") +
      '<div style="color:#71717a;margin-top:8px">The previous build keeps serving; this overlay updates when the source is fixed.</div></div>';
    if (!old) document.body.appendChild(box);
    box.querySelector("#gotsx-overlay-x").onclick = () => { box.remove(); overlay = null; };
    overlay = box;
  });
  /* SPA navigation morphs <body>: put the overlay back if a build is still failing */
  document.addEventListener("gotsx:navigated", () => { if (overlay && !overlay.isConnected) document.body.appendChild(overlay); });
  /* dev: console.error / console.warn also show up in the gotsx dev terminal */
  for (const level of ["error", "warn"]) {
    const orig = console[level].bind(console);
    console[level] = (...a) => { orig(...a); try { report({ type: "console." + level, message: a.map((x) => (x instanceof Error ? x.message : typeof x === "string" ? x : JSON.stringify(x))).join(" ").slice(0, 2000), url: location.href }); } catch { /* ignore */ } };
  }
}

reconcile();
window.__gotsxNavigate = navigate;
window.__gotsxReconcile = reconcile;
