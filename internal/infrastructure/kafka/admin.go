package kafka

import (
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
)

func EnsureTopics(brokers []string) error {

	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}
	var controllerConn *kafka.Conn
	controllerConn, err = kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	// 3. Описываем нужные топики (из твоего конфига)
	topicConfigs := []kafka.TopicConfig{
		{Topic: "pet_events", NumPartitions: 3, ReplicationFactor: 1},
		{Topic: "pet_events_dlq", NumPartitions: 1, ReplicationFactor: 1},
		{Topic: "pet_saga_responses", NumPartitions: 3, ReplicationFactor: 1},
	}

	// 4. Создаем (если их нет, Kafka сама проверит)
	err = controllerConn.CreateTopics(topicConfigs...)
	return err
}
