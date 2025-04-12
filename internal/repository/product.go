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
	query := `
		INSERT INTO product (id, type, reception_id)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, product.ID, product.Type, product.ReceptionID)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}
