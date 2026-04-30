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

func (s *ShardManager) Ping() error {

	for i, dbConn := range s.shards {
		// getting sql.DB from GORM wrapper
		sqlDB, err := dbConn.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB for shard %d: %w", i, err)
		}
		// pinging each shard to ensure connectivity
		if err := sqlDB.Ping(); err != nil {
			return fmt.Errorf("shard %d is unreachable: %w", i, err)
		}
	}
	return nil
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
	shardIndex := s.GetShardIndex(id)

	db, ok := s.shards[shardIndex]
	if !ok {
		return nil, fmt.Errorf("shard %d not found", shardIndex)
	}

	return db, nil
}

func (s *ShardManager) GetShardIndex(id uint64) int {
	numShards := len(s.shards)
	if numShards == 0 {
		return 0
	}
	return int(id % uint64(numShards))
}

func (s *ShardManager) GetShardName(id uint64) string {
	shardIndex := s.GetShardIndex(id)
	return fmt.Sprintf("shard_%d", shardIndex)
}
