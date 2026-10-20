package jsonkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

// TestGetOutOfRange 测试数组下标越界时不会panic
func TestGetOutOfRange(t *testing.T) {
	js := `{"list":[1,2,3],"obj":{"a":1}}`
	cases := []string{
		"/list[3]",  // 下标等于长度，历史实现会panic
		"/list[10]", // 远超长度
		"/obj/a[0]", // 非数组使用下标
	}
	for _, path := range cases {
		if v := Get(js, path); v != nil {
			t.Errorf("Get(%s) = %v, want nil", path, v)
		}
	}
	// 负数下标按"无下标"处理（历史行为），只需保证不panic
	_ = Get(js, "/list[-1]")
}

// TestGetNormal 测试正常的路径取值
func TestGetNormal(t *testing.T) {
	js := `{"list":[{"name":"a"},{"name":"b"}],"n":1}`
	if v := Get(js, "/list[1]/name"); v != "b" {
		t.Errorf("Get /list[1]/name = %v", v)
	}
	// jsonkit默认开启UseNumber，数字会被解析为json.Number
	if v := Get(js, "/n"); v == nil || fmt.Sprint(v) != "1" {
		t.Errorf("Get /n = %v", v)
	}
	// 非法的json应返回nil
	if v := Get("not a json", "/a"); v != nil {
		t.Errorf("Get with invalid json = %v", v)
	}
}

// TestMarshalSorted 测试key排序编码
func TestMarshalSorted(t *testing.T) {
	data := map[string]interface{}{"b": 1, "a": 2}
	b, err := MarshalSorted(data)
	if err != nil {
		t.Fatalf("MarshalSorted failed: %v", err)
	}
	if bytes.HasSuffix(b, []byte("\n")) {
		t.Fatalf("MarshalSorted result should not end with a newline: %q", string(b))
	}
	expect := `{"a":2,"b":1}`
	if string(b) != expect {
		t.Fatalf("MarshalSorted = %s, want %s", string(b), expect)
	}
}

// TestNumSafeDecode 测试数字安全解码
func TestNumSafeDecode(t *testing.T) {
	// 超出JS安全整数范围的数字应该被转换为字符串
	js := `{"big":12345678901234567890,"normal":123}`
	var m map[string]interface{}
	if err := NumSafeDecode(js, &m); err != nil {
		t.Fatalf("NumSafeDecode failed: %v", err)
	}
	if _, ok := m["big"].(string); !ok {
		t.Fatalf("big number should be converted to string, got %T", m["big"])
	}
	if n, ok := m["normal"].(float64); !ok || n != 123 {
		t.Fatalf("normal number should be float64, got %T(%v)", m["normal"], m["normal"])
	}

	// 非法参数不应该panic
	if err := NumSafeDecode(js, nil); err == nil {
		t.Error("NumSafeDecode with nil dst should return error")
	}
	var notPtr map[string]interface{}
	if err := NumSafeDecode(js, notPtr); err == nil {
		t.Error("NumSafeDecode with non-pointer dst should return error")
	}
	var ptr *map[string]interface{}
	if err := NumSafeDecode(js, ptr); err == nil {
		t.Error("NumSafeDecode with nil pointer should return error")
	}
}

// TestUnmarshalUseNumber 测试使用UseNumber时的行为
func TestUnmarshalUseNumber(t *testing.T) {
	var m map[string]interface{}
	if err := Unmarshal([]byte(`{"a":1}`), &m); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if _, ok := m["a"].(json.Number); !ok {
		t.Logf("number type is %T", m["a"])
	}
}
