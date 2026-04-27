package models

import "time"

type OutboxEvent struct {
	ID        string    `gorm:"column:id"`      // event`s` UUID
	PetID     uint64    `gorm:"column:pet_id"`  // for sharding and routing
	Type      string    `gorm:"column:type"`    // e.g. "HealthMetricCreated"
	Payload   []byte    `gorm:"column:payload"` // event data JSON
	Topic     string    `gorm:"column:topic"`
	Status    string    `gorm:"column:status"` // "pending", "processed", "failed"
	CreatedAt time.Time `gorm:"column:created_at"`
}
