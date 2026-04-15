package db

import (
	"fmt"

	"pethealth/internal/config"

	"gorm.io/gorm"
)

// manages pools of connections to different shards
type ShardManager struct {
	shards map[int]*gorm.DB
}

func NewShardManager(shardConfigs []config.ShardConfig) (*ShardManager, error) {
	shards := make(map[int]*gorm.DB)

	for i, cfg := range shardConfigs {

		dbConn, err := NewPostgresDB(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to shard %d: %w", i, err)
		}
		shards[i] = dbConn
	}

	return &ShardManager{shards: shards}, nil
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
