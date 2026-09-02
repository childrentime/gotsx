package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func isWindows() bool { return runtime.GOOS == "windows" }

// tailwindAsset: the release asset name of the Tailwind v4 standalone CLI for this platform
func tailwindAsset() (string, error) {
	arch := map[string]string{"amd64": "x64", "arm64": "arm64"}[runtime.GOARCH]
	if arch == "" {
		return "", fmt.Errorf("no Tailwind standalone build for %s/%s; download it manually and set GOTSX_TAILWIND", runtime.GOOS, runtime.GOARCH)
	}
	switch runtime.GOOS {
	case "darwin":
		return "tailwindcss-macos-" + arch, nil
	case "linux":
		return "tailwindcss-linux-" + arch, nil
	case "windows":
		if arch != "x64" {
			return "", fmt.Errorf("no Tailwind standalone build for windows/%s", runtime.GOARCH)
		}
		return "tailwindcss-windows-x64.exe", nil
	}
	return "", fmt.Errorf("no Tailwind standalone build for %s; set GOTSX_TAILWIND", runtime.GOOS)
}

// downloadTailwind fetches the standalone binary into <dir>/tailwindcss[.exe] (cross-platform replacement for scripts/get-tailwind.sh)
func downloadTailwind(dir string) error {
	asset, err := tailwindAsset()
	if err != nil {
		return err
	}
	url := "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/" + asset
	dst := filepath.Join(dir, "tailwindcss"+exeSuffix())
	fmt.Printf("gotsx: downloading %s → %s\n", url, dst)
	cl := &http.Client{Timeout: 5 * time.Minute}
	resp, err := cl.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp := dst + ".part"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return err
	}
	f.Close()
	if err := os.Rename(tmp, dst); err != nil {
		return err
	}
	fmt.Println("gotsx: tailwind ready:", dst)
	return nil
}
