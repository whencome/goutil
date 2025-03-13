package encrypt

import (
    "crypto/md5"
    "encoding/hex"
    "io"
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
