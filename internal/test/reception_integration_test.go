package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kosttiik/pvz-service/internal/handlers"
	"github.com/kosttiik/pvz-service/internal/middleware"
	"github.com/kosttiik/pvz-service/internal/models"
	"github.com/kosttiik/pvz-service/internal/utils"
	"github.com/kosttiik/pvz-service/pkg/database"
	"github.com/kosttiik/pvz-service/pkg/logger"
	"github.com/kosttiik/pvz-service/pkg/redis"
)

func init() {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "pvz_db")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("JWT_SECRET", "test_secret")

	if err := logger.Init(); err != nil {
		panic(err)
	}
}

func TestReceptionWorkflow(t *testing.T) {
	if err := database.Connect(); err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	if err := redis.Connect(); err != nil {
		t.Fatalf("Failed to connect to redis: %v", err)
	}
	defer redis.Close()
	defer logger.Close()

	utils.Migrate()

	// Очищаем таблицы перед тестом
	_, err := database.DB.Exec(context.Background(), "TRUNCATE pvz, reception, product CASCADE")
	if err != nil {
		t.Fatalf("Failed to cleanup tables: %v", err)
	}

	t.Log("Getting moderator token...")
	moderatorToken := getModeratorToken(t)

	t.Log("Creating PVZ...")
	pvz := createPVZ(t, moderatorToken)
	t.Logf("Created PVZ with ID: %s", pvz.ID)

	t.Log("Getting employee token...")
	employeeToken := getEmployeeToken(t)

	t.Log("Creating reception...")
	reception := createReception(t, employeeToken, pvz.ID.String())
	t.Logf("Created reception with ID: %s", reception.ID)

	t.Log("Adding products...")
	for i := range 50 {
		product := addProduct(t, employeeToken, pvz.ID.String())
		t.Logf("Added product %d with ID: %s", i+1, product.ID)
	}

	t.Log("Closing reception...")
	closedReception := closeReception(t, employeeToken, pvz.ID.String())
	t.Logf("Closed reception with status: %s", closedReception.Status)

	// Сверяем статус приёмки
	if closedReception.Status != models.StatusClosed {
		t.Errorf("Reception should be closed, got status: %s", closedReception.Status)
	}
}

func getEmployeeToken(t *testing.T) string {
	body := map[string]string{"role": "employee"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handlers.DummyLoginHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get token, status: %d", w.Code)
	}

	var response struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return response.Token
}

func getModeratorToken(t *testing.T) string {
	body := map[string]string{"role": "moderator"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handlers.DummyLoginHandler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to get moderator token, status: %d", w.Code)
	}

	var response struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return response.Token
}

func createPVZ(t *testing.T, token string) models.PVZ {
	body := map[string]string{"city": "Москва"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/pvz", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w := httptest.NewRecorder()

	middleware.AuthMiddleware(middleware.RoleMiddleware("moderator")(handlers.CreatePVZHandler))(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create PVZ, status: %d", w.Code)
	}

	var pvz models.PVZ
	if err := json.NewDecoder(w.Body).Decode(&pvz); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return pvz
}

func createReception(t *testing.T, token string, pvzID string) models.Reception {
	body := map[string]string{"pvzId": pvzID}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/receptions", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w := httptest.NewRecorder()

	middleware.AuthMiddleware(middleware.RoleMiddleware("employee")(handlers.CreateReceptionHandler))(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to create reception, status: %d", w.Code)
	}

	var reception models.Reception
	if err := json.NewDecoder(w.Body).Decode(&reception); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return reception
}

func addProduct(t *testing.T, token string, pvzID string) models.Product {
	body := map[string]string{
		"type":  "электроника",
		"pvzId": pvzID,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w := httptest.NewRecorder()

	middleware.AuthMiddleware(
		middleware.RoleMiddleware("employee")(
			func(w http.ResponseWriter, r *http.Request) {
				handlers.AddProductHandler(w, r)
			},
		),
	)(w, req)

	if w.Code != http.StatusCreated {
		var errResp struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(w.Body).Decode(&errResp); err == nil {
			t.Fatalf("Failed to add product, status: %d, message: %s", w.Code, errResp.Message)
		} else {
			t.Fatalf("Failed to add product, status: %d", w.Code)
		}
	}

	var product models.Product
	if err := json.NewDecoder(w.Body).Decode(&product); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return product
}

func closeReception(t *testing.T, token string, pvzID string) models.Reception {
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/pvz/%s/close_last_reception", pvzID), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	w := httptest.NewRecorder()

	middleware.AuthMiddleware(middleware.RoleMiddleware("employee")(handlers.CloseReceptionHandler))(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Failed to close reception, status: %d", w.Code)
	}

	var reception models.Reception
	if err := json.NewDecoder(w.Body).Decode(&reception); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return reception
}
