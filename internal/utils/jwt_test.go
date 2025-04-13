package utils

import (
	"os"
	"testing"
	"time"

	"github.com/kosttiik/pvz-service/internal/models"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test_secret")
	code := m.Run()
	os.Unsetenv("JWT_SECRET")
	os.Exit(code)
}

func TestGenerateAndParseJWT(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		role    string
		wantErr bool
	}{
		{
			name:    "Valid moderator token",
			userID:  "123e4567-e89b-12d3-a456-426614174000",
			role:    string(models.Moderator),
			wantErr: false,
		},
		{
			name:    "Valid employee token",
			userID:  "123e4567-e89b-12d3-a456-426614174001",
			role:    string(models.Employee),
			wantErr: false,
		},
		{
			name:    "Empty userID",
			userID:  "",
			role:    string(models.Employee),
			wantErr: false,
		},
		{
			name:    "Invalid role",
			userID:  "123e4567-e89b-12d3-a456-426614174002",
			role:    "invalid_role",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateJWT(tt.userID, tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateJWT() returned empty token")
				return
			}

			claims, err := ParseJWT(token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if claims.UserID != tt.userID {
					t.Errorf("ParseJWT() got UserID = %v, want %v", claims.UserID, tt.userID)
				}
				if string(claims.Role) != tt.role {
					t.Errorf("ParseJWT() got Role = %v, want %v", claims.Role, tt.role)
				}
				if claims.ExpiresAt < time.Now().Unix() {
					t.Error("ParseJWT() token has already expired")
				}
			}
		})
	}
}

func TestParseJWT_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Invalid format",
			token:   "hello.world",
			wantErr: true,
		},
		{
			name:    "Cringe token",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.cringe.signature",
			wantErr: true,
		},
		{
			name:    "Wrong signature",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseJWT(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJWT() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
