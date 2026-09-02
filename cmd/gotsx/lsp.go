package main

// A small Language Server (LSP over stdio): on every open/change/save it re-analyzes the app the file
// belongs to (parse + type-check + both backends, in memory, with unsaved buffers overlaid) and publishes
// diagnostics. That is the dialect-specific half editors lack; TypeScript's own tooling keeps doing
// completion and navigation via app/.gen/*.d.ts.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/childrentime/gotsx/compiler"
)

type rpcMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type lspServer struct {
	out       *bufio.Writer
	mu        sync.Mutex
	docs      map[string]string            // path → unsaved text
	published map[string][]string          // appDir → files that currently carry diagnostics
	checkers  map[string]*compiler.Checker // appDir → last analysis (hover / definition answer from it)
}

func runLSP() error {
	s := &lspServer{out: bufio.NewWriter(os.Stdout), docs: map[string]string{}, published: map[string][]string{}, checkers: map[string]*compiler.Checker{}}
	in := bufio.NewReader(os.Stdin)
	for {
		msg, err := readMessage(in)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if done := s.handle(msg); done {
			return nil
		}
	}
}

func readMessage(r *bufio.Reader) (*rpcMessage, error) {
	length := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if k, v, ok := strings.Cut(line, ":"); ok && strings.EqualFold(strings.TrimSpace(k), "Content-Length") {
			length, _ = strconv.Atoi(strings.TrimSpace(v))
		}
	}
	if length < 0 {
		return nil, fmt.Errorf("lsp: missing Content-Length")
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	var m rpcMessage
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("lsp: bad message: %w", err)
	}
	return &m, nil
}

func (s *lspServer) send(v any) {
	b, _ := json.Marshal(v)
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Fprintf(s.out, "Content-Length: %d\r\n\r\n", len(b))
	s.out.Write(b)
	s.out.Flush()
}

func (s *lspServer) reply(id json.RawMessage, result any) {
	s.send(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
}

func (s *lspServer) replyErr(id json.RawMessage, code int, msg string) {
	s.send(map[string]any{"jsonrpc": "2.0", "id": id, "error": rpcError{Code: code, Message: msg}})
}

func (s *lspServer) notify(method string, params any) {
	s.send(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

// handle returns true when the client asked us to exit
func (s *lspServer) handle(m *rpcMessage) bool {
	switch m.Method {
	case "initialize":
		s.reply(m.ID, map[string]any{
			"capabilities": map[string]any{
				"textDocumentSync":   map[string]any{"openClose": true, "change": 1, "save": map[string]any{"includeText": false}},
				"hoverProvider":      true,
				"definitionProvider": true,
			},
			"serverInfo": map[string]any{"name": "gotsx", "version": version()},
		})
	case "initialized", "$/setTrace", "$/cancelRequest", "workspace/didChangeConfiguration", "workspace/didChangeWatchedFiles":
	case "shutdown":
		s.reply(m.ID, nil)
	case "exit":
		return true
	case "textDocument/didOpen":
		var p struct {
			TextDocument struct {
				URI  string `json:"uri"`
				Text string `json:"text"`
			} `json:"textDocument"`
		}
		json.Unmarshal(m.Params, &p)
		if path := uriToPath(p.TextDocument.URI); path != "" {
			s.docs[path] = p.TextDocument.Text
			s.analyze(path)
		}
	case "textDocument/didChange":
		var p struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
			ContentChanges []struct {
				Text string `json:"text"`
			} `json:"contentChanges"`
		}
		json.Unmarshal(m.Params, &p)
		if path := uriToPath(p.TextDocument.URI); path != "" && len(p.ContentChanges) > 0 {
			s.docs[path] = p.ContentChanges[len(p.ContentChanges)-1].Text // full sync
			s.analyze(path)
		}
	case "textDocument/didSave":
		var p struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
		}
		json.Unmarshal(m.Params, &p)
		if path := uriToPath(p.TextDocument.URI); path != "" {
			delete(s.docs, path) // disk is now the truth (and host.json may have been regenerated)
			s.analyze(path)
		}
	case "textDocument/didClose":
		var p struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
		}
		json.Unmarshal(m.Params, &p)
		if path := uriToPath(p.TextDocument.URI); path != "" {
			delete(s.docs, path)
			s.analyze(path)
		}
	case "textDocument/hover", "textDocument/definition":
		var p struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
			Position struct {
				Line      int `json:"line"`
				Character int `json:"character"`
			} `json:"position"`
		}
		json.Unmarshal(m.Params, &p)
		path := uriToPath(p.TextDocument.URI)
		c := s.checkerFor(path)
		if c == nil {
			s.reply(m.ID, nil)
			return false
		}
		line, col := p.Position.Line+1, p.Position.Character+1
		if m.Method == "textDocument/hover" {
			h := c.HoverAt(path, line, col)
			if h == nil {
				s.reply(m.ID, nil)
				return false
			}
			s.reply(m.ID, map[string]any{"contents": map[string]any{"kind": "markdown", "value": h.Text}})
			return false
		}
		def := c.DefinitionAt(path, line, col)
		if def == nil {
			s.reply(m.ID, nil)
			return false
		}
		pos := map[string]int{"line": def.Line - 1, "character": max(def.Col-1, 0)}
		s.reply(m.ID, map[string]any{"uri": pathToURI(def.File), "range": map[string]any{"start": pos, "end": pos}})
	default:
		if len(m.ID) > 0 { // unknown request: answer so the client does not hang
			s.replyErr(m.ID, -32601, "method not found: "+m.Method)
		}
	}
	return false
}

