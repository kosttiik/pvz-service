package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/kosttiik/pvz-service/internal/dto"
	"github.com/kosttiik/pvz-service/internal/models"
	"github.com/kosttiik/pvz-service/internal/repository"
	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/database"
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

	resp := DummyLoginResponse{Token: token}
	utils.WriteJSON(w, resp, http.StatusOK)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userRepo := repository.NewUserRepository(database.DB)

	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, "Invalid request", http.StatusBadRequest)
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

	w.WriteHeader(http.StatusCreated)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userRepo := repository.NewUserRepository(database.DB)

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		utils.WriteError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := utils.CheckPassword(req.Password, user.Password); err != nil {
		utils.WriteError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateJWT(user.ID.String(), user.Role)
	if err != nil {
		utils.WriteError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, map[string]string{"token": token}, http.StatusOK)
}
