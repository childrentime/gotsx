from playwright.sync_api import sync_playwright
import time, os, sys, json
app = sys.argv[1]
port = sys.argv[2] if len(sys.argv) > 2 else "3472"
idx = os.path.join(app, "app/pages/index.server.tsx")
src = open(idx).read()
with sync_playwright() as p:
    b = p.chromium.launch(); pg = b.new_page(bypass_csp=True)
    pg.goto(f"http://127.0.0.1:{port}/"); pg.wait_for_timeout(500)
    assert pg.locator("#gotsx-overlay").count() == 0
    # break the page
    open(idx, "w").write(src.replace("const list = todos.list();", "const list = todos.nope();"))
    for _ in range(60):
        if os.path.exists(os.path.join(app, ".gotsx/diagnostics.json")): break
        time.sleep(0.25)
    d = json.load(open(os.path.join(app, ".gotsx/diagnostics.json")))
    print("diagnostics:", d["errors"][0]["file"].split("/")[-1], d["errors"][0]["line"], d["errors"][0]["msg"][:80])
    pg.wait_for_selector("#gotsx-overlay", timeout=8000)
    txt = pg.locator("#gotsx-overlay").inner_text()
    assert "nope" in txt and "index.server.tsx" in txt, txt
    print("overlay shown")
    # browser console error forwarded to the terminal
    pg.evaluate("console.error('hello from browser')"); pg.wait_for_timeout(300)
    # fix it
    open(idx, "w").write(src)
    pg.wait_for_selector("#gotsx-overlay", state="detached", timeout=15000)
    assert not os.path.exists(os.path.join(app, ".gotsx/diagnostics.json"))
    print("overlay cleared; reload after rebuild:")
    pg.wait_for_timeout(3000)
    print("page title after fix:", pg.title())
    print("OVERLAY_OK")
    b.close()
