package kafkaproducer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"go-word-cloud/shared/config"
	"go-word-cloud/shared/domain"
	"io/ioutil"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/riferrei/srclient"
)

type FeedKafkaProducer struct {
	Producer *kafka.Producer
	Schema   *srclient.Schema
}

func NewFeedKafkaProducer() *FeedKafkaProducer {
	// Create a producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": config.AppConfig.Kafka.BootstrapServers})
	if err != nil {
		panic(err)
	}

	// Fetch the latest version of the schema, or create a new one if it is the first
	topic := config.AppConfig.Kafka.Topic
	schemaRegistryClient := srclient.CreateSchemaRegistryClient(config.AppConfig.Kafka.SchemaRegistryUrl)
	schema, err := schemaRegistryClient.GetLatestSchema(topic)
	if schema == nil {
		schemaBytes, err := ioutil.ReadFile(config.AppConfig.Kafka.AvroSchemaFile)
		if err != nil {
			panic(fmt.Sprintf("Error reading the avsc file %s", err))
		}

		schema, err = schemaRegistryClient.CreateSchema(topic, string(schemaBytes), srclient.Avro)
		if err != nil {
			panic(fmt.Sprintf("Error creating the schema %s", err))
		}
	}

	return &FeedKafkaProducer{
		Producer: p,
		Schema:   schema,
	}
}

func (fkp *FeedKafkaProducer) Send(feedItem *domain.FeedItem) {
	go func() {
		for event := range fkp.Producer.Events() {
			switch ev := event.(type) {
			case *kafka.Message:
				message := ev
				if ev.TopicPartition.Error != nil {
					log.Printf("Error delivering the message '%s': %v\n", message.Key, ev.TopicPartition.Error)
				} else {
					log.Printf("Message '%s' delivered successfully!\n", message.Key)
				}
			}
		}
	}()

	err := fkp.Producer.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &config.AppConfig.Kafka.Topic,
				Partition: kafka.PartitionAny,
			},
			Key:   []byte(feedItem.Id),
			Value: fkp.serializeRecord(feedItem),
		},
		nil,
	)
	if err != nil {
		log.Printf("Error producing the message %s\n", feedItem.Id)
	}

	fkp.Producer.Flush(15 * 1000)
}

func (fkp *FeedKafkaProducer) serializeRecord(feedItem *domain.FeedItem) []byte {
	// Serialize the record using the schema provided by the client,
	// making sure to include the schema id as part of the record.
	schemaIDBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(schemaIDBytes, uint32(fkp.Schema.ID()))
	value, _ := json.Marshal(feedItem)
	native, _, _ := fkp.Schema.Codec().NativeFromTextual(value)
	valueBytes, _ := fkp.Schema.Codec().BinaryFromNative(nil, native)

	var recordValue []byte
	recordValue = append(recordValue, byte(0))
	recordValue = append(recordValue, schemaIDBytes...)
	recordValue = append(recordValue, valueBytes...)

	return recordValue
}
