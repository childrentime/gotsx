#!/usr/bin/env python3
"""Reproducible SSR benchmark: the same 50-product page in gotsx, Go html/template, Gin, templ, Next.js, Astro and Hono.

For each contender: build (timed) → start → cold start (ms to first 200) → load test with bench/load
(keep-alive, C connections, D seconds after a warm-up) while sampling RSS → optional client-weight run
(Playwright: bytes of HTML / JS / CSS a browser downloads). Writes results/results.json and results/results.md.

Usage: python3 bench/run.py [--only gotsx,stdlib,...] [--conc 64] [--duration 10] [--single-core] [--no-client]
Requirements: Go; Node + npm (Next, Astro); Bun (Hono); Python Playwright for --client (optional).
"""
import argparse, json, os, shutil, subprocess, sys, threading, time, urllib.request, gzip, io, platform

HERE = os.path.dirname(os.path.abspath(__file__))
os.chdir(HERE)

CONTENDERS = {
    # name: dict(build=[cmd...], run=[cmd...], cwd, port, size_paths)
    "gotsx":  dict(cwd="gotsx",  build=[["sh", "-c", "../../.tools/gotsx-stable build . >/dev/null 2>&1 || go run ../../cmd/gotsx build . && go build -o .gotsx/app ."]],
                   run=["./.gotsx/app", "-addr", ":{port}"], size=[".gotsx/app"], note="production mode: CSP nonce, security headers, request logging off"),
    "stdlib": dict(cwd="stdlib", build=[["go", "build", "-o", "stdlib", "."]], run=["./stdlib", "-addr", ":{port}"], size=["stdlib"], note="net/http + html/template"),
    "gin":    dict(cwd="gin",    build=[["go", "build", "-o", "gin", "."]], run=["./gin", "-addr", ":{port}"], size=["gin"], note="gin release mode + html/template"),
    "templ":  dict(cwd="templ",  build=[["go", "build", "-o", "templ", "."]], run=["./templ", "-addr", ":{port}"], size=["templ"], note="a-h/templ generated Go"),
    "nextjs": dict(cwd="nextjs", build=[["npx", "next", "build"]], run=["npx", "next", "start", "-p", "{port}"], size=[".next", "node_modules"], node=True, note="App Router, force-dynamic, next start"),
    "astro":  dict(cwd="astro",  build=[["npx", "astro", "build"]], run=["node", "dist/server/entry.mjs"], env={"PORT": "{port}", "HOST": "127.0.0.1"}, size=["dist", "node_modules"], node=True, note="output: server, node adapter, preact island"),
    "hono":   dict(cwd="hono",   build=[], run=["bun", "run", "index.tsx"], env={"PORT": "{port}"}, size=["node_modules"], node=True, note="hono/jsx on bun, SSR only"),
}

def du(paths, cwd):
    total = 0
    for p in paths:
        full = os.path.join(cwd, p)
        if os.path.isfile(full):
            total += os.path.getsize(full)
        elif os.path.isdir(full):
            for root, _, files in os.walk(full):
                for f in files:
                    try: total += os.path.getsize(os.path.join(root, f))
                    except OSError: pass
    return total

def wait_ready(port, timeout=60):
    t0 = time.time()
    while time.time() - t0 < timeout:
        try:
            with urllib.request.urlopen(f"http://127.0.0.1:{port}/", timeout=2) as r:
                if r.status == 200:
                    r.read()
                    return (time.time() - t0) * 1000
        except Exception:
            time.sleep(0.01)
    return None

def rss_kb(pid):
    total = 0
    try:
        out = subprocess.run(["ps", "-o", "rss=,pid=", "-p", str(pid)], capture_output=True, text=True).stdout.split()
        if out: total += int(out[0])
        kids = subprocess.run(["pgrep", "-P", str(pid)], capture_output=True, text=True).stdout.split()
        for k in kids: total += rss_kb(int(k))
    except Exception:
        pass
    return total

def kill_tree(proc):
    try:
        kids = subprocess.run(["pgrep", "-P", str(proc.pid)], capture_output=True, text=True).stdout.split()
        for k in kids: subprocess.run(["kill", "-9", k], capture_output=True)
        proc.kill(); proc.wait(timeout=5)
    except Exception:
        pass

def client_weight(port):
    try:
        from playwright.sync_api import sync_playwright
    except ImportError:
        return None
    with sync_playwright() as p:
        b = p.chromium.launch(); page = b.new_page()
        sizes = {"document": 0, "script": 0, "stylesheet": 0, "other": 0}; gz_js = io.BytesIO(); count = 0
        def on_resp(resp):
            nonlocal count
            try:
                body = resp.body(); count += 1
                kind = resp.request.resource_type
                key = kind if kind in sizes else "other"
                sizes[key] += len(body)
                if kind == "script": gz_js.write(body)
            except Exception:
                pass
        page.on("response", on_resp)
        page.goto(f"http://127.0.0.1:{port}/", wait_until="networkidle")
        page.click("button"); page.wait_for_timeout(100)
        interactive = "1" in page.inner_text("header button")
        b.close()
    js_gz = len(gzip.compress(gz_js.getvalue())) if gz_js.tell() else 0
    return dict(html_bytes=sizes["document"], js_bytes=sizes["script"], js_gzip_bytes=js_gz, css_bytes=sizes["stylesheet"], requests=count, counter_works=interactive)

