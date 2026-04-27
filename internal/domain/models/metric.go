package models

import (
	"time"
)

type HealthMetric struct {
	ID         string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ExternalID string    `json:"external_id" gorm:"uniqueIndex" validate:"required,uuid4"`
	PetID      uint64    `json:"pet_id" gorm:"index" validate:"required,gt=0"`
	Type       string    `json:"type" validate:"required,oneof=heart_rate temperature weight"` // e.g., "heart_rate", "temperature"
	Value      float64   `json:"value" validate:"required,gt=0"`
	ShardID    int       `json:"shard_id"`
	Timestamp  time.Time `json:"timestamp" validate:"required"`
}
