package models

import "time"

type OutboxEvent struct {
	ID        string // event`s` UUID
	Payload   []byte // event data JSON
	Topic     string
	Status    string // "pending", "processed", "failed"
	CreatedAt time.Time
}
