package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kosttiik/pvz-service/internal/models"
)

func TestPVZRepository(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewPVZRepository(pool)
	ctx := context.Background()

	// Очищаем таблицы перед выполнением тестов
	_, err := pool.Exec(ctx, "TRUNCATE pvz, reception, product CASCADE")
	if err != nil {
		t.Fatalf("Failed to cleanup tables: %v", err)
	}

	baseTime := time.Now().UTC()
	pvzID := uuid.New()
	_, err = pool.Exec(ctx,
		"INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)",
		pvzID, baseTime, "Москва")
	if err != nil {
		t.Fatalf("Failed to create test PVZ: %v", err)
	}

	receptionID := uuid.New()
	_, err = pool.Exec(ctx,
		"INSERT INTO reception (id, date_time, pvz_id, status) VALUES ($1, $2, $3, $4)",
		receptionID, baseTime, pvzID, models.StatusInProgress)
	if err != nil {
		t.Fatalf("Failed to create test reception: %v", err)
	}

	t.Run("GetPVZ", func(t *testing.T) {
		filter := GetPVZFilter{
			StartDate: nil,
			EndDate:   nil,
			Page:      1,
			Limit:     10,
		}

		pvzList, err := repo.GetPVZ(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to get PVZ list: %v", err)
		}

		if len(pvzList) == 0 {
			t.Error("Expected non-empty PVZ list")
		}
	})

	t.Run("GetPVZWithDateFilter", func(t *testing.T) {
		startDate := baseTime.Add(-1 * time.Hour).UTC()
		endDate := baseTime.Add(1 * time.Hour).UTC()
		filter := GetPVZFilter{
			StartDate: &startDate,
			EndDate:   &endDate,
			Page:      1,
			Limit:     10,
		}

		pvzList, err := repo.GetPVZ(ctx, filter)
		if err != nil {
			t.Fatalf("Failed to get PVZ list with date filter: %v", err)
		}

		if len(pvzList) == 0 {
			t.Error("Expected non-empty PVZ list")
		}

		for _, pvz := range pvzList {
			for _, reception := range pvz.Receptions {
				recTime := reception.Reception.DateTime.UTC()
				if recTime.Before(startDate) || recTime.After(endDate) {
					t.Errorf("Reception time %v outside range [%v, %v]", recTime, startDate, endDate)
				}
			}
		}
	})
}
