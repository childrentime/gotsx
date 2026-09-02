// Package host 是示例应用的宿主模块: 方言里 `import { models } from "host:data"` 的 Go 一侧。
// 编译后是直接的 Go 调用, 没有任何编组。
package host

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	gotsx "github.com/childrentime/gotsx/runtime"
)

type Model struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Author    string   `json:"author"`
	Desc      string   `json:"desc"`
	Likes     int      `json:"likes"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"createdAt"`
}

type ModelStore struct {
	mu    sync.Mutex
	items []Model
}

func (s *ModelStore) List() []Model {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Model, len(s.items))
	copy(out, s.items)
	return out
}

func (s *ModelStore) Search(q string) []Model {
	q = strings.ToLower(strings.TrimSpace(q))
	all := s.List()
	if q == "" {
		return all
	}
	out := []Model{}
	for _, m := range all {
		if strings.Contains(strings.ToLower(m.Title+" "+m.Desc+" "+strings.Join(m.Tags, " ")), q) {
			out = append(out, m)
		}
	}
	return out
}

func (s *ModelStore) Get(id string) (Model, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, m := range s.items {
		if m.ID == id {
			return m, nil
		}
	}
	return Model{}, fmt.Errorf("%w: model %q", gotsx.ErrNotFound, id)
}

func (s *ModelStore) Like(id string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == id {
			s.items[i].Likes++
			return s.items[i].Likes, nil
		}
	}
	return 0, fmt.Errorf("%w: model %q", gotsx.ErrNotFound, id)
}

type DataModule struct {
	Models *ModelStore `json:"models"`
}

// Stats / Trending: 慢查询(模拟数据库延迟), 页面把它们放进 <Suspense> —— 外壳先到, 这两个在各自的 goroutine 里并发填充
type Stats struct {
	Total int `json:"total"`
	Likes int `json:"likes"`
}

func (d *DataModule) Stats() Stats {
	time.Sleep(600 * time.Millisecond)
	var s Stats
	for _, m := range d.Models.List() {
		s.Total++
		s.Likes += m.Likes
	}
	return s
}

func (d *DataModule) Trending() []Model {
	time.Sleep(300 * time.Millisecond)
	all := d.Models.List()
	sort.Slice(all, func(i, j int) bool { return all[i].Likes > all[j].Likes })
	if len(all) > 3 {
		all = all[:3]
	}
	return all
}

type IntlModule struct{}

func (IntlModule) FmtNumber(n int) string {
	s := fmt.Sprint(n)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

func (IntlModule) FmtDate(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	return t.Format("2006年1月2日")
}

func (IntlModule) Now() string { return time.Now().Format("15:04:05.000") }

// 生成代码引用的入口
var (
	Data = &DataModule{Models: &ModelStore{items: seed()}}
	Intl = IntlModule{}
)

// Registry: 模块名 → 值 + 生成代码里的 Go 表达式
var Registry = map[string]gotsx.HostModule{
	"data": {Value: Data, Go: "host.Data"},
	"intl": {Value: Intl, Go: "host.Intl"},
}

func seed() []Model {
	return []Model{
		{"m1", "Foldable Phone Stand", "ziwei", "Prints in one piece, no supports; 0.3 mm hinge clearance", 1287, []string{"practical", "no-supports"}, "2026-08-02T10:00:00Z"},
		{"m2", "Parametric Cable Box", "lin", "OpenSCAD parametric, three sizes", 402, []string{"storage", "parametric"}, "2026-08-10T09:30:00Z"},
		{"m3", "AMS Desiccant Tray", "maker_wu", "Fits the AMS bay, snap-fit", 3310, []string{"Bambu", "AMS"}, "2026-07-21T14:12:00Z"},
		{"m4", "Low-poly Shiba", "kyo", "Multi-colour, 0.12 mm layers recommended", 958, []string{"decor", "multi-colour"}, "2026-08-18T20:45:00Z"},
		{"m5", "Hex Drawer Module", "ziwei", "Endlessly tileable desk storage", 2201, []string{"storage", "modular"}, "2026-06-30T08:00:00Z"},
		{"m6", "Keyboard Wrist Rest (TPU)", "lin", "TPU 95A, 15% infill", 133, []string{"TPU", "practical"}, "2026-08-25T11:20:00Z"},
	}
}
