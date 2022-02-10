package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaAdmin struct {
	Admin *kafka.AdminClient
}

func NewKafkaAdmin(broker string) (*KafkaAdmin, error) {
	adminClient, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		return nil, err
	}
	return &KafkaAdmin{Admin: adminClient}, nil
}

func (kafkaAdmin *KafkaAdmin) CreateTopics(topic string, numPartitions int, replicationFactor int) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	results, err := kafkaAdmin.Admin.CreateTopics(
		ctx,
		[]kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor}},
		kafka.SetAdminOperationTimeout(time.Second*60),
	)
	if err != nil {
		return fmt.Errorf("Failed to create topic '%s': %w", topic, err)
	}

	for _, result := range results {
		log.Printf("Created topic: %v\n", result.Topic)
	}
	return nil
}

func (kafkaAdmin *KafkaAdmin) IsTopicCreated(topic string) (bool, error) {
	metaData, err := kafkaAdmin.Admin.GetMetadata(
		&topic,
		true,
		5000,
	)
	if err != nil {
		return false, fmt.Errorf("Failed to get Kafka metadata: %w", err)
	}

	return metaData.Topics[topic].Topic != "", nil
}
