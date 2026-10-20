package zipfile

import (
    "archive/zip"
    "os"
    "path/filepath"
    "testing"
)

// TestZipAndUnzip 测试压缩与解压
func TestZipAndUnzip(t *testing.T) {
    srcDir := t.TempDir()
    subDir := filepath.Join(srcDir, "sub")
    if err := os.MkdirAll(subDir, 0755); err != nil {
        t.Fatalf("create dir failed: %v", err)
    }
    content := []byte("hello zip")
    if err := os.WriteFile(filepath.Join(subDir, "a.txt"), content, 0644); err != nil {
        t.Fatalf("write file failed: %v", err)
    }

    zipFile := filepath.Join(t.TempDir(), "test.zip")
    if err := Zip(srcDir, zipFile); err != nil {
        t.Fatalf("zip failed: %v", err)
    }

    // 压缩包信息
    fi, err := GetFileInfo(zipFile)
    if err != nil {
        t.Fatalf("get zip file info failed: %v", err)
    }
    if len(fi.SubFiles) == 0 {
        t.Fatal("zip file should contain files")
    }

    // 解压并校验内容
    destDir := t.TempDir()
    if err := Unzip(zipFile, destDir); err != nil {
        t.Fatalf("unzip failed: %v", err)
    }
    got, err := os.ReadFile(filepath.Join(destDir, filepath.Base(srcDir), "sub", "a.txt"))
    if err != nil {
        t.Fatalf("read unzipped file failed: %v", err)
    }
    if string(got) != string(content) {
        t.Fatalf("unexpected content: %s", got)
    }

    // 获取压缩包内指定文件的内容
    zipContent, err := GetFileContent(zipFile, filepath.Base(srcDir)+"/sub/a.txt")
    if err != nil {
        t.Fatalf("get file content failed: %v", err)
    }
    if string(zipContent) != string(content) {
        t.Fatalf("unexpected zip content: %s", zipContent)
    }
}

// TestZipNotExist 测试压缩不存在的路径时返回错误
func TestZipNotExist(t *testing.T) {
    zipFile := filepath.Join(t.TempDir(), "test.zip")
    if err := Zip(filepath.Join(t.TempDir(), "not-exists"), zipFile); err == nil {
        t.Fatal("zip a not-exist path should return error")
    }
}

// TestUnzipSlip 测试解压时不会因为路径穿越而写到目标目录之外
func TestUnzipSlip(t *testing.T) {
    tmpDir := t.TempDir()
    zipFile := filepath.Join(tmpDir, "evil.zip")
    f, err := os.Create(zipFile)
    if err != nil {
        t.Fatalf("create zip file failed: %v", err)
    }
    zw := zip.NewWriter(f)
    w, err := zw.Create("../evil.txt")
    if err != nil {
        t.Fatalf("create zip entry failed: %v", err)
    }
    if _, err := w.Write([]byte("evil")); err != nil {
        t.Fatalf("write zip entry failed: %v", err)
    }
    if err := zw.Close(); err != nil {
        t.Fatalf("close zip writer failed: %v", err)
    }
    _ = f.Close()

    destDir := filepath.Join(tmpDir, "out")
    if err := os.MkdirAll(destDir, 0755); err != nil {
        t.Fatalf("create dest dir failed: %v", err)
    }
    if err := Unzip(zipFile, destDir); err != nil {
        t.Fatalf("unzip failed: %v", err)
    }
    // 穿越的文件不应该被创建
    if _, err := os.Stat(filepath.Join(tmpDir, "evil.txt")); err == nil {
        t.Fatal("zip slip is not prevented")
    }
}
