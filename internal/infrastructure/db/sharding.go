package db

import (
	"fmt"

	"gorm.io/gorm"
)

// manages pools of connections to different shards
type ShardManager struct {
	shards map[int]*gorm.DB
}

func NewShardManager(shards map[int]*gorm.DB) *ShardManager {
	return &ShardManager{shards: shards}
}

// selects the appropriate shard based on the pet ID
func (s *ShardManager) GetShardById(id uint64) (*gorm.DB, error) {
	numShards := len(s.shards)
	if numShards == 0 {
		return nil, fmt.Errorf("no shards available")
	}

	// Simple logic: remainder of division
	shardIndex := int(id % uint64(numShards))

	db, ok := s.shards[shardIndex]
	if !ok {
		return nil, fmt.Errorf("shard %d not found", shardIndex)
	}

	return db, nil
}
