package models

import "time"

type Pet struct {
	ID        uint64    `json:"id" gorm:"primaryKey"`
	OwnerID   uint64    `json:"owner_id"`
	Name      string    `json:"name"`
	Species   string    `json:"species"`
	ShardID   int       `json:"shard_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
