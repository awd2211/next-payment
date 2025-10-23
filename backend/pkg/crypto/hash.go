package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"hash"
)

// Hash 计算哈希值
type Hash struct{}

// NewHash 创建Hash实例
func NewHash() *Hash {
	return &Hash{}
}

// MD5 计算MD5哈希
func (h *Hash) MD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// SHA1 计算SHA1哈希
func (h *Hash) SHA1(data []byte) string {
	hash := sha1.Sum(data)
	return hex.EncodeToString(hash[:])
}

// SHA256 计算SHA256哈希
func (h *Hash) SHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// SHA512 计算SHA512哈希
func (h *Hash) SHA512(data []byte) string {
	hash := sha512.Sum512(data)
	return hex.EncodeToString(hash[:])
}

// HMAC HMAC签名
type HMAC struct {
	key []byte
}

// NewHMAC 创建HMAC实例
func NewHMAC(key []byte) *HMAC {
	return &HMAC{key: key}
}

// Sign 使用HMAC签名数据
func (h *HMAC) Sign(data []byte, hashFunc func() hash.Hash) string {
	mac := hmac.New(hashFunc, h.key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// SignSHA256 使用HMAC-SHA256签名
func (h *HMAC) SignSHA256(data []byte) string {
	return h.Sign(data, sha256.New)
}

// SignSHA512 使用HMAC-SHA512签名
func (h *HMAC) SignSHA512(data []byte) string {
	return h.Sign(data, sha512.New)
}

// Verify 验证HMAC签名
func (h *HMAC) Verify(data []byte, signature string, hashFunc func() hash.Hash) bool {
	expectedSig := h.Sign(data, hashFunc)
	return hmac.Equal([]byte(expectedSig), []byte(signature))
}

// VerifySHA256 验证HMAC-SHA256签名
func (h *HMAC) VerifySHA256(data []byte, signature string) bool {
	return h.Verify(data, signature, sha256.New)
}

// VerifySHA512 验证HMAC-SHA512签名
func (h *HMAC) VerifySHA512(data []byte, signature string) bool {
	return h.Verify(data, signature, sha512.New)
}
