from playwright.sync_api import sync_playwright
import re, sys
port = sys.argv[1] if len(sys.argv) > 1 else "3461"
with sync_playwright() as p:
    b = p.chromium.launch(); pg = b.new_page(bypass_csp=True)
    reqs = []; pg.on("request", lambda r: reqs.append((r.method, r.url, r.headers.get("x-gotsx-action"))))
    resps = []; pg.on("response", lambda r: resps.append((r.status, r.url)))
    errs = []; pg.on("console", lambda m: errs.append(m.text) if m.type == "error" else None)
    pg.goto(f"http://127.0.0.1:{port}/models/m1"); pg.wait_for_timeout(500)
    print([b[:80] for b in pg.locator("button").evaluate_all("els => els.map(e => e.outerHTML)")])
    btn = pg.locator("button:has(svg path[d^='M19 14'])").first
    before = int(re.sub(r"\D", "", btn.inner_text()))
    btn.click(); pg.wait_for_timeout(600)
    after = int(re.sub(r"\D", "", btn.inner_text()))
    acts = [r for r in reqs if "/_gotsx/act/" in r[1]]
    print("before", before, "after", after, "act requests", acts, "act statuses", [r for r in resps if "/_gotsx/act/" in r[1]], "console errors", errs)
    assert after == before + 1 and acts and acts[0][0] == "POST" and acts[0][2] == "1"
    print("ACTION_OK")
    b.close()
