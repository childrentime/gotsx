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

/* ---------- SPA navigation ----------
   Full-page HTML morph (the Turbo / htmx-boost family): fetch the target document, morph <body> with idiomorph,
   islands survive by DOM identity. On top of that: a snapshot cache so back/forward is instant, scroll restoration,
   native same-page anchors, full <head> metadata merge, focus management + a route announcer, and streaming —
   the shell is morphed as soon as </html> arrives while Suspense fills keep streaming in. */
const stripHash = (u) => u.split("#")[0];

function mergeHead(doc) {
  document.title = doc.title;
  const head = document.head, fresh = doc.head;
  // per-page metadata is replaced wholesale; the viewport / charset metas and the framework's own bootstrap stay
  const sel = 'meta[name]:not([name="viewport"]), meta[property], link[rel="canonical"], link[rel="alternate"], script[type="application/ld+json"]';
  head.querySelectorAll(sel).forEach((n) => n.remove());
  fresh.querySelectorAll(sel).forEach((n) => head.appendChild(n.cloneNode(true)));
  // stylesheets the new page needs that this document lacks (old ones stay: no flash of unstyled content)
  const have = new Set(Array.from(head.querySelectorAll('link[rel="stylesheet"][href]'), (l) => l.href));
  fresh.querySelectorAll('link[rel="stylesheet"][href]').forEach((l) => { if (!have.has(l.href)) head.appendChild(l.cloneNode(true)); });
}
function keepIslands(oldNode, newNode) {
  if (oldNode.nodeType !== 1 || oldNode.tagName !== "GOTSX-ISLAND" || !oldNode.__gotsx || newNode.nodeType !== 1) return;
  oldNode.setAttribute("name", newNode.getAttribute("name"));
  oldNode.setAttribute("props", newNode.getAttribute("props") || "{}");
  return false;     // the island's subtree belongs to the runtime; morph leaves it alone
}

/* Top navigation progress bar */
const bar = () => document.getElementById("gotsx-bar");
let barTimer = null, barOn = false;
function barStart() {
  const b = bar(); if (!b) return;
  barOn = true;
  clearTimeout(barTimer); b.style.transition = "none"; b.style.width = "0"; b.style.opacity = "1";
  requestAnimationFrame(() => { b.style.transition = ""; b.style.width = "75%"; });
}
function barDone() {
  const b = bar(); if (!b || !barOn) return;
  barOn = false;
  b.style.width = "100%";
  barTimer = setTimeout(() => { b.style.opacity = "0"; b.style.width = "0"; }, 250);
}

/* Accessibility: move focus to the main landmark after a navigation and announce the new title to screen readers */
let announcer = null;
function announce(text) {
  if (!announcer) {
    announcer = document.createElement("div");
    announcer.id = "gotsx-announcer";
    announcer.setAttribute("role", "status");
    announcer.setAttribute("aria-live", "assertive");
    announcer.setAttribute("aria-atomic", "true");
    announcer.setAttribute("style", "position:absolute;width:1px;height:1px;margin:-1px;padding:0;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;border:0");
  }
  if (!announcer.isConnected) document.body.appendChild(announcer);
  announcer.textContent = "";
  setTimeout(() => { announcer.textContent = text; }, 0);
}
function focusMain() {
  const el = document.querySelector("[data-gotsx-focus]") || document.querySelector("main") || document.querySelector("h1");
  if (!el) return;
  if (!el.hasAttribute("tabindex")) el.setAttribute("tabindex", "-1");
  try { el.focus({ preventScroll: true }); } catch { /* not focusable */ }
}