def run_one(name, spec, port, conc, duration, single_core, with_client):
    cwd = os.path.join(HERE, spec["cwd"])
    res = dict(name=name, note=spec.get("note", ""))
    t0 = time.time()
    for cmd in spec["build"]:
        r = subprocess.run(cmd, cwd=cwd, capture_output=True, text=True)
        if r.returncode != 0:
            print(f"  build failed: {r.stdout[-800:]} {r.stderr[-800:]}", file=sys.stderr); res["error"] = "build failed"; return res
    res["build_s"] = round(time.time() - t0, 2)
    res["artifact_bytes"] = du(spec["size"], cwd)
    env = dict(os.environ)
    for k, v in spec.get("env", {}).items(): env[k] = v.format(port=port)
    if single_core and not spec.get("node"): env["GOMAXPROCS"] = "1"
    cmd = [a.format(port=port) for a in spec["run"]]
    proc = subprocess.Popen(cmd, cwd=cwd, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        cold = wait_ready(port)
        if cold is None:
            res["error"] = "did not start"; return res
        res["cold_start_ms"] = round(cold)
        peak = [0]; stop = threading.Event()
        def sampler():
            while not stop.is_set():
                peak[0] = max(peak[0], rss_kb(proc.pid)); time.sleep(0.2)
        th = threading.Thread(target=sampler, daemon=True); th.start()
        out = subprocess.run([os.path.join(HERE, "load", "load"), "-url", f"http://127.0.0.1:{port}/", "-c", str(conc), "-d", f"{duration}s", "-warmup", "2s"], capture_output=True, text=True)
        stop.set(); th.join()
        try:
            res.update(json.loads(out.stdout.strip().splitlines()[-1]))
        except Exception:
            res["error"] = "load failed: " + out.stderr[-300:]; return res
        res["peak_rss_mb"] = round(peak[0] / 1024, 1)
        if with_client:
            cw = client_weight(port)
            if cw: res.update(cw)
    finally:
        kill_tree(proc)
    return res

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--only", default="")
    ap.add_argument("--conc", type=int, default=64)
    ap.add_argument("--duration", type=int, default=10)
    ap.add_argument("--single-core", action="store_true", help="GOMAXPROCS=1 for the Go servers (Node/Bun are single-threaded anyway)")
    ap.add_argument("--no-client", action="store_true")
    a = ap.parse_args()
    names = [n for n in a.only.split(",") if n] or list(CONTENDERS)
    subprocess.run(["go", "build", "-o", "load/load", "./load"], check=True)
    for d, files in [("nextjs", ["products.json", "shared.css"]), ("hono", ["products.json", "shared.css"]), ("astro/src", ["products.json", "shared.css"])]:
        os.makedirs(d, exist_ok=True)
        for f in files: shutil.copy(os.path.join("data" if f.endswith(".json") else ".", f), os.path.join(d, f))
    results = []
    port = 3100
    for n in names:
        print(f"== {n}", flush=True)
        results.append(run_one(n, CONTENDERS[n], port, a.conc, a.duration, a.single_core, not a.no_client)); port += 1
        print("  ", {k: v for k, v in results[-1].items() if k in ("rps", "p50_ms", "p99_ms", "peak_rss_mb", "cold_start_ms", "js_bytes", "error")}, flush=True)
    os.makedirs("results", exist_ok=True)
    meta = dict(machine=platform.machine(), system=platform.platform(), cpu=subprocess.run(["sysctl", "-n", "machdep.cpu.brand_string"], capture_output=True, text=True).stdout.strip() if sys.platform == "darwin" else "",
                go=subprocess.run(["go", "version"], capture_output=True, text=True).stdout.strip(), node=subprocess.run(["node", "-v"], capture_output=True, text=True).stdout.strip(),
                bun=subprocess.run(["bun", "-v"], capture_output=True, text=True).stdout.strip(), conc=a.conc, duration=a.duration, single_core=a.single_core, date=time.strftime("%Y-%m-%d"))
    suffix = "-1core" if a.single_core else ""
    json.dump(dict(meta=meta, results=results), open(f"results/results{suffix}.json", "w"), indent=1)
    with open(f"results/results{suffix}.md", "w") as f:
        f.write(f"<!-- generated by bench/run.py on {meta['date']}: {meta['cpu']} · {meta['go']} · node {meta['node']} · bun {meta['bun']} · c={a.conc} d={a.duration}s single_core={a.single_core} -->\n")
        f.write("| framework | req/s | p50 | p99 | peak RSS | cold start | build | artifact | HTML | JS (gz) | note |\n|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|\n")
        for r in results:
            if "error" in r:
                f.write(f"| {r['name']} | error: {r['error']} |||||||||\n"); continue
            f.write(f"| {r['name']} | {r['rps']:,.0f} | {r['p50_ms']:.2f} ms | {r['p99_ms']:.2f} ms | {r['peak_rss_mb']} MB | {r['cold_start_ms']} ms | {r['build_s']} s | {r['artifact_bytes']/1e6:.1f} MB | {r.get('html_bytes', r.get('bytes_per_response',0))/1024:.1f} KB | {r.get('js_bytes',0)/1024:.1f} KB ({r.get('js_gzip_bytes',0)/1024:.1f}) | {r['note']} |\n")
    print(open(f"results/results{suffix}.md").read())

if __name__ == "__main__":
    main()
