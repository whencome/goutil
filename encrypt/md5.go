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
