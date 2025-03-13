package encrypt

import (
    "crypto"
    "crypto/md5"
    "crypto/sha1"
    "crypto/sha256"
    "crypto/sha512"
)

func Hash(plain []byte, algo crypto.Hash) []byte {
    var digest []byte
    switch algo {
    case crypto.MD5:
        digest = HashMD5(plain)
    case crypto.SHA1:
        digest = HashSHA1(plain)
    case crypto.SHA256:
        digest = HashSHA256(plain)
    case crypto.SHA512:
        digest = HashSHA512(plain)
    }
    return digest
}

func HashMD5(data []byte) []byte {
    h := md5.New()
    h.Write(data)
    return h.Sum(nil)
}

func HashSHA1(data []byte) []byte {
    h := sha1.New()
    h.Write(data)
    return h.Sum(nil)
}

func HashSHA256(data []byte) []byte {
    v := sha256.Sum256(data)
    return v[:]
}

func HashSHA512(data []byte) []byte {
    h := sha512.New()
    h.Write(data)
    return h.Sum(nil)
}
