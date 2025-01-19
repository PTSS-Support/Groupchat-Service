package middleware

import (
	"Groupchat-Service/internal/config"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"strings"
)

type JWTMiddlewareConfig struct {
	rsaPublicKey *rsa.PublicKey
	cookieName   string
}

func NewJWTMiddleware(cfg *config.Config) (gin.HandlerFunc, error) {
	publicKey, err := parseRawPublicKey(cfg.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	middlewareConfig := &JWTMiddlewareConfig{
		rsaPublicKey: publicKey,
		cookieName:   cfg.AccessTokenCookieName,
	}
	return middlewareConfig.handleRequest, nil
}

func parseRawPublicKey(rawKey string) (*rsa.PublicKey, error) {
	// Decode the base64-encoded key
	derKey, err := base64.StdEncoding.DecodeString(rawKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 public key: %v", err)
	}

	// Parse the DER-encoded key
	pub, err := x509.ParsePKIXPublicKey(derKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	// Ensure it's an RSA public key
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPub, nil
}

func (config *JWTMiddlewareConfig) handleRequest(c *gin.Context) {
	if strings.HasPrefix(c.Request.URL.Path, "/q/health") || c.Request.URL.Path == "/metrics" {
		c.Next()
		return
	}

	tokenString, err := getTokenFromCookie(c, config.cookieName)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		c.Abort()
		return
	}

	token, err := config.parseToken(tokenString)
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

func (config *JWTMiddlewareConfig) parseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return config.rsaPublicKey, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil
}

func getTokenFromCookie(c *gin.Context, cookieName string) (string, error) {
	tokenString, err := c.Cookie(cookieName)
	if err != nil {
		return "", fmt.Errorf("authorization cookie required")
	}
	return tokenString, nil
}

func validateTokenClaims(token *jwt.Token) (string, string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", "", fmt.Errorf("user ID not found in token")
	}

	groupID, ok := claims["group_id"].(string)
	if !ok {
		return "", "", fmt.Errorf("group ID not found in token")
	}

	return userID, groupID, nil
}

func setContextValues(c *gin.Context, userID, groupID string) {
	c.Set("userID", userID)
	c.Set("groupID", groupID)
}
