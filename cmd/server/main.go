package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/kosttiik/pvz-service/pkg/database"
)

func main() {
	err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
