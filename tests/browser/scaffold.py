from playwright.sync_api import sync_playwright
import re, sys
port = sys.argv[1] if len(sys.argv) > 1 else "3471"
B = f"http://127.0.0.1:{port}"
with sync_playwright() as p:
    b = p.chromium.launch(); pg = b.new_page(bypass_csp=True)
    errs = []; pg.on("console", lambda m: errs.append(m.text) if m.type == "error" else None)
    acts = []; pg.on("request", lambda r: acts.append((r.method, r.url)) if "/_gotsx/act/" in r.url else None)
    pg.goto(B + "/"); pg.wait_for_timeout(300)
    assert pg.title().startswith("Todos · newapp"), pg.title()
    assert pg.locator('meta[name=description]').count() == 1
    assert pg.locator('input[name=_csrf]').input_value() != ""
    # form post with CSRF → flash
    pg.fill('input[name=title]', 'Write docs'); pg.click('form[action="/todos"] button'); pg.wait_for_timeout(300)
    assert pg.url.endswith("/"), pg.url
    flash = pg.locator('.alert-ok').inner_text(); assert 'Write docs' in flash, flash
    pg.reload(); pg.wait_for_timeout(200)
    assert pg.locator('.alert').count() == 0, "flash must be consumed once"
    # validation error via form → error flash
    pg.fill('input[name=title]', '   '); pg.click('form[action="/todos"] button'); pg.wait_for_timeout(300)
    err = pg.locator('.alert-error').inner_text(); assert 'required' in err.lower(), err
    # toggle via typed action
    n_before = pg.locator('ul.todos li.done').count()
    pg.locator('ul.todos li:not(.done) button.btn-outline').first.click(); pg.wait_for_timeout(400)
    n_after = pg.locator('ul.todos li.done').count()
    assert n_after == n_before + 1, (n_before, n_after)
    assert any(u.endswith('/_gotsx/act/data/toggle') for _, u in acts), acts
    # remove via action
    total = pg.locator('ul.todos li').count()
    pg.locator('ul.todos li button[aria-label=Remove]').first.click(); pg.wait_for_timeout(400)
    assert pg.locator('ul.todos li').count() == total - 1
    assert any(u.endswith('/_gotsx/act/data/remove') for _, u in acts), acts
    # CSRF: posting without token → 403
    r = pg.request.post(B + "/todos", form={"title": "x"}, headers={"Origin": B})
    assert r.status == 403, r.status
    print("console errors:", errs)
    assert not errs
    print("SCAFFOLD_OK")
    b.close()
