package request

import (
    "net/http/httptest"
    "strings"
    "testing"
)

// TestGetSimpleParams 测试GET请求参数解析
func TestGetSimpleParams(t *testing.T) {
    r := httptest.NewRequest("GET", "/test?a=1&a=2&b=3", nil)
    m := GetSimpleParams(r)
    if m.GetString("a") != "1" {
        t.Errorf("a = %s, want 1", m.GetString("a"))
    }
    if m.GetString("b") != "3" {
        t.Errorf("b = %s, want 3", m.GetString("b"))
    }
    // 获取全部参数（数组形式）
    params := GetParams(r)
    if len(params["a"]) != 2 {
        t.Errorf("unexpected params: %v", params)
    }
}

// TestPostFormParams 测试表单POST参数解析
func TestPostFormParams(t *testing.T) {
    body := strings.NewReader("a=1&b=2")
    r := httptest.NewRequest("POST", "/test", body)
    r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    m := GetSimpleParams(r)
    if m.GetString("a") != "1" || m.GetString("b") != "2" {
        t.Errorf("unexpected params: %v", m)
    }
}

// TestPostJSONParams 测试JSON请求体参数解析
func TestPostJSONParams(t *testing.T) {
    body := strings.NewReader(`{"a":"1","obj":{"k":"v"},"arr":[1,2],"n":10}`)
    r := httptest.NewRequest("POST", "/test", body)
    r.Header.Set("Content-Type", "application/json; charset=utf-8")
    m := GetSimpleParams(r)
    if m.GetString("a") != "1" {
        t.Errorf("a = %s, want 1", m.GetString("a"))
    }
    if m.GetString("n") != "10" {
        t.Errorf("n = %s, want 10", m.GetString("n"))
    }
    // 对象与数组保留json字面量，便于调用方二次解析
    if m.GetString("obj") != `{"k":"v"}` {
        t.Errorf("obj = %s", m.GetString("obj"))
    }
    if m.GetString("arr") != `[1,2]` {
        t.Errorf("arr = %s", m.GetString("arr"))
    }
}

// TestPostMultipartParams 测试multipart表单解析以及异常请求不会panic
func TestPostMultipartParams(t *testing.T) {
    // 声明为multipart但内容非法的请求，历史实现会因MultipartForm为nil而panic
    r := httptest.NewRequest("POST", "/test", strings.NewReader("not a multipart body"))
    r.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
    m := GetSimpleParams(r)
    if m == nil {
        t.Error("params should not be nil")
    }
}

// TestNilRequest 测试空请求不会panic
func TestNilRequest(t *testing.T) {
    if m := GetSimpleParams(nil); len(m) != 0 {
        t.Errorf("unexpected params: %v", m)
    }
    if params := GetParams(nil); len(params) != 0 {
        t.Errorf("unexpected params: %v", params)
    }
    if params := PostParams(nil); len(params) != 0 {
        t.Errorf("unexpected params: %v", params)
    }
    if ct := ContentType(nil); ct != "application/x-www-form-urlencoded" {
        t.Errorf("unexpected content type: %s", ct)
    }
}

// TestContentType 测试Content-Type解析
func TestContentType(t *testing.T) {
    r := httptest.NewRequest("POST", "/test", nil)
    if ct := ContentType(r); ct != "application/x-www-form-urlencoded" {
        t.Errorf("default content type = %s", ct)
    }
    r.Header.Set("Content-Type", "Application/JSON; charset=utf-8")
    if ct := ContentType(r); ct != "application/json" {
        t.Errorf("content type = %s", ct)
    }
}
