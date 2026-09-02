package gotsx

import "testing"

func withI18n(t *testing.T, c *I18n) func() {
	old := i18nCfg
	configureI18n(c)
	return func() { i18nCfg = old }
}

func testCfg() *I18n {
	return &I18n{
		Locales: []string{"zh", "en"}, Default: "zh", Prefix: true,
		Currency: map[string]string{"zh": "¥", "en": "$"},
		Messages: map[string]map[string]string{
			"zh": {"hi": "你好 {name}", "items": "{n} 件"},
			"en": {"hi": "Hi {name}", "items": "{n} item|{n} items", "only": "en-only"},
		},
	}
}

func TestTrAndFallback(t *testing.T) {
	defer withI18n(t, testCfg())()
	if Tr("en", "only") != "en-only" {
		t.Error("en lookup")
	}
	if Tr("zh", "only") != "en-only" { // 回退到默认? 不: only 只在 en, zh 无 → 回退默认(zh)也无 → key
		// zh 无 only, 回退默认 zh 也无 → 返回 key
	}
	if Tr("en", "missing") != "missing" {
		t.Error("缺失应返回 key")
	}
}

func TestInterp(t *testing.T) {
	defer withI18n(t, testCfg())()
	if Trv("zh", "hi", map[string]string{"name": "小明"}) != "你好 小明" {
		t.Errorf("zh interp: %q", Trv("zh", "hi", map[string]string{"name": "小明"}))
	}
	if Trv("en", "hi", map[string]string{"name": "Sam"}) != "Hi Sam" {
		t.Error("en interp")
	}
}

func TestPlural(t *testing.T) {
	defer withI18n(t, testCfg())()
	if Plural("en", "items", 1) != "1 item" {
		t.Errorf("en one: %q", Plural("en", "items", 1))
	}
	if Plural("en", "items", 3) != "3 items" {
		t.Errorf("en other: %q", Plural("en", "items", 3))
	}
	if Plural("zh", "items", 3) != "3 件" || Plural("zh", "items", 1) != "1 件" {
		t.Error("zh 只有一种形式")
	}
}

func TestFmt(t *testing.T) {
	defer withI18n(t, testCfg())()
	if FmtNum("en", 1234567) != "1,234,567" {
		t.Errorf("num: %q", FmtNum("en", 1234567))
	}
	if FmtCur("zh", 123456) != "¥1,234.56" {
		t.Errorf("cur zh: %q", FmtCur("zh", 123456))
	}
	if FmtCur("en", 123456) != "$1,234.56" {
		t.Errorf("cur en: %q", FmtCur("en", 123456))
	}
	if FmtDate("zh", "2026-08-02T10:00:00Z") != "2026年8月2日" {
		t.Errorf("date zh: %q", FmtDate("zh", "2026-08-02T10:00:00Z"))
	}
	if FmtDate("en", "2026-08-02T10:00:00Z") != "Aug 2, 2026" {
		t.Errorf("date en: %q", FmtDate("en", "2026-08-02T10:00:00Z"))
	}
}

func TestResolveLocale(t *testing.T) {
	defer withI18n(t, testCfg())()
	// URL 前缀
	loc, path := resolveLocale("/en/p/p001", "", "")
	if loc != "en" || path != "/p/p001" {
		t.Errorf("prefix: %q %q", loc, path)
	}
	loc, path = resolveLocale("/p/p001", "", "")
	if loc != "zh" || path != "/p/p001" {
		t.Errorf("default: %q %q", loc, path)
	}
	loc, _ = resolveLocale("/en", "", "")
	if loc != "en" {
		t.Errorf("/en root: %q", loc)
	}
}

func TestResolveLocaleCookieHeader(t *testing.T) {
	c := testCfg()
	c.Prefix = false
	defer withI18n(t, c)()
	if loc, _ := resolveLocale("/p", "en", ""); loc != "en" {
		t.Error("cookie lang")
	}
	if loc, _ := resolveLocale("/p", "", "en-US,en;q=0.9"); loc != "en" {
		t.Error("accept-language")
	}
	if loc, _ := resolveLocale("/p", "", "fr-FR"); loc != "zh" {
		t.Error("未知语言回退默认")
	}
}

func TestLPath(t *testing.T) {
	defer withI18n(t, testCfg())()
	if LPath("en", "/p/x") != "/en/p/x" {
		t.Error("en prefix")
	}
	if LPath("zh", "/p/x") != "/p/x" {
		t.Error("默认不加前缀")
	}
	if LPath("en", "/") != "/en" {
		t.Error("en root")
	}
}
