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
    RSA_PRIVATE_BEGIN           = "-----BEGIN PRIVATE KEY-----\n" // PKCS8格式以及通用格式
    RSA_PRIVATE_END             = "\n-----END PRIVATE KEY-----"
    RSA_PKCS1_PRIVATE_BEGIN     = "-----BEGIN RSA PRIVATE KEY-----\n" // PKCS1格式
    RSA_PKCS1_PRIVATE_END       = "\n-----END RSA PRIVATE KEY-----"
    RSA_ENCRYPTED_PRIVATE_BEGIN = "-----BEGIN ENCRYPTED PRIVATE KEY-----\n" // 私钥加密内容
    RSA_ENCRYPTED_PRIVATE_END   = "\n-----END ENCRYPTED PRIVATE KEY-----"
    RSA_PKCS1_PUBLIC_BEGIN      = "-----BEGIN RSA PUBLIC KEY-----\n"
    RSA_PKCS1_PUBLIC_END        = "\n-----END RSA PUBLIC KEY-----"
    RSA_PUBLIC_BEGIN            = "-----BEGIN PUBLIC KEY-----\n"
    RSA_PUBLIC_END              = "\n-----END PUBLIC KEY-----"
)

// detectRsaPublicKeyFormat 探测公钥格式
// 第一个bool值表示是否探测成功
// 第二个是密钥格式，默认为PKCS#8
func detectRsaPublicKeyFormat(pubKey string) (bool, string) {
    var pubKeyBytes []byte
    pubKey = strings.TrimSpace(pubKey)
    if strings.HasPrefix(pubKey, "-----") {
        pubKeyBytes = []byte(pubKey)
    } else {
        pubKeyBytes = []byte(RSA_PUBLIC_BEGIN + pubKey + RSA_PUBLIC_END)
    }
    block, _ := pem.Decode(pubKeyBytes)
    if block == nil {
        return false, PKCS8
    }
    // try pkcs#8 format
    if _, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
        return true, PKCS8
    }
    // Attempt to parse as PKCS#1
    if _, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
        return true, PKCS1
    }
    return false, PKCS8
}

// detectRsaPrivateKeyFormat 探测私钥格式
// 第一个bool值表示是否探测成功
// 第二个是密钥格式，默认为PKCS#8
func detectRsaPrivateKeyFormat(privKey string) (bool, string) {
    var privKeyBytes []byte
    privKey = strings.TrimSpace(privKey)
    if strings.HasPrefix(privKey, "-----") {
        privKeyBytes = []byte(privKey)
    } else {
        privKeyBytes = []byte(RSA_PRIVATE_BEGIN + privKey + RSA_PRIVATE_END)
    }
    block, _ := pem.Decode(privKeyBytes)
    if block == nil {
        return false, PKCS8
    }
    // try pkcs#1 format
    if _, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
        return true, PKCS1
    }
    // try pkcs#8 format
    if _, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
        return true, PKCS8
    }
    return false, PKCS8
}

// FormatRsaPrivateKey 格式化私钥
func FormatRsaPrivateKey(privateKey string) ([]byte, string) {
    // 以下任意格式直接返回
    var wrapped = false
    if strings.HasPrefix(privateKey, RSA_PRIVATE_BEGIN) ||
        strings.HasPrefix(privateKey, RSA_PKCS1_PRIVATE_BEGIN) ||
        strings.HasPrefix(privateKey, RSA_ENCRYPTED_PRIVATE_BEGIN) {
        wrapped = true
    }
    _, format := detectRsaPrivateKeyFormat(privateKey)
    switch format {
    case PKCS1:
        if !wrapped {
            privateKey = RSA_PKCS1_PRIVATE_BEGIN + privateKey + RSA_PKCS1_PRIVATE_END
        }
    case PKCS8:
        if !wrapped {
            privateKey = RSA_PRIVATE_BEGIN + privateKey + RSA_PRIVATE_END
        }
    }
    return []byte(privateKey), format
}

// FormatRsaPublicKey 格式化公钥
func FormatRsaPublicKey(publicKey string) ([]byte, string) {
    // 以下任意格式直接返回
    var wrapped = false
    if strings.HasPrefix(publicKey, RSA_PUBLIC_BEGIN) ||
        strings.HasPrefix(publicKey, RSA_PKCS1_PUBLIC_BEGIN) {
        wrapped = true
    }
    _, format := detectRsaPublicKeyFormat(publicKey)
    switch format {
    case PKCS1:
        if !wrapped {
            publicKey = RSA_PKCS1_PUBLIC_BEGIN + publicKey + RSA_PKCS1_PUBLIC_END
        }
    case PKCS8:
        if !wrapped {
            publicKey = RSA_PUBLIC_BEGIN + publicKey + RSA_PUBLIC_END
        }
    }
    return []byte(publicKey), format
}

