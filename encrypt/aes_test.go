package encrypt

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"
)

func TestAESCBCRoundTrip(t *testing.T) {
	key := []byte("1234567890abcdef")
	data := []byte("hello, world! this is a test for aes cbc")
	cipherTxt, err := AESCBCEncrypt(data, key)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	plain, err := AESCBCDecrypt(cipherTxt, key)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if !bytes.Equal(plain, data) {
		t.Fatalf("decrypt got %q, want %q", plain, data)
	}
}

func TestAESCBCDecryptWithWrongKey(t *testing.T) {
	data := []byte("hello, world!")
	cipherTxt, err := AESCBCEncrypt(data, []byte("1234567890abcdef"))
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	// 使用错误的密钥解密，应该返回错误而不是无意义的数据（历史实现可能会panic）
	if _, err := AESCBCDecrypt(cipherTxt, []byte("abcdef1234567890")); err == nil {
		t.Fatal("decrypt with a wrong key should return error")
	}
	// 非法的密文长度
	if _, err := AESCBCDecrypt([]byte("short"), []byte("1234567890abcdef")); err == nil {
		t.Fatal("decrypt with a invalid ciphertext should return error")
	}
}

func TestAESECBRoundTrip(t *testing.T) {
	key := []byte("1234567890abcdef")
	data := []byte("hello, world!")
	cipherTxt, err := AesECBEncrypt(data, key)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}
	plain, err := AesECBDecrypt(cipherTxt, key)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if !bytes.Equal(plain, data) {
		t.Fatalf("decrypt got %q, want %q", plain, data)
	}
	if _, err := AesECBDecrypt(cipherTxt, []byte("abcdef1234567890")); err == nil {
		t.Fatal("decrypt with a wrong key should return error")
	}
}

func TestPKCS7PaddingEdge(t *testing.T) {
	// 非法分组长度不应panic，也不应修改入参
	org := []byte("abc")
	ret := PKCS7Padding(org, 0)
	if !bytes.Equal(ret, org) {
		t.Fatalf("PKCS7Padding with invalid block size got %v", ret)
	}
	// 空数据去填充不应panic
	if got := PKCS7UnPadding(nil); len(got) != 0 {
		t.Fatalf("PKCS7UnPadding(nil) = %v", got)
	}
	// 非法填充值不应panic
	invalid := []byte{1, 2, 3, 9}
	if got := PKCS7UnPadding(invalid); !bytes.Equal(got, invalid) {
		t.Fatalf("PKCS7UnPadding with invalid padding = %v", got)
	}
	// 正常填充
	padded := PKCS7Padding([]byte("1234567890abcdef"), 16)
	if len(padded) != 32 {
		t.Fatalf("unexpected padded length: %d", len(padded))
	}
	if got := PKCS7UnPadding(padded); !bytes.Equal(got, []byte("1234567890abcdef")) {
		t.Fatalf("PKCS7UnPadding got %q", got)
	}
}

// TestGetRsaKeyWithNonRSAKey 测试传入非RSA密钥时返回错误而不是panic
func TestGetRsaKeyWithNonRSAKey(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ecdsa key failed: %v", err)
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		t.Fatalf("marshal pkcs8 failed: %v", err)
	}
	pemData := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})
	if _, err := GetRsaPrivateKey(pemData); err == nil {
		t.Fatal("GetRsaPrivateKey with an ecdsa key should return error")
	}
	pubDer, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key failed: %v", err)
	}
	pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDer})
	if _, err := GetRsaPublicKey(pubPem); err == nil {
		t.Fatal("GetRsaPublicKey with an ecdsa key should return error")
	}
}

// TestRsaSignAndVerify 测试RSA签名与验签
func TestRsaSignAndVerify(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key failed: %v", err)
	}
	data := "hello, rsa"
	sign, err := GetRsaSHA256Sign(data, priv)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}
	if err := VerifyRsaSHA256Sign(data, sign, &priv.PublicKey); err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if err := VerifyRsaSHA256Sign(data+"x", sign, &priv.PublicKey); err == nil {
		t.Fatal("verify should fail for modified data")
	}
}
