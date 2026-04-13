package app

import (
	"fmt"
	"pethealth/internal/config"
	"pethealth/internal/infrastructure/db"

	"gorm.io/gorm"
)

type App struct {
	ShardManager *db.ShardManager
	Config       *config.Config
}

func NewApp() (*App, error) {
	cfg := config.Load()
	shards := make(map[int]*gorm.DB)

	for i, shardCfg := range cfg.Shards {

		conn, err := db.NewPostgresDB(shardCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to shard %d: %w", i, err)
		}
		shards[i] = conn
	}

	shardManager := db.NewShardManager(shards)

	return &App{
		ShardManager: shardManager,
		Config:       cfg,
	}, nil
}
