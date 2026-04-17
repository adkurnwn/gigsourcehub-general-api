package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (m *appMiddleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		hAuth := c.GetHeader("Authorization")
		if hAuth == "" {
			response := response.Error(http.StatusUnauthorized, "Unauthorized: Header authorization is required")
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		splitToken := strings.Split(hAuth, "Bearer ")
		if len(splitToken) != 2 {
			response := response.Error(http.StatusUnauthorized, "Unauthorized: Token is invalid")
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		// get token without 'Bearer '
		tokenString := splitToken[1]

		// validating token
		token, err := jwt.ParseWithClaims(tokenString, &domain.JWTClaimUser{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(m.secret), nil
		})

		// check validity token
		if token == nil || !token.Valid {
			if errors.Is(err, jwt.ErrTokenMalformed) {
				c.AbortWithStatusJSON(
					http.StatusUnauthorized,
					response.Error(http.StatusUnauthorized, "Unauthorized: Token is invalid"),
				)
				return
			}

			if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
				c.AbortWithStatusJSON(
					http.StatusUnauthorized,
					response.Error(http.StatusUnauthorized, "Unauthorized: Token signature invalid"),
				)
				return
			}

			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(
					http.StatusUnauthorized,
					response.Error(http.StatusUnauthorized, "Unauthorized: Token expired"),
				)
				return
			}

			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				response.Error(http.StatusUnauthorized, err.Error()),
			)
			return
		}

		claims, tokenOK := token.Claims.(*domain.JWTClaimUser)
		if !tokenOK {
			response := response.Error(http.StatusUnauthorized, "Unauthorized: Token data not valid")
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		c.Set("token_data", *claims)
		c.Next()
	}
}

func (m *appMiddleware) AuthRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Ensure token_data is set by Auth()
		claimsVal, exists := c.Get("token_data")
		if !exists {
			response := response.Error(http.StatusUnauthorized, "Unauthorized: Token data missing. Ensure Auth() is called first.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		claims, ok := claimsVal.(domain.JWTClaimUser)
		if !ok {
			response := response.Error(http.StatusUnauthorized, "Unauthorized: Invalid token data structure.")
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		// Retrieve the user's string role from the database cleanly
		roleName, err := m.repo.GetRoleNameByUserID(c.Request.Context(), claims.UserID)
		if err != nil {
			response := response.Error(http.StatusInternalServerError, "Internal Server Error: Unable to verify user role.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}

		// Check if the exact role string exists in our allowed array
		isAllowed := false
		for _, allowed := range allowedRoles {
			if roleName == allowed {
				isAllowed = true
				break
			}
		}

		if !isAllowed {
			response := response.Error(http.StatusForbidden, "Forbidden: You do not have the necessary permissions.")
			c.AbortWithStatusJSON(http.StatusForbidden, response)
			return
		}

		c.Next()
	}
}

func (m *appMiddleware) AuthAdmin() gin.HandlerFunc {
	return m.AuthRole("Admin")
}

func (m *appMiddleware) AuthSuperadmin() gin.HandlerFunc {
	return m.AuthRole("Superadmin")
}

func (m *appMiddleware) AuthEmployee() gin.HandlerFunc {
	return m.AuthRole("Employee")
}

func (m *appMiddleware) AuthCandidate() gin.HandlerFunc {
	return m.AuthRole("Candidate")
}

func (m *appMiddleware) AuthInternal() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Internal-Token")
		if token == "" || token != m.internalToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Error(http.StatusUnauthorized, "Unauthorized: Internal access only"))
			return
		}
		c.Next()
	}
}
