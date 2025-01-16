package middleware

import (
	"Groupchat-Service/internal/config"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"math/big"
	"net/http"
	"strings"
)

// JWKS represents the JSON Web Key Set structure
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string   `json:"kid"`
	Kty string   `json:"kty"`
	Alg string   `json:"alg"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
	X5c []string `json:"x5c"`
}

// JWTMiddlewareConfig holds the configuration for the JWT middleware
type JWTMiddlewareConfig struct {
	rsaPublicKey *rsa.PublicKey
	cookieName   string
}

// NewJWTMiddleware creates a new JWT middleware instance with the provided configuration
func NewJWTMiddleware(cfg *config.Config) (gin.HandlerFunc, error) {
	jwksJSON := cfg.JWKSJSON
	if jwksJSON == "" {
		return nil, fmt.Errorf("JWKS JSON is required")
	}

	// Parse and validate the JWKS JSON
	publicKey, err := parseJWKS(jwksJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %v", err)
	}

	middlewareConfig := &JWTMiddlewareConfig{
		rsaPublicKey: publicKey,
		cookieName:   cfg.AccessTokenCookieName,
	}

	return middlewareConfig.handleRequest, nil
}

func (config *JWTMiddlewareConfig) handleRequest(c *gin.Context) {
	// Skip JWT check for health and metrics endpoints
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

func parseJWKS(jwksJSON string) (*rsa.PublicKey, error) {
	var jwks JWKS
	if err := json.Unmarshal([]byte(jwksJSON), &jwks); err != nil {
		return nil, fmt.Errorf("failed to parse JWKS JSON: %v", err)
	}

	// Find the signing key (use: "sig")
	var signingKey *JWK
	for _, key := range jwks.Keys {
		if key.Use == "sig" && key.Alg == "RS256" {
			signingKey = &key
			break
		}
	}

	if signingKey == nil {
		return nil, fmt.Errorf("no suitable signing key found in JWKS")
	}

	// Decode the modulus and exponent
	nBytes, err := base64.RawURLEncoding.DecodeString(signingKey.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %v", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(signingKey.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %v", err)
	}

	// Convert bytes to big integers
	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
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
