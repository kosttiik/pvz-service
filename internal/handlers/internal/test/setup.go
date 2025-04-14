package test

import (
	"context"

	"github.com/kosttiik/pvz-service/internal/testutils"
	"github.com/kosttiik/pvz-service/pkg/database"
	"github.com/kosttiik/pvz-service/pkg/logger"
	"go.uber.org/zap"
)

var TestSetup struct {
	Cleanup func()
}

func Init() {
	TestSetup.Cleanup = testutils.SetupTestEnvironment()
	log := logger.Log

	log.Info("Cleaning test database")
	result, err := database.DB.Exec(context.Background(), "TRUNCATE users, pvz, reception, product CASCADE")
	if err != nil {
		log.Fatal("Failed to clean test database", zap.Error(err))
	}
	log.Info("Test database cleaned successfully", zap.Int64("rows_affected", result.RowsAffected()))
}
