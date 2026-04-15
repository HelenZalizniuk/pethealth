package models

import "time"

type OutboxEvent struct {
	ID        string // event`s` UUID
	PetID     uint64
	Type      string
	Payload   []byte // event data JSON
	Topic     string
	Status    string // "pending", "processed", "failed"
	CreatedAt time.Time
}
