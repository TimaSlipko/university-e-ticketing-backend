package middleware

import (
	"strings"

	"eticketing/internal/models"
	"eticketing/internal/utils"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeaderKey  = "authorization"
	AuthorizationTypeBearer = "bearer"
	AuthorizationPayloadKey = "authorization_payload"
)

func AuthMiddleware(jwtManager *utils.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeader := c.GetHeader(AuthorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			utils.UnauthorizedResponse(c, "Authorization header not provided")
			c.Abort()
			return
		}

		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			utils.UnauthorizedResponse(c, "Invalid authorization header format")
			c.Abort()
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != AuthorizationTypeBearer {
			utils.UnauthorizedResponse(c, "Unsupported authorization type")
			c.Abort()
			return
		}

		accessToken := fields[1]
		payload, err := jwtManager.ValidateToken(accessToken)
		if err != nil {
			utils.UnauthorizedResponse(c, "Invalid access token")
			c.Abort()
			return
		}

		if payload.Type != "access" {
			utils.UnauthorizedResponse(c, "Invalid token type")
			c.Abort()
			return
		}

		c.Set(AuthorizationPayloadKey, payload)
		c.Next()
	}
}

func RequireRole(roles ...models.UserType) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload, exists := c.Get(AuthorizationPayloadKey)
		if !exists {
			utils.UnauthorizedResponse(c, "Authorization payload not found")
			c.Abort()
			return
		}

		claims, ok := payload.(*utils.JWTClaims)
		if !ok {
			utils.UnauthorizedResponse(c, "Invalid authorization payload")
			c.Abort()
			return
		}

		// Check if user has required role
		for _, role := range roles {
			if claims.UserType == role {
				c.Next()
				return
			}
		}

		utils.ForbiddenResponse(c, "Insufficient permissions")
		c.Abort()
	}
}

func GetCurrentUser(c *gin.Context) (*utils.JWTClaims, error) {
	payload, exists := c.Get(AuthorizationPayloadKey)
	if !exists {
		return nil, utils.NewError("authorization payload not found")
	}

	claims, ok := payload.(*utils.JWTClaims)
	if !ok {
		return nil, utils.NewError("invalid authorization payload")
	}

	return claims, nil
}
