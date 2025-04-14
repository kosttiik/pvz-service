package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kosttiik/pvz-service/internal/routes"
	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/database"
	"github.com/kosttiik/pvz-service/pkg/logger"
	"github.com/kosttiik/pvz-service/pkg/redis"
	"go.uber.org/zap"
)

func main() {
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Close()

	log := logger.Log

	errChan := make(chan error, 2)

	go func() {
		if err := database.Connect(); err != nil {
			errChan <- fmt.Errorf("database connection failed: %w", err)
			return
		}
		log.Info("Successfully connected to database")
		errChan <- nil
	}()

	go func() {
		if err := redis.Connect(); err != nil {
			errChan <- fmt.Errorf("redis connection failed: %w", err)
			return
		}
		log.Info("Successfully connected to redis")
		errChan <- nil
	}()

	for range 2 {
		if err := <-errChan; err != nil {
			log.Fatal("Failed to initialize services", zap.Error(err))
		}
	}

	defer redis.Close()

	utils.Migrate()
	log.Info("Database migration completed")

	routes.SetupRoutes()
	log.Info("Routes initialized")

	addr := ":8080"
	log.Info("Starting server", zap.String("address", addr))
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("Server failed", zap.Error(err))
	}
}