// checkerFor: the last analysis of the app the file belongs to (analyzing now if there is none yet)
func (s *lspServer) checkerFor(path string) *compiler.Checker {
	appDir := appDirOf(path)
	if appDir == "" {
		return nil
	}
	if c := s.checkers[appDir]; c != nil {
		return c
	}
	s.analyze(path)
	return s.checkers[appDir]
}

// appDirOf: the app/ directory a source file belongs to (nearest ancestor named "app"), "" if none
func appDirOf(path string) string {
	for d := filepath.Dir(path); ; d = filepath.Dir(d) {
		if filepath.Base(d) == "app" {
			return d
		}
		if filepath.Dir(d) == d {
			return ""
		}
	}
}

func (s *lspServer) analyze(path string) {
	appDir := appDirOf(path)
	if appDir == "" {
		return
	}
	overlay := map[string]string{}
	for p, text := range s.docs {
		if strings.HasPrefix(p, appDir+string(filepath.Separator)) {
			overlay[p] = text
		}
	}
	diags, err := compiler.Analyze(appDir, overlay)
	if c, _, lerr := compiler.Load(appDir, overlay); lerr == nil { // keep the checked AST for hover / definition
		s.checkers[appDir] = c
	}
	byFile := map[string][]map[string]any{}
	if err != nil { // the whole app failed to load: pin the error on the current file
		byFile[path] = append(byFile[path], lspDiag(1, 1, err.Error()))
	}
	for _, d := range diags {
		f := d.File
		if f == "" {
			f = path
		}
		byFile[f] = append(byFile[f], lspDiag(d.Line, d.Col, d.Msg))
	}
	// clear files that had diagnostics last time but are clean now
	for _, f := range s.published[appDir] {
		if _, still := byFile[f]; !still {
			byFile[f] = []map[string]any{}
		}
	}
	var now []string
	files := make([]string, 0, len(byFile))
	for f := range byFile {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, f := range files {
		ds := byFile[f]
		s.notify("textDocument/publishDiagnostics", map[string]any{"uri": pathToURI(f), "diagnostics": ds})
		if len(ds) > 0 {
			now = append(now, f)
		}
	}
	s.published[appDir] = now
}

func lspDiag(line, col int, msg string) map[string]any {
	if line < 1 {
		line = 1
	}
	if col < 1 {
		col = 1
	}
	pos := map[string]int{"line": line - 1, "character": col - 1}
	return map[string]any{
		"range":    map[string]any{"start": pos, "end": pos},
		"severity": 1,
		"source":   "gotsx",
		"message":  msg,
	}
}

func uriToPath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return ""
	}
	p := u.Path
	if runtime.GOOS == "windows" && strings.HasPrefix(p, "/") && len(p) > 2 && p[2] == ':' {
		p = p[1:] // /C:/x → C:/x
	}
	return filepath.Clean(filepath.FromSlash(p))
}

func pathToURI(p string) string {
	p = filepath.ToSlash(p)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p // windows drive
	}
	return (&url.URL{Scheme: "file", Path: p}).String()
}
