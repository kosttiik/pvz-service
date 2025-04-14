package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kosttiik/pvz-service/internal/models"
	"github.com/kosttiik/pvz-service/internal/testutils"
)

func TestUserRepository(t *testing.T) {
	pool := testutils.SetupTestDB(t)
	defer pool.Close()

	repo := NewUserRepository(pool)
	ctx := context.Background()

	// Очищаем таблицы перед каждым тестом
	_, err := pool.Exec(ctx, "TRUNCATE users CASCADE")
	if err != nil {
		t.Fatalf("Failed to cleanup tables: %v", err)
	}

	t.Run("Create and GetByEmail", func(t *testing.T) {
		// Очищаем перед каждым тестом
		_, err := pool.Exec(ctx, "TRUNCATE users CASCADE")
		if err != nil {
			t.Fatalf("Failed to cleanup users table: %v", err)
		}

		user := &models.User{
			ID:       uuid.New(),
			Email:    "test@example.com",
			Password: "hashedpass",
			Role:     "employee",
		}

		if err := repo.Create(ctx, user); err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}

		got, err := repo.GetByEmail(ctx, user.Email)
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}

		if got.ID != user.ID {
			t.Errorf("Got user ID = %v, want %v", got.ID, user.ID)
		}
	})

	t.Run("GetByEmail_NonExistent", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		if err == nil {
			t.Error("Expected error for non-existent user")
		}
	})

	t.Run("Create_DuplicateEmail", func(t *testing.T) {
		user1 := &models.User{
			ID:       uuid.New(),
			Email:    "duplicate@example.com",
			Password: "pass1",
			Role:     "employee",
		}
		user2 := &models.User{
			ID:       uuid.New(),
			Email:    "duplicate@example.com",
			Password: "pass2",
			Role:     "employee",
		}

		if err := repo.Create(ctx, user1); err != nil {
			t.Fatalf("Failed to create first user: %v", err)
		}

		if err := repo.Create(ctx, user2); err == nil {
			t.Error("Expected error when creating user with duplicate email")
		}
	})
}
