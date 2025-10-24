package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/payment-platform/pkg/config"
)

// Config TLS配置
type Config struct {
	// 服务端配置
	CertFile string // 服务证书路径
	KeyFile  string // 服务私钥路径
	CAFile   string // CA证书路径（用于验证客户端）

	// 客户端配置
	ClientCertFile string // 客户端证书路径
	ClientKeyFile  string // 客户端私钥路径

	// 通用配置
	EnableMTLS       bool   // 是否启用mTLS
	ServerName       string // 服务器名称（用于SNI）
	InsecureSkipVerify bool // 跳过证书验证（仅测试）
}

// LoadFromEnv 从环境变量加载TLS配置
func LoadFromEnv() *Config {
	return &Config{
		EnableMTLS:         config.GetEnvBool("ENABLE_MTLS", false),
		CertFile:           config.GetEnv("TLS_CERT_FILE", ""),
		KeyFile:            config.GetEnv("TLS_KEY_FILE", ""),
		CAFile:             config.GetEnv("TLS_CA_FILE", ""),
		ClientCertFile:     config.GetEnv("TLS_CLIENT_CERT", ""),
		ClientKeyFile:      config.GetEnv("TLS_CLIENT_KEY", ""),
		ServerName:         config.GetEnv("TLS_SERVER_NAME", ""),
		InsecureSkipVerify: config.GetEnvBool("TLS_INSECURE_SKIP_VERIFY", false),
	}
}

// NewServerTLSConfig 创建服务端TLS配置（支持mTLS）
func NewServerTLSConfig(cfg *Config) (*tls.Config, error) {
	if !cfg.EnableMTLS {
		return nil, nil
	}

	// 加载服务端证书和私钥
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("加载服务端证书失败: %w", err)
	}

	// 加载CA证书（用于验证客户端）
	caCert, err := os.ReadFile(cfg.CAFile)
	if err != nil {
		return nil, fmt.Errorf("读取CA证书失败: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("解析CA证书失败")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert, // 要求客户端提供证书并验证
		MinVersion:   tls.VersionTLS12,               // 最低TLS 1.2
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		PreferServerCipherSuites: true,
	}

	return tlsConfig, nil
}

// NewClientTLSConfig 创建客户端TLS配置（支持mTLS）
func NewClientTLSConfig(cfg *Config) (*tls.Config, error) {
	if !cfg.EnableMTLS {
		return nil, nil
	}

	// 加载客户端证书和私钥
	cert, err := tls.LoadX509KeyPair(cfg.ClientCertFile, cfg.ClientKeyFile)
	if err != nil {
		return nil, fmt.Errorf("加载客户端证书失败: %w", err)
	}

	// 加载CA证书（用于验证服务端）
	caCert, err := os.ReadFile(cfg.CAFile)
	if err != nil {
		return nil, fmt.Errorf("读取CA证书失败: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("解析CA证书失败")
	}

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}

	// 设置ServerName（用于SNI）
	if cfg.ServerName != "" {
		tlsConfig.ServerName = cfg.ServerName
	}

	return tlsConfig, nil
}

// ValidateServerConfig 验证服务端TLS配置
func ValidateServerConfig(cfg *Config) error {
	if !cfg.EnableMTLS {
		return nil
	}

	if cfg.CertFile == "" {
		return fmt.Errorf("TLS_CERT_FILE 未配置")
	}
	if cfg.KeyFile == "" {
		return fmt.Errorf("TLS_KEY_FILE 未配置")
	}
	if cfg.CAFile == "" {
		return fmt.Errorf("TLS_CA_FILE 未配置")
	}

	// 检查文件是否存在
	if _, err := os.Stat(cfg.CertFile); os.IsNotExist(err) {
		return fmt.Errorf("证书文件不存在: %s", cfg.CertFile)
	}
	if _, err := os.Stat(cfg.KeyFile); os.IsNotExist(err) {
		return fmt.Errorf("私钥文件不存在: %s", cfg.KeyFile)
	}
	if _, err := os.Stat(cfg.CAFile); os.IsNotExist(err) {
		return fmt.Errorf("CA证书文件不存在: %s", cfg.CAFile)
	}

	return nil
}

// ValidateClientConfig 验证客户端TLS配置
func ValidateClientConfig(cfg *Config) error {
	if !cfg.EnableMTLS {
		return nil
	}

	if cfg.ClientCertFile == "" {
		return fmt.Errorf("TLS_CLIENT_CERT 未配置")
	}
	if cfg.ClientKeyFile == "" {
		return fmt.Errorf("TLS_CLIENT_KEY 未配置")
	}
	if cfg.CAFile == "" {
		return fmt.Errorf("TLS_CA_FILE 未配置")
	}

	// 检查文件是否存在
	if _, err := os.Stat(cfg.ClientCertFile); os.IsNotExist(err) {
		return fmt.Errorf("客户端证书文件不存在: %s", cfg.ClientCertFile)
	}
	if _, err := os.Stat(cfg.ClientKeyFile); os.IsNotExist(err) {
		return fmt.Errorf("客户端私钥文件不存在: %s", cfg.ClientKeyFile)
	}
	if _, err := os.Stat(cfg.CAFile); os.IsNotExist(err) {
		return fmt.Errorf("CA证书文件不存在: %s", cfg.CAFile)
	}

	return nil
}
