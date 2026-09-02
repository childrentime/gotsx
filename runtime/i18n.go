package gotsx

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// I18n 是可选的国际化配置。消息目录: locale → key → 文案。
// 文案支持 {name} 占位符;复数用 | 分隔(英语 one|other, 中文只写一条)。
type I18n struct {
	Locales  []string                     // 支持的语言, 如 ["zh","en"]
	Default  string                       // 默认语言(URL 前缀模式下默认语言不加前缀)
	Prefix   bool                         // true: 语言走 URL 前缀 /en/...(SEO 友好); false: 走 cookie lang + Accept-Language
	Messages map[string]map[string]string // locale → key → 文案
	Currency map[string]string            // locale → 货币符号(如 zh→¥, en→$); 不做汇率换算
}

var i18nCfg *I18n

func configureI18n(c *I18n) {
	if c != nil && c.Default == "" && len(c.Locales) > 0 {
		c.Default = c.Locales[0]
	}
	i18nCfg = c
}

func i18nActive() *I18n {
	if i18nCfg != nil {
		return i18nCfg
	}
	return &I18n{Default: "zh", Locales: []string{"zh"}, Messages: map[string]map[string]string{}}
}

func (c *I18n) has(loc string) bool {
	for _, l := range c.Locales {
		if l == loc {
			return true
		}
	}
	return false
}

// resolveLocale: 从 URL 前缀 / cookie / Accept-Language 解析语言, 返回 (locale, 去前缀后的 path)
func resolveLocale(path, cookieLang, acceptLang string) (string, string) {
	c := i18nActive()
	if c.Prefix {
		p := strings.TrimPrefix(path, "/")
		seg := p
		if i := strings.IndexByte(p, '/'); i >= 0 {
			seg = p[:i]
		}
		if seg != "" && seg != c.Default && c.has(seg) {
			rest := "/" + strings.TrimPrefix(p, seg)
			rest = "/" + strings.TrimLeft(rest, "/")
			if rest == "" {
				rest = "/"
			}
			return seg, rest
		}
		return c.Default, path
	}
	if cookieLang != "" && c.has(cookieLang) {
		return cookieLang, path
	}
	for _, part := range strings.Split(acceptLang, ",") {
		tag := strings.TrimSpace(part)
		if i := strings.IndexByte(tag, ';'); i >= 0 {
			tag = tag[:i]
		}
		tag = strings.ToLower(tag)
		if c.has(tag) {
			return tag, path
		}
		if i := strings.IndexByte(tag, '-'); i >= 0 && c.has(tag[:i]) {
			return tag[:i], path
		}
	}
	return c.Default, path
}

func lookup(locale, key string) (string, bool) {
	c := i18nActive()
	if m, ok := c.Messages[locale]; ok {
		if v, ok := m[key]; ok {
			return v, true
		}
	}
	if locale != c.Default {
		if m, ok := c.Messages[c.Default]; ok {
			if v, ok := m[key]; ok {
				return v, true
			}
		}
	}
	return key, false // 兜底: 返回 key 本身(便于发现缺翻译)
}

func interp(msg string, vars map[string]string) string {
	if !strings.Contains(msg, "{") {
		return msg
	}
	for k, v := range vars {
		msg = strings.ReplaceAll(msg, "{"+k+"}", v)
	}
	return msg
}

// Tr: 简单查表
func Tr(locale, key string) string { m, _ := lookup(locale, key); return m }

// Trv: 带 {name} 插值
func Trv(locale, key string, vars map[string]string) string {
	m, _ := lookup(locale, key)
	return interp(m, vars)
}

// pluralCategory: 极简 CLDR 规则(够覆盖 zh/en/大多数)
func pluralCategory(locale string, n float64) string {
	switch locale {
	case "zh", "ja", "ko", "vi", "th", "id":
		return "other"
	default: // en 等
		if n == 1 {
			return "one"
		}
		return "other"
	}
}

// Plural: 文案用 "one|other" 分隔(中文只写一条), {n} 替换为数字
func Plural(locale, key string, n float64) string {
	m, _ := lookup(locale, key)
	forms := strings.Split(m, "|")
	pick := forms[len(forms)-1]
	if pluralCategory(locale, n) == "one" && len(forms) > 1 {
		pick = forms[0]
	}
	return strings.ReplaceAll(pick, "{n}", Num(n))
}

func groupThousands(intPart string) string {
	neg := strings.HasPrefix(intPart, "-")
	intPart = strings.TrimPrefix(intPart, "-")
	var out []byte
	for i, c := range []byte(intPart) {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	s := string(out)
	if neg {
		s = "-" + s
	}
	return s
}

// FmtNum: 本地化数字(千分位)
func FmtNum(locale string, n float64) string {
	s := strconv.FormatFloat(n, 'f', -1, 64)
	dot := strings.IndexByte(s, '.')
	if dot < 0 {
		return groupThousands(s)
	}
	return groupThousands(s[:dot]) + s[dot:]
}

// FmtCur: 本地化金额(cents → 符号 + 千分位, 不换算汇率)
func FmtCur(locale string, cents float64) string {
	sym := "¥"
	if c := i18nActive(); c.Currency != nil {
		if s, ok := c.Currency[locale]; ok {
			sym = s
		}
	}
	neg := ""
	if cents < 0 {
		neg = "-"
		cents = -cents
	}
	whole := int64(cents) / 100
	frac := int64(cents) % 100
	return fmt.Sprintf("%s%s%s.%02d", neg, sym, groupThousands(strconv.FormatInt(whole, 10)), frac)
}

var monthEn = []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

// FmtDate: 本地化日期
func FmtDate(locale, iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		if t, err = time.Parse("2006-01-02", iso); err != nil {
			return iso
		}
	}
	switch locale {
	case "zh":
		return fmt.Sprintf("%d年%d月%d日", t.Year(), int(t.Month()), t.Day())
	default:
		return fmt.Sprintf("%s %d, %d", monthEn[int(t.Month())], t.Day(), t.Year())
	}
}

// LPath: 给路径加语言前缀(默认语言不加, 与 canonical 一致)
func LPath(locale, path string) string {
	c := i18nActive()
	if !c.Prefix || locale == c.Default || locale == "" {
		return path
	}
	if path == "/" {
		return "/" + locale
	}
	return "/" + locale + path
}
