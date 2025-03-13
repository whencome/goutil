package encrypt

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "fmt"
    "io"
)

// AESCBCDecrypt aes CBC模式解密
// AES算法支持的密钥长度为128位（16字节）、192位（24字节）和256位（32字节）
func AESCBCDecrypt(cipherTxt []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    if len(cipherTxt) < aes.BlockSize {
        return nil, fmt.Errorf("ciphertext too short")
    }

    // IV是密文的第一个块
    iv := cipherTxt[:aes.BlockSize]
    cipherTxt = cipherTxt[aes.BlockSize:]

    // 检查密文长度是否为块大小的倍数
    if len(cipherTxt)%aes.BlockSize != 0 {
        return nil, fmt.Errorf("ciphertext is not a multiple of the block size")
    }

    mode := cipher.NewCBCDecrypter(block, iv)

    // 解密
    plaintext := make([]byte, len(cipherTxt))
    mode.CryptBlocks(plaintext, cipherTxt)

    // 删除填充
    plaintext = PKCS7UnPadding(plaintext)
    return plaintext, nil
}

// AESCBCEncrypt AES CBC模式加密
// AES算法支持的密钥长度为128位（16字节）、192位（24字节）和256位（32字节）
func AESCBCEncrypt(org []byte, key []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    // 对数据进行PKCS7填充
    org = PKCS7Padding(org, block.BlockSize())

    // 初始化向量IV必须是唯一的，但不需要保密
    ciphertext := make([]byte, aes.BlockSize+len(org))
    iv := ciphertext[:aes.BlockSize]
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
        return nil, err
    }

    // 使用CBC模式加密
    mode := cipher.NewCBCEncrypter(block, iv)
    mode.CryptBlocks(ciphertext[aes.BlockSize:], org)

    return ciphertext, nil
}

// AesECBEncrypt 实现AES-128-ECB加密
// AES算法支持的密钥长度为128位（16字节）、192位（24字节）和256位（32字节）。
func AesECBEncrypt(data []byte, key []byte) ([]byte, error) {
    // 获取加密算法
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    bs := block.BlockSize()
    data = PKCS7Padding(data, bs)
    if len(data)%bs != 0 {
        return nil, fmt.Errorf("need a multiple of the blocksize")
    }

    encrypted := make([]byte, len(data))
    for i := 0; i < len(data); i += bs {
        block.Encrypt(encrypted[i:i+bs], data[i:i+bs])
    }

    return encrypted, nil
}

// AesECBDecrypt 实现AES-128-ECB解密
func AesECBDecrypt(encrypted []byte, key []byte) ([]byte, error) {
    // 获取加密算法
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    bs := block.BlockSize()
    if len(encrypted)%bs != 0 {
        return nil, fmt.Errorf("encrypted data is not a multiple of the blocksize")
    }

    data := make([]byte, len(encrypted))
    for i := 0; i < len(encrypted); i += bs {
        block.Decrypt(data[i:i+bs], encrypted[i:i+bs])
    }

    data = PKCS7UnPadding(data)
    return data, nil
}