// 检查算法是否支持
func isRsaAlgorithmSupported(algo crypto.Hash) bool {
    if algo == crypto.MD5 ||
        algo == crypto.SHA1 ||
        algo == crypto.SHA256 ||
        algo == crypto.SHA512 {
        return true
    }
    return false
}

// DetectRsaPrivateKey 探测并解析RSA私钥
func DetectRsaPrivateKey(privateKey []byte) (*rsa.PrivateKey, string, error) {
    // 解析私钥
    var rsaPrivateKey *rsa.PrivateKey
    var format string
    var err error
    block, _ := pem.Decode(privateKey)
    if block == nil {
        return nil, format, fmt.Errorf("rsaSign pem.Decode error")
    }
    if block.Type == "RSA PRIVATE KEY" { // PKCS#1格式
        format = PKCS1
        rsaPrivateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
        if err != nil {
            return nil, format, err
        }
    } else if block.Type == "PRIVATE KEY" { // PKCS#8格式
        format = PKCS8
        k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
        if err != nil {
            return nil, format, err
        }
        rsaKey, ok := k.(*rsa.PrivateKey)
        if !ok {
            return nil, format, errors.New("illegal rsa private key")
        }
        rsaPrivateKey = rsaKey
    } else {
        return nil, format, fmt.Errorf("unsupported rsa key type")
    }
    return rsaPrivateKey, format, nil
}

// GetRsaPrivateKey 获取rsa私钥信息
func GetRsaPrivateKey(privateKey []byte) (*rsa.PrivateKey, error) {
    privateKey, format := FormatRsaPrivateKey(string(privateKey))
    block, _ := pem.Decode(privateKey)
    if block == nil {
        return nil, fmt.Errorf("rsaSign pem.Decode error")
    }
    var priKey *rsa.PrivateKey
    var err error
    switch format {
    case PKCS1:
        priKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
    case PKCS8:
        theKey, e := x509.ParsePKCS8PrivateKey(block.Bytes)
        if e != nil {
            err = e
        } else {
            priKey = theKey.(*rsa.PrivateKey)
        }
    default:
        err = errors.New("unsupported private key format")
    }
    return priKey, err
}

// GetRsaPublicKey 获取rsa公钥信息
func GetRsaPublicKey(publicKey []byte) (*rsa.PublicKey, error) {
    publicKey, format := FormatRsaPublicKey(string(publicKey))
    block, _ := pem.Decode(publicKey)
    if block == nil {
        return nil, errors.New("decode public key failed")
    }
    var pubKey *rsa.PublicKey
    var err error
    switch format {
    case PKCS1:
        pubKey, err = x509.ParsePKCS1PublicKey(block.Bytes)
    case PKCS8:
        pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
        if err != nil {
            return nil, errors.New("parse public key failed")
        }
        pubKey = pubInterface.(*rsa.PublicKey) //pub:公钥对象
    default:
        err = errors.New("unsupported public key format")
    }
    return pubKey, err
}

// RsaSignRaw 获取RSA签名
func RsaSignRaw(plain []byte, key []byte, algo crypto.Hash) (string, error) {
    // 私钥检测
    privateKey, _, err := DetectRsaPrivateKey(key)
    if err != nil {
        return "", err
    }
    // 算法检测
    if !isRsaAlgorithmSupported(algo) {
        return "", fmt.Errorf("algorithm not supported")
    }
    // 对原始数据进行hash
    digest := Hash(plain, algo)
    // 生成签名
    s, err := rsa.SignPKCS1v15(rand.Reader, privateKey, algo, digest)
    if err != nil {
        return "", err
    }
    hashVal := base64.StdEncoding.EncodeToString(s)
    return hashVal, nil
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
    return data, nil
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

// RsaEncrypt 加密
func RsaEncrypt(origData []byte, publicKey []byte) ([]byte, error) {
    pub, err := GetRsaPublicKey(publicKey)
    if err != nil {
        return nil, err
    }
    return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// RsaDecrypt 解密
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
