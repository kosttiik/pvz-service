package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kosttiik/pvz-service/internal/handlers"
	"github.com/kosttiik/pvz-service/internal/middleware"
	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/database"
)

func main() {
	err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	utils.Migrate()

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})

	http.HandleFunc("/dummyLogin", handlers.DummyLoginHandler)

	http.HandleFunc("/pvz", middleware.AuthMiddleware(
		middleware.RoleMiddleware("moderator")(handlers.CreatePVZHandler)),
	)

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
