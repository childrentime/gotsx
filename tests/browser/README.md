# Browser suites

Real-Chromium checks of the demos and of a fresh `gotsx new` scaffold, using Python Playwright:

| suite | what it drives |
|---|---|
| `example_action.py` | the Like button: a typed action round trip (`POST /_gotsx/act/data/like`, marker header, count updates) |
| `shop.py` | 41 assertions: meta titles, JSON-LD, feed/related actions, cart, wishlist, checkout validation and flash, `/zh`, 404s, CSRF |
| `admin.py` | 33 assertions: session login/logout, 401/403, user CRUD through actions, field errors, toasts, 404 page |
| `scaffold.py` | a scaffolded app: meta, CSRF form + flash, validation error, toggle/remove actions, cross-site post refused |
| `dev_overlay.py` | `gotsx dev`: error overlay appears on a broken build and clears on the fix, console errors forwarded |

```bash
pip install playwright && playwright install chromium
python3 tests/browser/run.py           # builds everything, runs all suites (≈2 min)
python3 tests/browser/run.py shop      # one suite
```

`make test` does not run these (they need a browser); CI runs them in the `browser` job.
