package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

// UserClaims represents the expected JWT claims
type UserClaims struct {
	UserID string   `json:"sub"`
	Roles  []string `json:"realm_access.roles"`
	jwt.StandardClaims
}

// AuthMiddleware validates JWT tokens and ensures proper authorization
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, http.StatusUnauthorized, "No authorization header")
			return
		}

		// Remove 'Bearer ' prefix
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			respondWithError(w, http.StatusUnauthorized, "Invalid token format")
			return
		}

		// Parse and validate the token
		claims, err := validateToken(tokenString)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, fmt.Sprintf("Invalid token: %v", err))
			return
		}

		// Check required roles
		if !hasRequiredRoles(claims.Roles, []string{"core_user", "family_member"}) {
			respondWithError(w, http.StatusForbidden, "Insufficient permissions")
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken parses and validates the JWT token
func validateToken(tokenString string) (*UserClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Replace this with your actual JWT secret or public key
		return []byte("your-jwt-secret"), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// hasRequiredRoles checks if the user has any of the required roles
func hasRequiredRoles(userRoles []string, requiredRoles []string) bool {
	for _, required := range requiredRoles {
		for _, role := range userRoles {
			if role == required {
				return true
			}
		}
	}
	return false
}

// respondWithError sends an error response in JSON format
func respondWithError(w http.ResponseWriter, code int, message string) {
	response := map[string]string{"error": message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}
