package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kosttiik/pvz-service/internal/models"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgresql://postgres:postgres@localhost:5432/pvz_db?sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping db: %v", err)
	}

	return pool
}

func TestReceptionRepository(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewReceptionRepository(pool)
	ctx := context.Background()

	// Очищаем таблицы перед всем
	_, err := pool.Exec(ctx, "TRUNCATE pvz, reception CASCADE")
	if err != nil {
		t.Fatalf("Failed to cleanup tables: %v", err)
	}

	// Создаем тестовое пвз
	pvzID := uuid.New()
	_, err = pool.Exec(ctx,
		"INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)",
		pvzID, time.Now(), "Москва")
	if err != nil {
		t.Fatalf("Failed to create test PVZ: %v", err)
	}

	receptionID := uuid.New()

	t.Run("HasOpenReception", func(t *testing.T) {
		reception := &models.Reception{
			ID:       receptionID,
			DateTime: time.Now(),
			PvzID:    pvzID.String(),
			Status:   models.StatusInProgress,
		}

		if err := repo.Create(ctx, reception); err != nil {
			t.Fatalf("Failed to create reception: %v", err)
		}

		hasOpen, err := repo.HasOpenReception(ctx, pvzID.String())
		if err != nil {
			t.Fatalf("Failed to check open reception: %v", err)
		}

		if !hasOpen {
			t.Error("Expected to have open reception")
		}
	})

	t.Run("GetLastOpenReception", func(t *testing.T) {
		reception, err := repo.GetLastOpenReception(ctx, pvzID.String())
		if err != nil {
			t.Fatalf("Failed to get last open reception: %v", err)
		}

		if reception.ID != receptionID {
			t.Error("Got wrong reception ID")
		}
	})

	t.Run("CloseLastReception", func(t *testing.T) {
		// Создаем тестовую приёмку
		newReceptionID := uuid.New()
		_, err := pool.Exec(ctx,
			"INSERT INTO reception (id, date_time, pvz_id, status) VALUES ($1, $2, $3, $4)",
			newReceptionID, time.Now(), pvzID, models.StatusInProgress)
		if err != nil {
			t.Fatalf("Failed to create test reception: %v", err)
		}

		reception, err := repo.CloseLastReception(ctx, pvzID.String())
		if err != nil {
			t.Fatalf("Failed to close reception: %v", err)
		}

		if reception.Status != models.StatusClosed {
			t.Errorf("Expected status %s, got %s", models.StatusClosed, reception.Status)
		}
	})

	t.Run("GetLastOpenReception_NotFound", func(t *testing.T) {
		nonExistentPVZID := uuid.New().String()
		_, err := repo.GetLastOpenReception(ctx, nonExistentPVZID)
		if err == nil {
			t.Error("Expected error for non-existent PVZ")
		}
	})

	t.Run("CloseLastReception_NoOpenReception", func(t *testing.T) {
		nonExistentPVZID := uuid.New().String()
		_, err := repo.CloseLastReception(ctx, nonExistentPVZID)
		if err == nil {
			t.Error("Expected error when no open reception exists")
		}
	})

	t.Run("Create_DuplicateID", func(t *testing.T) {
		reception := &models.Reception{
			ID:       receptionID, // Используем существующий ID
			DateTime: time.Now(),
			PvzID:    pvzID.String(),
			Status:   models.StatusInProgress,
		}
		err := repo.Create(ctx, reception)
		if err == nil {
			t.Error("Expected error when creating reception with duplicate ID")
		}
	})

	t.Run("HasOpenReception_NonExistentPVZ", func(t *testing.T) {
		hasOpen, err := repo.HasOpenReception(ctx, uuid.New().String())
		if err != nil {
			t.Fatalf("Failed to check open reception: %v", err)
		}
		if hasOpen {
			t.Error("Should not have open reception for non-existent PVZ")
		}
	})

	t.Run("CloseReception_StatusChange", func(t *testing.T) {
		reception, err := repo.CloseLastReception(ctx, pvzID.String())
		if err != nil {
			t.Fatalf("Failed to close reception: %v", err)
		}
		if reception.Status != models.StatusClosed {
			t.Errorf("Got status = %v, want %v", reception.Status, models.StatusClosed)
		}
	})
}
