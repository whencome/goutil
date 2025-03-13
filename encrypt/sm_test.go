package encrypt

import (
    "testing"
)

// 定义sm2私钥和公钥
var sm2PriKey = `-----BEGIN PRIVATE KEY-----
MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgE49ljZvg5QDOUxkf
848wojFUxE9q5J0iYJ/AZZOfOHCgCgYIKoEcz1UBgi2hRANCAAQ+HIktKsUEbgDV
1uwmyYI/Fl5xQqSWPkJpMCwHFdBW7MZh/u8YKo1yBamjj7poqeMc4yepA5H0OSJP
i7Sy7KaZ
-----END PRIVATE KEY-----`
var sm2PubKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEPhyJLSrFBG4A1dbsJsmCPxZecUKk
lj5CaTAsBxXQVuzGYf7vGCqNcgWpo4+6aKnjHOMnqQOR9DkiT4u0suymmQ==
-----END PUBLIC KEY-----`

var sm2PriKeyHex = `MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgE49ljZvg5QDOUxkf848wojFUxE9q5J0iYJ/AZZOfOHCgCgYIKoEcz1UBgi2hRANCAAQ+HIktKsUEbgDV1uwmyYI/Fl5xQqSWPkJpMCwHFdBW7MZh/u8YKo1yBamjj7poqeMc4yepA5H0OSJPi7Sy7KaZ`
var sm2PubKeyHex = `MFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEPhyJLSrFBG4A1dbsJsmCPxZecUKklj5CaTAsBxXQVuzGYf7vGCqNcgWpo4+6aKnjHOMnqQOR9DkiT4u0suymmQ==`

// TestSM2 测试SM2
func TestSM2(t *testing.T) {
    data := "hello, world!"
    priKey, err := LoadSM2PrivateKey([]byte(sm2PriKey), "")
    if err != nil {
        t.Logf("load sm2 private key fail: %v\n", err)
        t.Fail()
    }
    pubKey, err := LoadSM2PublicKey([]byte(sm2PubKey))
    if err != nil {
        t.Logf("load sm2 public key fail: %v\n", err)
        t.Fail()
    }
    sign, err := SM2Sign(priKey, data)
    if err != nil {
        t.Logf("sm2 sign fail: %v\n", err)
        t.Fail()
    }
    t.Logf("sign result: %s => %s\n", data, sign)
    valid, err := SM2Verify(pubKey, data, sign)
    if err != nil {
        t.Logf("sm2 verify fail: %v\n", err)
        t.Fail()
    }
    if !valid {
        t.Logf("sm2 verify fail: not passed\n")
        t.Fail()
    }
    t.Logf("sm2 verify success\n")
}

// TestSM2V1 测试SM2
func TestSM2V1(t *testing.T) {
    data := "hello, world!"
    priKey, err := LoadSM2PrivateKey([]byte(sm2PriKeyHex), "")
    if err != nil {
        t.Logf("load sm2 private key fail: %v\n", err)
        t.Fail()
    }
    pubKey, err := LoadSM2PublicKey([]byte(sm2PubKeyHex))
    if err != nil {
        t.Logf("load sm2 public key fail: %v\n", err)
        t.Fail()
    }
    sign, err := SM2Sign(priKey, data)
    if err != nil {
        t.Logf("sm2 sign fail: %v\n", err)
        t.Fail()
    }
    t.Logf("sign result: %s => %s\n", data, sign)
    valid, err := SM2Verify(pubKey, data, sign)
    if err != nil {
        t.Logf("sm2 verify fail: %v\n", err)
        t.Fail()
    }
    if !valid {
        t.Logf("sm2 verify fail: not passed\n")
        t.Fail()
    }
    t.Logf("sm2 verify success\n")
}

// TestSM3 测试SM3
func TestSM3(t *testing.T) {
    data := "hello, world!"
    hash := SM3(data)
    t.Logf("%s => %s\n", data, hash)
}

// TestSM4CBC 测试SM4 cbc模式加解密
func TestSM4CBC(t *testing.T) {
    iv := "1234567890abcdef"
    key := "aaaabbbbccccdddd"
    data := "hello, world!"
    err := SetSM4IV([]byte(iv))
    if err != nil {
        t.Logf("set sm4 iv fail: %v\n", err)
        t.Fail()
    }
    cipher, err := SM4CBCEncrypt(data, key)
    if err != nil {
        t.Logf("do sm4 cbc encrypt fail: %v\n", err)
        t.Fail()
    }
    t.Logf("encrypt: %s => %s\n", data, cipher)
    plain, err := SM4CBCDecrypt(cipher, key)
    if err != nil {
        t.Logf("do sm4 cbc decrypt fail: %v\n", err)
        t.Fail()
    }
    t.Logf("decrypt: %s => %s\n", cipher, plain)
    if plain != data {
        t.Logf("compare to origin data fail: expect %s; got %s\n", data, plain)
        t.Fail()
    }
    t.Logf("success\n")
}

// TestSM4ECB 测试SM4 ecb模式加解密
func TestSM4ECB(t *testing.T) {
    iv := "1234567890abcdef"
    key := "aaaabbbbccccdddd"
    data := "hello, world!"
    err := SetSM4IV([]byte(iv))
    if err != nil {
        t.Logf("set sm4 iv fail: %v\n", err)
        t.Fail()
    }
    cipher, err := SM4ECBEncrypt(data, key)
    if err != nil {
        t.Logf("do sm4 cbc encrypt fail: %v\n", err)
        t.Fail()
    }
    t.Logf("encrypt: %s => %s\n", data, cipher)
    plain, err := SM4ECBDecrypt(cipher, key)
    if err != nil {
        t.Logf("do sm4 cbc decrypt fail: %v\n", err)
        t.Fail()
    }
    t.Logf("decrypt: %s => %s\n", cipher, plain)
    if plain != data {
        t.Logf("compare to origin data fail: expect %s; got %s\n", data, plain)
        t.Fail()
    }
    t.Logf("success\n")
}
