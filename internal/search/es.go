package search

import (
	"os"
	"github.com/elastic/go-elasticsearch/v8"
)

var ESClient *elasticsearch.Client

func InitES() {
	cfg := elasticsearch.Config{
		Addresses: []string{os.Getenv("ES_ADDR")},
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	ESClient = client
}

func GetESClient() *elasticsearch.Client {
	return ESClient
}
