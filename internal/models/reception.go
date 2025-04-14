package models

import (
	"time"

	"github.com/google/uuid"
)

type ReceptionStatus string

// Статусы приемки
const (
	StatusInProgress ReceptionStatus = "in_progress"
	StatusClosed     ReceptionStatus = "close"
)

var ValidReceptionStatuses = map[ReceptionStatus]bool{
	StatusInProgress: true,
	StatusClosed:     true,
}

type Reception struct {
	ID       uuid.UUID       `json:"id"`
	DateTime time.Time       `json:"dateTime"`
	PvzID    string          `json:"pvzId"`
	Status   ReceptionStatus `json:"status"`
}

func (s ReceptionStatus) IsValid() bool {
	return s == StatusInProgress || s == StatusClosed
}
