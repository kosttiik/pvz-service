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
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
        INSERT INTO reception (id, pvz_id, status)
        VALUES ($1, $2, $3)
    `
	if _, err := tx.Exec(ctx, query, reception.ID, reception.PvzID, reception.Status); err != nil {
		return fmt.Errorf("failed to create reception: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
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

func (r *ReceptionRepository) CloseLastReception(ctx context.Context, pvzID string) (*models.Reception, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
        UPDATE reception 
        SET status = $1
        WHERE id = (
            SELECT id FROM reception
            WHERE pvz_id = $2 AND status = $3
            ORDER BY date_time DESC
            LIMIT 1
        )
        RETURNING id, date_time, pvz_id, status
    `

	reception := &models.Reception{}
	err = tx.QueryRow(ctx, query, models.StatusClosed, pvzID, models.StatusInProgress).Scan(
		&reception.ID,
		&reception.DateTime,
		&reception.PvzID,
		&reception.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no open reception found")
		}
		return nil, fmt.Errorf("failed to close reception: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return reception, nil
}
