package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/kosttiik/pvz-service/internal/models"
	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/database"
)

var allowedCities = map[string]bool{
	"Москва":          true,
	"Санкт-Петербург": true,
	"Казань":          true,
}

func CreatePVZHandler(w http.ResponseWriter, r *http.Request) {
	claims := utils.GetUserFromContext(r.Context())
	if claims == nil {
		utils.WriteError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if string(claims.Role) != "moderator" {
		utils.WriteError(w, "Forbidden", http.StatusForbidden)
		return
	}

	var input struct {
		City string `json:"city"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if !allowedCities[input.City] {
		utils.WriteError(w, "City not allowed", http.StatusBadRequest)
		return
	}

	id := uuid.New()
	pvz := models.PVZ{
		ID:               id,
		City:             input.City,
		RegistrationDate: time.Now().UTC(),
	}

	_, err := database.DB.Exec(r.Context(),
		"INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)",
		pvz.ID, pvz.RegistrationDate, pvz.City,
	)

	if err != nil {
		utils.WriteError(w, "Failed to create pvz", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, pvz, http.StatusCreated)
}
