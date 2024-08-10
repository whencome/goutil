package encrypt

import (
    "crypto"
    "crypto/rand"
    "crypto/rsa"
    "crypto/sha1"
    "crypto/sha256"
    "crypto/x509"
    "encoding/base64"
    "encoding/pem"
    "errors"
    "fmt"
    "strings"
)

const (
    RSA_PRIVATE_BEGIN = "-----BEGIN RSA PRIVATE KEY-----\n"
    RSA_PRIVATE_END   = "\n-----END RSA PRIVATE KEY-----"
    RSA_PUBLIC_BEGIN  = "-----BEGIN PUBLIC KEY-----\n"
    RSA_PUBLIC_END    = "\n-----END PUBLIC KEY-----"
)

// FormatRsaPrivateKey 格式化私钥
func FormatRsaPrivateKey(privateKey string) []byte {
    if !strings.HasPrefix(privateKey, RSA_PRIVATE_BEGIN) {
        privateKey = RSA_PRIVATE_BEGIN + privateKey
    }
    if !strings.HasSuffix(privateKey, RSA_PRIVATE_END) {
        privateKey = privateKey + RSA_PRIVATE_END
    }
    return []byte(privateKey)
}

// FormatRsaPublicKey 格式化公钥
func FormatRsaPublicKey(publicKey string) []byte {
    if !strings.HasPrefix(publicKey, RSA_PUBLIC_BEGIN) {
        publicKey = RSA_PUBLIC_BEGIN + publicKey
    }
    if !strings.HasSuffix(publicKey, RSA_PUBLIC_END) {
        publicKey = publicKey + RSA_PUBLIC_END
    }
    return []byte(publicKey)
}

// 检查算法是否支持
func isRsaAlgorithmSupported(algo crypto.Hash) bool {
    if algo == crypto.SHA1 ||
        algo == crypto.SHA256 ||
        algo == crypto.SHA512 {
        return true
    }
    return false
}

// GetRsaPrivateKey 获取rsa私钥信息
func GetRsaPrivateKey(privateKey []byte) (*rsa.PrivateKey, error) {
    privateKey = FormatRsaPrivateKey(string(privateKey))
    block, _ := pem.Decode(privateKey)
    if block == nil {
        return nil, fmt.Errorf("rsaSign pem.Decode error")
    }
    rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
        return nil, err
    }
    return rsaPrivateKey, nil
}

// GetRsaPublicKey 获取rsa公钥信息
func GetRsaPublicKey(publicKey []byte) (*rsa.PublicKey, error) {
    publicKey = FormatRsaPublicKey(string(publicKey))
    block, _ := pem.Decode(publicKey)
    if block == nil {
        return nil, errors.New("decode public key failed")
    }
    pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        return nil, errors.New("parse public key failed")
    }
    pub := pubInterface.(*rsa.PublicKey) //pub:公钥对象
    return pub, nil
}

// GetRsaSign 获取RSA签名
func GetRsaSign(digest []byte, privateKey *rsa.PrivateKey, algo crypto.Hash) (string, error) {
    if !isRsaAlgorithmSupported(algo) {
        return "", fmt.Errorf("algorithm not supported")
    }
    s, err := rsa.SignPKCS1v15(rand.Reader, privateKey, algo, digest)
    if err != nil {
        return "", err
    }
    data := base64.StdEncoding.EncodeToString(s)
    return string(data), nil
}

// GetRsaSHA1Sign 获取RSA签名(sha1)
func GetRsaSHA1Sign(origData string, privateKey *rsa.PrivateKey) (string, error) {
    h := sha1.New()
    h.Write([]byte(origData))
    digest := h.Sum(nil)
    return GetRsaSign(digest, privateKey, crypto.SHA1)
}

// GetRsaSHA256Sign 获取Rsa签名（sha256）
func GetRsaSHA256Sign(origData string, privateKey *rsa.PrivateKey) (string, error) {
    digest := sha256.Sum256([]byte(origData))
    return GetRsaSign(digest[:], privateKey, crypto.SHA256)
}

// VerifyRsaSign 验证Rsa签名
func VerifyRsaSign(data []byte, sign []byte, publicKey *rsa.PublicKey, algo crypto.Hash) error {
    if !isRsaAlgorithmSupported(algo) {
        return fmt.Errorf("algorithm not supported")
    }
    err := rsa.VerifyPKCS1v15(publicKey, algo, data, sign)
    if err != nil {
        return err
    }
    return nil
}

// VerifyRsaSHA1Sign 验证Rsa签名
func VerifyRsaSHA1Sign(data, sign string, publicKey *rsa.PublicKey) error {
    // 计算hash
    h := sha1.New()
    h.Write([]byte(data))
    digest := h.Sum(nil)
    // 解析签名
    signBytes, err := base64.StdEncoding.DecodeString(sign)
    if err != nil {
        return err
    }
    // 验签
    return VerifyRsaSign(digest, signBytes, publicKey, crypto.SHA1)
}

// VerifyRsaSHA256Sign 验证Rsa签名
func VerifyRsaSHA256Sign(data, sign string, publicKey *rsa.PublicKey) error {
    // 计算hash
    digest := sha256.Sum256([]byte(data))
    // 解析签名
    signBytes, err := base64.StdEncoding.DecodeString(sign)
    if err != nil {
        return err
    }
    // 验签
    return VerifyRsaSign(digest[:], signBytes, publicKey, crypto.SHA256)
}

// 加密
func RsaEncrypt(origData []byte, publicKey []byte) ([]byte, error) {
    pub, err := GetRsaPublicKey(publicKey)
    if err != nil {
        return nil, err
    }
    return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// 解密
func RsaDecrypt(cipherText []byte, privateKey []byte) ([]byte, error) {
    priv, err := GetRsaPrivateKey(privateKey)
    if err != nil {
        return nil, err
    }
    return rsa.DecryptPKCS1v15(rand.Reader, priv, cipherText)
}

// RsaSign 对给定的数据（data）使用私钥priKey进行签名
func RsaSign(data, priKey string) (string, error) {
    privateKey, err := GetRsaPrivateKey([]byte(priKey))
    if err != nil {
        return "", err
    }
    return GetRsaSHA1Sign(data, privateKey)
}

// Rsa2Sign 对给定的数据（data）使用私钥priKey进行签名
func Rsa2Sign(data, priKey string) (string, error) {
    privateKey, err := GetRsaPrivateKey([]byte(priKey))
    if err != nil {
        return "", err
    }
    return GetRsaSHA256Sign(data, privateKey)
}

// RsaVerifySign verify rsa sign
func RsaVerifySign(data, sign, pubKey string) error {
    publicKey, err := GetRsaPublicKey([]byte(pubKey))
    if err != nil {
        return err
    }
    return VerifyRsaSHA1Sign(data, sign, publicKey)
}

// Rsa2VerifySign verify rsa2 sign
func Rsa2VerifySign(data, sign, pubKey string) error {
    publicKey, err := GetRsaPublicKey([]byte(pubKey))
    if err != nil {
        return err
    }
    return VerifyRsaSHA256Sign(data, sign, publicKey)
}
