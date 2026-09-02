// Package host 是示例应用的宿主模块: 方言里 `import { models } from "host:data"` 的 Go 一侧。
// 编译后是直接的 Go 调用, 没有任何编组。
package host

import (
	"fmt"
	"strings"
	"sync"
	"time"

	gotsx "gotsx/runtime"
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
		{"m1", "可折叠手机支架", "ziwei", "无需支撑, 一体打印, 铰链公差 0.3mm", 1287, []string{"实用", "无支撑"}, "2026-08-02T10:00:00Z"},
		{"m2", "参数化线缆收纳盒", "lin", "OpenSCAD 参数化, 三种尺寸", 402, []string{"收纳", "参数化"}, "2026-08-10T09:30:00Z"},
		{"m3", "AMS 干燥剂盒", "maker_wu", "适配 AMS 内舱, 卡扣固定", 3310, []string{"Bambu", "AMS"}, "2026-07-21T14:12:00Z"},
		{"m4", "低多边形柴犬", "kyo", "多色, 建议 0.12 层高", 958, []string{"摆件", "多色"}, "2026-08-18T20:45:00Z"},
		{"m5", "六角抽屉模块", "ziwei", "可无限拼接的桌面收纳", 2201, []string{"收纳", "模块化"}, "2026-06-30T08:00:00Z"},
		{"m6", "键盘腕托 (TPU)", "lin", "TPU 95A, 填充 15%", 133, []string{"TPU", "实用"}, "2026-08-25T11:20:00Z"},
	}
}
