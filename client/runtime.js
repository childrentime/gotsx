/* gotsx 客户端运行时: signals + 精确 DOM 绑定 + 走位 hydrate。
   编译器把 JSX 编成 el/t/text/cond/each 调用; 服务端 HTML 里的 <!--$--> / <!--[--> 标记
   让 hydrate 不需要 diff —— 同一个编译器产出的结构, 按顺序认领节点即可。 */

let current = null;          // 正在追踪依赖的计算
let owner = null;            // 当前所有者(用于销毁嵌套 effect, 防止 effect 更新已卸载的 DOM)
const pending = new Set();
let flushing = false;

export function signal(v) {
  const subs = new Set();
  const get = () => {
    if (current) { subs.add(current); current.deps.push(subs); }
    return v;
  };
  const set = (nv) => {
    if (typeof nv === "function") nv = nv(v);
    if (Object.is(nv, v)) return;
    v = nv;
    for (const s of [...subs]) schedule(s);
  };
  return [get, set];
}

function disposeChildren(e) {
  if (e.children && e.children.length) {
    for (const c of e.children) dispose(c);
    e.children = [];
  }
}
function dispose(e) {
  disposeChildren(e);
  for (const d of e.deps) d.delete(e);
  e.deps = [];
  e.disposed = true;
}

export function effect(fn) {
  const e = { fn, deps: [], children: [], run: null, disposed: false };
  if (owner) owner.children.push(e);   // 登记到所有者, 以便随其一起销毁
  e.run = () => {
    if (e.disposed) return;
    disposeChildren(e);                // 重跑前先销毁上一轮建立的嵌套 effect
    for (const d of e.deps) d.delete(e);
    e.deps = [];
    const pc = current, po = owner;
    current = e; owner = e;
    try { e.fn(); } finally { current = pc; owner = po; }
  };
  e.run();
  return e;
}

export function memo(fn) {
  const [get, set] = signal(undefined);
  effect(() => set(fn()));
  return get;
}

export function untrack(fn) {
  const prev = current;
  current = null;
  try { return fn(); } finally { current = prev; }
}

/* onMount: useEffect(fn, []) 的语义 —— 挂载后跑一次, 不建立 signal 依赖 */
export function onMount(fn) { queueMicrotask(() => untrack(fn)); }

/* 与 Go 后端逐字节对齐语义的内建 */
export function sort(xs, cmp) { return xs.slice().sort((a, b) => cmp(a, b)); }   // 拷贝后排序, 不改原数组
export function at(xs, i) { i = i < 0 ? xs.length + i : i; return i >= 0 && i < xs.length ? xs[i] : (typeof xs[0] === "number" ? 0 : typeof xs[0] === "string" ? "" : undefined); }
export function reverse(xs) { return xs.slice().reverse(); }
export function objectKeys(o) { return Object.keys(o).sort(); }                 // 排序, 与 Go map 一致
export function objectValues(o) { return objectKeys(o).map((k) => o[k]); }
export function jsonLd(s) { const el = document.createElement("script"); el.type = "application/ld+json"; el.textContent = s; return el; }

