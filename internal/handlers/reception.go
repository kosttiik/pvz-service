package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kosttiik/pvz-service/internal/metrics"
	"github.com/kosttiik/pvz-service/internal/models"
	"github.com/kosttiik/pvz-service/internal/repository"
	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/database"
	"github.com/kosttiik/pvz-service/pkg/logger"
	"go.uber.org/zap"
)

func CreateReceptionHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.Log
	claims := utils.GetUserFromContext(r.Context())
	if claims == nil {
		log.Warn("Unauthorized attempt to create reception")
		utils.WriteError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if string(claims.Role) != "employee" {
		log.Warn("Not employee attempted to create reception",
			zap.String("userID", claims.UserID),
			zap.String("role", string(claims.Role)))
		utils.WriteError(w, "Forbidden", http.StatusForbidden)
		return
	}

	var input struct {
		PvzID string `json:"pvzId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		log.Warn("Failed to decode reception creation request", zap.Error(err))
		utils.WriteError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if _, err := uuid.Parse(input.PvzID); err != nil {
		utils.WriteError(w, "Invalid PVZ ID format", http.StatusBadRequest)
		return
	}

	receptionRepo := repository.NewReceptionRepository(database.DB)

	hasOpen, err := receptionRepo.HasOpenReception(r.Context(), input.PvzID)
	if err != nil {
		fmt.Printf("Error checking open reception: %v\n", err)
		utils.WriteError(w, "Failed to check open reception", http.StatusInternalServerError)
		return
	}

	if hasOpen {
		utils.WriteError(w, "PVZ already has an open reception", http.StatusBadRequest)
		return
	}

	reception := models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now().UTC(),
		PvzID:    input.PvzID,
		Status:   models.StatusInProgress,
	}

	if err := receptionRepo.Create(r.Context(), &reception); err != nil {
		fmt.Printf("Error creating reception: %v\n", err)
		utils.WriteError(w, "Failed to create reception", http.StatusInternalServerError)
		return
	}

	log.Info("Reception created successfully",
		zap.String("id", reception.ID.String()),
		zap.String("pvzId", reception.PvzID),
		zap.String("createdBy", claims.UserID))

	utils.WriteJSON(w, reception, http.StatusCreated)
	metrics.OrderReceiptsCreatedTotal.Inc()
}

func AddProductHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.Log
	claims := utils.GetUserFromContext(r.Context())
	if claims == nil {
		log.Warn("Unauthorized attempt to add product")
		utils.WriteError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if string(claims.Role) != "employee" {
		utils.WriteError(w, "Forbidden", http.StatusForbidden)
		return
	}

	var input struct {
		Type  string `json:"type"`
		PvzID string `json:"pvzId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.WriteError(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if !models.ValidProduct[input.Type] {
		utils.WriteError(w, "Invalid product type", http.StatusBadRequest)
		return
	}

	receptionRepo := repository.NewReceptionRepository(database.DB)
	reception, err := receptionRepo.GetLastOpenReception(r.Context(), input.PvzID)
	if err != nil {
		utils.WriteError(w, fmt.Sprintf("Failed to get last open reception: %v", err), http.StatusBadRequest)
		return
	}

	product := models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now().UTC(),
		Type:        input.Type,
		ReceptionID: reception.ID.String(),
	}

	productRepo := repository.NewProductRepository(database.DB)
	if err := productRepo.Create(r.Context(), &product); err != nil {
		utils.WriteError(w, fmt.Sprintf("Failed to create product: %v", err), http.StatusInternalServerError)
		return
	}

	log.Info("Product added successfully",
		zap.String("id", product.ID.String()),
		zap.String("type", product.Type),
		zap.String("receptionId", product.ReceptionID),
		zap.String("addedBy", claims.UserID))

	utils.WriteJSON(w, product, http.StatusCreated)
	metrics.ProductsAddedTotal.Inc()
}

func CloseReceptionHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.Log
	claims := utils.GetUserFromContext(r.Context())
	if claims == nil {
		log.Warn("Unauthorized attempt to close reception")
		utils.WriteError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if string(claims.Role) != "employee" {
		utils.WriteError(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Извлекаем pvzId из URL
	pvzID := strings.TrimPrefix(r.URL.Path, "/pvz/")
	pvzID = strings.TrimSuffix(pvzID, "/close_last_reception")

	if _, err := uuid.Parse(pvzID); err != nil {
		utils.WriteError(w, "Invalid PVZ ID", http.StatusBadRequest)
		return
	}

	receptionRepo := repository.NewReceptionRepository(database.DB)
	reception, err := receptionRepo.CloseLastReception(r.Context(), pvzID)
	if err != nil {
		if err.Error() == "no open reception found" {
			utils.WriteError(w, "No open reception found", http.StatusBadRequest)
			return
		}
		utils.WriteError(w, "Failed to close reception", http.StatusInternalServerError)
		return
	}

	log.Info("Reception closed successfully",
		zap.String("id", reception.ID.String()),
		zap.String("pvzId", reception.PvzID),
		zap.String("closedBy", claims.UserID))

	utils.WriteJSON(w, reception, http.StatusOK)
}

func DeleteLastProductHandler(w http.ResponseWriter, r *http.Request) {
	claims := utils.GetUserFromContext(r.Context())
	if claims == nil {
		utils.WriteError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if string(claims.Role) != "employee" {
		utils.WriteError(w, "Forbidden", http.StatusForbidden)
		return
	}

	pvzID := strings.TrimPrefix(r.URL.Path, "/pvz/")
	pvzID = strings.TrimSuffix(pvzID, "/delete_last_product")

	if _, err := uuid.Parse(pvzID); err != nil {
		utils.WriteError(w, "Invalid PVZ ID", http.StatusBadRequest)
		return
	}

	receptionRepo := repository.NewReceptionRepository(database.DB)
	reception, err := receptionRepo.GetLastOpenReception(r.Context(), pvzID)
	if err != nil {
		utils.WriteError(w, "No open reception found", http.StatusBadRequest)
		return
	}

	productRepo := repository.NewProductRepository(database.DB)
	if err := productRepo.DeleteLastFromReception(r.Context(), reception.ID.String()); err != nil {
		utils.WriteError(w, "Failed to delete product", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
