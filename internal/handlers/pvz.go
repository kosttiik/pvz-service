package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"strconv"

	"github.com/google/uuid"
	"github.com/kosttiik/pvz-service/internal/metrics"
	"github.com/kosttiik/pvz-service/internal/models"
	"github.com/kosttiik/pvz-service/internal/repository"
	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/database"
	"github.com/kosttiik/pvz-service/pkg/logger"
	"go.uber.org/zap"
)

// форматировние ответа согласно API
type GetPVZListResponse struct {
	PVZ        models.PVZ          `json:"pvz"`
	Receptions []ReceptionResponse `json:"receptions"`
}

type ReceptionResponse struct {
	Reception models.Reception `json:"reception"`
	Products  []models.Product `json:"products"`
}

func CreatePVZHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.Log
	claims := utils.GetUserFromContext(r.Context())
	if claims == nil {
		log.Warn("Unauthorized attempt to create PVZ")
		utils.WriteError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if string(claims.Role) != "moderator" {
		log.Warn("Not moderator attempted to create PVZ",
			zap.String("userID", claims.UserID),
			zap.String("role", string(claims.Role)))
		utils.WriteError(w, "Forbidden", http.StatusForbidden)
		return
	}

	var input struct {
		City string `json:"city"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Warn("Failed to decode PVZ creation request", zap.Error(err))
		utils.WriteError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if !models.AllowedCities[input.City] {
		utils.WriteError(w, "City not allowed", http.StatusBadRequest)
		return
	}

	id := uuid.New()
	pvz := models.PVZ{
		ID:               id,
		City:             input.City,
		RegistrationDate: time.Now().UTC(),
	}

	errChan := make(chan error, 1)

	go func() {
		_, err := database.DB.Exec(r.Context(),
			"INSERT INTO pvz (id, registration_date, city) VALUES ($1, $2, $3)",
			pvz.ID, pvz.RegistrationDate, pvz.City,
		)
		errChan <- err
	}()

	select {
	case err := <-errChan:
		if err != nil {
			utils.WriteError(w, "Failed to create pvz", http.StatusInternalServerError)
			return
		}
		metrics.PvzCreatedTotal.Inc()
	case <-r.Context().Done():
		utils.WriteError(w, "Request timeout", http.StatusGatewayTimeout)
		return
	}

	log.Info("PVZ created successfully",
		zap.String("id", pvz.ID.String()),
		zap.String("city", pvz.City),
		zap.String("createdBy", claims.UserID))

	utils.WriteJSON(w, pvz, http.StatusCreated)
}

func GetPVZListHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.Log
	claims := utils.GetUserFromContext(r.Context())
	if claims == nil {
		log.Warn("Unauthorized attempt to get PVZ list")
		utils.WriteError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if string(claims.Role) != "employee" && string(claims.Role) != "moderator" {
		utils.WriteError(w, "Forbidden", http.StatusForbidden)
		return
	}

	query := r.URL.Query()

	var filter repository.GetPVZFilter

	if startDate := query.Get("startDate"); startDate != "" {
		parsedTime, err := time.Parse(time.RFC3339, startDate)
		if err != nil {
			utils.WriteError(w, "Invalid format of start date", http.StatusBadRequest)
			return
		}

		filter.StartDate = &parsedTime
	}

	if endDate := query.Get("endDate"); endDate != "" {
		parsedTime, err := time.Parse(time.RFC3339, endDate)
		if err != nil {
			utils.WriteError(w, "Invalid format of end date", http.StatusBadRequest)
			return
		}

		filter.EndDate = &parsedTime
	}

	if filter.StartDate != nil && filter.EndDate != nil {
		if filter.EndDate.Before(*filter.StartDate) {
			utils.WriteError(w, "End date cannot be before start date", http.StatusBadRequest)
			return
		}
	}

	filter.Page = 1
	if page := query.Get("page"); page != "" {
		pageNum, err := strconv.Atoi(page)
		if err != nil || pageNum < 1 {
			utils.WriteError(w, "Invalid page number", http.StatusBadRequest)
			return
		}
		filter.Page = pageNum
	}

	filter.Limit = 10
	if limit := query.Get("limit"); limit != "" {
		limitNum, err := strconv.Atoi(limit)
		if err != nil || limitNum < 1 || limitNum > 30 {
			utils.WriteError(w, "Invalid limit", http.StatusBadRequest)
			return
		}
		filter.Limit = limitNum
	}

	log.Debug("Getting PVZ list",
		zap.Any("filter", filter),
		zap.String("requestedBy", claims.UserID))

	pvzRepo := repository.NewPVZRepository(database.DB)
	pvzList, err := pvzRepo.GetPVZ(r.Context(), filter)
	if err != nil {
		log.Error("Failed to get PVZ list",
			zap.Error(err),
			zap.Any("filter", filter))
		utils.WriteError(w, "Failed to get PVZ list", http.StatusInternalServerError)
		return
	}

	log.Debug("Successfully retrieved PVZ list",
		zap.Int("count", len(pvzList)),
		zap.String("requestedBy", claims.UserID))

	response := make([]GetPVZListResponse, 0)
	for _, pvz := range pvzList {
		pvzResponse := GetPVZListResponse{
			PVZ:        pvz.PVZ,
			Receptions: make([]ReceptionResponse, 0),
		}

		for _, rec := range pvz.Receptions {
			receptionResp := ReceptionResponse{
				Reception: rec.Reception,
				Products:  rec.Products,
			}
			pvzResponse.Receptions = append(pvzResponse.Receptions, receptionResp)
		}

		response = append(response, pvzResponse)
	}

	utils.WriteJSON(w, response, http.StatusOK)
}
