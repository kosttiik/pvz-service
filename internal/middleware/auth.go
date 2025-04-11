package middleware

import (
	"net/http"
	"strings"

	"slices"

	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/cache"
	"github.com/kosttiik/pvz-service/pkg/redis"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	tokenCache := cache.NewTokenCache(redis.Client)

	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			utils.WriteError(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.WriteError(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ParseJWT(parts[1])
		if err != nil {
			utils.WriteError(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Проверям существует ли токен в редисе
		cachedToken, err := tokenCache.Get(r.Context(), claims.UserID)
		if err != nil || cachedToken != parts[1] {
			utils.WriteError(w, "Token has been revoked or expired", http.StatusUnauthorized)
			return
		}

		ctx := utils.SetUserContext(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func RoleMiddleware(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims := utils.GetUserFromContext(r.Context())
			if claims == nil {
				utils.WriteError(w, "Unauthorized - no claims in context", http.StatusUnauthorized)
				return
			}

			hasRole := slices.Contains(roles, string(claims.Role))

			if !hasRole {
				utils.WriteError(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}
