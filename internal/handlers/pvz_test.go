package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kosttiik/pvz-service/internal/models"
)

func TestCreatePVZHandler(t *testing.T) {
	tests := []struct {
		name       string
		city       string
		role       string
		wantStatus int
	}{
		{
			name:       "Valid PVZ creation",
			city:       "Москва",
			role:       "moderator",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "Invalid role",
			city:       "Москва",
			role:       "employee",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Invalid city",
			city:       "Маскваааааа",
			role:       "moderator",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "No auth token",
			city:       "Москва",
			role:       "",
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{"city": tt.city}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewBuffer(jsonBody))
			if tt.role != "" {
				req = getTestToken(t, tt.role, req)
			}
			w := httptest.NewRecorder()

			CreatePVZHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreatePVZHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if w.Code == http.StatusCreated {
				var pvz models.PVZ
				if err := json.NewDecoder(w.Body).Decode(&pvz); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if pvz.City != tt.city {
					t.Errorf("Got city = %v, want %v", pvz.City, tt.city)
				}
			}
		})
	}
}

func TestGetPVZListHandler(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		query      string
		setupData  bool
		wantStatus int
	}{
		{
			name:       "Valid request employee",
			role:       "employee",
			query:      "?page=1&limit=10",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Valid request moderator",
			role:       "moderator",
			query:      "?page=1&limit=10",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid page",
			role:       "employee",
			query:      "?page=0&limit=10",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid limit",
			role:       "employee",
			query:      "?page=1&limit=50",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid date format",
			role:       "employee",
			query:      "?startDate=invalid-date",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "End date before start date",
			role:       "employee",
			query:      "?startDate=2025-01-02T00:00:00Z&endDate=2024-01-01T00:00:00Z",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "With data",
			role:       "employee",
			query:      "?page=1&limit=10",
			setupData:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "No auth",
			role:       "",
			query:      "?page=1&limit=10",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid role",
			role:       "invalid",
			query:      "?page=1&limit=10",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Zero limit",
			role:       "employee",
			query:      "?page=1&limit=0",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupData {
				createTestPVZ(t)
			}
			req := httptest.NewRequest(http.MethodGet, "/pvz"+tt.query, nil)
			if tt.role != "" {
				req = getTestToken(t, tt.role, req)
			}

			w := httptest.NewRecorder()
			GetPVZListHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetPVZListHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if w.Code == http.StatusOK {
				var response []struct {
					PVZ        models.PVZ `json:"pvz"`
					Receptions []struct {
						Reception models.Reception `json:"reception"`
						Products  []models.Product `json:"products"`
					} `json:"receptions"`
				}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
			}
		})
	}
}
