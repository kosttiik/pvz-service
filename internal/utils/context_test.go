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
			ctx = SetUserContext(ctx, tt.claims)
			got := GetUserFromContext(ctx)

			if tt.claims == nil {
				if got != nil {
					t.Errorf("GetUserFromContext() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("GetUserFromContext() = nil, want non-nil")
			}

			if got.UserID != tt.claims.UserID {
				t.Errorf("UserID = %v, want %v", got.UserID, tt.claims.UserID)
			}

			if got.Role != tt.claims.Role {
				t.Errorf("Role = %v, want %v", got.Role, tt.claims.Role)
			}
		})
	}

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		claims := &models.Claims{UserID: "test", Role: models.Employee}
		ctx = SetUserContext(ctx, claims)
		cancel()

		got := GetUserFromContext(ctx)
		if got == nil {
			t.Error("Claims should persist after context cancellation")
		}
	})

	t.Run("Nested contexts", func(t *testing.T) {
		ctx := context.Background()
		claims1 := &models.Claims{UserID: "user1", Role: models.Employee}
		claims2 := &models.Claims{UserID: "user2", Role: models.Moderator}

		ctx = SetUserContext(ctx, claims1)
		innerCtx := SetUserContext(ctx, claims2)

		got1 := GetUserFromContext(ctx)
		got2 := GetUserFromContext(innerCtx)

		if got1.UserID != claims1.UserID {
			t.Errorf("Outer context: got %v, want %v", got1.UserID, claims1.UserID)
		}
		if got2.UserID != claims2.UserID {
			t.Errorf("Inner context: got %v, want %v", got2.UserID, claims2.UserID)
		}
	})
}