/* Prefetch: on hover / touchstart pull the target HTML into a cache so the click is instant (Core Web Vitals) */
const prefetchCache = new Map();
function prefetch(url) {
  if (prefetchCache.has(url)) return;
  const pr = fetch(url, { headers: { "X-Gotsx-Nav": "1", "Purpose": "prefetch" } })
    .then(async (r) => ({ ok: r.ok && (r.headers.get("content-type") || "").includes("text/html"), html: await r.text() }))
    .catch(() => ({ ok: false }));
  prefetchCache.set(url, pr);
  if (prefetchCache.size > 40) prefetchCache.delete(prefetchCache.keys().next().value); // bounded
}
function prefetchable(a) {
  return a && a.origin === location.origin && !a.target && !a.hasAttribute("download") &&
    a.getAttribute("href") && a.dataset.noPrefetch === undefined && stripHash(a.href) !== stripHash(location.href);
}
let hoverTimer;
document.addEventListener("pointerover", (e) => {
  const a = e.target.closest && e.target.closest("a[href]");
  if (!prefetchable(a)) return;
  clearTimeout(hoverTimer);
  loadMorph();
  hoverTimer = setTimeout(() => prefetch(stripHash(a.dataset.raw ? a.href : localize(a.href))), 60);   // 60 ms: a hover, not a pass-over
});
document.addEventListener("pointerout", () => clearTimeout(hoverTimer));
document.addEventListener("touchstart", (e) => {
  const a = e.target.closest && e.target.closest("a[href]");
  if (prefetchable(a)) prefetch(stripHash(a.href));
}, { passive: true });

/* Snapshot cache + scroll restoration: back/forward morph from the last HTML seen (instant, no network), restore the
   scroll position saved on the history entry, then revalidate against the server in the background */
const snapshots = new Map();
function remember(url, html) {
  snapshots.delete(url);
  snapshots.set(url, html);
  if (snapshots.size > 10) snapshots.delete(snapshots.keys().next().value);
}
function saveScroll() {
  try { history.replaceState(Object.assign({}, history.state || {}, { gotsx: true, scroll: [window.scrollX, window.scrollY] }), ""); } catch { /* ignore */ }
}
if ("scrollRestoration" in history) history.scrollRestoration = "manual";
const jump = (x, y) => window.scrollTo({ left: x, top: y, behavior: "instant" });   // restoration is never animated, whatever scroll-behavior the page sets (older engines animate)
// with manual restoration the browser no longer restores scroll on reload / full-page back: keep the position in
// sessionStorage (it survives reloads and same-tab traversals) and re-apply it when the same URL loads again
const scrollKey = () => "gotsx:scroll:" + location.href;
addEventListener("pagehide", () => { saveScroll(); try { sessionStorage.setItem(scrollKey(), JSON.stringify([window.scrollX, window.scrollY])); } catch { /* storage off */ } });
try {
  const saved = sessionStorage.getItem(scrollKey());
  if (saved) { sessionStorage.removeItem(scrollKey()); const [x, y] = JSON.parse(saved); if (y || x) jump(x, y); }
  else if (history.state && history.state.scroll) jump(history.state.scroll[0], history.state.scroll[1]);
} catch { /* ignore */ }
remember(stripHash(location.href), "<!DOCTYPE html>" + document.documentElement.outerHTML);

const applyFills = () => document.querySelectorAll("template[data-gotsx-fill]").forEach((t) => window.__gotsxFill && window.__gotsxFill(t.getAttribute("data-gotsx-fill")));
function applyChunk(chunk) {   // a streamed <template data-gotsx-fill> (the server's inline script is not needed: we call the fill ourselves)
  const d = new DOMParser().parseFromString(chunk, "text/html");
  d.querySelectorAll("template[data-gotsx-fill]").forEach((t) => {
    document.body.appendChild(document.adoptNode(t));
    if (window.__gotsxFill) window.__gotsxFill(t.getAttribute("data-gotsx-fill"));
  });
}

