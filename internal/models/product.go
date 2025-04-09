package models

import (
	"time"

	"github.com/google/uuid"
)

var ValidProduct = map[string]bool{
	"электроника": true,
	"одежда":      true,
	"обувь":       true,
}

type Product struct {
	ID          uuid.UUID `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"` // электроника, одежда, обувь
	ReceptionID string    `json:"receptionId"`
}
