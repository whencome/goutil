package zipfile

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Zip 创建zip文件
// srcFile could be a single file or a directory
// 说明：历史实现忽略了filepath.Walk返回的错误，压缩过程中出现错误时依旧返回nil，
// 此处已修正为返回真实错误，并避免在遍历过程中积累未关闭的文件句柄
func Zip(srcFile string, destZip string) (err error) {
	zf, err := os.Create(destZip)
	if err != nil {
		return err
	}
	// 确保文件被关闭，并把关闭错误一并返回
	defer func() {
		if closeErr := zf.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	archive := zip.NewWriter(zf)

	// 遍历并写入文件内容
	walkErr := filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info == nil {
			return fmt.Errorf("get file info of [%s] failed", path)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(strings.TrimPrefix(path, filepath.Dir(srcFile)+string(filepath.Separator)))
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		// 立即复制并关闭，避免使用defer导致文件句柄在遍历结束时才释放
		_, copyErr := io.Copy(writer, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})

	// 必须关闭zip写入器，否则压缩包的中央目录不会写入，生成的文件无法使用
	closeErr := archive.Close()
	if walkErr != nil {
		return walkErr
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}

// Unzip 将zip文件解压到指定目录
// 说明：为避免路径穿越攻击（zip slip），条目路径中如果包含".."或者绝对路径，该条目会被跳过；
// 单个条目解压失败不会中断整个解压过程（保持历史行为）
func Unzip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		err := copyFromZippedFile(f, destDir)
		if err != nil {
			continue
		}
	}
	return nil
}

// isUnsafePath 判断zip条目路径是否安全（防止路径穿越）
func isUnsafePath(name string) bool {
	if name == "" {
		return true
	}
	// 统一使用/分隔，再按层级检查
	clean := filepath.ToSlash(filepath.Clean(strings.ReplaceAll(name, "\\", "/")))
	if strings.HasPrefix(clean, "../") || clean == ".." {
		return true
	}
	if strings.Contains(clean, "/../") {
		return true
	}
	// 绝对路径
	if filepath.IsAbs(name) {
		return true
	}
	if len(clean) >= 2 && clean[1] == ':' { // Windows盘符
		return true
	}
	return false
}

func copyFromZippedFile(f *zip.File, destDir string) error {
	if isUnsafePath(f.Name) {
		return fmt.Errorf("unsafe file path in zip: %s", f.Name)
	}
	fpath := filepath.Join(destDir, f.Name)
	// 二次校验，确保最终路径仍在目标目录之内
	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return err
	}
	absPath, err := filepath.Abs(fpath)
	if err != nil {
		return err
	}
	if absPath != absDest && !strings.HasPrefix(absPath, absDest+string(filepath.Separator)) {
		return fmt.Errorf("unsafe file path in zip: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(fpath, os.ModePerm)
	}
	if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
		return err
	}

	inFile, err := f.Open()
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	if _, err = io.Copy(outFile, inFile); err != nil {
		return err
	}
	return nil
}
