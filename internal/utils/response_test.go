package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kosttiik/pvz-service/internal/dto"
)

func TestWriteError(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		status     int
		wantBody   string
		wantStatus int
	}{
		{
			name:       "Bad Request Error",
			message:    "invalid input",
			status:     http.StatusBadRequest,
			wantBody:   `{"message":"invalid input"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Internal Server Error",
			message:    "server error",
			status:     http.StatusInternalServerError,
			wantBody:   `{"message":"server error"}`,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "Empty Message",
			message:    "",
			status:     http.StatusBadRequest,
			wantBody:   `{"message":""}`,
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

			if w.Header().Get("Content-Type") != "application/json" {
				t.Error("WriteError() Content-Type header not set correctly")
			}

			if w.Body.String() != tt.wantBody+"\n" {
				t.Errorf("WriteError() body = %v, want %v", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	type testStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name       string
		data       any
		status     int
		wantBody   string
		wantStatus int
		wantErr    bool
	}{
		{
			name: "Valid struct",
			data: testStruct{
				Name:  "test",
				Value: 123,
			},
			status:     http.StatusOK,
			wantBody:   `{"name":"test","value":123}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "String data",
			data:       "hello",
			status:     http.StatusOK,
			wantBody:   `"hello"`,
			wantStatus: http.StatusOK,
		},
		{
			name: "Error response",
			data: dto.ErrorResponse{
				Message: "error occurred",
			},
			status:     http.StatusBadRequest,
			wantBody:   `{"message":"error occurred"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteJSON(w, tt.data, tt.status)

			if w.Code != tt.wantStatus {
				t.Errorf("WriteJSON() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if w.Header().Get("Content-Type") != "application/json" {
				t.Error("WriteJSON() Content-Type header not set correctly")
			}

			var got interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Errorf("WriteJSON() produced invalid JSON: %v", err)
			}

			want := tt.wantBody + "\n"
			if w.Body.String() != want {
				t.Errorf("WriteJSON() body = %v, want %v", w.Body.String(), want)
			}
		})
	}
}
