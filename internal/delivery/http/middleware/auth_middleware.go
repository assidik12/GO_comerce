package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/assidik12/go-restfull-api/internal/pkg/jwt"
	"github.com/julienschmidt/httprouter"
)

type AuthMiddleware struct {
	Handler http.Handler
}

type contextKey string

const (
	UserIDKey contextKey = "user_id"
	RoleKey   contextKey = "role"
)

func NewAuthMiddleware(handler http.Handler) *AuthMiddleware {
	return &AuthMiddleware{Handler: handler}
}

func (middleware AuthMiddleware) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	middleware.Handler.ServeHTTP(writer, request)
}

func (middleware AuthMiddleware) Middleware(role string, next httprouter.Handle, key string) httprouter.Handle {

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Authorization header format is invalid", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwt.NewJWTService(key).ValidateToken(tokenString)

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if !isRoleAllowed(claims.Role, role) {
			http.Error(w, "Forbidden: You don't have permission to access this resource", http.StatusForbidden)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)

		next(w, r.WithContext(ctx), ps)
	}

}

// Helper function to check if user role is allowed
func isRoleAllowed(userRole string, allowedRoles string) bool {

	return strings.EqualFold(userRole, allowedRoles)
}
