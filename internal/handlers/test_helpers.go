package handlers

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kosttiik/pvz-service/internal/models"
	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/cache"
	"github.com/kosttiik/pvz-service/pkg/database"
	"github.com/kosttiik/pvz-service/pkg/redis"
)

func getTestToken(t *testing.T, role string, req *http.Request) *http.Request {
	userID := uuid.New().String()
	token, err := utils.GenerateJWT(userID, role)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Храним токен в редисе
	tokenCache := cache.NewTokenCache(redis.Client)
	ctx := context.Background()
	if err := tokenCache.Set(ctx, userID, token); err != nil {
		t.Fatalf("Failed to store token in cache: %v", err)
	}

	// Добавляем токен в заголовк авторизации
	req.Header.Set("Authorization", "Bearer "+token)

	// Добавляем контекст с данными пользователя
	claims := &models.Claims{
		UserID: userID,
		Role:   models.Role(role),
	}
	return req.WithContext(utils.SetUserContext(req.Context(), claims))
}

// Функция для создания тестового пвз в бд
func createTestPVZ(t *testing.T) string {
	pvzID := uuid.New()
	_, err := database.DB.Exec(context.Background(),
		"INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)",
		pvzID, time.Now(), "Москва")
	if err != nil {
		t.Fatalf("Failed to create test PVZ: %v", err)
	}
	return pvzID.String()
}
