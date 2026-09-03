"""SPA navigation on the example app: no full reload, streaming morph, head merge, scroll restoration on back,
same-page anchors left to the browser, focus + route announcer."""
from playwright.sync_api import sync_playwright
import sys, time
port = sys.argv[1] if len(sys.argv) > 1 else "3461"
B = f"http://127.0.0.1:{port}"
passed = []
def ok(name, cond, extra=""):
    assert cond, f"{name} FAILED {extra}"
    passed.append(name)
with sync_playwright() as p:
    b = p.chromium.launch(); ctx = b.new_context(bypass_csp=True, viewport={"width": 1000, "height": 500}); pg = ctx.new_page()
    errs = []; pg.on("console", lambda m: errs.append(m.text) if m.type == "error" else None)
    docs = []; pg.on("request", lambda r: docs.append(r.url) if r.resource_type in ("document", "fetch") and "/_gotsx/" not in r.url and "/public/" not in r.url else None)
    pg.goto(B + "/kitchen"); pg.wait_for_timeout(400)
    pg.evaluate("window.__marker = 42")
    kitchen_title = pg.title()
    # scroll down, then SPA-navigate home via the nav link
    pg.evaluate("window.scrollTo({top: document.body.scrollHeight, behavior: 'instant'})"); pg.wait_for_timeout(150)
    y_before = pg.evaluate("window.scrollY")
    ok("scrolled on kitchen", y_before > 100, y_before)
    t0 = time.time()
    pg.click('nav a[href="/"]')
    pg.wait_for_function(f"document.title !== {kitchen_title!r}", timeout=5000)
    t_shell = time.time() - t0
    ok("no full reload", pg.evaluate("window.__marker") == 42)
    ok("title merged", pg.title().startswith("Models"), pg.title())
    ok("description merged", pg.locator('meta[name=description]').count() == 1 and "gotsx example" in pg.locator('meta[name=description]').get_attribute("content"))
    # streaming: the shell is morphed while the 600 ms Suspense query is still running; the fill lands afterwards
    ok("fallback present right after morph", pg.locator("gotsx-suspense:not([data-ready])").count() >= 1)
    pg.wait_for_selector("gotsx-suspense[data-ready]", timeout=4000)
    t_fill = time.time() - t0
    ok("shell before the slow boundary", t_shell < t_fill - 0.2, f"shell {t_shell:.2f}s fill {t_fill:.2f}s")
    pg.wait_for_function("document.querySelectorAll('gotsx-suspense:not([data-ready])').length === 0", timeout=4000)
    ok("fill streamed in", True)
    ok("scrolled to top on push", pg.evaluate("window.scrollY") == 0)
    ok("focus on main", pg.evaluate("document.activeElement && document.activeElement.tagName") == "MAIN")
    pg.wait_for_timeout(50)
    ok("route announced", pg.evaluate("document.getElementById('gotsx-announcer') && document.getElementById('gotsx-announcer').textContent") == pg.title())
    # back: instant from the snapshot (no document request) and the scroll position comes back
    n_docs = len(docs)
    pg.go_back(); pg.wait_for_function(f"document.title === {kitchen_title!r}", timeout=5000)
    pg.wait_for_function(f"Math.abs(window.scrollY - {y_before}) < 5", timeout=3000)
    pg.wait_for_timeout(600)   # the silent revalidation morph must not move it either
    ok("back restores scroll", abs(pg.evaluate("window.scrollY") - y_before) < 5, pg.evaluate("window.scrollY"))
    ok("back still no reload", pg.evaluate("window.__marker") == 42)
    # forward restores the home page from its snapshot, then back again
    pg.go_forward(); pg.wait_for_function("document.title.startsWith('Models')", timeout=5000)
    ok("forward works", pg.evaluate("window.__marker") == 42)
    pg.go_back(); pg.wait_for_function(f"document.title === {kitchen_title!r}", timeout=5000); pg.wait_for_timeout(300)
    # same-page anchor: no fetch, native hash
    pg.evaluate("const a = document.createElement('a'); a.href = '#bottom'; a.id = 'anchor'; a.textContent = 'down'; document.body.prepend(a); const t = document.createElement('div'); t.id = 'bottom'; t.style.marginTop = '2000px'; document.body.appendChild(t)")
    pg.evaluate("window.scrollTo({top: 0, behavior: 'instant'})")
    n_docs = len(docs)
    pg.click("#anchor"); pg.wait_for_function("window.scrollY > 500", timeout=3000)
    ok("anchor is native", len(docs) == n_docs and pg.evaluate("location.hash") == "#bottom", (len(docs) - n_docs, pg.evaluate("window.scrollY")))
    # a bare "#" link is not a navigation
    pg.evaluate("const a = document.createElement('a'); a.href = '#'; a.id = 'bare'; a.textContent = 'top'; document.body.prepend(a)")
    n_docs = len(docs); pg.click("#bare"); pg.wait_for_timeout(200)
    ok("bare # not navigated", len(docs) == n_docs and not pg.evaluate("location.pathname").endswith("/"), pg.evaluate("location.href"))
    # a full reload keeps the scroll position even though restoration is manual
    pg.goto(B + "/kitchen"); pg.wait_for_timeout(300)
    pg.evaluate("window.scrollTo({top: 200, behavior: 'instant'})"); pg.wait_for_timeout(100)
    pg.reload(); pg.wait_for_timeout(500)
    ok("reload keeps scroll", abs(pg.evaluate("window.scrollY") - 200) < 5, pg.evaluate("window.scrollY"))
    # GET form navigation is SPA too
    pg.goto(B + "/"); pg.wait_for_timeout(300); pg.evaluate("window.__marker = 7")
    if pg.locator('form[method="get"] input[name="q"]').count() > 0:
        pg.fill('form[method="get"] input[name="q"]', "stand"); pg.press('form[method="get"] input[name="q"]', "Enter")
        pg.wait_for_function("location.search.includes('q=stand')", timeout=5000); pg.wait_for_timeout(200)
        ok("GET form is SPA", pg.evaluate("window.__marker") == 7)
    ok("no console errors", not errs, errs)
    print(f"SPA_OK ({len(passed)} assertions)")
    b.close()
