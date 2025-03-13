package encrypt

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "github.com/tjfoc/gmsm/sm2"
    "github.com/tjfoc/gmsm/sm3"
    "github.com/tjfoc/gmsm/sm4"
    "github.com/tjfoc/gmsm/x509"
    "os"
    "strings"
)

const (
    SM2_PRIVATE_BEGIN = "-----BEGIN PRIVATE KEY-----\n"
    SM2_PRIVATE_END   = "\n-----END PRIVATE KEY-----"
    SM2_PUBLIC_BEGIN  = "-----BEGIN PUBLIC KEY-----\n"
    SM2_PUBLIC_END    = "\n-----END PUBLIC KEY-----"
)

// FormatSM2PrivateKey 格式化SM2私钥
func FormatSM2PrivateKey(privateKey string) []byte {
    if !strings.HasPrefix(privateKey, SM2_PRIVATE_BEGIN) {
        privateKey = SM2_PRIVATE_BEGIN + privateKey
    }
    if !strings.HasSuffix(privateKey, SM2_PRIVATE_END) {
        privateKey = privateKey + SM2_PRIVATE_END
    }
    return []byte(privateKey)
}

// FormatSM2PublicKey 格式化公钥
func FormatSM2PublicKey(publicKey string) []byte {
    if !strings.HasPrefix(publicKey, SM2_PUBLIC_BEGIN) {
        publicKey = SM2_PUBLIC_BEGIN + publicKey
    }
    if !strings.HasSuffix(publicKey, SM2_PUBLIC_END) {
        publicKey = publicKey + SM2_PUBLIC_END
    }
    return []byte(publicKey)
}

// LoadSM2PrivateKeyFromPEM 从pem文件加载PKCS#8格式私钥（推荐）
func LoadSM2PrivateKeyFromPEM(path string, password string) (*sm2.PrivateKey, error) {
    // 读取pem内容
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    return LoadSM2PrivateKey(data, password)
}

// LoadSM2PrivateKey 从私钥内容加载PKCS#8格式私钥
func LoadSM2PrivateKey(keyData []byte, password string) (*sm2.PrivateKey, error) {
    // 密码
    var pwd []byte = nil
    if password != "" {
        pwd = []byte(password)
    }
    // 解析数据内容，主要用于区分是pem内容还是hex内容
    k := string(keyData)
    if strings.HasPrefix(k, SM2_PRIVATE_BEGIN) { // pem内容
        return x509.ReadPrivateKeyFromPem(keyData, pwd)
    }
    formattedPrivateKey := FormatSM2PrivateKey(string(keyData))
    return x509.ReadPrivateKeyFromPem(formattedPrivateKey, pwd)
}

// LoadSM2PublicKeyFromPEM 从pem文件加载公钥
func LoadSM2PublicKeyFromPEM(path string) (*sm2.PublicKey, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    return LoadSM2PublicKey(data)
}

// LoadSM2PublicKey 从pem文件加载公钥
func LoadSM2PublicKey(keyData []byte) (*sm2.PublicKey, error) {
    k := string(keyData)
    if strings.HasPrefix(k, SM2_PUBLIC_BEGIN) { // pem内容
        return x509.ReadPublicKeyFromPem(keyData)
    }
    formattedPublicKey := FormatSM2PublicKey(string(keyData))
    return x509.ReadPublicKeyFromPem(formattedPublicKey)
}

// SM2Sign 实现SM2签名
func SM2Sign(privateKey *sm2.PrivateKey, data string) (string, error) {
    hashVal, err := privateKey.Sign(rand.Reader, []byte(data), nil)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(hashVal), nil
}

// SM2Verify 实现SM2验签
func SM2Verify(publicKey *sm2.PublicKey, data, sign string) (bool, error) {
    signBytes, err := hex.DecodeString(sign)
    if err != nil {
        return false, err
    }
    return publicKey.Verify([]byte(data), signBytes), nil
}

// HashSM3 实现SM3哈希摘要
func HashSM3(data []byte) []byte {
    h := sm3.New()
    h.Write(data)
    return h.Sum(nil)
}

// SM3 实现SM3哈希
func SM3(data string) string {
    return hex.EncodeToString(HashSM3([]byte(data)))
}

// ReverseByteSlice 将字节切片顺序反转
func ReverseByteSlice(s []byte) []byte {
    for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
        s[i], s[j] = s[j], s[i] // 双指针交换元素
    }
    return s
}

// SetSM4IV 设置sm4全局向量
// 注意：iv必须是16字节
func SetSM4IV(iv []byte) error {
    return sm4.SetIV(iv)
}

// SM4CBCEncrypt sm4 cbc模式加密
func SM4CBCEncrypt(data, key string) (string, error) {
    k := []byte(key)
    if len(k) != 16 {
        return "", fmt.Errorf("invalid key")
    }
    cipher, err := sm4.Sm4Cbc(k, []byte(data), true) // true表示加密
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(cipher), nil
}

// SM4CBCDecrypt sm4 cbc模式解密
func SM4CBCDecrypt(cipher, key string) (string, error) {
    k := []byte(key)
    if len(k) != 16 {
        return "", fmt.Errorf("invalid key")
    }
    d, err := hex.DecodeString(cipher)
    if err != nil {
        return "", err
    }
    plain, err := sm4.Sm4Cbc(k, d, false) // false表示解密
    if err != nil {
        return "", err
    }
    return string(plain), nil
}

// SM4ECBEncrypt sm4 ecb模式加密
func SM4ECBEncrypt(data, key string) (string, error) {
    k := []byte(key)
    if len(k) != 16 {
        return "", fmt.Errorf("invalid key")
    }
    hashVal, err := sm4.Sm4Ecb(k, []byte(data), true)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(hashVal), nil
}

// SM4ECBDecrypt sm4 ecb模式解密
func SM4ECBDecrypt(cipher, key string) (string, error) {
    k := []byte(key)
    if len(k) != 16 {
        return "", fmt.Errorf("invalid key")
    }
    // 获取密文字节数组
    d, err := hex.DecodeString(cipher)
    if err != nil {
        return "", err
    }
    plain, err := sm4.Sm4Ecb(k, d, false)
    if err != nil {
        return "", err
    }
    return string(plain), nil
}
