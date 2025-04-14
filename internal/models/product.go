package models

import (
	"time"

	"github.com/google/uuid"
)

// Разрешенные товары согласно апи
var ValidProduct = map[string]bool{
	"электроника": true,
	"одежда":      true,
	"обувь":       true,
}

type Product struct {
	ID          uuid.UUID `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"`
	ReceptionID string    `json:"receptionId"`
}
