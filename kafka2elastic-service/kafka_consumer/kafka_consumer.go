package kafkaconsumer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"go-word-cloud/shared/config"
	"go-word-cloud/shared/domain"
	"go-word-cloud/shared/elasticsearch"
	"io/ioutil"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/riferrei/srclient"
)

type FeedKafkaCustomer struct {
	Consumer *kafka.Consumer
	SrClient *srclient.SchemaRegistryClient
	EsClient *elasticsearch.EsClient
}

func NewFeedKafkaCustomer() *FeedKafkaCustomer {
	// Create the consumer
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": config.AppConfig.Kafka.BootstrapServers,
		"group.id":          config.AppConfig.Kafka.Consumer.ConsumerGroupId,
		"auto.offset.reset": config.AppConfig.Kafka.Consumer.AutoOffsetReset,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create customer %s", err))
	}

	mapping, err := ioutil.ReadFile(config.AppConfig.Elastic.MappingFile)
	if err != nil {
		panic(fmt.Sprintf("Error reading the elasticsearch mapping file %s", err))
	}

	return &FeedKafkaCustomer{
		Consumer: c,
		SrClient: srclient.CreateSchemaRegistryClient(config.AppConfig.Kafka.SchemaRegistryUrl),
		EsClient: elasticsearch.NewEsClient(
			config.AppConfig.Elastic.Url,
			config.AppConfig.Elastic.Index,
			string(mapping),
		),
	}
}

func (fkc *FeedKafkaCustomer) Receive() {
	fkc.Consumer.SubscribeTopics([]string{config.AppConfig.Kafka.Topic}, nil)

	for {
		msg, err := fkc.Consumer.ReadMessage(-1)
		if err != nil {
			log.Printf("Error consuming the message: %v (%v)\n", err, msg)
		}

		// Deserialize the record
		feedItem := fkc.deserializeRecord(msg)
		log.Printf("Data deserialized: %v\n", feedItem.Id)

		// Send to elastic
		fkc.EsClient.Index(config.AppConfig.Elastic.Index, feedItem.Id, *feedItem)
	}
}

func (fkc *FeedKafkaCustomer) deserializeRecord(msg *kafka.Message) *domain.FeedItem {
	schema := fkc.retrieveSchema(msg)
	native, _, _ := schema.Codec().NativeFromBinary(msg.Value[5:])
	value, _ := schema.Codec().TextualFromNative(nil, native)
	data := domain.FeedItem{}
	json.Unmarshal(value, &data)
	return &data
}

func (fkc *FeedKafkaCustomer) retrieveSchema(msg *kafka.Message) *srclient.Schema {
	schemaID := binary.BigEndian.Uint32(msg.Value[1:5])
	schema, err := fkc.SrClient.GetSchema(int(schemaID))
	if err != nil {
		panic(fmt.Sprintf("Error getting the schema with id '%d' %s", schemaID, err))
	}
	return schema
}
