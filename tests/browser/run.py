#!/usr/bin/env python3
"""Browser suites (Python Playwright): the demos and a fresh scaffold, driven through a real Chromium.

    python3 tests/browser/run.py            # everything: builds the CLI + demos, scaffolds an app, runs 5 suites
    python3 tests/browser/run.py shop admin # a subset

Needs Go, `pip install playwright && playwright install chromium`. Ports 3491-3495 must be free.
Each suite is a standalone script (python3 tests/browser/<name>.py <port>) so a single one can be rerun by hand.
"""
import os, shutil, subprocess, sys, tempfile, time, urllib.request, signal

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
HERE = os.path.dirname(os.path.abspath(__file__))
CLI = os.path.join(ROOT, ".tools", "gotsx-test")
ENV = dict(os.environ, SESSION_SECRET="browser-test", SHOP_NOLAG="1")

def sh(cmd, cwd=ROOT, env=ENV, check=True):
    r = subprocess.run(cmd, cwd=cwd, env=env, capture_output=True, text=True)
    if check and r.returncode != 0:
        raise SystemExit(f"$ {' '.join(cmd)}\n{r.stdout}{r.stderr}")
    return r

def wait(port, path="/", timeout=60):
    t0 = time.time()
    while time.time() - t0 < timeout:
        try:
            with urllib.request.urlopen(f"http://127.0.0.1:{port}{path}", timeout=2) as r:
                if r.status < 500: return True
        except Exception:
            time.sleep(0.2)
    return False

def serve(binary, cwd, port, extra=()):
    p = subprocess.Popen([binary, "-addr", f"127.0.0.1:{port}", *extra], cwd=cwd, env=ENV, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    if not wait(port):
        p.kill(); raise SystemExit(f"{binary} did not start on {port}")
    return p

def run_suite(name, *args):
    r = subprocess.run([sys.executable, os.path.join(HERE, name + ".py"), *args], cwd=ROOT, env=ENV, capture_output=True, text=True)
    tail = (r.stdout + r.stderr).strip().splitlines()[-1:] or [""]
    print(f"  {name:16s} {'PASS' if r.returncode == 0 else 'FAIL'}  {tail[0][:100]}", flush=True)
    if r.returncode != 0:
        print(r.stdout[-1500:], r.stderr[-1500:])
    return r.returncode == 0

def main():
    only = set(sys.argv[1:])
    want = lambda n: not only or n in only
    print("== build", flush=True)
    os.makedirs(os.path.join(ROOT, ".tools"), exist_ok=True)
    sh(["go", "build", "-o", CLI, "./cmd/gotsx"])
    tmp = tempfile.mkdtemp(prefix="gotsx-browser-")
    bins = {}
    for app in ("example", "shop", "admin"):
        if app != "example" and want(app) or app == "example" and (want("example_action") or want("spa")):
            sh([CLI, "build", app])
            bins[app] = os.path.join(tmp, app + "-bin")
            sh(["go", "build", "-o", bins[app], "./" + app])
    scaffold = os.path.join(tmp, "newapp")
    if want("scaffold") or want("dev_overlay"):
        sh([CLI, "new", scaffold, "--replace", ROOT])
        sh([CLI, "build", scaffold])
        bins["scaffold"] = os.path.join(scaffold, ".gotsx", "app")
        sh(["go", "build", "-o", bins["scaffold"], "."], cwd=scaffold)
    print("== run", flush=True)
    ok = True
    procs = []
    try:
        if want("example_action") or want("spa"):
            procs.append(serve(bins["example"], os.path.join(ROOT, "example"), 3491))
            if want("example_action"): ok &= run_suite("example_action", "3491")
            if want("spa"): ok &= run_suite("spa", "3491")
        if want("shop"):
            procs.append(serve(bins["shop"], os.path.join(ROOT, "shop"), 3492)); ok &= run_suite("shop", "3492")
        if want("admin"):
            procs.append(serve(bins["admin"], os.path.join(ROOT, "admin"), 3493)); ok &= run_suite("admin", "3493")
        if want("scaffold"):
            procs.append(serve(bins["scaffold"], scaffold, 3494)); ok &= run_suite("scaffold", "3494")
        if want("dev_overlay"):
            dev = subprocess.Popen([CLI, "dev", scaffold, "-addr", "127.0.0.1:3495"], cwd=scaffold, env=ENV, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            procs.append(dev)
            if not wait(3495, timeout=120):
                raise SystemExit("gotsx dev did not start")
            ok &= run_suite("dev_overlay", scaffold, "3495")
            dev.send_signal(signal.SIGINT); dev.wait(timeout=10)
    finally:
        for p in procs:
            if p.poll() is None:
                p.kill()
        shutil.rmtree(tmp, ignore_errors=True)
    print("== " + ("ALL PASS" if ok else "FAILURES"))
    sys.exit(0 if ok else 1)

if __name__ == "__main__":
    main()
