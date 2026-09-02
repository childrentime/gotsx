# Working on a gotsx app as an agent (gotsx {{version}})

1. **Orient**: read `AGENTS.md`, `app/.gen/host.d.ts` (what Go exposes, including actions) and the page you are changing.
   Never edit `gen/` or `app/.gen/`.
2. **Edit** `app/**`, `host/**`, `main.go`, `public/**`.
3. **Check** fast: `gotsx check --json` (or plain). Fix until it prints nothing. The message table is in `errors.md`.
4. **Run**: if `.gotsx/dev.json` exists and its `pid` is alive, a dev server is already running at its `url`;
   it rebuilds on save. Otherwise start `gotsx dev` in the background. After saving, `.gotsx/diagnostics.json`
   appears on a failed build and disappears on success; the terminal running `gotsx dev` prints browser errors as `[browser] …`.
5. **Verify** with HTTP, not by reading code: `curl -s http://localhost:3000/path | head`, check status codes and
   markup; for islands, exercise the action endpoints (`curl -X POST -H 'X-Gotsx-Action: 1' -H 'Origin: http://localhost:3000'
   -H 'Content-Type: application/json' -d '["1"]' http://localhost:3000/_gotsx/act/data/toggle`) or use a browser tool.
6. **Host changes**: after editing `host/`, a rebuild regenerates `host.d.ts`; only then can pages see new methods.
7. **Finish**: `gotsx build && go build -o /dev/null . && go test ./...` must pass. Commit `app/`, `host/`, `main.go`,
   `public/`, `AGENTS.md`. The scaffold's `.gitignore` excludes `gen/`, `app/.gen/` and `.gotsx/`: they are outputs,
   so CI and deploys run `gotsx build` before `go build` (a `go install github.com/childrentime/gotsx/cmd/gotsx@<version>` step).

Design rules for UI work: keep the design system classes and tokens (`conventions.md` → Styling); no emoji as
icons; no inline color literals; keep pages accessible (labels, `role="status"` on flashes, focus styles come free).

Things that look like they should work but do not: React hooks other than `useState/useEffect/useMemo`, context,
refs, portals, `className`, `dangerouslySetInnerHTML`, `async` server components, `getServerSideProps`,
`import` of npm packages, `fetch` in pages, classes, generics, `any`, `JSON.parse` to an unknown shape.
