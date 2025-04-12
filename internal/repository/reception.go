package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kosttiik/pvz-service/internal/models"
)

type ReceptionRepository struct {
	db *pgxpool.Pool
}

func NewReceptionRepository(db *pgxpool.Pool) *ReceptionRepository {
	return &ReceptionRepository{db: db}
}

func (r *ReceptionRepository) HasOpenReception(ctx context.Context, pvzID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM reception 
			WHERE pvz_id = $1 AND status = 'in_progress'
		)
	`
	var exists bool
	err := r.db.QueryRow(ctx, query, pvzID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check open reception: %w", err)
	}
	return exists, nil
}

func (r *ReceptionRepository) Create(ctx context.Context, reception *models.Reception) error {
	query := `
		INSERT INTO reception (id, pvz_id, status)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, reception.ID, reception.PvzID, reception.Status)
	if err != nil {
		return fmt.Errorf("failed to create reception: %w", err)
	}
	return nil
}

func (r *ReceptionRepository) GetLastOpenReception(ctx context.Context, pvzID string) (*models.Reception, error) {
	query := `
		SELECT id, date_time, pvz_id, status
		FROM reception
		WHERE pvz_id = $1 AND status = 'in_progress'
		ORDER BY date_time DESC
		LIMIT 1
	`
	reception := &models.Reception{}
	err := r.db.QueryRow(ctx, query, pvzID).Scan(
		&reception.ID,
		&reception.DateTime,
		&reception.PvzID,
		&reception.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no open reception found")
		}
		return nil, fmt.Errorf("failed to get reception: %w", err)
	}
	return reception, nil
}
