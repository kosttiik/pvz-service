package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/kosttiik/pvz-service/internal/dto"
	"github.com/kosttiik/pvz-service/internal/models"
	"github.com/kosttiik/pvz-service/internal/repository"
	"github.com/kosttiik/pvz-service/internal/utils"
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

var userRepo = &repository.UserRepository{}

func DummyLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req DummyLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	token, ok := dummyTokens[req.Role]
	if !ok {
		utils.WriteError(w, "Invalid role", http.StatusBadRequest)
		return
	}

	resp := DummyLoginResponse{Token: token}
	json.NewEncoder(w).Encode(&resp)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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
	json.NewEncoder(w).Encode(user)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	token, err := utils.GenerateJWT(user.Email, user.Role)
	if err != nil {
		utils.WriteError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}
