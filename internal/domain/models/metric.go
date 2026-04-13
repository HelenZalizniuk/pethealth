package models

import "time"

type HealthMetric struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	PetID     uint64    `json:"pet_id" gorm:"index"`
	Type      string    `json:"type"` // e.g., "heart_rate", "temperature"
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}
