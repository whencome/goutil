package xhttp

import (
    "bytes"
    "fmt"
    "io"
    "mime/multipart"
    "os"
    "path"
    "path/filepath"
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

// Download 下载指定地址的文件到本地
func Download(remoteUrl, localPath string) error {
    resp, err := NewClient().Get(remoteUrl)
    if err != nil {
        return nil
    }
    defer resp.Body.Close()
    // Create the file
    localDir := path.Dir(localPath)
    if err := os.MkdirAll(localDir, 0755); err != nil {
        return err
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

    // Read the response body
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    return string(respBody), nil
}
