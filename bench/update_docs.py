#!/usr/bin/env python3
"""Rewrite the benchmark sections of bench/README.md, README.md and README.zh.md from bench/results/*.json.
Run after the Benchmark workflow has committed fresh results: python3 bench/update_docs.py"""
import json, os, re
HERE = os.path.dirname(os.path.abspath(__file__)); ROOT = os.path.dirname(HERE)
allr = json.load(open(os.path.join(HERE, "results/results.json")))
one = json.load(open(os.path.join(HERE, "results/results-1core.json")))
meta = allr["meta"]
by = {r["name"]: r for r in allr["results"]}; by1 = {r["name"]: r for r in one["results"]}
order = ["gotsx", "templ", "stdlib", "gin", "hono", "astro", "nextjs"]
label = {"gotsx": "gotsx", "templ": "templ", "stdlib": "html/template", "gin": "Gin", "hono": "Hono (Bun)", "astro": "Astro 7", "nextjs": "Next.js 16"}
machine = f"{meta['cpu']} · {meta['go'].replace('go version ', '')} · node {meta['node']} · bun {meta['bun']} · {meta['conc']} connections · {meta['duration']} s per contender · {meta['date']}"

def table_md(path):
    return open(path).read().split("\n", 1)[1].strip()

def replace_block(path, start, end, body):
    s = open(path).read()
    new = re.sub(re.escape(start) + r".*?" + re.escape(end), lambda m: start + "\n" + body + "\n" + end, s, flags=re.S)
    if new == s and start not in s:
        raise SystemExit(f"{path}: markers {start} not found")
    open(path, "w").write(new)

g, t, s_, x, a = by["gotsx"], by["templ"], by["stdlib"], by["nextjs"], by["astro"]
ratio = lambda r: f"{g['rps']/r['rps']:.1f}×"
takeaways = f"""**Takeaways**

- **Throughput / latency**: gotsx and templ, the two compiled-Go stacks, are within {abs(1-g['rps']/t['rps'])*100:.0f}% of each other on 4 cores ({g['rps']:,.0f} vs {t['rps']:,.0f} req/s); on one core gotsx is the fastest of the seven ({by1['gotsx']['rps']:,.0f} req/s, p50 {by1['gotsx']['p50_ms']:.1f} ms) because a page is straight-line writes with ~140 allocations. gotsx is {ratio(s_)} `html/template` (stdlib; Gin is the same template engine), {ratio(a)} Astro and {ratio(x)} Next.js, whose p50 is {x['p50_ms']:.0f} ms.
- **Memory**: the Go binaries peak at {min(by[n]['peak_rss_mb'] for n in ('gotsx','templ','stdlib','gin')):.0f}–{max(by[n]['peak_rss_mb'] for n in ('gotsx','templ','stdlib','gin')):.0f} MB under load; Hono/Bun {by['hono']['peak_rss_mb']:.0f} MB, Astro {a['peak_rss_mb']:.0f} MB, Next.js {x['peak_rss_mb']:.0f} MB.
- **Cold start**: Go binaries answer within {max(by[n]['cold_start_ms'] for n in ('gotsx','templ','stdlib','gin'))} ms of launch; Astro {a['cold_start_ms']} ms, Next.js {x['cold_start_ms']} ms.
- **Artifact**: {min(by[n]['artifact_bytes'] for n in ('gotsx','templ','stdlib'))/1e6:.0f}–{by['gin']['artifact_bytes']/1e6:.0f} MB static binaries versus {a['artifact_bytes']/1e6:.0f} MB (Astro) / {x['artifact_bytes']/1e6:.0f} MB (Next.js) of `node_modules` + build output.
- **What the browser downloads**: ~14 KB of HTML everywhere except Next.js ({x['html_bytes']/1024:.0f} KB with the RSC payload). First-load JS, gzipped: Next.js {x['js_gzip_bytes']/1024:.0f} KB (React runtime + page chunks), gotsx {g['js_gzip_bytes']/1024:.1f} KB (signals runtime + loader + the island; the morphing library for SPA navigation loads on hover or first navigation), Astro {a['js_gzip_bytes']/1024:.1f} KB (Preact + island); the template stacks ship no framework JS but also have no hydration story.
- **Build**: templ and plain Go build in under a second; gotsx adds `hostgen` (a `go run`) and the dialect compiler; Gin's {by['gin']['build_s']:.0f} s is a cold module download on the runner; Next.js builds in {x['build_s']:.0f} s."""
block = f"""**GitHub-hosted runner (`ubuntu-latest`): {machine}.** Raw data: [`results/`](results/). Runner hardware varies between runs (AMD EPYC or Intel Xeon); ratios are stable, absolute numbers move ±30%.

All cores (Go servers use all 4 vCPUs; Node and Bun are single-threaded):

{table_md(os.path.join(HERE, 'results/results.md'))}

Go servers pinned to one core (`GOMAXPROCS=1`) — the per-core comparison (JS column not re-measured):

{table_md(os.path.join(HERE, 'results/results-1core.md'))}

{takeaways}"""
replace_block(os.path.join(HERE, "README.md"), "<!-- BENCH:ALL -->", "<!-- /BENCH:ALL -->", block)

def summary(zh):
    rows = [("req/s", lambda r, r1: f"{r['rps']:,.0f}"), ("p50", lambda r, r1: f"{r['p50_ms']:.1f} ms"),
            ("峰值内存" if zh else "peak RSS", lambda r, r1: f"{r['peak_rss_mb']:.0f} MB"), ("冷启动" if zh else "cold start", lambda r, r1: f"{r['cold_start_ms']} ms"),
            ("单核 req/s" if zh else "req/s on 1 core", lambda r, r1: f"{r1['rps']:,.0f}")]
    out = ["| | " + " | ".join(label[n] for n in order) + " |", "|---|" + "---:|" * len(order)]
    for name, f in rows:
        cells = []
        for n in order:
            v = f(by[n], by1[n]); cells.append(f"**{v}**" if n == "gotsx" and name.startswith(("req/s", "单核")) else v)
        out.append(f"| {name} | " + " | ".join(cells) + " |")
    note = (f"runner: {meta['cpu']}, {meta['conc']} 连接, {meta['date']}" if zh else f"runner: {meta['cpu']}, {meta['conc']} connections, {meta['date']}")
    return "\n".join(out) + f"\n\n<sub>{note}</sub>"
replace_block(os.path.join(ROOT, "README.md"), "<!-- BENCH:SUMMARY -->", "<!-- /BENCH:SUMMARY -->", summary(False))
replace_block(os.path.join(ROOT, "README.zh.md"), "<!-- BENCH:SUMMARY -->", "<!-- /BENCH:SUMMARY -->", summary(True))
print("docs updated from", machine)
