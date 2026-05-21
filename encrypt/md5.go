package encrypt

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// Md5 生成md5 hash
func Md5(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// Md5Short 16位MD5
func Md5Short(str string) string {
	hexStr := Md5(str)
	return string([]byte(hexStr)[8:24])
}

func Md5Raw(data []byte) []byte {
	h := md5.New()
	h.Write(data)
	return h.Sum(nil)
}

// Md5Bytes 对给定的字节进行MD5处理
func Md5Bytes(data []byte) string {
	h := md5.New()
	h.Write(data)
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

// Md5ShortBytes 16位MD5
func Md5ShortBytes(data []byte) string {
	hexStr := Md5Bytes(data)
	return string([]byte(hexStr)[8:24])
}

// Md5File 生成文件的MD5值
func Md5File(filePath string) (string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 创建md5 hash接口实例
	hash := md5.New()
	// 将文件内容拷贝到hash实例中
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// 获取MD5值并打印出来
	return hex.EncodeToString(hash.Sum(nil)), nil
}
