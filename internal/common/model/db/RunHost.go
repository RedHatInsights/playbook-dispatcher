package db

import (
	"time"

	"github.com/google/uuid"
)

type RunHost struct {
	ID    uuid.UUID `gorm:"type:uuid"`
	RunID uuid.UUID `gorm:"type:uuid"`

	InventoryID *uuid.UUID `gorm:"type:uuid"`
	Host        string

	SatSequence *int

	Status string
	Log    string

	CreatedAt time.Time
	UpdatedAt time.Time
}
