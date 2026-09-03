from playwright.sync_api import sync_playwright
import re, sys
port = sys.argv[1] if len(sys.argv) > 1 else "3481"
B = f"http://127.0.0.1:{port}"
passed = []
def ok(name, cond, extra=""):
    assert cond, f"{name} FAILED {extra}"
    passed.append(name)
with sync_playwright() as p:
    b = p.chromium.launch(); ctx = b.new_context(bypass_csp=True); pg = ctx.new_page()
    errs = []; pg.on("console", lambda m: errs.append(m.text) if m.type == "error" else None)
    acts = []; pg.on("response", lambda r: acts.append((r.status, r.url.split(B)[1])) if "/_gotsx/act/" in r.url else None)
    # home
    pg.goto(B + "/"); pg.wait_for_timeout(1500)
    ok("home title from meta", pg.title() == "Global goods, factory-direct · gomu", pg.title())
    ok("home description", "factory-direct" in pg.locator('meta[name=description]').get_attribute("content"))
    ok("home json-ld in body", pg.locator('body script[type="application/ld+json"]').count() == 1)
    ok("feed via action", any(u == "/_gotsx/act/catalog/feed" and s == 200 for s, u in acts), acts)
    ok("feed cards rendered", pg.locator("a[href^='/p/']").count() > 10)
    # product page
    href = pg.locator("a[href^='/p/']").first.get_attribute("href")
    pg.goto(B + href); pg.wait_for_timeout(1200)
    h1 = pg.locator("h1").inner_text()
    ok("product title from meta", pg.title() == h1 + " · gomu", pg.title())
    ok("product og:type", pg.locator('meta[property="og:type"]').get_attribute("content") == "product")
    ok("product og:image", "/img/p/" in pg.locator('meta[property="og:image"]').get_attribute("content"))
    ok("product canonical", pg.locator('link[rel=canonical]').get_attribute("href").endswith(href))
    ok("related via action", any(u == "/_gotsx/act/catalog/related" and s == 200 for s, u in acts), acts)
    # pick every variant's first option, add to cart
    groups = pg.locator("div.mt-5.space-y-5 > div > div.flex.flex-wrap")
    for i in range(groups.count()):
        groups.nth(i).locator("button").first.click()
    pg.locator("button.btn-primary.btn-lg").click(); pg.wait_for_timeout(900)
    ok("add to cart message", "Added to cart" in pg.locator("span.text-success").inner_text())
    ok("cart add via action", any(u == "/_gotsx/act/cart/add" and s == 200 for s, u in acts), acts)
    ok("cart badge updated", pg.locator("a[href='/cart'] span.absolute").inner_text().strip() == "1")
    # stock error path (400 with message) through the same action, over HTTP
    r = ctx.request.post(B + "/_gotsx/act/cart/add", data=f'["{href.split("/")[-1]}", "x", 99999]', headers={"Content-Type": "application/json", "X-Gotsx-Action": "1", "Origin": B})
    ok("stock error is 400 with message", r.status == 400 and "left in stock" in r.json()["error"], (r.status, r.text()))
    # wishlist toggle on a category listing
    cat = pg.locator("nav a[href^='/c/']").first.get_attribute("href")
    pg.goto(B + cat); pg.wait_for_timeout(800)
    ok("category title from meta", pg.title().endswith(" · gomu") and pg.title() != "gomu", pg.title())
    ok("category description count", re.search(r"\d+ items", pg.locator('meta[name=description]').get_attribute("content")) is not None)
    wb = pg.locator("button[aria-label='Add to wishlist']").first
    wb.click(); pg.wait_for_timeout(700)
    ok("wish toggled on", wb.get_attribute("aria-pressed") == "true")
    ok("wish via action", any(u == "/_gotsx/act/wish/toggle" and s == 200 for s, u in acts), acts)
    wb.click(); pg.wait_for_timeout(700)
    ok("wish toggled off", wb.get_attribute("aria-pressed") == "false")
    # cart page: quantity + remove
    pg.goto(B + "/cart"); pg.wait_for_timeout(600)
    ok("cart title from meta", pg.title() == "Cart · gomu", pg.title())
    ok("cart noindex", pg.locator('meta[name=robots][content=noindex]').count() == 1)
    pg.locator("button[aria-label=Increase]").first.click(); pg.wait_for_timeout(900)
    ok("qty via action", any(u == "/_gotsx/act/cart/setQty" and s == 200 for s, u in acts), acts)
    ok("qty is 2", pg.locator("span.w-9").first.inner_text().strip() == "2")
    pg.locator("button[aria-label=Remove]").first.click(); pg.wait_for_timeout(900)
    ok("cart empty after remove", "Your cart is empty" in pg.locator("main").inner_text())
    # re-add and checkout
    pg.goto(B + href); pg.wait_for_timeout(1000)
    groups = pg.locator("div.mt-5.space-y-5 > div > div.flex.flex-wrap")
    for i in range(groups.count()):
        groups.nth(i).locator("button").first.click()
    pg.locator("button.btn-primary.btn-lg").click(); pg.wait_for_timeout(900)
    pg.goto(B + "/checkout"); pg.wait_for_timeout(600)
    ok("checkout title from meta", pg.title() == "Checkout · gomu", pg.title())
    pg.locator("button.btn-primary.btn-lg").click(); pg.wait_for_timeout(1500)
    ok("checkout 422 with fields", any(u == "/_gotsx/act/orders/place" and s == 422 for s, u in acts), acts)
    ok("three field errors shown", pg.locator("p.text-xs.text-destructive").count() == 3)
    inputs = pg.locator("div.space-y-4 input.input")  # the checkout form only (the header has a search input)
    inputs.nth(0).fill("Ada Lovelace"); inputs.nth(1).fill("+1 415 555 0132"); inputs.nth(2).fill("1 Analytical Engine Way, London")
    pg.locator("button.btn-primary.btn-lg").click()
    for _ in range(40):  # SPA navigation (pushState): poll the URL instead of waiting for a load event
        if re.search(r"/orders/ORD\d+", pg.url): break
        pg.wait_for_timeout(200)
    pg.wait_for_timeout(800)
    ok("order placed via action", any(u == "/_gotsx/act/orders/place" and s == 200 for s, u in acts), acts)
    oid = pg.url.split("/")[-1]
    ok("order title from meta", pg.title() == f"Order {oid} · gomu", pg.title())
    ok("flash shown once", "confirmed" in pg.locator(".alert-ok").inner_text() and "Ada" in pg.locator(".alert-ok").inner_text())
    pg.reload(); pg.wait_for_timeout(500)
    ok("flash consumed", pg.locator(".alert").count() == 0)
    ok("cart badge cleared", pg.locator("a[href='/cart'] span.absolute").count() == 0)
    # zh locale
    pg.goto(B + "/zh"); pg.wait_for_timeout(1200)
    ok("zh title", pg.title() == "全球好物 · 工厂直发 · gomu", pg.title())
    ok("zh html lang", pg.locator("html").get_attribute("lang") == "zh")
    pg.goto(B + "/zh/cart"); pg.wait_for_timeout(400)
    ok("zh cart title", pg.title() == "购物车 · gomu", pg.title())
    # 404 + not-found order + noindex search
    r = pg.goto(B + "/nope"); pg.wait_for_timeout(300)
    ok("404 status + page", r.status == 404 and pg.title() == "Page not found · gomu" and pg.locator("a.badge").count() > 3, (r.status, pg.title()))
    r = pg.goto(B + "/orders/NOPE")
    ok("unknown order → 404 page", r.status == 404 and pg.title() == "Page not found · gomu", (r.status, pg.title()))
    pg.goto(B + "/search?q=lamp"); pg.wait_for_timeout(500)
    ok("search noindex + title", pg.locator('meta[name=robots][content=noindex]').count() == 1 and pg.title().startswith("Results for"), pg.title())
    # CSRF: action without header → 403
    r = ctx.request.post(B + "/_gotsx/act/wish/toggle", data='["x"]', headers={"Content-Type": "application/json", "Origin": B})
    ok("action without X-Gotsx-Action → 403", r.status == 403, r.status)
    # the browser logs "Failed to load resource" for the deliberate 422 (empty checkout) and the two 404 visits
    real = [e for e in errs if not (e.startswith("Failed to load resource") and ("422" in e or "404" in e))]
    ok("no console errors (beyond the deliberate 422/404 responses)", not real, errs)
    ok("no non-2xx actions except the expected 422", all(s == 200 or (s == 422 and u.endswith("/orders/place")) for s, u in acts), acts)
    print("\n".join("PASS " + n for n in passed)); print(f"SHOP_OK ({len(passed)} assertions)")
    b.close()
