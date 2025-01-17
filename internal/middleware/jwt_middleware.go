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
	certPEM := cfg.PublicKey
	if certPEM == "" {
		return nil, fmt.Errorf("public key certificate is required")
	}

	publicKey, err := parsePublicKey(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	middlewareConfig := &JWTMiddlewareConfig{
		rsaPublicKey: publicKey,
		cookieName:   cfg.AccessTokenCookieName,
	}
	return middlewareConfig.handleRequest, nil
}

func parsePublicKey(certPEM string) (*rsa.PublicKey, error) {
	certData, err := base64.StdEncoding.DecodeString(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to decode certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return publicKey, nil
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

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		return "", "", fmt.Errorf("user ID claim missing")
	}

	groupID, ok := claims["groupId"].(string)
	if !ok || groupID == "" {
		return "", "", fmt.Errorf("group ID claim missing")
	}

	return userID, groupID, nil
}

func setContextValues(c *gin.Context, userID, groupID string) {
	c.Set("userID", userID)
	c.Set("groupID", groupID)
}
