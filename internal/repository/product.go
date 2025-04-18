package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kosttiik/pvz-service/internal/models"
)

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO product (id, type, reception_id)
		VALUES ($1, $2, $3)
	`
	if _, err := tx.Exec(ctx, query, product.ID, product.Type, product.ReceptionID); err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *ProductRepository) DeleteLastFromReception(ctx context.Context, receptionID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
        DELETE FROM product 
        WHERE id = (
            SELECT id 
            FROM product 
            WHERE reception_id = $1 
            ORDER BY date_time DESC 
            LIMIT 1
        )
        RETURNING id`

	var deletedID string
	err = tx.QueryRow(ctx, query, receptionID).Scan(&deletedID)
	if err != nil {
		return fmt.Errorf("failed to delete last product: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
