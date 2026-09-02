// Package host is the admin demo's host module: auth sessions + a user directory (CRUD).
// Writes (login / create / update / delete) are Go methods called from actions; the dialect only reads.
package host

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	gotsx "github.com/childrentime/gotsx/runtime"
)

// ---------- auth ----------

type Account struct {
	User string
	Pass string
	Name string
	Role string
}

type AuthModule struct {
	mu       sync.Mutex
	accounts map[string]Account
	sessions map[string]string // sid → user
}

type Session struct {
	User string `json:"user"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func (a *AuthModule) Login(user, pass string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	acc, ok := a.accounts[strings.TrimSpace(strings.ToLower(user))]
	if !ok || acc.Pass != pass {
		return "", fmt.Errorf("wrong username or password")
	}
	b := make([]byte, 16)
	rand.Read(b)
	sid := hex.EncodeToString(b)
	a.sessions[sid] = acc.User
	return sid, nil
}

func (a *AuthModule) Logout(sid string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.sessions, sid)
}

// Current: the signed-in session for the dialect (an empty Session when not signed in)
func (a *AuthModule) Current(sid string) Session {
	a.mu.Lock()
	defer a.mu.Unlock()
	if u, ok := a.sessions[sid]; ok {
		if acc, ok := a.accounts[u]; ok {
			return Session{User: acc.User, Name: acc.Name, Role: acc.Role}
		}
	}
	return Session{}
}

func (a *AuthModule) IsAuthed(sid string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	_, ok := a.sessions[sid]
	return ok
}

// ---------- user directory ----------

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`   // admin | editor | viewer
	Status   string `json:"status"` // active | disabled
	Dept     string `json:"dept"`
	Created  string `json:"created"`
	LastSeen string `json:"lastSeen"`
}

