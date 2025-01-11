package middleware

import (
	"fmt"
	"github.com/MicahParks/keyfunc/v2"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strings"
	"time"
)

func JWTMiddleware(jwksURL string) gin.HandlerFunc {
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{
		RefreshInterval: time.Hour,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create JWKS from URL: %v", err))
	}

	return func(c *gin.Context) {
		// Bypass JWT middleware for health check endpoints
		if strings.HasPrefix(c.Request.URL.Path, "/q/health") {
			c.Next()
			return
		}

		cookieName, err := getCookieName()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		tokenString, err := getTokenFromCookie(c, cookieName)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		token, err := parseToken(tokenString, jwks)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		userID, groupID, err := validateTokenClaims(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		setContextValues(c, userID, groupID)
		c.Next()
	}
}

func getCookieName() (string, error) {
	cookieName := os.Getenv("ACCESS_TOKEN_COOKIE_NAME")
	if cookieName == "" {
		return "", fmt.Errorf("ACCESS_TOKEN_COOKIE_NAME not set")
	}
	return cookieName, nil
}

func getTokenFromCookie(c *gin.Context, cookieName string) (string, error) {
	tokenString, err := c.Cookie(cookieName)
	if err != nil {
		return "", fmt.Errorf("authorization cookie required")
	}
	return tokenString, nil
}

func parseToken(tokenString string, jwks *keyfunc.JWKS) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, jwks.Keyfunc)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil
}

func validateTokenClaims(token *jwt.Token) (string, string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["userID"].(string)
	if !ok {
		return "", "", fmt.Errorf("user ID not found in token")
	}

	groupID, ok := claims["groupID"].(string)
	if !ok {
		return "", "", fmt.Errorf("group ID not found in token")
	}

	return userID, groupID, nil
}

func setContextValues(c *gin.Context, userID, groupID string) {
	c.Set("userID", userID)
	c.Set("groupID", groupID)
}
