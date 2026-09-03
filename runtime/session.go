package gotsx

// Signed-cookie sessions: the value is JSON signed with HMAC-SHA256(Options.SessionSecret), HttpOnly + SameSite=Lax.
// Pages read it through PageProps.session / flash / csrf; actions read and write it through req.Session(); changes save automatically.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

const sessionCookie = "gotsx_session"

// Flash: a one-shot message (gone once read); pages receive it in PageProps.flash
type Flash struct {
	Kind string `json:"kind"` // ok | error | info …
	Text string `json:"text"`
}

type sessionData struct {
	Values map[string]string `json:"v,omitempty"`
	Flash  []Flash           `json:"f,omitempty"`
	CSRF   string            `json:"c,omitempty"`
}

// Session is the session view for one request; Set / Delete / Flash / Clear mark it modified so it is saved before the response
type Session struct {
	data     sessionData
	modified bool
}

func (s *Session) Get(k string) string { return s.data.Values[k] }
func (s *Session) Set(k, v string) {
	if s.data.Values == nil {
		s.data.Values = map[string]string{}
	}
	s.data.Values[k] = v
	s.modified = true
}
func (s *Session) Delete(k string) {
	delete(s.data.Values, k)
	s.modified = true
}
func (s *Session) Clear() {
	s.data = sessionData{}
	s.modified = true
}

// Flash queues a message; it appears in PageProps.flash on the next page render
func (s *Session) Flash(kind, text string) {
	s.data.Flash = append(s.data.Flash, Flash{Kind: kind, Text: text})
	s.modified = true
}

// Values: all key/value pairs (a read-only copy; nil for an empty session — reading a missing key yields the zero value, no allocation)
func (s *Session) Values() map[string]string {
	if len(s.data.Values) == 0 {
		return nil
	}
	out := make(map[string]string, len(s.data.Values))
	for k, v := range s.data.Values {
		out[k] = v
	}
	return out
}

// CSRF: this session's CSRF token (generated on first use), for classic <form method="post">: <input type="hidden" name="_csrf" value={csrf} />
func (s *Session) CSRF() string {
	if s.data.CSRF == "" {
		s.data.CSRF = randomHex(16)
		s.modified = true
	}
	return s.data.CSRF
}

func (s *server) sessionKey() []byte {
	if s.opt.SessionSecret == "" {
		return []byte(s.autoSecret) // generated once in newServer (read-only here: request goroutines share it)
	}
	return []byte(s.opt.SessionSecret)
}

func (s *server) sign(payload string) string {
	m := hmac.New(sha256.New, s.sessionKey())
	m.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// loadSession parses and verifies the cookie; a bad or missing one → an empty session
func (s *server) loadSession(r *http.Request) *Session {
	sess := &Session{}
	ck, err := r.Cookie(sessionCookie)
	if err != nil || ck.Value == "" {
		return sess
	}
	payload, sig, ok := strings.Cut(ck.Value, ".")
	if !ok || !hmac.Equal([]byte(s.sign(payload)), []byte(sig)) {
		return sess
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil || json.Unmarshal(raw, &sess.data) != nil {
		sess.data = sessionData{}
	}
	return sess
}

func (sess *Session) save(w http.ResponseWriter, r *http.Request) {
	if !sess.modified {
		return
	}
	srv := serverOf(r)
	if srv == nil {
		return
	}
	raw, _ := json.Marshal(sess.data)
	payload := base64.RawURLEncoding.EncodeToString(raw)
	if srv.opt.SessionSecret == "" {
		srv.secretWarn.Do(func() {
			log.Printf("gotsx: a session cookie was issued with a per-process random key (Options.SessionSecret is empty): sessions, flash messages and CSRF tokens will not survive a restart or be shared between replicas — set SESSION_SECRET in production")
		})
	}
	if len(payload) > 3800 { // browsers drop cookies over ~4 KB: the session would silently vanish
		log.Printf("gotsx: session cookie is %d bytes (limit ~4 KB): keep sessions small, store data on the server", len(payload))
	}
	maxAge := srv.opt.SessionMaxAge
	if maxAge == 0 {
		maxAge = 30 * 24 * time.Hour
	}
	secure := IsHTTPS(r)
	empty := len(sess.data.Values) == 0 && len(sess.data.Flash) == 0 && sess.data.CSRF == ""
	c := &http.Cookie{Name: sessionCookie, Value: payload + "." + srv.sign(payload), Path: "/", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: int(maxAge / time.Second)}
	if empty {
		c.Value, c.MaxAge = "", -1
	}
	http.SetCookie(w, c)
	sess.modified = false
}

// VerifyCSRF checks a classic form POST — the _csrf field or the X-CSRF-Token header must equal the session's token
func VerifyCSRF(r *http.Request) bool {
	srv := serverOf(r)
	if srv == nil {
		return false
	}
	want := srv.loadSession(r).data.CSRF
	if want == "" {
		return false
	}
	got := r.Header.Get("X-CSRF-Token")
	if got == "" {
		got = r.FormValue("_csrf")
	}
	return hmac.Equal([]byte(got), []byte(want))
}

// SessionOf reads/writes the session inside a plain Options.Actions handler; call Save before writing the body
func SessionOf(r *http.Request) *Session {
	if srv := serverOf(r); srv != nil {
		return srv.loadSession(r)
	}
	return &Session{}
}

// Save writes the session back to the response (for Options.Actions handlers; action methods save automatically)
func (s *Session) Save(w http.ResponseWriter, r *http.Request) { s.save(w, r) }

func serverOf(r *http.Request) *server {
	s, _ := r.Context().Value(ctxServer).(*server)
	return s
}

// IsHTTPS reports whether the request arrived over TLS, directly or through a proxy / CDN
// (X-Forwarded-Proto, Forwarded: proto=https, Cloudflare's CF-Visitor). Use it for the Secure flag of your own cookies.
func IsHTTPS(r *http.Request) bool {
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		return true
	}
	if fwd := r.Header.Get("Forwarded"); strings.Contains(strings.ToLower(fwd), "proto=https") {
		return true
	}
	return strings.Contains(r.Header.Get("CF-Visitor"), `"https"`)
}
