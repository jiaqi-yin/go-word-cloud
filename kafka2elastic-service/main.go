package main

import (
	"fmt"
	feedkafkaconsumer "go-word-cloud/kafka2elastic-service/kafka_consumer"
	"go-word-cloud/shared/config"
	_ "go-word-cloud/shared/config"
	"go-word-cloud/shared/kafka"
	"go-word-cloud/shared/util"
	"net/http"
	"time"
)

func init() {
	waitForKafkaTopics()
	waitForElasticSearch()
}

func waitForKafkaTopics() {
	retry(func() error {
		kafkaAdmin, err := kafka.NewKafkaAdmin(config.AppConfig.Kafka.BootstrapServers)
		if err != nil {
			return err
		}

		isTopicCreated, err := kafkaAdmin.IsTopicCreated(config.AppConfig.Kafka.Topic)
		if err != nil {
			return err
		}

		if !isTopicCreated {
			return fmt.Errorf("Kafka topic '%s' is not ready yet", config.AppConfig.Kafka.Topic)
		}

		return nil
	})
}

func waitForElasticSearch() {
	retry(func() error {
		resp, err := http.Head(config.AppConfig.Elastic.Url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		return nil
	})
}

func retry(operation func() error) {
	retry := util.NewRetry(
		time.Second*time.Duration(config.AppConfig.Retry.SleepTime),
		time.Second*time.Duration(config.AppConfig.Retry.Timeout),
		operation,
	)
	err := retry.Do()
	if err != nil {
		panic(err)
	}
}

func main() {
	consumer := feedkafkaconsumer.NewFeedKafkaCustomer()
	defer consumer.Consumer.Close()

	consumer.Receive()
}
