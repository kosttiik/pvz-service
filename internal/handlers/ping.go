package handlers

import (
	"fmt"
	"net/http"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "pong")
}
