package tls

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WrapHTTPServer 为HTTP服务器启用mTLS
func WrapHTTPServer(server *http.Server, tlsConfig *tls.Config) *http.Server {
	if tlsConfig == nil {
		return server
	}

	server.TLSConfig = tlsConfig
	return server
}

// StartHTTPSServer 启动HTTPS服务器（支持mTLS）
func StartHTTPSServer(addr string, handler *gin.Engine, tlsConfig *tls.Config, certFile, keyFile string) error {
	server := &http.Server{
		Addr:      addr,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	if tlsConfig != nil {
		return server.ListenAndServeTLS(certFile, keyFile)
	}

	return server.ListenAndServe()
}

// GetClientCertInfo 从请求中提取客户端证书信息
func GetClientCertInfo(c *gin.Context) (map[string]string, error) {
	if c.Request.TLS == nil {
		return nil, fmt.Errorf("非TLS连接")
	}

	if len(c.Request.TLS.PeerCertificates) == 0 {
		return nil, fmt.Errorf("客户端未提供证书")
	}

	cert := c.Request.TLS.PeerCertificates[0]
	info := map[string]string{
		"CommonName":   cert.Subject.CommonName,
		"Organization": "",
		"Issuer":       cert.Issuer.CommonName,
		"NotBefore":    cert.NotBefore.String(),
		"NotAfter":     cert.NotAfter.String(),
	}

	if len(cert.Subject.Organization) > 0 {
		info["Organization"] = cert.Subject.Organization[0]
	}

	return info, nil
}

// MTLSMiddleware mTLS验证中间件（可选，记录客户端证书信息）
func MTLSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.TLS != nil && len(c.Request.TLS.PeerCertificates) > 0 {
			cert := c.Request.TLS.PeerCertificates[0]
			// 将客户端服务名存入context
			c.Set("client_service", cert.Subject.CommonName)
		}
		c.Next()
	}
}
