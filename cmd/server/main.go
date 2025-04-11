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
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}

	if err := redis.Connect(); err != nil {
		log.Fatalf("Failed to connect to redis: %v", err)
	}

	defer redis.Close()

	utils.Migrate()

	routes.SetupRoutes()

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
