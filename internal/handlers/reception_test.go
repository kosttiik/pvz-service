package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/kosttiik/pvz-service/internal/models"
)

func TestCreateReceptionHandler(t *testing.T) {
	// Создаем тестовое пвз
	pvzID := createTestPVZ(t)

	tests := []struct {
		name       string
		pvzID      string
		role       string
		wantStatus int
	}{
		{
			name:       "Valid reception creation",
			pvzID:      pvzID,
			role:       "employee",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "Invalid role",
			pvzID:      uuid.New().String(),
			role:       "moderator",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "Invalid PVZ ID",
			pvzID:      "invalid-uuid",
			role:       "employee",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{"pvzId": tt.pvzID}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer(jsonBody))

			// Получаем тестовый токен
			req = getTestToken(t, tt.role, req)

			w := httptest.NewRecorder()
			CreateReceptionHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CreateReceptionHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if w.Code == http.StatusCreated {
				var reception models.Reception
				if err := json.NewDecoder(w.Body).Decode(&reception); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if reception.PvzID != tt.pvzID {
					t.Errorf("Got pvzID = %v, want %v", reception.PvzID, tt.pvzID)
				}
			}
		})
	}
}

func createTestReception(t *testing.T, pvzID string) string {
	// Создаем приемку с ролью работника пвз
	body := map[string]string{"pvzId": pvzID}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer(jsonBody))
	req = getTestToken(t, "employee", req)
	w := httptest.NewRecorder()

	CreateReceptionHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create test reception: status = %v", w.Code)
	}

	var reception models.Reception
	if err := json.NewDecoder(w.Body).Decode(&reception); err != nil {
		t.Fatalf("Failed to decode reception response: %v", err)
	}
	return reception.ID.String()
}

func TestAddProductHandler(t *testing.T) {
	pvzID := createTestPVZ(t)
	_ = createTestReception(t, pvzID) // Сначала создаем приемку

	tests := []struct {
		name        string
		productType string
		role        string
		wantStatus  int
	}{
		{
			name:        "Valid product",
			productType: "электроника",
			role:        "employee",
			wantStatus:  http.StatusCreated,
		},
		{
			name:        "Invalid product type",
			productType: "invalid",
			role:        "employee",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "Invalid role",
			productType: "электроника",
			role:        "moderator",
			wantStatus:  http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{
				"type":  tt.productType,
				"pvzId": pvzID,
			}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(jsonBody))
			req = getTestToken(t, tt.role, req)
			w := httptest.NewRecorder()

			AddProductHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("AddProductHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestDeleteLastProductHandler(t *testing.T) {
	pvzID := createTestPVZ(t)
	_ = createTestReception(t, pvzID) // Сначала создаем приемку

	// Создаем тестовый продукт
	body := map[string]string{
		"type":  "электроника",
		"pvzId": pvzID,
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(jsonBody))
	req = getTestToken(t, "employee", req)
	w := httptest.NewRecorder()
	AddProductHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create test product: status = %v", w.Code)
	}

	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{
			name:       "Valid deletion",
			role:       "employee",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid role",
			role:       "moderator",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/pvz/%s/delete_last_product", pvzID), nil)
			req = getTestToken(t, tt.role, req)
			w := httptest.NewRecorder()

			DeleteLastProductHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("DeleteLastProductHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestCloseReceptionHandler(t *testing.T) {
	pvzID := createTestPVZ(t)
	_ = createTestReception(t, pvzID)

	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{
			name:       "Valid closure",
			role:       "employee",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid role",
			role:       "moderator",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/pvz/%s/close_last_reception", pvzID), nil)
			req = getTestToken(t, tt.role, req)
			w := httptest.NewRecorder()

			CloseReceptionHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("CloseReceptionHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if w.Code == http.StatusOK {
				var reception models.Reception
				if err := json.NewDecoder(w.Body).Decode(&reception); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if reception.Status != models.StatusClosed {
					t.Errorf("Got status = %v, want %v", reception.Status, models.StatusClosed)
				}
			}
		})
	}
}
