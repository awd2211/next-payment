package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
)

// RSACrypto RSA加密工具
type RSACrypto struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewRSACrypto 创建RSA加密实例
func NewRSACrypto(bits int) (*RSACrypto, error) {
	if bits < 2048 {
		return nil, errors.New("RSA密钥长度不能小于2048位")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	return &RSACrypto{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

// NewRSACryptoFromKeys 从已有密钥创建RSA加密实例
func NewRSACryptoFromKeys(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) *RSACrypto {
	return &RSACrypto{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

// Encrypt 使用公钥加密数据
func (c *RSACrypto) Encrypt(plaintext []byte) (string, error) {
	if c.publicKey == nil {
		return "", errors.New("公钥未设置")
	}

	ciphertext, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		c.publicKey,
		plaintext,
		nil,
	)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 使用私钥解密数据
func (c *RSACrypto) Decrypt(ciphertext string) ([]byte, error) {
	if c.privateKey == nil {
		return nil, errors.New("私钥未设置")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, err
	}

	return rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		c.privateKey,
		data,
		nil,
	)
}

// Sign 使用私钥签名数据
func (c *RSACrypto) Sign(data []byte) (string, error) {
	if c.privateKey == nil {
		return "", errors.New("私钥未设置")
	}

	hash := sha256.Sum256(data)
	signature, err := rsa.SignPKCS1v15(rand.Reader, c.privateKey, 0, hash[:])
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// Verify 使用公钥验证签名
func (c *RSACrypto) Verify(data []byte, signature string) error {
	if c.publicKey == nil {
		return errors.New("公钥未设置")
	}

	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(c.publicKey, 0, hash[:], sig)
}

// ExportPrivateKey 导出私钥为PEM格式
func (c *RSACrypto) ExportPrivateKey() (string, error) {
	if c.privateKey == nil {
		return "", errors.New("私钥未设置")
	}

	privBytes := x509.MarshalPKCS1PrivateKey(c.privateKey)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	return string(privPEM), nil
}

// ExportPublicKey 导出公钥为PEM格式
func (c *RSACrypto) ExportPublicKey() (string, error) {
	if c.publicKey == nil {
		return "", errors.New("公钥未设置")
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(c.publicKey)
	if err != nil {
		return "", err
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	return string(pubPEM), nil
}

// LoadPrivateKey 从PEM格式加载私钥
func LoadPrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("无效的PEM格式")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// LoadPublicKey 从PEM格式加载公钥
func LoadPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("无效的PEM格式")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("不是RSA公钥")
	}

	return rsaPub, nil
}
