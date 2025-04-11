package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kosttiik/pvz-service/internal/routes"
	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/database"
	"github.com/kosttiik/pvz-service/pkg/redis"
)

func main() {
	errChan := make(chan error, 2)

	go func() {
		errChan <- database.Connect()
	}()

	go func() {
		errChan <- redis.Connect()
	}()

	for range 2 {
		if err := <-errChan; err != nil {
			log.Fatalf("Failed to initialize services: %v", err)
		}
	}

	defer redis.Close()

	utils.Migrate()

	routes.SetupRoutes()

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
