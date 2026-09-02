// Package host: 后台管理应用的宿主模块 —— 认证会话 + 用户目录(CRUD)。
// 写操作(登录/增删改)是 Go 方法, 由 action 调用, 方言侧只读。
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

	gotsx "gotsx/runtime"
)

// ---------- 认证 ----------

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
		return "", fmt.Errorf("用户名或密码错误")
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

// Current: 供 dialect 读取当前登录者(未登录返回空 Session)
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

// ---------- 用户目录 ----------

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

// Query: 供 dialect / api 读取, 支持搜索/角色过滤/排序/分页
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
	return User{}, fmt.Errorf("%w: 用户 %s", gotsx.ErrNotFound, id)
}

var emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
var validRole = map[string]bool{"admin": true, "editor": true, "viewer": true}

// Validate: 服务端字段校验, 返回 字段→错误信息
func (m *UsersModule) validate(name, email, role, dept string, selfID string) map[string]string {
	errs := map[string]string{}
	if len(strings.TrimSpace(name)) < 2 {
		errs["name"] = "姓名至少 2 个字"
	}
	if !emailRe.MatchString(strings.TrimSpace(email)) {
		errs["email"] = "邮箱格式不正确"
	} else {
		for _, u := range m.list {
			if u.ID != selfID && strings.EqualFold(u.Email, strings.TrimSpace(email)) {
				errs["email"] = "该邮箱已被占用"
			}
		}
	}
	if !validRole[role] {
		errs["role"] = "角色不合法"
	}
	if strings.TrimSpace(dept) == "" {
		errs["dept"] = "请填写部门"
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
		Created: time.Now().Format("2006-01-02"), LastSeen: "刚刚",
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
	return User{}, map[string]string{"_": "用户不存在"}
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

// ---------- 仪表盘统计 ----------

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
		{"总用户数", fmt.Sprintf("%d", total), "+12%", true, "👥"},
		{"活跃用户", fmt.Sprintf("%d", active), "+8%", true, "✅"},
		{"管理员", fmt.Sprintf("%d", admins), "0%", true, "🛡️"},
		{"本月新增", "23", "+34%", true, "📈"},
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
		{"陈晓", "创建了用户 «李雷»", "2 分钟前", "➕"},
		{"王芳", "修改了 «销售部» 的权限", "18 分钟前", "✏️"},
		{"系统", "自动禁用了 3 个长期未登录账号", "1 小时前", "🔒"},
		{"赵敏", "导出了用户报表", "3 小时前", "📤"},
		{"陈晓", "登录了后台", "5 小时前", "🔑"},
	}
}

// ---------- 种子数据 ----------

var (
	Auth  = &AuthModule{accounts: seedAccounts(), sessions: map[string]string{}}
	Users = &UsersModule{list: seedUsers()}
	Stats = StatsModule{}
)

func seedAccounts() map[string]Account {
	return map[string]Account{
		"admin": {User: "admin", Pass: "admin123", Name: "超级管理员", Role: "admin"},
		"demo":  {User: "demo", Pass: "demo", Name: "演示账号", Role: "viewer"},
	}
}

func NewSID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func seedUsers() []User {
	names := []string{"李雷", "韩梅梅", "张伟", "王芳", "刘洋", "陈静", "杨帆", "赵敏", "孙浩", "周杰", "吴强", "郑爽", "冯磊", "蒋雯", "沈括", "韩寒", "许巍", "邓超", "曹操", "范冰"}
	depts := []string{"研发部", "销售部", "市场部", "人事部", "财务部", "运营部"}
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
			Created: fmt.Sprintf("2026-0%d-1%d", (i%8)+1, i%10), LastSeen: []string{"刚刚", "5 分钟前", "1 小时前", "今天", "昨天", "3 天前"}[i%6],
		}
	}
	return out
}

var Registry = map[string]gotsx.HostModule{
	"auth":  {Value: Auth, Go: "host.Auth"},
	"users": {Value: Users, Go: "host.Users"},
	"stats": {Value: Stats, Go: "host.Stats"},
}
