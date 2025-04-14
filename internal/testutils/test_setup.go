package testutils

import (
	"os"

	"github.com/kosttiik/pvz-service/pkg/database"
	"github.com/kosttiik/pvz-service/pkg/logger"
	"github.com/kosttiik/pvz-service/pkg/redis"
	"go.uber.org/zap"
)

func SetupTestEnvironment() func() {
	setupEnvVars()

	if err := logger.Init(); err != nil {
		panic(err)
	}
	log := logger.Log
	log.Info("Test logger initialized")

	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to test database", zap.Error(err))
	}
	log.Info("Connected to test database")

	if err := redis.Connect(); err != nil {
		log.Fatal("Failed to connect to test redis", zap.Error(err))
	}
	log.Info("Connected to test redis")

	return func() {
		log.Info("Cleaning up test environment")
		redis.Close()
		logger.Close()
	}
}

func setupEnvVars() {
	testEnv := map[string]string{
		"JWT_SECRET":  "test_secret",
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_USER":     "postgres",
		"DB_PASSWORD": "postgres",
		"DB_NAME":     "pvz_db",
		"REDIS_HOST":  "localhost",
		"LOG_LEVEL":   "debug",
	}

	for k, v := range testEnv {
		os.Setenv(k, v)
	}
}
