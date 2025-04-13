package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kosttiik/pvz-service/internal/models"
)

func TestProductRepository(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewProductRepository(pool)
	ctx := context.Background()

	// Создаем тестовый пвз и приемку
	pvzID := uuid.New()
	_, err := pool.Exec(ctx,
		"INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)",
		pvzID, time.Now(), "Москва")
	if err != nil {
		t.Fatalf("Failed to create test PVZ: %v", err)
	}

	receptionID := uuid.New()
	_, err = pool.Exec(ctx,
		"INSERT INTO reception (id, date_time, pvz_id, status) VALUES ($1, $2, $3, $4)",
		receptionID, time.Now(), pvzID, models.StatusInProgress)
	if err != nil {
		t.Fatalf("Failed to create test reception: %v", err)
	}

	productID := uuid.New()

	t.Run("Create", func(t *testing.T) {
		product := &models.Product{
			ID:          productID,
			DateTime:    time.Now(),
			Type:        "электроника",
			ReceptionID: receptionID.String(),
		}

		if err := repo.Create(ctx, product); err != nil {
			t.Fatalf("Failed to create product: %v", err)
		}
	})

	t.Run("DeleteLastFromReception", func(t *testing.T) {
		if err := repo.DeleteLastFromReception(ctx, receptionID.String()); err != nil {
			t.Fatalf("Failed to delete last product: %v", err)
		}
	})
}
