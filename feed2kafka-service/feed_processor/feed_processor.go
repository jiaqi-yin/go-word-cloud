package feedprocessor

import (
	feedparser "go-word-cloud/feed2kafka-service/feed_parser"
	feedkafkaproducer "go-word-cloud/feed2kafka-service/kafka_producer"
	"go-word-cloud/shared/domain"
	"log"
)

type FeedProcessor struct {
	ItemChan chan *domain.FeedItem
	ExitChan chan bool
	Parser   *feedparser.FeedParser
	Producer *feedkafkaproducer.FeedKafkaProducer
}

func (fp *FeedProcessor) Parse() {
	pageNo := 1
	for {
		feedItems, err := fp.Parser.GetPagedFeedItems(pageNo)
		if err != nil {
			log.Printf("Error when fetching the RSS content: %v\n", err.Error())
			close(fp.ItemChan)
			break
		}
		for _, feedItem := range feedItems {
			fp.ItemChan <- feedItem
		}
		pageNo++
	}
}

func (fp *FeedProcessor) Produce() {
	for feedItem := range fp.ItemChan {
		fp.Producer.Send(feedItem)
	}
	fp.ExitChan <- true
	close(fp.ExitChan)
	fp.Producer.Producer.Close()
}
