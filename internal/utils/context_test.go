package utils

import (
	"context"
	"testing"

	"github.com/kosttiik/pvz-service/internal/models"
)

func TestUserContext(t *testing.T) {
	tests := []struct {
		name   string
		claims *models.Claims
	}{
		{
			name: "Valid claims",
			claims: &models.Claims{
				UserID: "test-user-id",
				Role:   models.Employee,
			},
		},
		{
			name:   "Nil claims",
			claims: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			newCtx := SetUserContext(ctx, tt.claims)
			if newCtx == nil {
				t.Error("SetUserContext returned nil context")
				return
			}

			gotClaims := GetUserFromContext(newCtx)
			if tt.claims == nil {
				if gotClaims != nil {
					t.Error("GetUserFromContext returned non-nil claims for nil input")
				}
				return
			}

			if gotClaims == nil {
				t.Error("GetUserFromContext returned nil claims for non-nil input")
				return
			}

			if gotClaims.UserID != tt.claims.UserID {
				t.Errorf("GetUserFromContext UserID = %v, want %v", gotClaims.UserID, tt.claims.UserID)
			}

			if gotClaims.Role != tt.claims.Role {
				t.Errorf("GetUserFromContext Role = %v, want %v", gotClaims.Role, tt.claims.Role)
			}
		})
	}

	t.Run("Get from empty context", func(t *testing.T) {
		ctx := context.Background()
		if claims := GetUserFromContext(ctx); claims != nil {
			t.Error("GetUserFromContext() should return nil for empty context")
		}
	})
}
