package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// Exists 检查给定的文件名是否存在
// 说明：历史实现在权限不足等非"文件不存在"的错误场景下返回true，容易造成误判，
// 此处修正为仅在能够正常获取文件信息时返回true
func Exists(fileName string) bool {
	if fileName == "" {
		return false
	}
	_, err := os.Stat(fileName)
	return err == nil
}

// DirExists 检查给定的路径是否存在，且是一个目录
func DirExists(path string) bool {
	if path == "" {
		return false
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// CopyDir 复制目录
func CopyDir(srcDir, dstDir string) error {
	path, dirName := filepath.Split(srcDir)
	sDir := Dir{ParentName: path, DirName: dirName, FileDelSign: true}
	return sDir.CopyDir(dstDir)
}

// FormatSize 格式化文件大小
// 字节的单位转换 保留两位小数
func FormatSize(fileSize int64) string {
	if fileSize < 1024 {
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}
