package xhttp

import (
    "bytes"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
    "path"
    "path/filepath"
    "strings"
)

// String 将任意类型转换为string
func String(v any) string {
    if v == nil {
        return ""
    }
    switch v.(type) {
    case byte:
        return string(v.(byte))
    case []byte:
        return string(v.([]byte))
    case rune:
        return string(v.(rune))
    case []rune:
        return string(v.([]rune))
    default:
        return fmt.Sprintf("%v", v)
    }
}

// BuildUrl 构造请求地址
func BuildUrl(host, path string) string {
    if !strings.HasSuffix(host, "/") {
        host += "/"
    }
    if len(path) >= 1 {
        if strings.HasPrefix(path, "/") {
            path = path[1:]
        }
    }
    return host + path
}

// Download 下载指定地址的文件到本地
// 注意：历史实现中请求失败时会返回nil（错误被吞掉），此处已修正为返回真实错误
func Download(remoteUrl, localPath string) error {
    resp, err := NewClient().Get(remoteUrl)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    // 非2xx状态码视为失败，避免把错误页面写入本地文件
    if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
        return fmt.Errorf("download [%s] failed: http status %d", remoteUrl, resp.StatusCode)
    }
    // Create the file
    localDir := path.Dir(localPath)
    if localDir != "" && localDir != "." {
        if err := os.MkdirAll(localDir, 0755); err != nil {
            return err
        }
    }
    file, err := os.Create(localPath)
    if err != nil {
        return err
    }
    defer file.Close()
    // Write the response body to the file
    _, err = io.Copy(file, resp.Body)
    if err != nil {
        return err
    }
    return nil
}

// Upload 实现本地文件上传到远端
func Upload(localFile string, serverUrl string, fileField string) (string, error) {
    // Open the file
    file, err := os.Open(localFile)
    if err != nil {
        return "", err
    }
    defer file.Close()

    // Create a new multipart form
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    // Add the file to the form
    part, err := writer.CreateFormFile(fileField, filepath.Base(localFile))
    if err != nil {
        return "", err
    }
    _, err = io.Copy(part, file)
    if err != nil {
        return "", err
    }

    // Close the form
    err = writer.Close()
    if err != nil {
        return "", err
    }

    // Create the request
    resp, err := NewClient().
        WithHeader("Content-Type", writer.FormDataContentType()).
        WithBody(body).
        Post(serverUrl)
    if err != nil {
        return "", err
    }
    // 必须关闭响应正文，否则会造成连接泄漏
    defer resp.Body.Close()

    // Read the response body
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(respBody), nil
}

// GetHeaders 获取给定地址的头信息，不返回body部分内容
func GetHeaders(url string) (http.Header, error) {
    resp, err := NewClient().Head(url)
    if err != nil {
        return nil, err
    }
    _ = resp.Body.Close()
    return resp.Header, nil
}
