package db

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	RunStatusRunning = "running"
	RunStatusSuccess = "success"
	RunStatusFailure = "failure"
	RunStatusTimeout = "timeout"
)

type Run struct {
	ID      uuid.UUID `gorm:"type:uuid"`
	Account string

	Recipient     uuid.UUID `gorm:"type:uuid"`
	CorrelationID uuid.UUID `gorm:"type:uuid"`
	URL           string

	Status string
	Labels Labels
	Events []byte

	CreatedAt time.Time
	UpdatedAt time.Time
	Timeout   int
}

type Labels map[string]string

func (l Labels) Value() (driver.Value, error) {
	value, err := json.Marshal(l)
	return string(value), err
}

func (l *Labels) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &l); err != nil {
		return err
	}

	return nil
}
