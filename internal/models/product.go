package models

import "time"

type Product struct {
	ID          string    `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"` // электроника, одежда, обувь
	ReceptionID string    `json:"receptionId"`
}