type UserPage struct {
	Items    []User `json:"items"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	Pages    int    `json:"pages"`
	PageList []int  `json:"pageList"`
	Active   int    `json:"active"`
	Admins   int    `json:"admins"`
}

type UsersModule struct {
	mu   sync.Mutex
	list []User
	seq  int
}

const usersPerPage = 8

func (m *UsersModule) counts() (active, admins int) {
	for _, u := range m.list {
		if u.Status == "active" {
			active++
		}
		if u.Role == "admin" {
			admins++
		}
	}
	return
}

// Query: read by the dialect / API; search, role filter, sort, pagination
func (m *UsersModule) Query(q, role, sortBy string, page int) UserPage {
	m.mu.Lock()
	defer m.mu.Unlock()
	q = strings.ToLower(strings.TrimSpace(q))
	var hit []User
	for _, u := range m.list {
		if role != "" && role != "all" && u.Role != role {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(u.Name+" "+u.Email+" "+u.Dept), q) {
			continue
		}
		hit = append(hit, u)
	}
	switch sortBy {
	case "name":
		sort.SliceStable(hit, func(i, j int) bool { return hit[i].Name < hit[j].Name })
	case "created":
		sort.SliceStable(hit, func(i, j int) bool { return hit[i].Created > hit[j].Created })
	default: // recent
		sort.SliceStable(hit, func(i, j int) bool { return hit[i].LastSeen > hit[j].LastSeen })
	}
	total := len(hit)
	pages := (total + usersPerPage - 1) / usersPerPage
	if pages == 0 {
		pages = 1
	}
	if page < 1 {
		page = 1
	}
	if page > pages {
		page = pages
	}
	a := (page - 1) * usersPerPage
	b := a + usersPerPage
	if b > total {
		b = total
	}
	var list []int
	for i := 1; i <= pages; i++ {
		list = append(list, i)
	}
	active, admins := m.counts()
	items := []User{}
	if a < b {
		items = append(items, hit[a:b]...)
	}
	return UserPage{Items: items, Total: total, Page: page, Pages: pages, PageList: list, Active: active, Admins: admins}
}

func (m *UsersModule) All() UserPage {
	m.mu.Lock()
	defer m.mu.Unlock()
	active, admins := m.counts()
	items := append([]User{}, m.list...)
	return UserPage{Items: items, Total: len(items), Page: 1, Pages: 1, Active: active, Admins: admins}
}

func (m *UsersModule) Get(id string) (User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.list {
		if u.ID == id {
			return u, nil
		}
	}
	return User{}, fmt.Errorf("%w: user %s", gotsx.ErrNotFound, id)
}

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
var validRole = map[string]bool{"admin": true, "editor": true, "viewer": true}

// validate: server-side field validation, returns field → message
func (m *UsersModule) validate(name, email, role, dept string, selfID string) map[string]string {
	errs := map[string]string{}
	if len(strings.TrimSpace(name)) < 2 {
		errs["name"] = "Name must be at least 2 characters"
	}
	if !emailRe.MatchString(strings.TrimSpace(email)) {
		errs["email"] = "Invalid email address"
	} else {
		for _, u := range m.list {
			if u.ID != selfID && strings.EqualFold(u.Email, strings.TrimSpace(email)) {
				errs["email"] = "This email is already in use"
			}
		}
	}
	if !validRole[role] {
		errs["role"] = "Invalid role"
	}
	if strings.TrimSpace(dept) == "" {
		errs["dept"] = "Department is required"
	}
	return errs
}

func (m *UsersModule) Create(name, email, role, dept string) (User, map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if errs := m.validate(name, email, role, dept, ""); len(errs) > 0 {
		return User{}, errs
	}
	m.seq++
	u := User{
		ID: fmt.Sprintf("u%03d", m.seq), Name: strings.TrimSpace(name), Email: strings.TrimSpace(email),
		Role: role, Status: "active", Dept: strings.TrimSpace(dept),
		Created: time.Now().Format("2006-01-02"), LastSeen: "just now",
	}
	m.list = append([]User{u}, m.list...)
	return u, nil
}

func (m *UsersModule) Update(id, name, email, role, dept, status string) (User, map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if errs := m.validate(name, email, role, dept, id); len(errs) > 0 {
		return User{}, errs
	}
	for i := range m.list {
		if m.list[i].ID == id {
			m.list[i].Name = strings.TrimSpace(name)
			m.list[i].Email = strings.TrimSpace(email)
			m.list[i].Role = role
			m.list[i].Dept = strings.TrimSpace(dept)
			if status == "active" || status == "disabled" {
				m.list[i].Status = status
			}
			return m.list[i], nil
		}
	}
	return User{}, map[string]string{"_": "User not found"}
}

func (m *UsersModule) Delete(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.list {
		if m.list[i].ID == id {
			m.list = append(m.list[:i], m.list[i+1:]...)
			return true
		}
	}
	return false
}

// ---------- dashboard stats ----------

type StatsModule struct{}

type Stat struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Delta string `json:"delta"`
	Up    bool   `json:"up"`
	Icon  string `json:"icon"`
}

func (StatsModule) Cards() []Stat {
	active, admins := Users.counts()
	Users.mu.Lock()
	total := len(Users.list)
	Users.mu.Unlock()
	return []Stat{
		{"Total users", fmt.Sprintf("%d", total), "+12%", true, "👥"},
		{"Active users", fmt.Sprintf("%d", active), "+8%", true, "✅"},
		{"Admins", fmt.Sprintf("%d", admins), "0%", true, "🛡️"},
		{"New this month", "23", "+34%", true, "📈"},
	}
}

type Activity struct {
	Who  string `json:"who"`
	What string `json:"what"`
	When string `json:"when"`
	Icon string `json:"icon"`
}

func (StatsModule) Recent() []Activity {
	return []Activity{
		{"Alice Chen", "created user “Liam Lee”", "2 minutes ago", "➕"},
		{"Fiona Wang", "changed permissions for “Sales”", "18 minutes ago", "✏️"},
		{"System", "auto-disabled 3 dormant accounts", "1 hour ago", "🔒"},
		{"Mia Zhao", "exported the user report", "3 hours ago", "📤"},
		{"Alice Chen", "signed in to the console", "5 hours ago", "🔑"},
	}
}

// ---------- seed data ----------

var (
	Auth  = &AuthModule{accounts: seedAccounts(), sessions: map[string]string{}}
	Users = &UsersModule{list: seedUsers()}
	Stats = StatsModule{}
)

func seedAccounts() map[string]Account {
	return map[string]Account{
		"admin": {User: "admin", Pass: "admin123", Name: "Super Admin", Role: "admin"},
		"demo":  {User: "demo", Pass: "demo", Name: "Demo Account", Role: "viewer"},
	}
}

func NewSID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func seedUsers() []User {
	names := []string{"Liam Lee", "Hannah Meyer", "Zack Wang", "Fiona Wang", "Leo Liu", "Claire Chen", "Frank Yang", "Mia Zhao", "Sam Sun", "Jay Zhou", "Quinn Wu", "Sophie Zheng", "Felix Feng", "Wendy Jiang", "Kai Shen", "Henry Han", "Xavier Xu", "Chad Deng", "Cody Cao", "Bing Fan"}
	depts := []string{"Engineering", "Sales", "Marketing", "HR", "Finance", "Operations"}
	roles := []string{"admin", "editor", "viewer", "editor", "viewer", "viewer"}
	out := make([]User, len(names))
	for i, n := range names {
		status := "active"
		if i%7 == 6 {
			status = "disabled"
		}
		out[i] = User{
			ID: fmt.Sprintf("u%03d", i+1), Name: n,
			Email: fmt.Sprintf("%s@gotsx.dev", []string{"lilei", "hmm", "zhw", "wf", "ly", "cj", "yf", "zm", "sh", "zj", "wq", "zs", "fl", "jw", "sk", "hh", "xw", "dc", "cc", "fb"}[i]),
			Role:  roles[i%len(roles)], Status: status, Dept: depts[i%len(depts)],
			Created: fmt.Sprintf("2026-0%d-1%d", (i%8)+1, i%10), LastSeen: []string{"just now", "5 minutes ago", "1 hour ago", "today", "yesterday", "3 days ago"}[i%6],
		}
	}
	return out
}

var Registry = map[string]gotsx.HostModule{
	"auth":  {Value: Auth, Go: "host.Auth"},
	"users": {Value: Users, Go: "host.Users"},
	"stats": {Value: Stats, Go: "host.Stats"},
}
