package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPingHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	PingHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("PingHandler() status = %v, want %v", w.Code, http.StatusOK)
	}

	want := "pong\n"
	if got := w.Body.String(); got != want {
		t.Errorf("PingHandler() response = %v, want %v", got, want)
	}
}
