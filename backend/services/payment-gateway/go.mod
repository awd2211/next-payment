module github.com/payment-platform/services/payment-gateway

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/payment-platform/pkg v0.0.0
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
)

replace github.com/payment-platform/pkg => ../../pkg
