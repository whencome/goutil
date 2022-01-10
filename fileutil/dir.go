package fileutil

import (
	"bufio"
	"errors"
	"github.com/whencome/xlog"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//WalkFile 文件对象
type WalkFile struct {
	Fname string
	Fmode os.FileMode
}

//Dir 目录对象
type Dir struct {
	//父目录
	ParentName string
	//目录的名字
	DirName string
	//目录下所有的目录
	DirList []WalkFile
	//目录下所有文件
	FileList []WalkFile
	//文件已经存在时 是否删除
	FileDelSign bool
}

//visit 浏览文件 区分为目录和文件
func (my *Dir) visit(path string, f os.FileInfo, err error) error {
	tmpPath := strings.Replace(path, my.ParentName, "", 1)
	if f == nil {
		return err
	}
	if f.IsDir() {
		my.DirList = append(my.DirList, WalkFile{Fname: tmpPath, Fmode: f.Mode()})
	} else if (f.Mode() & os.ModeSymlink) > 0 {
		return errors.New("symbol link not supported")
	} else {
		my.FileList = append(my.FileList, WalkFile{Fname: tmpPath, Fmode: f.Mode()})
	}
	return nil
}

//CreateDir 创建目录
func (my *Dir) CreateDir(dest string, mode os.FileMode) error {
	//如果目录存在就直接返回
	if _, err := os.Stat(dest); err != nil {
		mkErr := os.MkdirAll(dest, mode)
		if mkErr != nil {
			return mkErr
		}
	}
	return nil
}

//CopyFile 复制文件
func (my *Dir) CopyFile(src string, dst string, mode os.FileMode) error {
	//判断文件是否存在，如果存在根据FileDelSign决定是否删除
	if _, err := os.Stat(dst); err == nil {
		if my.FileDelSign {
			if rError := os.Remove(dst); rError != nil {
				return rError
			}
		} else {
			return nil
		}
	}
	//拷贝文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		err := srcFile.Close()
		if err != nil {
			xlog.Errorf("close file failed %v\n", err)
		}
	}()
	//通过srcFile，获取到READER
	reader := bufio.NewReader(srcFile)
	//打开dstFileName
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, mode)
	if err != nil {
		return err
	}
	//通过dstFile，获取到WRITER
	writer := bufio.NewWriter(dstFile)

	defer func() {
		writer.Flush()
		err := dstFile.Close()
		if err != nil {
			xlog.Errorf("close file failed %v\n", err)
		}
	}()

	_, copyErr := io.Copy(writer, reader)
	if copyErr != nil {
		return copyErr
	}
	//修改权限
	chmodErr := os.Chmod(dst, mode)
	if chmodErr != nil {
		return chmodErr
	}
	return nil
}

//CopyDir 复制目录
func (my *Dir) CopyDir(destParentDir string) error {
	//遍历目录
	err := filepath.Walk(filepath.Join(my.ParentName, my.DirName), func(path string, f os.FileInfo, err error) error {
		return my.visit(path, f, err)
	})
	if err != nil {
		return err
	}
	//create dir
	for _, dirName := range my.DirList {
		destDir := filepath.Join(destParentDir, dirName.Fname)
		err := my.CreateDir(destDir, dirName.Fmode)
		if err != nil {
			return err
		}
	}
	//create file
	for _, fileName := range my.FileList {
		destFile := filepath.Join(destParentDir, fileName.Fname)
		srcFile := filepath.Join(my.ParentName, fileName.Fname)
		err := my.CopyFile(srcFile, destFile, fileName.Fmode)
		if err != nil {
			return err
		}
	}
	return nil
}
