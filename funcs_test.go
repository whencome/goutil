package goutil

import (
	"reflect"
	"strings"
	"testing"
)

func TestBool(t *testing.T) {
	cases := []struct {
		in   interface{}
		want bool
	}{
		{true, true},
		{false, false},
		{"true", true},
		{"TRUE", true},
		{" false ", false},
		{"1", true},
		{"0", false},
		{"1.0", true},
		{"abc", false},
		{1, true},
		{0, false},
		{nil, false},
	}
	for _, c := range cases {
		if got := Bool(c.in); got != c.want {
			t.Errorf("Bool(%#v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestNumberConversionWithSpace(t *testing.T) {
	if got := Int64(" 123 "); got != 123 {
		t.Errorf("Int64(\" 123 \") = %d, want 123", got)
	}
	if got := Uint64(" 456 "); got != 456 {
		t.Errorf("Uint64(\" 456 \") = %d, want 456", got)
	}
	if got := Float64(" 1.5 "); got != 1.5 {
		t.Errorf("Float64(\" 1.5 \") = %v, want 1.5", got)
	}
}

func TestSubString(t *testing.T) {
	str := "hello,世界"
	if got := SubString(str, 0, 5); got != "hello" {
		t.Errorf("SubString(%q,0,5) = %q", str, got)
	}
	if got := SubString(str, 6, 8); got != "世界" {
		t.Errorf("SubString(%q,6,8) = %q", str, got)
	}
	// 越界不应该panic
	if got := SubString(str, 0, 100); got != str {
		t.Errorf("SubString(%q,0,100) = %q", str, got)
	}
	if got := SubString(str, 20, 30); got != "" {
		t.Errorf("SubString with out of range should return empty, got %q", got)
	}
	if got := SubString(str, 5, 2); got != "" {
		t.Errorf("SubString with start>end should return empty, got %q", got)
	}
	if got := SubString(str, -2, 100); got != "世界" {
		t.Errorf("SubString with negative start = %q", got)
	}
}

func TestRandString(t *testing.T) {
	s1 := RandString(16)
	s2 := RandString(16)
	if len(s1) != 16 || len(s2) != 16 {
		t.Fatalf("unexpected length: %d, %d", len(s1), len(s2))
	}
	// 历史实现使用秒级种子，同一秒内生成的结果完全相同；修复后不应该出现此问题
	if s1 == s2 {
		t.Fatalf("RandString generated duplicated value in the same second: %s", s1)
	}
	if RandString(0) != "" {
		t.Error("RandString(0) should return empty string")
	}
}

func TestRandStringSecure(t *testing.T) {
	s1, err := RandStringSecure(32)
	if err != nil {
		t.Fatalf("RandStringSecure failed: %v", err)
	}
	s2, err := RandStringSecure(32)
	if err != nil {
		t.Fatalf("RandStringSecure failed: %v", err)
	}
	if len(s1) != 32 || s1 == s2 {
		t.Fatalf("unexpected result: %s, %s", s1, s2)
	}
}

func TestDeepCopyArray(t *testing.T) {
	arr := [3]int{1, 2, 3}
	cp := DeepCopy(arr)
	cpArr, ok := cp.([3]int)
	if !ok {
		t.Fatalf("DeepCopy of array returns %T", cp)
	}
	if !reflect.DeepEqual(arr, cpArr) {
		t.Fatalf("DeepCopy of array mismatch: %v", cpArr)
	}
}

func TestDeepCopyNil(t *testing.T) {
	var m map[string]interface{}
	if got := DeepCopy(m); got != nil {
		// 针对nil map，历史实现会返回空map，此处允许返回nil或者空map，但不能panic
		if mm, ok := got.(map[string]interface{}); !ok || len(mm) != 0 {
			t.Fatalf("unexpected DeepCopy result: %#v", got)
		}
	}
	var p *int
	if got := DeepCopy(p); got != nil {
		t.Fatalf("DeepCopy of nil pointer should return nil, got %#v", got)
	}
}

func TestCopyStruct(t *testing.T) {
	type Src struct {
		Name  string
		Age   int
		Extra string
	}
	type Dst struct {
		Name string
		Age  int
	}
	src := &Src{Name: "jack", Age: 18, Extra: "x"}
	dst := &Dst{}
	CopyStruct(src, dst)
	if dst.Name != "jack" || dst.Age != 18 {
		t.Fatalf("CopyStruct failed: %+v", dst)
	}
	// 非法参数不应该panic
	CopyStruct(nil, dst)
	CopyStruct(src, nil)
	CopyStruct(src, Dst{})
	CopyStruct(src, 123)
}

func TestSFiltered(t *testing.T) {
	s := S{0, "", nil, "a", 1}
	filtered := s.Filtered()
	if len(filtered) != 2 {
		t.Fatalf("Filtered got %d items, want 2: %v", len(filtered), filtered)
	}
}

func TestSContainsNonString(t *testing.T) {
	s := S{1, "a", []int{1}}
	// 传入非字符串类型时不应该panic
	if s.Contains([]int{1}) != true {
		t.Error("Contains should find the slice value")
	}
	if s.Contains(struct{}{}) {
		t.Error("Contains should not find the struct value")
	}
}

func TestKVPairs(t *testing.T) {
	p := KVPairs{{K: "b", V: "2"}, {K: "a", V: "1"}}
	p.Sort()
	if p[0].K != "a" {
		t.Fatalf("Sort failed: %v", p)
	}
	// 历史实现中Reverse没有效果
	p.Reverse()
	if p[0].K != "b" {
		t.Fatalf("Reverse failed: %v", p)
	}
	// Add为历史遗留的空操作，Append应返回新的列表
	np := p.Append("c", "3")
	if len(np) != 3 || len(p) != 2 {
		t.Fatalf("Append failed: p=%v, np=%v", p, np)
	}
	if np.String() != "b=2&a=1&c=3" {
		t.Fatalf("unexpected string: %s", np.String())
	}
}

func TestUrlValuesAndQuery(t *testing.T) {
	m := M{
		"name": "jack",
		"ids":  []int{1, 2, 3},
	}
	uv := m.UrlValues()
	if uv.Get("name") != "jack" {
		t.Errorf("UrlValues name = %q", uv.Get("name"))
	}
	if len(uv["ids"]) != 3 {
		t.Errorf("UrlValues ids = %v", uv["ids"])
	}
	q := m.UrlQuery()
	if !strings.Contains(q, "name=jack") || !strings.Contains(q, "ids=1") {
		t.Errorf("unexpected query: %s", q)
	}
}
