package main

import (
	feedparser "go-word-cloud/feed2kafka-service/feed_parser"
	feedprocessor "go-word-cloud/feed2kafka-service/feed_processor"
	feedkafkaproducer "go-word-cloud/feed2kafka-service/kafka_producer"
	"go-word-cloud/shared/config"
	"go-word-cloud/shared/domain"
	"go-word-cloud/shared/kafka"
	"go-word-cloud/shared/util"
	"log"
	"net/http"
	"time"
)

func init() {
	createTopics()
	waitForSchemaRegistry()
}

func createTopics() {
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
			if err := kafkaAdmin.CreateTopics(config.AppConfig.Kafka.Topic, config.AppConfig.Kafka.NumPartitions, config.AppConfig.Kafka.ReplicationFactor); err != nil {
				return err
			}
		}

		return nil
	})
}

func waitForSchemaRegistry() {
	retry(func() error {
		resp, err := http.Head(config.AppConfig.Kafka.SchemaRegistryUrl)
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
	fp := &feedprocessor.FeedProcessor{
		ItemChan: make(chan *domain.FeedItem, 10),
		ExitChan: make(chan bool, 1),
		Parser:   feedparser.NewFeedParser(config.AppConfig.Feed.Url),
		Producer: feedkafkaproducer.NewFeedKafkaProducer(),
	}

	// Goroutine to fetch the RSS feed content
	go fp.Parse()

	// Goroutine to send the feed items to Kafka
	go fp.Produce()

	for {
		_, ok := <-fp.ExitChan
		if !ok {
			break
		}
	}
	log.Println("Exit")
}
