package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kosttiik/pvz-service/internal/dto"
)

func TestWriteError(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		status     int
		wantStatus int
	}{
		{
			name:       "Bad Request Error",
			message:    "invalid input",
			status:     http.StatusBadRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Internal Server Error",
			message:    "server error",
			status:     http.StatusInternalServerError,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Empty Message",
			message:    "",
			status:     http.StatusBadRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Unicode message",
			message:    "ошибка",
			status:     http.StatusBadRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Special characters",
			message:    "error: !@#$%^&*()",
			status:     http.StatusBadRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "JSON special chars",
			message:    `test "quotes" and \backslashes\`,
			status:     http.StatusBadRequest,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteError(w, tt.message, tt.status)

			if w.Code != tt.wantStatus {
				t.Errorf("WriteError() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Content-Type = %v, want application/json", contentType)
			}

			var gotErr dto.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&gotErr); err != nil {
				t.Errorf("Failed to decode error response: %v", err)
			}

			if gotErr.Message != tt.message {
				t.Errorf("WriteError() message = %v, want %v", gotErr.Message, tt.message)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name       string
		data       any
		status     int
		wantStatus int
		wantBody   string
	}{
		{
			name: "Valid struct",
			data: struct {
				Name  string `json:"name"`
				Value int    `json:"value"`
			}{
				Name:  "test",
				Value: 123,
			},
			status:     http.StatusOK,
			wantStatus: http.StatusOK,
			wantBody:   `{"name":"test","value":123}`,
		},
		{
			name:       "String data",
			data:       "hello",
			status:     http.StatusOK,
			wantStatus: http.StatusOK,
			wantBody:   `"hello"`,
		},
		{
			name: "Error response",
			data: dto.ErrorResponse{
				Message: "error occurred",
			},
			status:     http.StatusBadRequest,
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"message":"error occurred"}`,
		},
		{
			name:       "Nil data",
			data:       nil,
			status:     http.StatusOK,
			wantStatus: http.StatusOK,
			wantBody:   "null",
		},
		{
			name: "Complex nested struct",
			data: struct {
				Items []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"items"`
			}{
				Items: []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				}{{ID: "1", Name: "test"}},
			},
			status:     http.StatusOK,
			wantStatus: http.StatusOK,
			wantBody:   `{"items":[{"id":"1","name":"test"}]}`,
		},
		{
			name:       "Map data",
			data:       map[string]interface{}{"key": "value"},
			status:     http.StatusOK,
			wantStatus: http.StatusOK,
			wantBody:   `{"key":"value"}`,
		},
		{
			name:       "Array data",
			data:       []string{"one", "two"},
			status:     http.StatusOK,
			wantStatus: http.StatusOK,
			wantBody:   `["one","two"]`,
		},
		{
			name:       "Number data",
			data:       123,
			status:     http.StatusOK,
			wantStatus: http.StatusOK,
			wantBody:   `123`,
		},
		{
			name:       "Empty slice",
			data:       []string{},
			status:     http.StatusOK,
			wantStatus: http.StatusOK,
			wantBody:   `[]`,
		},
		{
			name:       "Boolean true",
			data:       true,
			status:     http.StatusOK,
			wantStatus: http.StatusOK,
			wantBody:   `true`,
		},
		{
			name:       "Boolean false",
			data:       false,
			status:     http.StatusOK,
			wantStatus: http.StatusOK,
			wantBody:   `false`,
		},
		{
			name:       "Float number",
			data:       123.45,
			status:     http.StatusOK,
			wantStatus: http.StatusOK,
			wantBody:   `123.45`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteJSON(w, tt.data, tt.status)

			if w.Code != tt.wantStatus {
				t.Errorf("WriteJSON() status = %v, want %v", w.Code, tt.wantStatus)
			}

			got := strings.TrimSpace(w.Body.String())
			if got != tt.wantBody {
				t.Errorf("WriteJSON() body = %v, want %v", got, tt.wantBody)
			}
		})
	}

	t.Run("Headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		data := map[string]string{"key": "value"}
		WriteJSON(w, data, http.StatusOK)

		if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
			t.Errorf("Content-Type = %v, want application/json", contentType)
		}
	})
}
