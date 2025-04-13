package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kosttiik/pvz-service/pkg/database"
	"github.com/kosttiik/pvz-service/pkg/redis"
)

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test_secret")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "pvz_db")
	os.Setenv("REDIS_HOST", "localhost")

	if err := database.Connect(); err != nil {
		panic(err)
	}
	if err := redis.Connect(); err != nil {
		panic(err)
	}
	code := m.Run()
	redis.Close()
	os.Exit(code)
}

func TestDummyLoginHandler(t *testing.T) {
	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{"Valid employee", "employee", http.StatusOK},
		{"Valid moderator", "moderator", http.StatusOK},
		{"Invalid role", "invalid", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := map[string]string{"role": tt.role}
			jsonBody, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPost, "/dummyLogin", bytes.NewBuffer(jsonBody))
			w := httptest.NewRecorder()

			DummyLoginHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var response struct {
					Token string `json:"token"`
				}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("failed to decode response: %v", err)
				}
				if response.Token == "" {
					t.Error("token is empty")
				}
			}
		})
	}
}