let navSeq = 0, navAbort = null;
async function navigate(url, push = true, restore = null) {
  const seq = ++navSeq;
  if (navAbort) navAbort.abort();
  navAbort = new AbortController();
  const key = stripHash(url);
  const hash = url.includes("#") ? url.slice(url.indexOf("#") + 1) : "";
  const morphReady = loadMorph();
  let Idiomorph = null, settled = false;
  const fail = () => { barDone(); location.href = url; };
  const swapDoc = (html) => {
    const doc = new DOMParser().parseFromString(html, "text/html");
    const swap = () => {
      mergeHead(doc);
      Idiomorph.morph(document.body, doc.body, { morphStyle: "outerHTML", callbacks: { beforeNodeMorphed: keepIslands } });
      reconcile();
    };
    if (document.startViewTransition && !settled) {
      const vt = document.startViewTransition(swap);
      vt.ready.catch(() => {});
      return vt.finished.catch(() => {});
    }
    swap();
    return Promise.resolve();
  };
  const afterSwap = () => {
    if (settled) return;
    settled = true;
    if (restore) jump(restore[0], restore[1]);
    else if (hash && document.getElementById(hash)) document.getElementById(hash).scrollIntoView();
    else if (push) jump(0, 0);
    if (push) focusMain();
    announce(document.title);
    document.dispatchEvent(new CustomEvent("gotsx:navigated", { detail: { url } }));
  };
  // 1. back/forward with a snapshot: morph it right away, then revalidate silently
  const snap = !push && snapshots.get(key);
  if (snap) {
    try { Idiomorph = await morphReady; } catch { fail(); return; }
    if (seq !== navSeq) return;
    await swapDoc(snap);
    applyFills();
    afterSwap();
  } else {
    barStart();
    if (push) { saveScroll(); history.pushState({ gotsx: true }, "", url); }
  }
  // 2. fetch: the prefetch cache, or a streamed response whose shell is morphed as soon as </html> is in
  let html = "";
  try {
    const cached = prefetchCache.get(key);
    if (cached) {
      const res = await cached;
      prefetchCache.delete(key);
      if (seq !== navSeq) return;
      if (!res.ok) { fail(); return; }
      html = res.html;
    } else {
      const res = await fetch(url, { headers: { "X-Gotsx-Nav": "1" }, signal: navAbort.signal });
      if (seq !== navSeq) return;
      if (!res.ok || !(res.headers.get("content-type") || "").includes("text/html")) { if (!snap) fail(); return; }
      if (!res.body || !res.body.getReader) html = await res.text();
      else {
        const reader = res.body.getReader();
        const dec = new TextDecoder();
        let buf = "", shell = null;
        for (;;) {
          const { value, done } = await reader.read();
          if (seq !== navSeq) { reader.cancel().catch(() => {}); return; }
          if (value) buf += dec.decode(value, { stream: true });
          if (shell === null) {
            const i = buf.indexOf("</html>");
            if (i >= 0 || done) {
              const end = i >= 0 ? i + 7 : buf.length;
              shell = buf.slice(0, end); buf = buf.slice(end); html = shell;
              if (!Idiomorph) { try { Idiomorph = await morphReady; } catch { fail(); return; } }
              if (seq !== navSeq) return;
              barDone();
              await swapDoc(shell);
              if (seq !== navSeq) return;
              applyFills();
              afterSwap();
              if (restore && settled) jump(restore[0], restore[1]);   // revalidation after a snapshot restore: keep the position
            }
          }
          if (shell !== null) {
            // each complete Suspense fill is applied as it lands; the terminator is the server's own
            // `</template><script…>__gotsxFill(…)</script>`, which escaped page content can never contain
            for (;;) {
              const j = buf.indexOf("</template><script");
              if (j < 0) break;
              const k = buf.indexOf("</script>", j);
              if (k < 0) break;
              const chunk = buf.slice(0, k + 9); buf = buf.slice(k + 9); html += chunk;
              applyChunk(chunk);
            }
          }
          if (done) break;
        }
        remember(key, html);
        return;
      }
    }
  } catch (e) { if (e.name !== "AbortError" && !snap) fail(); return; }
  if (seq !== navSeq) return;
  if (!Idiomorph) { try { Idiomorph = await morphReady; } catch { fail(); return; } }
  if (seq !== navSeq) return;
  barDone();
  await swapDoc(html);
  if (seq !== navSeq) return;
  applyFills();
  afterSwap();
  if (restore && settled) jump(restore[0], restore[1]);
  remember(key, html);
}

document.addEventListener("click", (e) => {
  const a = e.target.closest && e.target.closest("a[href]");
  if (!a || e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
  if (a.origin !== location.origin || a.target || a.hasAttribute("download")) return;
  const raw = a.getAttribute("href") || "";
  if (raw === "#" || ((a.hash !== "" || raw.startsWith("#")) && stripHash(a.href) === stripHash(location.href))) { saveScroll(); return; }   // same-page anchor (or a bare "#"): native scroll + hashchange
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
addEventListener("popstate", (e) => {
  const st = e.state || history.state;
  if (!st || !st.gotsx) { if (location.hash) return; }   // hash-only history entries: the browser handles them
  navigate(location.href, false, (st && st.scroll) || null);
});

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
