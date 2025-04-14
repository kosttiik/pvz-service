package utils

import (
	"context"

	"github.com/kosttiik/pvz-service/pkg/database"
	"github.com/kosttiik/pvz-service/pkg/logger"
	"go.uber.org/zap"
)

func Migrate() {
	log := logger.Log
	connection := database.DB
	ctx := context.Background()

	tx, err := connection.Begin(ctx)
	if err != nil {
		log.Fatal("Failed to start migration transaction", zap.Error(err))
	}
	defer tx.Rollback(ctx)

	log.Info("Starting database migration")

	sql := `
CREATE TABLE IF NOT EXISTS pvz (
	id UUID PRIMARY KEY,
	registration_date TIMESTAMP NOT NULL DEFAULT now(),
	city VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS reception (
	id UUID PRIMARY KEY,
	date_time TIMESTAMP NOT NULL DEFAULT now(),
	pvz_id UUID REFERENCES pvz(id),
	status VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS product (
	id UUID PRIMARY KEY,
	date_time TIMESTAMP NOT NULL DEFAULT now(),
	type VARCHAR(50) NOT NULL,
	reception_id UUID REFERENCES reception(id)
);

CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY,
	email VARCHAR(255) UNIQUE NOT NULL,
	password VARCHAR(255) NOT NULL,
	role VARCHAR(50) NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
`
	if _, err := tx.Exec(ctx, sql); err != nil {
		log.Fatal("Failed to execute migrations", zap.Error(err))
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatal("Failed to commit migrations", zap.Error(err))
	}

	log.Info("Database migration completed successfully")
}
