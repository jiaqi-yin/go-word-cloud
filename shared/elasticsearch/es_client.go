package elasticsearch

import (
	"context"
	"log"
	"time"

	"github.com/olivere/elastic/v7"
)

type EsClient struct {
	Client *elastic.Client
}

func NewEsClient(esUrl, esIndex, mapping string) *EsClient {
	ctx := context.Background()

	// Create an ES client
	client, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(esUrl),
		elastic.SetHealthcheckInterval(5*time.Second),
	)
	if err != nil {
		panic(err)
	}
	info, _, err := client.Ping(esUrl).Do(ctx)
	if err != nil {
		panic(err)
	}
	log.Printf("Connected to elasticsearch v%s\n", info.Version.Number)

	// Check if the specified elasticsearch index exists
	// If not, create a new index
	exists, err := client.IndexExists(esIndex).Do(ctx)
	if err != nil {
		panic(err)
	}
	if !exists {
		index, err := client.CreateIndex(esIndex).BodyJson(mapping).Do(ctx)
		if err != nil {
			panic(err)
		}
		if !index.Acknowledged {
			// Not acknowledged
			panic("index not acknowledged")
		}
	}

	return &EsClient{
		Client: client,
	}
}

func (client *EsClient) Index(esIndex string, id string, doc interface{}) {
	_, err := client.Client.Index().Index(esIndex).Id(id).BodyJson(doc).Do(context.TODO())
	if err != nil {
		panic(err)
	}
	log.Printf("Indexed record id: %v\n", id)
}
