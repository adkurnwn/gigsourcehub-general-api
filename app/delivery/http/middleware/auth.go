package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/adkurnwn/gigsourcehub-general-api/domain/model/response"
	"github.com/adkurnwn/gigsourcehub-general-api/helpers"
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

		// Enrichment for Activity Logging
		ip := c.GetHeader("CF-Connecting-IP")
		if ip == "" {
			ip = c.GetHeader("X-Forwarded-For")
			if ip != "" {
				// X-Forwarded-For can be a comma-separated list
				ip = strings.Split(ip, ",")[0]
			} else {
				ip = c.ClientIP()
			}
		}

		// Inject into Request Context for Usecases
		ctx := c.Request.Context()
		ctx = helpers.SetActorID(ctx, claims.UserID)
		ctx = helpers.SetIPAddress(ctx, ip)
		ctx = helpers.SetUserAgent(ctx, c.Request.UserAgent())
		ctx = helpers.SetEndpoint(ctx, c.Request.Method+" "+c.Request.URL.Path)
		c.Request = c.Request.WithContext(ctx)

		c.Set("token_data", *claims)

		// Real-time account status check for active sessions
		status, errStatus := m.repo.GetUserAccountStatus(c.Request.Context(), claims.UserID)
		if errStatus != nil {
			response := response.Error(http.StatusInternalServerError, "Internal Server Error: Unable to verify account status.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}

		if status == "Blocked" {
			response := response.Error(http.StatusForbidden, "Your account has been blocked.")
			c.AbortWithStatusJSON(http.StatusForbidden, response)
			return
		}

		// Real-time verification check
		verifiedAt, errVerified := m.repo.GetUserVerifiedAt(c.Request.Context(), claims.UserID)
		if errVerified != nil {
			response := response.Error(http.StatusInternalServerError, "Internal Server Error: Unable to verify account verification status.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}

		if verifiedAt == nil {
			response := response.Error(http.StatusForbidden, "Please verify your email address before continuing.")
			c.AbortWithStatusJSON(http.StatusForbidden, response)
			return
		}

		// Real-time password reset check
		mustReset, errReset := m.repo.GetUserMustResetPassword(c.Request.Context(), claims.UserID)
		if errReset != nil {
			response := response.Error(http.StatusInternalServerError, "Internal Server Error: Unable to verify password status.")
			c.AbortWithStatusJSON(http.StatusInternalServerError, response)
			return
		}

		if mustReset {
			path := c.Request.URL.Path
			method := c.Request.Method
			// Allow only GET for profile/me and PUT for password reset
			isAllowed := (method == "GET" && (strings.HasSuffix(path, "/profile") || strings.HasSuffix(path, "/me"))) ||
				(method == "PUT" && strings.HasSuffix(path, "/profile/password"))

			if !isAllowed {
				response := response.Error(http.StatusForbidden, "Password reset required before continuing.")
				c.AbortWithStatusJSON(http.StatusForbidden, response)
				return
			}
		}

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
