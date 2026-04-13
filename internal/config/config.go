package config

import "os"

type ShardConfig struct {
	MasterDSN  string
	ReplicaDSN string
}

type Config struct {
	Shards []ShardConfig
}

func Load() *Config {
	return &Config{
		Shards: []ShardConfig{
			{
				MasterDSN:  os.Getenv("SHARD_0_MASTER_DSN"),
				ReplicaDSN: os.Getenv("SHARD_0_REPLICA_DSN"),
			},
			{
				MasterDSN:  os.Getenv("SHARD_1_MASTER_DSN"),
				ReplicaDSN: os.Getenv("SHARD_1_REPLICA_DSN"),
			},
		},
	}
}
