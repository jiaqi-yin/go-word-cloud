package feedparser

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"go-word-cloud/shared/domain"

	"github.com/mmcdole/gofeed"
	"golang.org/x/net/html"
)

type FeedParser struct {
	Parser  *gofeed.Parser
	FeedUrl string
}

func (fp *FeedParser) GetPagedFeedItems(pageNo int) ([]*domain.FeedItem, error) {
	var feedItems []*domain.FeedItem

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	feed, err := fp.Parser.ParseURLWithContext(fmt.Sprintf("%s?paged=%v", fp.FeedUrl, pageNo), ctx)
	if err != nil {
		return nil, err
	}

	log.Printf("Got feed items from Page %v\n", pageNo)
	for _, item := range feed.Items {
		description := item.Description
		doc, err := html.Parse(strings.NewReader(description))
		if err != nil {
			log.Printf("Error when parsing item description: %v", err.Error())
		} else {
			description = parseText(doc)
		}
		feedItem := &domain.FeedItem{
			Id:          item.GUID,
			Title:       item.Title,
			Description: description,
			Categories:  item.Categories,
		}
		feedItems = append(feedItems, feedItem)
	}

	return feedItems, nil
}

func parseText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var ret string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		ret += parseText(c) + " "
	}
	return strings.Join(strings.Fields(ret), " ")
}

func NewFeedParser(feedUrl string) *FeedParser {
	return &FeedParser{
		Parser:  gofeed.NewParser(),
		FeedUrl: feedUrl,
	}
}
