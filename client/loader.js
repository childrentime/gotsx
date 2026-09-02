/* gotsx 岛加载器 + SPA 跳转。页面本身零 JS; 岛按需 import(); 跳转 = 拉 HTML → idiomorph → 岛按 DOM 同一性存活 */
import { mount } from "./runtime.js";
import { Idiomorph } from "./idiomorph.esm.js";

const manifest = () => (window.__GOTSX && window.__GOTSX.islands) || {};

/* i18n 前缀模式: 把内链自动补上当前语言前缀(默认语言不补), 保持同语言导航 */
function localize(url) {
  const i = (window.__GOTSX && window.__GOTSX.i18n) || {};
  if (!i.prefix || !i.locale || i.locale === i.default) return url;
  let u;
  try { u = new URL(url, location.origin); } catch { return url; }
  if (u.origin !== location.origin) return url;
  const p = u.pathname;
  if (p === "/" + i.locale || p.startsWith("/" + i.locale + "/")) return url;   // 已带前缀
  for (const skip of ["/_gotsx", "/api", "/actions", "/public", "/img", "/robots", "/sitemap", "/manifest", "/icon", "/healthz", "/readyz"])
    if (p.startsWith(skip)) return url;
  u.pathname = "/" + i.locale + (p === "/" ? "" : p);
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
  barDone();
  const doc = new DOMParser().parseFromString(html, "text/html");
  if (push) history.pushState({ gotsx: true }, "", url);
  const swap = () => {
    mergeHead(doc);
    Idiomorph.morph(document.body, doc.body, { morphStyle: "outerHTML", callbacks: { beforeNodeMorphed: keepIslands } });
    reconcile();
  };
  if (document.startViewTransition) {
    const vt = document.startViewTransition(swap);
    vt.ready.catch(() => {});                       // 过渡被跳过(如紧接着又一次跳转)不算错
    try { await vt.finished; } catch { /* 同上 */ }
  } else swap();
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

reconcile();
window.__gotsxNavigate = navigate;
