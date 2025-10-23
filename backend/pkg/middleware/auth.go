package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/payment-platform/pkg/auth"
)

type ContextKey string

const (
	ClaimsKey ContextKey = "claims"
)

var (
	ErrMissingAuthHeader = errors.New("missing authorization header")
	ErrInvalidAuthFormat = errors.New("invalid authorization format")
)

// AuthMiddleware creates JWT authentication middleware
func AuthMiddleware(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": ErrMissingAuthHeader.Error()})
			c.Abort()
			return
		}

		// Bearer <token>
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, gin.H{"error": ErrInvalidAuthFormat.Error()})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Store claims in context
		ctx := context.WithValue(c.Request.Context(), ClaimsKey, claims)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequirePermission checks if user has required permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Request.Context().Value(ClaimsKey).(*auth.Claims)
		if !ok {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if !claims.HasPermission(permission) {
			c.JSON(403, gin.H{"error": "forbidden: insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole checks if user has required role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Request.Context().Value(ClaimsKey).(*auth.Claims)
		if !ok {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if !claims.HasRole(role) {
			c.JSON(403, gin.H{"error": "forbidden: insufficient role"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdminType ensures only admin users can access
func RequireAdminType() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Request.Context().Value(ClaimsKey).(*auth.Claims)
		if !ok {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if claims.UserType != "admin" {
			c.JSON(403, gin.H{"error": "forbidden: admin only"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireMerchantType ensures only merchant users can access
func RequireMerchantType() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Request.Context().Value(ClaimsKey).(*auth.Claims)
		if !ok {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		if claims.UserType != "merchant" {
			c.JSON(403, gin.H{"error": "forbidden: merchant only"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetClaims retrieves claims from gin context
func GetClaims(c *gin.Context) (*auth.Claims, error) {
	claims, ok := c.Request.Context().Value(ClaimsKey).(*auth.Claims)
	if !ok {
		return nil, errors.New("claims not found in context")
	}
	return claims, nil
}
