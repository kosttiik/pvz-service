package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/kosttiik/pvz-service/internal/dto"
	"github.com/kosttiik/pvz-service/internal/models"
	"github.com/kosttiik/pvz-service/internal/repository"
	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/cache"
	"github.com/kosttiik/pvz-service/pkg/database"
	"github.com/kosttiik/pvz-service/pkg/logger"
	"github.com/kosttiik/pvz-service/pkg/redis"
	"go.uber.org/zap"
)

type DummyLoginRequest struct {
	Role string `json:"role"`
}

type DummyLoginResponse struct {
	Token string `json:"token"`
}

var dummyTokens = map[string]string{
	"moderator": "moderator-token",
	"employee":  "employee-token",
}

func DummyLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if !models.ValidRoles[req.Role] {
		utils.WriteError(w, "Invalid role", http.StatusBadRequest)
		return
	}

	dummyUserID := uuid.New().String()
	token, err := utils.GenerateJWT(dummyUserID, req.Role)
	if err != nil {
		utils.WriteError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	tokenCache := cache.NewTokenCache(redis.Client)
	if err := tokenCache.Set(r.Context(), dummyUserID, token); err != nil {
		utils.WriteError(w, "Failed to manage session", http.StatusInternalServerError)
		return
	}

	resp := DummyLoginResponse{Token: token}
	utils.WriteJSON(w, resp, http.StatusOK)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.Log
	ctx := r.Context()

	userRepo := repository.NewUserRepository(database.DB)

	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Failed to decode register request", zap.Error(err))
		utils.WriteError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		utils.WriteError(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Проверяем есть ли уже такой пользователь с эмейлом
	var count int
	err := database.DB.QueryRow(ctx,
		"SELECT COUNT(*) FROM users WHERE email = $1",
		req.Email).Scan(&count)
	if err != nil {
		utils.WriteError(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		utils.WriteError(w, "Email already registered", http.StatusBadRequest)
		return
	}

	if !models.ValidRoles[req.Role] {
		utils.WriteError(w, "Invalid role", http.StatusBadRequest)
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.WriteError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Создаем нового пользователя
	user := models.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Password: hashedPassword,
		Role:     req.Role,
	}

	if err := userRepo.Create(ctx, &user); err != nil {
		utils.WriteError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Info("User registered successfully",
		zap.String("userID", user.ID.String()),
		zap.String("email", user.Email),
		zap.String("role", user.Role))

	w.WriteHeader(http.StatusCreated)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.Log
	ctx := r.Context()
	userRepo := repository.NewUserRepository(database.DB)
	tokenCache := cache.NewTokenCache(redis.Client)

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warn("Failed to decode login request", zap.Error(err))
		utils.WriteError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		log.Info("Login failed - user not found",
			zap.String("email", req.Email),
			zap.Error(err))
		utils.WriteError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := utils.CheckPassword(req.Password, user.Password); err != nil {
		utils.WriteError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Инвалидим все токены юзера
	if err := tokenCache.Invalidate(ctx, user.ID.String()); err != nil {
		utils.WriteError(w, "Failed to manage session", http.StatusInternalServerError)
		return
	}

	token, err := utils.GenerateJWT(user.ID.String(), user.Role)
	if err != nil {
		utils.WriteError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Кэшируем токен юзера
	if err := tokenCache.Set(ctx, user.ID.String(), token); err != nil {
		utils.WriteError(w, "Failed to manage session", http.StatusInternalServerError)
		return
	}

	log.Info("User logged in successfully",
		zap.String("userID", user.ID.String()),
		zap.String("email", user.Email),
		zap.String("role", user.Role))

	utils.WriteJSON(w, map[string]string{"token": token}, http.StatusOK)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.Log
	ctx := r.Context()
	claims := utils.GetUserFromContext(ctx)
	if claims == nil {
		utils.WriteError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	tokenCache := cache.NewTokenCache(redis.Client)
	if err := tokenCache.Invalidate(ctx, claims.UserID); err != nil {
		utils.WriteError(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	log.Info("User logged out successfully",
		zap.String("userID", claims.UserID))

	w.WriteHeader(http.StatusNoContent)
}
