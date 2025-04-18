package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kosttiik/pvz-service/internal/dto"
	handlertest "github.com/kosttiik/pvz-service/internal/handlers/internal/test"
	"github.com/kosttiik/pvz-service/internal/models"
	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/database"
)

func init() {
	handlertest.Init()
}

func TestDummyLoginHandler(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{"Valid employee", "employee", http.StatusOK},
		{"Valid moderator", "moderator", http.StatusOK},
		{"Invalid role", "invalid", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{"role": tt.role}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(jsonBody))
			w := httptest.NewRecorder()

			DummyLoginHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response struct {
					Token string `json:"token"`
				}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if response.Token == "" {
					t.Error("token is empty")
				}
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	duplicateUser := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "somepassword",
		Role:     "employee",
	}

	tests := []struct {
		name       string
		input      dto.RegisterRequest
		setupUser  bool
		wantStatus int
	}{
		{
			name: "Valid registration",
			input: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "employee",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "Invalid role",
			input: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "invalid",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Empty email",
			input: dto.RegisterRequest{
				Email:    "",
				Password: "password123",
				Role:     "employee",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Empty password",
			input: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "",
				Role:     "employee",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "Duplicate email",
			input: dto.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "employee",
			},
			setupUser:  true,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := database.DB.Exec(context.Background(), "TRUNCATE users CASCADE"); err != nil {
				t.Fatalf("Failed to cleanup users table: %v", err)
			}

			if tt.setupUser {
				_, err := database.DB.Exec(context.Background(),
					"INSERT INTO users (id, email, password, role) VALUES ($1, $2, $3, $4)",
					duplicateUser.ID, duplicateUser.Email, duplicateUser.Password, duplicateUser.Role)
				if err != nil {
					t.Fatalf("Failed to setup test user: %v", err)
				}
			}

			jsonBody, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
			w := httptest.NewRecorder()

			RegisterHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("RegisterHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	password := "testpass123"
	hashedPassword, _ := utils.HashPassword(password)
	testUser := models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: hashedPassword,
		Role:     "employee",
	}

	ctx := context.Background()
	database.DB.Exec(ctx, "TRUNCATE users CASCADE")

	_, err := database.DB.Exec(ctx, `INSERT INTO users (id, email, password, role) 
		VALUES ($1, $2, $3, $4)`, testUser.ID, testUser.Email, testUser.Password, testUser.Role)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	tests := []struct {
		name       string
		input      dto.LoginRequest
		wantStatus int
	}{
		{
			name: "Valid login",
			input: dto.LoginRequest{
				Email:    testUser.Email,
				Password: password,
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Wrong password",
			input: dto.LoginRequest{
				Email:    testUser.Email,
				Password: "wrongpass",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "Non-existent user",
			input: dto.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: password,
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.input)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
			w := httptest.NewRecorder()

			LoginHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("LoginHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response struct {
					Token string `json:"token"`
				}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if response.Token == "" {
					t.Error("Expected non-empty token")
				}
			}
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	tests := []struct {
		name       string
		setupAuth  bool
		wantStatus int
	}{
		{
			name:       "Valid logout",
			setupAuth:  true,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "Unauthorized",
			setupAuth:  false,
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/logout", nil)
			if tt.setupAuth {
				req = getTestToken(t, "employee", req)
			}
			w := httptest.NewRecorder()

			LogoutHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("LogoutHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}