/* i18n(客户端): 与 Go 后端逐一对应。目录是当前语言的那份(服务端注入到 __GOTSX.i18n) */
function _i18n() { return (window.__GOTSX && window.__GOTSX.i18n) || { locale: "", default: "", prefix: false, messages: {}, currency: {} }; }
function _msg(key) { const m = _i18n().messages || {}; return key in m ? m[key] : key; }
export function tr(locale, key) { return _msg(key); }
export function trv(locale, key, vars) { let s = _msg(key); for (const k in vars) s = s.split("{" + k + "}").join(vars[k]); return s; }
const _oneLocales = { zh: 1, ja: 1, ko: 1, vi: 1, th: 1, id: 1 };
export function plural(locale, key, n) {
  const forms = _msg(key).split("|");
  const pick = (!_oneLocales[locale] && n === 1 && forms.length > 1) ? forms[0] : forms[forms.length - 1];
  return pick.split("{n}").join(String(n));
}
function _group(intStr) {
  const neg = intStr[0] === "-"; if (neg) intStr = intStr.slice(1);
  let out = ""; for (let i = 0; i < intStr.length; i++) { if (i > 0 && (intStr.length - i) % 3 === 0) out += ","; out += intStr[i]; }
  return (neg ? "-" : "") + out;
}
export function fmtNum(locale, n) { const s = String(n); const d = s.indexOf("."); return d < 0 ? _group(s) : _group(s.slice(0, d)) + s.slice(d); }
export function fmtCur(locale, cents) {
  const sym = (_i18n().currency || {})[locale] || "¥"; const neg = cents < 0; if (neg) cents = -cents;
  const whole = Math.floor(cents / 100), frac = Math.round(cents % 100);
  return (neg ? "-" : "") + sym + _group(String(whole)) + "." + String(frac).padStart(2, "0");
}
const _monthEn = ["", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
export function fmtDate(locale, iso) {
  const d = new Date(iso); if (isNaN(d)) return iso;
  return locale === "zh" ? `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日` : `${_monthEn[d.getMonth() + 1]} ${d.getDate()}, ${d.getFullYear()}`;
}
export function lpath(locale, path) {
  const c = _i18n(); if (!c.prefix || locale === c.default || !locale) return path;
  return path === "/" ? "/" + locale : "/" + locale + path;
}

function schedule(e) {
  if (e.disposed) return;
  pending.add(e);
  if (flushing) return;
  flushing = true;
  queueMicrotask(() => {
    flushing = false;
    const list = [...pending];
    pending.clear();
    for (const e of list) if (!e.disposed) e.run();
  });
}

/* ---------- hydrate 游标 ---------- */
let hy = null;   // { node }: 下一个待认领的节点; null = 创建模式

function skipWs() {
  while (hy.node && hy.node.nodeType === 3 && hy.node.data.trim() === "" && !(hy.node.nextSibling && hy.node.nextSibling.nodeType === 8)) hy.node = hy.node.nextSibling;
}
function claimEl(tag) {
  let n = hy.node;
  while (n && !(n.nodeType === 1 && n.tagName.toLowerCase() === tag)) n = n.nextSibling;
  if (!n) { n = document.createElement(tag); if (hy.node) hy.node.parentNode.insertBefore(n, hy.node); }
  hy.node = n.nextSibling;
  return n;
}
function claimText(s) {
  const n = hy.node;
  if (n && n.nodeType === 3) { hy.node = n.nextSibling; if (n.data !== s) n.data = s; return n; }
  const t = document.createTextNode(s);
  if (n) n.parentNode.insertBefore(t, n);
  return t;
}
function claimComment(data) {
  let n = hy.node;
  while (n && !(n.nodeType === 8 && n.data === data)) n = n.nextSibling;
  if (!n) throw new Error("hydrate: 找不到标记 <!--" + data + "-->");
  hy.node = n.nextSibling;
  return n;
}

const attrMap = { className: "class", htmlFor: "for", charSet: "charset", tabIndex: "tabindex" };
function setAttr(e, name, v) {
  name = attrMap[name] || name;
  if (typeof v === "boolean" && (name.startsWith("aria-") || name.startsWith("data-"))) v = String(v);
  if (name === "value" && "value" in e) { if (e.value !== String(v ?? "")) e.value = v ?? ""; return; }
  if (name === "checked" || name === "disabled" || name === "selected" || name === "readonly" || name === "hidden") {
    if (v) e.setAttribute(name, ""); else e.removeAttribute(name);
    if (name in e) e[name] = !!v;
    return;
  }
  if (v === false || v == null) e.removeAttribute(name);
  else e.setAttribute(name, v === true ? "" : String(v));
}
function bindAttr(e, name, v) {
  if (name.startsWith("on") && typeof v === "function") { e.addEventListener(name.slice(2).toLowerCase(), v); return; }
  if (typeof v === "function") effect(() => setAttr(e, name, v()));
  else if (!hy) setAttr(e, name, v);          // hydrate 时静态属性已在 HTML 里
}

function flat(v, out = []) {
  if (v == null || v === false || v === true) return out;
  if (typeof v === "function") return flat(v(), out);        // children thunk
  if (Array.isArray(v)) { for (const x of v) flat(x, out); return out; }
  if (typeof v === "string" || typeof v === "number") { out.push(document.createTextNode(String(v))); return out; }
  out.push(v);
  return out;
}

/* el(tag, attrs, kids): kids 是返回子节点数组的函数 —— 延迟到父元素认领之后再执行, hydrate 才能按序走位 */
const SVG = new Set(["svg", "path", "circle", "rect", "line", "polyline", "polygon", "g", "defs", "use", "text", "ellipse", "clipPath", "mask", "linearGradient", "stop"]);
export function el(tag, attrs, kids) {
  const e = hy ? claimEl(tag) : (SVG.has(tag) ? document.createElementNS("http://www.w3.org/2000/svg", tag) : document.createElement(tag));
  if (attrs) for (const k in attrs) bindAttr(e, k, attrs[k]);
  if (kids) {
    if (hy) {
      const save = hy;
      hy = { node: e.firstChild };
      flat(kids());                          // 顺序执行(含 children thunk), 逐个认领
      hy = save;
    } else {
      for (const n of flat(kids())) e.appendChild(n);
    }
  }
  return e;
}

/* 跨岛事件: 岛之间不共享 signal, 用 DOM 事件总线沟通 */
export function emit(name, detail) { document.dispatchEvent(new CustomEvent(name, { detail })); }
export function on(name, fn) { document.addEventListener(name, (e) => fn(e.detail)); }

/* children 可能是 thunk(组件调用时延迟求值) */
export function kids(c) { return typeof c === "function" ? c() : c; }

/* 静态文本 */
export function t(s) {
  return hy ? claimText(s) : document.createTextNode(s);
}

/* 响应式文本: 服务端标记 <!--$-->…<!--/--> */
export function text(fn) {
  let node;
  if (hy) {
    claimComment("$");
    const n = hy.node;
    if (n && n.nodeType === 3) { node = n; hy.node = n.nextSibling; }
    else { node = document.createTextNode(""); n.parentNode.insertBefore(node, n); }
    claimComment("/");
  } else node = document.createTextNode("");
  effect(() => { const v = fn(); const s = v == null || v === false ? "" : String(v); if (node.data !== s) node.data = s; });
  return node;
}

/* 响应式区块(条件 / 三元 / 列表): start / end 注释标记之间的内容随 effect 重建。
   重建时按 DOM 范围(start 与 end 之间)整体删除, 而不是追踪节点数组 ——
   这样即使嵌套块又往里插了节点, 也能被干净清除(否则深层嵌套会残留旧节点)。 */
function block(build) {
  const first = !!hy;
  const start = hy ? claimComment("[") : document.createComment("[");
  let end, initialNodes = [], initial = true;
  effect(() => {
    const nodes = build();          // cond/each 内部已 flat
    if (initial) {
      initial = false;
      end = hy ? claimComment("]") : document.createComment("]");
      initialNodes = nodes;         // 创建模式: 由调用方插入 [start, ...nodes, end]
      return;
    }
    const parent = end.parentNode;
    if (!parent) return;            // 所在区域已被上层删除: 什么都不做(所有者机制已停掉本 effect)
    let n = start.nextSibling;      // 删除 start 与 end 之间的一切(含嵌套块插入的节点)
    while (n && n !== end) { const nx = n.nextSibling; parent.removeChild(n); n = nx; }
    for (const x of nodes) parent.insertBefore(x, end);
  });
  return first ? [] : [start, ...initialNodes, end];
}

export function cond(test, a, b) {
  return block(() => {
    const v = test();
    const branch = v ? a : b;
    return untrack(() => {
      const save = hy;
      if (!hy) { /* 创建模式 */ }
      const out = branch ? flat(branch()) : [];
      hy = save;
      return out;
    });
  });
}

export function each(list, item) {
  return block(() => {
    const xs = list() || [];
    return untrack(() => flat(xs.map((x, i) => item(x, i))));
  });
}

/* 岛的挂载: hydrate=true 时按 root 里的现有 DOM 走位, 否则重建 */
export function mount(Comp, props, root, hydrate) {
  if (hydrate) {
    hy = { node: root.firstChild };
    try { Comp(props); } finally { hy = null; }
  } else {
    root.replaceChildren(...flat(Comp(props)));
  }
}
