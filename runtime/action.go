package gotsx

// Typed actions: a host module lists Actions in its Registry entry and the compiler generates (a) the typed
// call inside islands — await toggle(id) becomes POST /_gotsx/act/<module>/<name> — and (b) gen.HostActions,
// where the server decodes the arguments by Go type and calls the method.
// This file is the runtime half: routing, same-origin checks, argument decoding, error → HTTP status, session saving.

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
)

// Req is the optional first parameter of an action method (*gotsx.Req): the current request, cookies, locale, session.
// Signature: func (m *M) Toggle(req *gotsx.Req, id string) (Todo, error) — req does not appear in the TS signature.
type Req struct {
	W       http.ResponseWriter
	R       *http.Request
	Cookies map[string]string
	Locale  string
	sess    *Session
	srv     *server
}

// Session returns the signed-cookie session (see session.go); changes are saved automatically before the response is written
func (q *Req) Session() *Session {
	if q.sess == nil {
		if q.srv == nil || q.R == nil { // a Req built by hand (unit tests): a detached in-memory session
			q.sess = &Session{}
		} else {
			q.sess = q.srv.loadSession(q.R)
		}
	}
	return q.sess
}

// SetCookie sets a cookie on the response
func (q *Req) SetCookie(c *http.Cookie) { http.SetCookie(q.W, c) }

// HostAction is one action registered by generated code (gen.HostActions); pass the list to Options.HostActions
type HostAction struct {
	Module string
	Name   string
	Fn     func(req *Req, args []json.RawMessage) (any, error)
}

// ValidationError: a business / validation failure — in an island the caught e.fields maps field → message (HTTP 422);
// a message alone is 400, or the status code set in Status
type ValidationError struct {
	Message string
	Fields  map[string]string
	Status  int // 0 → 422 with fields / 400 without; Unauthorized / Forbidden set 401 / 403
}

// Error returns the message with the field errors appended, readable when a form handler shows it directly
// ("validation failed: title: Title is required")
func (e *ValidationError) Error() string {
	if len(e.Fields) == 0 {
		return e.Message
	}
	keys := make([]string, 0, len(e.Fields))
	for k := range e.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+": "+e.Fields[k])
	}
	if e.Message == "" || e.Message == "validation failed" {
		return strings.Join(parts, "; ")
	}
	return e.Message + ": " + strings.Join(parts, "; ")
}

// Invalid: a set of field errors (the message defaults to "validation failed")
func Invalid(fields map[string]string) error {
	return &ValidationError{Message: "validation failed", Fields: fields}
}

// Fail: a business error with a message (HTTP 400); wrap gotsx.ErrNotFound with fmt.Errorf("%w", …) for a 404
func Fail(msg string) error { return &ValidationError{Message: msg} }

// Unauthorized: not signed in (HTTP 401); Forbidden: no permission (HTTP 403) — for authorization inside actions
func Unauthorized(msg string) error {
	return &ValidationError{Message: msg, Status: http.StatusUnauthorized}
}
func Forbidden(msg string) error { return &ValidationError{Message: msg, Status: http.StatusForbidden} }

// Arg decodes the i-th argument (used by generated code)
func Arg[T any](args []json.RawMessage, i int, dst *T) error {
	if i >= len(args) {
		return &ValidationError{Message: fmt.Sprintf("missing argument %d", i)}
	}
	if err := json.Unmarshal(args[i], dst); err != nil {
		return &ValidationError{Message: fmt.Sprintf("argument %d: %v", i, err)}
	}
	return nil
}

const actionPrefix = "/_gotsx/act/"

// actionHandler: POST /_gotsx/act/{module}/{name} → same-origin + header checks → decode → call → JSON
func (s *server) actionHandler(acts map[string]HostAction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		key := strings.TrimPrefix(r.URL.Path, actionPrefix)
		act, ok := acts[key]
		if !ok || r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, `{"error":"unknown action"}`)
			return
		}
		// The compiled stub always sends X-Gotsx-Action; requiring it (even with DisableCSRF, which only relaxes the
		// origin check) keeps HTML forms — which cannot set headers — from reaching actions cross-site.
		if r.Header.Get("X-Gotsx-Action") == "" || (!s.opt.DisableCSRF && !SameOrigin(r)) {
			w.WriteHeader(http.StatusForbidden)
			io.WriteString(w, `{"error":"cross-origin request rejected"}`)
			return
		}
		var args []json.RawMessage
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&args); err != nil && err != io.EOF {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, `{"error":"bad request body"}`)
			return
		}
		req := &Req{W: w, R: r, Cookies: cookieMap(r), Locale: s.localeOf(r), srv: s}
		res, err := func() (res any, err error) {
			defer func() {
				if rec := recover(); rec != nil {
					if e, ok := rec.(error); ok {
						err = e
					} else {
						err = fmt.Errorf("panic: %v", rec)
					}
				}
			}()
			return act.Fn(req, args)
		}()
		if req.sess != nil {
			req.sess.save(w, r) // Set-Cookie must precede the body
		}
		if err != nil {
			s.writeActionError(w, r, key, err)
			return
		}
		body, _ := json.Marshal(map[string]any{"ok": true, "data": res})
		w.Write(body)
	}
}

func (s *server) writeActionError(w http.ResponseWriter, r *http.Request, key string, err error) {
	var ve *ValidationError
	var he *HostError
	switch {
	case errors.As(err, &ve):
		status := http.StatusBadRequest
		if len(ve.Fields) > 0 {
			status = http.StatusUnprocessableEntity
		}
		if ve.Status != 0 {
			status = ve.Status
		}
		w.WriteHeader(status)
		b, _ := json.Marshal(map[string]any{"error": ve.Message, "fields": ve.Fields})
		w.Write(b)
	case errors.Is(err, ErrNotFound) || (errors.As(err, &he) && errors.Is(he.Err, ErrNotFound)):
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"error":"not found"}`)
	default:
		id, _ := r.Context().Value(ctxReqID).(string)
		log.Printf("action %s failed id=%s: %v", key, id, err)
		w.WriteHeader(http.StatusInternalServerError)
		msg := "internal error"
		if s.opt.Dev {
			msg = err.Error()
		}
		b, _ := json.Marshal(map[string]any{"error": msg})
		w.Write(b)
	}
}

func cookieMap(r *http.Request) map[string]string {
	m := map[string]string{}
	for _, ck := range r.Cookies() {
		m[ck.Name] = ck.Value
	}
	return m
}

// localeOf: the same locale resolution pages use (for t() inside actions)
func (s *server) localeOf(r *http.Request) string {
	if i18nCfg == nil {
		return ""
	}
	lang := ""
	if ck, err := r.Cookie("lang"); err == nil {
		lang = ck.Value
	}
	ref := r.Header.Get("Referer")
	path := "/"
	if i := strings.Index(ref, "://"); i >= 0 {
		if j := strings.IndexByte(ref[i+3:], '/'); j >= 0 {
			path = ref[i+3+j:]
		}
	}
	loc, _ := resolveLocale(path, lang, r.Header.Get("Accept-Language"))
	return loc
}
