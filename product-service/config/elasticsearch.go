package config

import (
	"log"

	elasticsearch "github.com/elastic/go-elasticsearch/v7"
)

// InitElasticsearch initializes and returns an Elasticsearch client.
func InitElasticsearch() (*elasticsearch.Client, error) {
	cfg := NewConfig() // Ambil dari config.NewConfig()

	esCfg := elasticsearch.Config{
		Addresses: []string{
			cfg.ElasticSearch.Url, // Sudah dari viper
		},
		Username: cfg.ElasticSearch.Username,
		Password: cfg.ElasticSearch.Password,
	}

	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		log.Printf("[Elasticsearch] Failed to create client: %v", err)
		return nil, err
	}

	res, err := es.Info()
	if err != nil {
		log.Printf("[Elasticsearch] Connection test failed: %v", err)
		return nil, err
	}
	defer res.Body.Close()

	log.Println("[Elasticsearch] Connected successfully")
	return es, nil
}
