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
	ServerPort         string
	KafkaBrokers       []string
	KafkaTopic         string
	KafkaDLQTopic      string
	KafkaConsumerGroup string
	Shards             []ShardConfig
}

func Load() *Config {
	port := getEnv("SERVER_PORT", "8080")
	topic := getEnv("KAFKA_TOPIC", "pet_events")
	dlqTopic := getEnv("KAFKA_DLQ_TOPIC", "pet_events_dlq")
	group := getEnv("KAFKA_CONSUMER_GROUP", "health-service-group")

	brokersEnv := os.Getenv("KAFKA_BROKERS")
	if brokersEnv == "" {
		brokersEnv = "localhost:9092"
	}
	brokers := strings.Split(brokersEnv, ",")

	return &Config{
		ServerPort:         port,
		KafkaBrokers:       brokers,
		KafkaTopic:         topic,
		KafkaDLQTopic:      dlqTopic,
		KafkaConsumerGroup: group,
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

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
