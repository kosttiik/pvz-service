package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/cache"
	"github.com/kosttiik/pvz-service/pkg/logger"
	"github.com/kosttiik/pvz-service/pkg/redis"
	"go.uber.org/zap"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	tokenCache := cache.NewTokenCache(redis.Client)

	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.Log
		log.Debug("Processing request",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method),
			zap.String("remote_addr", r.RemoteAddr))

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Warn("No authorization token provided")
			utils.WriteError(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Warn("Invalid authorization header")
			utils.WriteError(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ParseJWT(parts[1])
		if err != nil {
			log.Warn("Invalid token",
				zap.Error(err),
				zap.String("path", r.URL.Path))
			utils.WriteError(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Проверям существует ли токен в редисе
		cachedToken, err := tokenCache.Get(r.Context(), claims.UserID)
		if err != nil || cachedToken != parts[1] {
			log.Warn("Token not found in cache or invalid",
				zap.String("userID", claims.UserID),
				zap.Error(err))
			utils.WriteError(w, "Token has been revoked or expired", http.StatusUnauthorized)
			return
		}

		log.Debug("Request authorized",
			zap.String("userID", claims.UserID),
			zap.String("role", string(claims.Role)),
			zap.String("path", r.URL.Path))

		log.Debug("User authenticated successfully",
			zap.String("userID", claims.UserID),
			zap.String("role", string(claims.Role)))

		ctx := utils.SetUserContext(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func RoleMiddleware(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			log := logger.Log
			claims := utils.GetUserFromContext(r.Context())
			if claims == nil {
				log.Warn("No claims found in context",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method))
				utils.WriteError(w, "Unauthorized - no claims in context", http.StatusUnauthorized)
				return
			}

			hasRole := slices.Contains(roles, string(claims.Role))

			if !hasRole {
				log.Warn("Access denied - invalid role",
					zap.String("userID", claims.UserID),
					zap.String("userRole", string(claims.Role)),
					zap.Strings("requiredRoles", roles),
					zap.String("path", r.URL.Path))
				utils.WriteError(w, "Forbidden", http.StatusForbidden)
				return
			}

			log.Debug("Role check passed",
				zap.String("userID", claims.UserID),
				zap.String("role", string(claims.Role)),
				zap.String("path", r.URL.Path))
			next.ServeHTTP(w, r)
		}
	}
}
