// gotsx: 方言编译器 + dev 循环。
//
//	gotsx build <appdir>   编译 app/ → gen/
//	gotsx dev   <appdir>   编译 + go build + 运行, 监视 app/ 改动自动重来
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gotsx/client"
	"gotsx/compiler"
)

type appConfig struct {
	Module      string `json:"module"`
	HostPackage string `json:"hostPackage"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("用法: gotsx build|dev <appdir>")
		os.Exit(2)
	}
	dir, _ := filepath.Abs(os.Args[2])
	cfg := readConfig(dir)
	switch os.Args[1] {
	case "build":
		if err := build(dir, cfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "dev":
		dev(dir, cfg)
	default:
		fmt.Println("未知命令", os.Args[1])
		os.Exit(2)
	}
}

func readConfig(dir string) appConfig {
	var cfg appConfig
	b, err := os.ReadFile(filepath.Join(dir, "gotsx.json"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "缺少 gotsx.json:", err)
		os.Exit(1)
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		fmt.Fprintln(os.Stderr, "gotsx.json:", err)
		os.Exit(1)
	}
	return cfg
}

func build(dir string, cfg appConfig) error {
	t := time.Now()
	// 宿主类型: 应用自带 cmd/hostgen 就先跑一下
	if _, err := os.Stat(filepath.Join(dir, "cmd", "hostgen")); err == nil {
		cmd := exec.Command("go", "run", "./cmd/hostgen", "app/.gen")
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("hostgen 失败: %v\n%s", err, out)
		}
	}
	if err := runTailwind(dir); err != nil {
		return err
	}
	rep, err := compiler.Build(compiler.Config{
		AppDir:     filepath.Join(dir, "app"),
		OutDir:     filepath.Join(dir, "gen"),
		Module:     cfg.Module,
		HostPkg:    cfg.HostPackage,
		RuntimePkg: "gotsx/runtime",
		ClientFS:   client.FS,
	})
	if err != nil {
		return err
	}
	fmt.Printf("gotsx: 编译 %d 个模块 → 路由 %v, 岛 %v (%s)\n", len(rep.Modules), rep.Routes, rep.Islands, time.Since(t).Round(time.Millisecond))
	return nil
}

func dev(dir string, cfg appConfig) {
	var proc *exec.Cmd
	stop := func() {
		if proc != nil && proc.Process != nil {
			proc.Process.Kill()
			proc.Wait()
			proc = nil
		}
	}
	gen := 0
	run := func() {
		// 先构建, 成功了才换进程: 编译错误期间旧服务继续跑
		if err := build(dir, cfg); err != nil {
			fmt.Fprintln(os.Stderr, "\n"+err.Error()+"\n(旧版本继续运行)")
			return
		}
		gen++
		bin := filepath.Join(dir, ".gotsx", fmt.Sprintf("app-%d", gen%2)) // 交替文件名, 不覆盖正在运行的二进制
		os.MkdirAll(filepath.Dir(bin), 0o755)
		t := time.Now()
		gb := exec.Command("go", "build", "-o", bin, ".")
		gb.Dir = dir
		if out, err := gb.CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "\ngo build 失败:\n%s\n(旧版本继续运行)\n", out)
			return
		}
		fmt.Printf("gotsx: go build %s\n", time.Since(t).Round(time.Millisecond))
		stop()
		proc = exec.Command(bin, append([]string{"-dev"}, os.Args[3:]...)...)
		proc.Dir = dir
		proc.Stdout, proc.Stderr = os.Stdout, os.Stderr
		if err := proc.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "启动失败:", err)
		}
	}
	run()
	last := latestMtime(filepath.Join(dir, "app"), filepath.Join(dir, "host"), filepath.Join(dir, "public"))
	for {
		time.Sleep(400 * time.Millisecond)
		now := latestMtime(filepath.Join(dir, "app"), filepath.Join(dir, "host"), filepath.Join(dir, "public"))
		if now.After(last) {
			last = now
			fmt.Println("gotsx: 源码有改动, 重新编译…")
			run()
		}
	}
}

func latestMtime(dirs ...string) time.Time {
	var latest time.Time
	for _, d := range dirs {
		filepath.WalkDir(d, func(p string, e os.DirEntry, err error) error {
			if err != nil || strings.Contains(p, "/.gen") {
				return nil
			}
			if info, err := e.Info(); err == nil && info.ModTime().After(latest) {
				latest = info.ModTime()
			}
			return nil
		})
	}
	return latest
}

// Tailwind: 有 app/tailwind.css 就用 standalone CLI(不需要 Node)扫描 app/**/*.tsx 生成 public/tailwind.css。
// 二进制查找顺序: $GOTSX_TAILWIND → <repo>/.tools/tailwindcss → PATH 里的 tailwindcss
func runTailwind(dir string) error {
	in := filepath.Join(dir, "app", "tailwind.css")
	if _, err := os.Stat(in); err != nil {
		return nil
	}
	bin := os.Getenv("GOTSX_TAILWIND")
	if bin == "" {
		for d := dir; d != "/" && d != "."; d = filepath.Dir(d) {
			if p := filepath.Join(d, ".tools", "tailwindcss"); fileExists(p) {
				bin = p
				break
			}
		}
	}
	if bin == "" {
		if p, err := exec.LookPath("tailwindcss"); err == nil {
			bin = p
		}
	}
	if bin == "" {
		fmt.Fprintln(os.Stderr, "gotsx: 找到 app/tailwind.css 但没有 tailwindcss 二进制(设置 GOTSX_TAILWIND 或放到 .tools/), 跳过")
		return nil
	}
	t := time.Now()
	os.MkdirAll(filepath.Join(dir, "public"), 0o755)
	cmd := exec.Command(bin, "-i", in, "-o", filepath.Join(dir, "public", "tailwind.css"), "--minify")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("tailwind 失败: %v\n%s", err, out)
	}
	fmt.Printf("gotsx: tailwind %s\n", time.Since(t).Round(time.Millisecond))
	return nil
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}
