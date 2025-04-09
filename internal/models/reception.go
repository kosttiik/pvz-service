package models

import (
	"time"

	"github.com/google/uuid"
)

// Статусы приемки
const (
	StatusInProgress = "in_progress"
	StatusClosed     = "close"
)

type Reception struct {
	ID       uuid.UUID `json:"id"`
	DateTime time.Time `json:"dateTime"`
	PvzID    string    `json:"pvzId"`
	Status   string    `json:"status"`
}
