from playwright.sync_api import sync_playwright
import sys
port = sys.argv[1] if len(sys.argv) > 1 else "3482"
B = f"http://127.0.0.1:{port}"
passed = []
def ok(name, cond, extra=""):
    assert cond, f"{name} {extra}"
    passed.append(name)
with sync_playwright() as p:
    b = p.chromium.launch(); ctx = b.new_context(bypass_csp=True); pg = ctx.new_page()
    # Chromium logs a "Failed to load resource" line for every non-2xx response; the 422s / 404 here are deliberate
    errs = []; pg.on("console", lambda m: errs.append(m.text) if m.type == "error" and not m.text.startswith("Failed to load resource") else None)
    pg.on("pageerror", lambda e: errs.append("pageerror: " + str(e)))
    acts = []; pg.on("response", lambda r: acts.append((r.url.split("/_gotsx/act/")[1], r.status)) if "/_gotsx/act/" in r.url else None)
    pg.on("dialog", lambda d: d.accept())
    # 1. unauthenticated → /login
    pg.goto(B + "/"); pg.wait_for_timeout(300)
    ok("unauth redirects to /login", pg.url.endswith("/login"), pg.url)
    ok("login title from meta", pg.title() == "Sign in · gotsx admin", pg.title())
    ok("login description from meta", pg.locator('meta[name=description]').count() == 1)
    pg.goto(B + "/users"); pg.wait_for_timeout(200)
    ok("unauth /users redirects", pg.url.endswith("/login"), pg.url)
    # 2. wrong password
    pg.fill('input[name=pass]', 'nope'); pg.click('button:has-text("Sign in")'); pg.wait_for_timeout(500)
    ok("wrong password shows error", "Wrong username or password" in pg.locator('.alert-error').inner_text())
    ok("login action 422", ("auth/login", 422) in acts, acts)
    # 3. correct login → dashboard + flash
    pg.fill('input[name=pass]', 'admin123'); pg.click('button:has-text("Sign in")'); pg.wait_for_url(B + "/", timeout=5000); pg.wait_for_timeout(300)
    ok("login action 200", ("auth/login", 200) in acts, acts)
    ok("dashboard title from meta", pg.title() == "Dashboard · gotsx admin", pg.title())
    ok("flash after login", "Welcome back, Super Admin" in pg.locator('.alert-ok').first.inner_text())
    ok("header shows profile", pg.locator('header').inner_text().find("Super Admin") >= 0)
    pg.reload(); pg.wait_for_timeout(300)
    ok("flash consumed once", pg.locator('.alert-ok').count() == 0)
    # 4. users page: list via action
    pg.goto(B + "/users"); pg.wait_for_selector('table.table tbody tr', timeout=5000); pg.wait_for_timeout(200)
    ok("users title from meta", pg.title() == "Users · gotsx admin", pg.title())
    ok("list action 200", ("users/all", 200) in acts, acts)
    before = int(pg.locator('.card .text-2xl').first.inner_text())
    # 5. validation error
    pg.click('button:has-text("New user")'); pg.wait_for_selector('[role=dialog]')
    pg.fill('[role=dialog] input >> nth=0', 'x'); pg.fill('[role=dialog] input >> nth=1', 'bad'); pg.click('[role=dialog] button:has-text("Save")'); pg.wait_for_timeout(500)
    dlg = pg.locator('[role=dialog]').inner_text()
    ok("field errors shown", "Name must be at least 2 characters" in dlg and "Invalid email address" in dlg and "Department is required" in dlg, dlg)
    ok("create action 422", ("users/create", 422) in acts, acts)
    # 6. create
    pg.fill('[role=dialog] input >> nth=0', 'Test Person'); pg.fill('[role=dialog] input >> nth=1', 'test.person@gotsx.dev'); pg.fill('[role=dialog] input >> nth=2', 'QA')
    pg.click('[role=dialog] button:has-text("Save")'); pg.wait_for_timeout(600)
    ok("create action 200", ("users/create", 200) in acts, acts)
    ok("toast after create", "User created" in pg.locator('.alert-ok').last.inner_text())
    pg.fill('input[placeholder^=Search]', 'Test Person'); pg.wait_for_timeout(300)
    ok("created row visible", pg.locator('tr:has-text("Test Person")').count() == 1)
    ok("total went up", int(pg.locator('.card .text-2xl').first.inner_text()) == before + 1)
    # 7. edit
    pg.locator('tr:has-text("Test Person") button[aria-label=Edit]').click(); pg.wait_for_selector('[role=dialog]')
    ok("edit modal prefilled", pg.locator('[role=dialog] input >> nth=0').input_value() == "Test Person")
    pg.fill('[role=dialog] input >> nth=0', 'Test Person Two'); pg.click('[role=dialog] button:has-text("Save")'); pg.wait_for_timeout(600)
    ok("update action 200", ("users/update", 200) in acts, acts)
    ok("row renamed", pg.locator('tr:has-text("Test Person Two")').count() == 1)
    # 8. delete
    pg.locator('tr:has-text("Test Person Two") button[aria-label=Delete]').click(); pg.wait_for_timeout(600)
    ok("remove action 200", ("users/remove", 200) in acts, acts)
    ok("toast after delete", "User deleted" in pg.locator('.alert-ok').last.inner_text())
    ok("row gone", pg.locator('tr:has-text("Test Person Two")').count() == 0)
    # 9. logout
    pg.click('button[aria-label="Sign out"]'); pg.wait_for_url(B + "/login", timeout=5000); pg.wait_for_timeout(300)
    ok("logout action 200", ("auth/logout", 200) in acts, acts)
    ok("signed-out flash", "signed out" in pg.locator('.alert-info').inner_text())
    pg.goto(B + "/"); pg.wait_for_timeout(200)
    ok("session cleared", pg.url.endswith("/login"), pg.url)
    # 10. action without session → rejected
    r = ctx.request.post(B + "/_gotsx/act/users/all", data="[]", headers={"Origin": B, "X-Gotsx-Action": "1", "Content-Type": "application/json"})
    ok("unauth action rejected", r.status == 401 and "not signed in" in r.text(), (r.status, r.text()))
    # 11. viewer role cannot edit
    pg.fill('input[name=user]', 'demo'); pg.fill('input[name=pass]', 'demo'); pg.click('button:has-text("Sign in")'); pg.wait_for_url(B + "/", timeout=5000)
    pg.goto(B + "/users"); pg.wait_for_selector('table.table tbody tr', timeout=5000)
    ok("viewer has no New user button", pg.locator('button:has-text("New user")').count() == 0)
    r = ctx.request.post(B + "/_gotsx/act/users/remove", data='["u001"]', headers={"Origin": B, "X-Gotsx-Action": "1", "Content-Type": "application/json"})
    ok("viewer remove rejected", r.status == 403 and "role" in r.text(), (r.status, r.text()))
    pg.goto(B + "/nope"); pg.wait_for_timeout(200)
    ok("404 page from _404 with meta", pg.title() == "Page not found · gotsx admin" and "nothing at /nope" in pg.inner_text("main"), pg.title())
    ok("no console errors", not errs, errs)
    print("PASSED:", len(passed)); print("\n".join(" - " + x for x in passed)); print("ADMIN_OK")
    b.close()
