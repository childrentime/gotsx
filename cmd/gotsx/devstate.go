package main

// The visible state of gotsx dev (for editors, agents and a second gotsx dev):
//   .gotsx/dev.json         {pid, port, url, started}; removed when the process exits
//   .gotsx/diagnostics.json diagnostics of a failed build (browser overlay + agents read it); removed on success

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/childrentime/gotsx/compiler"
)

type devState struct {
	PID     int    `json:"pid"`
	Port    int    `json:"port"`
	URL     string `json:"url"`
	Started string `json:"started"`
}

func devStatePath(dir string) string { return filepath.Join(dir, ".gotsx", "dev.json") }
func diagPath(dir string) string     { return filepath.Join(dir, ".gotsx", "diagnostics.json") }

// readDevState: only a running gotsx dev (live pid) counts
func readDevState(dir string) (*devState, bool) {
	raw, err := os.ReadFile(devStatePath(dir))
	if err != nil {
		return nil, false
	}
	var st devState
	if json.Unmarshal(raw, &st) != nil || st.PID == 0 {
		return nil, false
	}
	return &st, pidAlive(st.PID)
}

func pidAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if isWindows() {
		return true // FindProcess already checked
	}
	err = p.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}

func writeDevState(dir string, port int) error {
	if err := os.MkdirAll(filepath.Join(dir, ".gotsx"), 0o755); err != nil {
		return err
	}
	st := devState{PID: os.Getpid(), Port: port, URL: fmt.Sprintf("http://localhost:%d", port), Started: time.Now().Format(time.RFC3339)}
	raw, _ := json.MarshalIndent(st, "", "  ")
	return os.WriteFile(devStatePath(dir), raw, 0o644)
}

// portOf: -addr :3000 / --addr=:3000 from the app arguments; default 3000
func portOf(appArgs []string) int {
	addr := ":3000"
	for i, a := range appArgs {
		switch {
		case (a == "-addr" || a == "--addr") && i+1 < len(appArgs):
			addr = appArgs[i+1]
		case strings.HasPrefix(a, "-addr=") || strings.HasPrefix(a, "--addr="):
			addr = a[strings.Index(a, "=")+1:]
		}
	}
	var port int
	if i := strings.LastIndexByte(addr, ':'); i >= 0 {
		fmt.Sscanf(addr[i+1:], "%d", &port)
	}
	if port == 0 {
		port = 3000
	}
	return port
}

// goErrLine: `path/file.go:12:5: message` — the path may carry a Windows drive letter (C:\x\file.go)
var goErrLine = regexp.MustCompile(`^((?:[A-Za-z]:)?[^:]+\.go):(\d+):(\d+):\s*(.*)$`)

type diagFile struct {
	Title  string                `json:"title"`
	Errors []compiler.Diagnostic `json:"errors"`
	Text   string                `json:"text"`
}

// writeDiagnostics turns a failed build into structured diagnostics. err is the build() error (dialect diagnostics / hostgen / go build output)
func writeDiagnostics(dir string, title string, err error) {
	df := diagFile{Title: title, Text: err.Error()}
	if diags, aerr := compiler.Analyze(filepath.Join(dir, "app"), nil); aerr == nil && len(diags) > 0 {
		df.Errors = diags
	}
	if len(df.Errors) == 0 {
		df.Errors = parseGoErrors(dir, err.Error())
	}
	if len(df.Errors) == 0 { // a hostgen panic: the `panic: hostgen: …` line says what is wrong, the goroutine dump does not
		for _, ln := range strings.Split(err.Error(), "\n") {
			if strings.HasPrefix(ln, "panic: ") {
				df.Errors = []compiler.Diagnostic{{Msg: strings.TrimPrefix(ln, "panic: ")}}
				break
			}
		}
	}
	if len(df.Errors) == 0 {
		df.Errors = []compiler.Diagnostic{{Msg: strings.TrimSpace(err.Error())}}
	}
	os.MkdirAll(filepath.Join(dir, ".gotsx"), 0o755)
	raw, _ := json.MarshalIndent(df, "", "  ")
	os.WriteFile(diagPath(dir), raw, 0o644)
}

func clearDiagnostics(dir string) { os.Remove(diagPath(dir)) }

// parseGoErrors: the file.go:12:34: message lines in go build / hostgen output
func parseGoErrors(dir, text string) []compiler.Diagnostic {
	var out []compiler.Diagnostic
	for _, ln := range strings.Split(text, "\n") {
		m := goErrLine.FindStringSubmatch(strings.TrimSpace(ln))
		if m == nil {
			continue
		}
		var line, col int
		fmt.Sscanf(m[2]+" "+m[3], "%d %d", &line, &col)
		parts := []string{m[1], m[2], m[3], m[4]}
		file := parts[0]
		if !filepath.IsAbs(file) {
			file = filepath.Join(dir, file)
		}
		out = append(out, compiler.Diagnostic{File: file, Line: line, Col: col, Msg: strings.TrimSpace(parts[3])})
	}
	return out
}
