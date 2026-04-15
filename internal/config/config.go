package config

import "os"

type ShardConfig struct {
	MasterDSN  string
	ReplicaDSN string
}

type Config struct {
	Shards     []ShardConfig
	ServerPort string
}

func Load() *Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080" // default port
	}

	return &Config{
		ServerPort: port,
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
