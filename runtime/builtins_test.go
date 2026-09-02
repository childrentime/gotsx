package gotsx

import (
	"math"
	"testing"
)

func TestNum(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{3, "3"}, {-3, "-3"}, {3.5, "3.5"}, {0, "0"}, {1000000, "1000000"},
		{2.5, "2.5"}, {-0.25, "-0.25"},
	}
	for _, c := range cases {
		if got := Num(c.in); got != c.want {
			t.Errorf("Num(%v)=%q want %q", c.in, got, c.want)
		}
	}
	if Num(math.NaN()) != "NaN" {
		t.Error("NaN")
	}
	if Num(math.Inf(1)) != "Infinity" || Num(math.Inf(-1)) != "-Infinity" {
		t.Error("Inf")
	}
}

func TestSliceAndAt(t *testing.T) {
	xs := []int{1, 2, 3, 4, 5}
	eq := func(a []int, b ...int) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	if !eq(Slice(xs, 1, 3), 2, 3) {
		t.Error("slice 1..3")
	}
	if !eq(Slice(xs, -2), 4, 5) {
		t.Error("slice -2")
	}
	if !eq(Slice(xs, 10)) {
		t.Error("slice OOB")
	}
	if At(xs, 2) != 3 || At(xs, 99) != 0 || At(xs, -1) != 0 {
		t.Errorf("At(方括号): %d %d %d", At(xs, 2), At(xs, 99), At(xs, -1))
	}
	if AtWrap(xs, -1) != 5 || AtWrap(xs, -2) != 4 || AtWrap(xs, 99) != 0 {
		t.Errorf("AtWrap(.at): %d %d %d", AtWrap(xs, -1), AtWrap(xs, -2), AtWrap(xs, 99))
	}
}

func TestArrayHelpers(t *testing.T) {
	xs := []int{3, 1, 2}
	if Join(Map(Sort(xs, func(a, b int) float64 { return float64(a - b) }), func(x int) string { return Num(float64(x)) }), ",") != "1,2,3" {
		t.Error("sort")
	}
	if xs[0] != 3 {
		t.Error("sort 不应改原数组")
	}
	if Reduce(xs, func(acc int, x int) int { return acc + x }, 0) != 6 {
		t.Error("reduce")
	}
	rev := Reverse(xs)
	if rev[0] != 2 || xs[0] != 3 {
		t.Error("reverse 应返回新数组")
	}
	if !Includes(xs, 2) || Includes(xs, 9) {
		t.Error("includes")
	}
	if IndexOf(xs, 1) != 1 || IndexOf(xs, 9) != -1 {
		t.Error("indexOf")
	}
	if Len(Filter(xs, func(x int) bool { return x > 1 })) != 2 {
		t.Error("filter")
	}
}

func TestStringHelpers(t *testing.T) {
	if Upper("aB") != "AB" || Lower("aB") != "ab" || Trim("  x  ") != "x" {
		t.Error("case/trim")
	}
	if StrSlice("héllo", 0, 2) != "hé" {
		t.Errorf("StrSlice rune-aware: %q", StrSlice("héllo", 0, 2))
	}
	if CharAt("abc", 1) != "b" || CharAt("abc", 9) != "" {
		t.Error("charAt")
	}
	if PadStart("7", 3, "0") != "007" {
		t.Errorf("padStart: %q", PadStart("7", 3, "0"))
	}
	if PadEnd("7", 3, "-") != "7--" {
		t.Errorf("padEnd: %q", PadEnd("7", 3, "-"))
	}
	if StrLen("héllo") != 5 {
		t.Error("StrLen rune count")
	}
}

func TestObjectKeys(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1, "c": 3}
	ks := ObjectKeys(m)
	if len(ks) != 3 || ks[0] != "a" || ks[1] != "b" || ks[2] != "c" {
		t.Errorf("ObjectKeys 应排序: %v", ks)
	}
	vs := ObjectValues(m)
	if vs[0] != 1 || vs[1] != 2 || vs[2] != 3 {
		t.Errorf("ObjectValues 按键排序: %v", vs)
	}
}

func TestOrAndMath(t *testing.T) {
	if Or("", "x") != "x" || Or("a", "x") != "a" {
		t.Error("Or")
	}
	if OrNum(0, 5) != 5 || OrNum(3, 5) != 3 {
		t.Error("OrNum")
	}
	if Mod(7, 3) != 1 || Sign(-2) != -1 || Sign(0) != 0 || Trunc(3.9) != 3 {
		t.Error("math")
	}
	if ToFixed(3.14159, 2) != "3.14" {
		t.Errorf("toFixed: %q", ToFixed(3.14159, 2))
	}
}

func TestToNum(t *testing.T) {
	if ToNum("42") != 42 || ToNum("3.5") != 3.5 || ToNum(true) != 1 {
		t.Error("ToNum")
	}
	if !math.IsNaN(ToNum("abc")) {
		t.Error("ToNum NaN")
	}
}
