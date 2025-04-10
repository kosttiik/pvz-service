package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kosttiik/pvz-service/internal/routes"
	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/cache"
	"github.com/kosttiik/pvz-service/pkg/database"
)

func main() {
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}

	if err := cache.ConnectRedis(); err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	utils.Migrate()

	routes.SetupRoutes()

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
