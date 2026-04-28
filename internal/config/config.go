package config

import (
	"os"
	"strings"
)

type ShardConfig struct {
	MasterDSN  string
	ReplicaDSN string
}

type Config struct {
	Shards       []ShardConfig
	ServerPort   string
	KafkaBrokers []string // Brokers list
	KafkaTopic   string   // Metrics topic
}

func Load() *Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080" // default port
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "health_alerts"
	}

	brokersEnv := os.Getenv("KAFKA_BROKERS")
	brokers := strings.Split(brokersEnv, ",")

	return &Config{
		ServerPort:   port,
		KafkaBrokers: brokers,
		KafkaTopic:   topic,
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
