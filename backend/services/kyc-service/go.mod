module payment-platform/kyc-service

go 1.24.0

toolchain go1.24.6

require (
	github.com/gin-gonic/gin v1.11.0
	github.com/google/uuid v1.6.0
	github.com/payment-platform/pkg v0.0.0
	github.com/prometheus/client_golang v1.19.1
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.1
	gorm.io/gorm v1.25.12
)

replace github.com/payment-platform/pkg => ../../pkg
