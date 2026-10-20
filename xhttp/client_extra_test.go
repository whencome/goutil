package xhttp

import (
    "io"
    "net/http"
    "net/http/httptest"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

// TestPutWithParams 测试PUT请求的参数作为表单正文提交
// 历史实现中PUT请求的参数会被直接丢弃
func TestPutWithParams(t *testing.T) {
    var gotBody, gotContentType string
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPut {
            t.Errorf("expected PUT request, got %s", r.Method)
        }
        body, _ := io.ReadAll(r.Body)
        gotBody = string(body)
        gotContentType = r.Header.Get("Content-Type")
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    resp, err := NewClient().WithParams(map[string]interface{}{"a": 1, "b": "2"}).Put(server.URL + "/test")
    if err != nil {
        t.Fatalf("put request failed: %v", err)
    }
    defer resp.Body.Close()
    if gotBody == "" {
        t.Fatal("request params should be sent as form body for PUT")
    }
    if !strings.Contains(gotBody, "a=1") || !strings.Contains(gotBody, "b=2") {
        t.Errorf("unexpected body: %s", gotBody)
    }
    if gotContentType != "application/x-www-form-urlencoded" {
        t.Errorf("unexpected content type: %s", gotContentType)
    }
}

// TestPatchWithParams 测试PATCH请求
func TestPatchWithParams(t *testing.T) {
    var gotBody string
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPatch {
            t.Errorf("expected PATCH request, got %s", r.Method)
        }
        body, _ := io.ReadAll(r.Body)
        gotBody = string(body)
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    resp, err := NewClient().WithParams(map[string]interface{}{"a": "1"}).Patch(server.URL + "/test")
    if err != nil {
        t.Fatalf("patch request failed: %v", err)
    }
    defer resp.Body.Close()
    if !strings.Contains(gotBody, "a=1") {
        t.Errorf("unexpected body: %s", gotBody)
    }

    // Request 方法同样支持 PATCH
    resp2, err := NewClient().WithParams(map[string]interface{}{"a": "2"}).Request(http.MethodPatch, server.URL+"/test")
    if err != nil {
        t.Fatalf("patch request via Request failed: %v", err)
    }
    defer resp2.Body.Close()
    if !strings.Contains(gotBody, "a=2") {
        t.Errorf("unexpected body: %s", gotBody)
    }
}

// TestDeleteWithParams 测试DELETE请求的参数拼接到URL中
func TestDeleteWithParams(t *testing.T) {
    var gotQuery string
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodDelete {
            t.Errorf("expected DELETE request, got %s", r.Method)
        }
        gotQuery = r.URL.RawQuery
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    resp, err := NewClient().WithParams(map[string]interface{}{"id": 10}).Delete(server.URL + "/test")
    if err != nil {
        t.Fatalf("delete request failed: %v", err)
    }
    defer resp.Body.Close()
    if gotQuery != "id=10" {
        t.Errorf("unexpected query: %s", gotQuery)
    }
}

// TestRawClientWithoutHeaders 测试直接构造的Client（Headers为nil）不会panic
func TestRawClientWithoutHeaders(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    c := &Client{Client: &http.Client{}}
    resp, err := c.Post(server.URL + "/test")
    if err != nil {
        t.Fatalf("post request failed: %v", err)
    }
    defer resp.Body.Close()
}

// TestDownload 测试文件下载
func TestDownload(t *testing.T) {
    content := "hello, download"
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/not-found" {
            w.WriteHeader(http.StatusNotFound)
            return
        }
        _, _ = w.Write([]byte(content))
    }))
    defer server.Close()

    dir := t.TempDir()
    localFile := filepath.Join(dir, "sub", "test.txt")
    if err := Download(server.URL+"/file", localFile); err != nil {
        t.Fatalf("download failed: %v", err)
    }
    b, err := os.ReadFile(localFile)
    if err != nil {
        t.Fatalf("read downloaded file failed: %v", err)
    }
    if string(b) != content {
        t.Errorf("unexpected content: %s", string(b))
    }

    // 请求失败时历史实现会返回nil（吞掉错误），此处必须返回错误
    if err := Download(server.URL+"/not-found", filepath.Join(dir, "404.txt")); err == nil {
        t.Fatal("download a 404 response should return error")
    }
    if err := Download("http://127.0.0.1:1/not-exist", filepath.Join(dir, "err.txt")); err == nil {
        t.Fatal("download with an unreachable address should return error")
    }
}

// TestBuildUrl 测试URL拼接
func TestBuildUrl(t *testing.T) {
    cases := []struct {
        host, path, want string
    }{
        {"http://a.com", "b/c", "http://a.com/b/c"},
        {"http://a.com/", "/b/c", "http://a.com/b/c"},
        {"http://a.com/api", "/b", "http://a.com/api/b"},
    }
    for _, c := range cases {
        if got := BuildUrl(c.host, c.path); got != c.want {
            t.Errorf("BuildUrl(%s,%s) = %s, want %s", c.host, c.path, got, c.want)
        }
    }
}
