package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
)

// 检查给定的文件名是否存在
func Exists(fileName string) bool {
	_, err := os.Stat(fileName)
	return !os.IsNotExist(err)
}

// 复制目录
func CopyDir(srcDir, dstDir string) error {
	path, dirName := filepath.Split(srcDir)
	sDir := Dir{ParentName: path, DirName: dirName, FileDelSign: true}
	return sDir.CopyDir(dstDir)
}

// 格式化文件大小
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
