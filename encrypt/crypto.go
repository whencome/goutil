package encrypt

import "bytes"

// PKCS7Padding 数据填充，PKCS5的分组是以8为单位
// PKCS7的分组长度为1-255
// 说明：当blockSize非法（小于等于0或大于255）时，无法进行填充，此时会返回原始数据的副本，
// 避免历史实现中因取模0导致的panic
func PKCS7Padding(org []byte, blockSize int) []byte {
    if blockSize <= 0 || blockSize > 255 {
        // 非法分组长度，返回原始数据的副本
        ret := make([]byte, len(org))
        copy(ret, org)
        return ret
    }
    pad := blockSize - len(org)%blockSize
    padArr := bytes.Repeat([]byte{byte(pad)}, pad)
    ret := make([]byte, 0, len(org)+pad)
    ret = append(ret, org...)
    return append(ret, padArr...)
}

// PKCS7UnPadding 去除填充
// 说明：历史实现在数据为空或者填充值非法时会发生切片越界panic，
// 此处对非法数据做保护，非法填充时原样返回数据（不修改入参内存）
func PKCS7UnPadding(org []byte) []byte {
    l := len(org)
    if l == 0 {
        return org
    }
    pad := int(org[l-1])
    if pad == 0 || pad > l {
        // 非法的填充值，原样返回
        return org
    }
    return org[:l-pad]
}

// isValidPKCS7Padding 校验PKCS7填充是否合法
func isValidPKCS7Padding(data []byte) bool {
    l := len(data)
    if l == 0 {
        return false
    }
    pad := int(data[l-1])
    if pad == 0 || pad > l {
        return false
    }
    for i := l - pad; i < l; i++ {
        if int(data[i]) != pad {
            return false
        }
    }
    return true
}
