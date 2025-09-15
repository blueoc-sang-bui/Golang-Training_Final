package search

import (
	"github.com/elastic/go-elasticsearch/v8"
)

func NewESClient(addr string) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{addr},
	}
	return elasticsearch.NewClient(cfg)
}

func IndexPost(es *elasticsearch.Client, postID string, title string, content string, tags []string) error {
	// TODO: Implement indexing logic
	return nil
}

func SearchPosts(es *elasticsearch.Client, query string) ([]map[string]interface{}, error) {
	// TODO: Implement search logic
	return nil, nil
}
