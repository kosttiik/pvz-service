package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/cache"
	"github.com/kosttiik/pvz-service/pkg/redis"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test_secret")
	os.Setenv("REDIS_HOST", "localhost")
	if err := redis.Connect(); err != nil {
		panic(err)
	}
	code := m.Run()
	redis.Close()
	os.Exit(code)
}

func TestAuthMiddleware(t *testing.T) {
	tokenCache := cache.NewTokenCache(redis.Client)
	ctx := context.Background()

	tests := []struct {
		name       string
		token      string
		userID     string
		wantStatus int
	}{
		{"No token", "", "", http.StatusUnauthorized},
		{"Invalid token", "Bearer invalid", "", http.StatusUnauthorized},
		{"Valid token", "", "test-user", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Valid token" {
				token, err := utils.GenerateJWT(tt.userID, "employee")
				if err != nil {
					t.Fatalf("Failed to generate token: %v", err)
				}
				tt.token = "Bearer " + token

				// Сохраняем токен в редисе для теста
				if err := tokenCache.Set(ctx, tt.userID, token); err != nil {
					t.Fatalf("Failed to store token: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}
			w := httptest.NewRecorder()

			handler := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}
		})

		// Очищаем кэш после каждого теста
		if tt.userID != "" {
			if err := tokenCache.Invalidate(ctx, tt.userID); err != nil {
				t.Logf("Failed to cleanup token: %v", err)
			}
		}
	}
}
