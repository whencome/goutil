package encrypt

import "bytes"

// PKCS7Padding 数据填充，PKCS5的分组是以8为单位
// PKCS7的分组长度为1-255
func PKCS7Padding(org []byte, blockSize int) []byte {
    pad := blockSize - len(org)%blockSize
    padArr := bytes.Repeat([]byte{byte(pad)}, pad)
    return append(org, padArr...)
}

// PKCS7UnPadding 去除填充
func PKCS7UnPadding(org []byte) []byte {
    l := len(org)
    pad := org[l-1]
    return org[:l-int(pad)]
}
