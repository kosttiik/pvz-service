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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{"city": tt.city}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewBuffer(jsonBody))
			req = getTestToken(t, tt.role, req)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/pvz"+tt.query, nil)
			req = getTestToken(t, tt.role, req)

			w := httptest.NewRecorder()
			GetPVZListHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("GetPVZListHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}
