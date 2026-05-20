package zipfile

import (
	"archive/zip"
	"io"
	"os"
	"strings"
)

// FileInf 定义文件信息
type FileInf struct {
	Path     string // 全路径，含文件名
	Name     string
	Size     int64
	IsDir    bool
	SubFiles []*FileInf
}

// GetFileInfo 获取zip文件信息
func GetFileInfo(f string) (*FileInf, error) {
	zfi, err := os.Stat(f)
	if err != nil {
		return nil, err
	}
	// 构建结果
	rs := &FileInf{
		Path:     f,
		Name:     zfi.Name(),
		Size:     zfi.Size(),
		IsDir:    zfi.IsDir(),
		SubFiles: make([]*FileInf, 0),
	}
	// 打开一个zip格式文件
	r, err := zip.OpenReader(f)
	if err != nil {
		return rs, err
	}
	defer r.Close()
	// 迭代压缩文件中的文件，打印出文件中的内容
	for _, f := range r.File {
		fileName := f.Name
		fileName = strings.Trim(fileName, "/")
		fi := f.FileInfo()
		fInf := &FileInf{
			Name:     fileName,
			Size:     fi.Size(),
			IsDir:    fi.IsDir(),
			SubFiles: make([]*FileInf, 0),
		}
		rs.SubFiles = append(rs.SubFiles, fInf)
	}
	return rs, nil
}

// GetFileContent 获取zip文件钟指定文件的内容
func GetFileContent(p, name string) ([]byte, error) {
	// 打开一个zip格式文件
	r, err := zip.OpenReader(p)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	// 迭代压缩文件中的文件，打印出文件中的内容
	for _, f := range r.File {
		fileName := f.Name
		fileName = strings.Trim(fileName, "/")
		if fileName != name {
			continue
		}
		r, e := f.Open()
		if e != nil {
			return nil, e
		}
		defer r.Close()
		return io.ReadAll(r)
	}
	return []byte{}, nil
}
