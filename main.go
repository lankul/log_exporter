package main

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

var es *elasticsearch.Client

func initElasticsearch(config *Config) {
	cfg := elasticsearch.Config{
		Addresses: config.Elasticsearch.Addresses,
		Username:  config.Elasticsearch.Username,
		Password:  config.Elasticsearch.Password,
	}
	var err error
	es, err = elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	//info, err := es.Info()
	//if err != nil {
	//	log.Fatalf("Error getting Elasticsearch info: %s", err)
	//}
	//defer info.Body.Close()
	//
	//log.Printf("Elasticsearch connection established: %s", info)
}

func main() {
	config, err := loadConfig("./config")
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}

	initElasticsearch(config)

	metrics := make(map[string]*prometheus.Desc)
	for _, metricConfig := range config.Metrics {
		labelNames := make([]string, 0, len(metricConfig.StaticLabels)+len(metricConfig.DynamicLabels))
		for labelName := range metricConfig.StaticLabels {
			labelNames = append(labelNames, labelName)
		}
		for labelName := range metricConfig.DynamicLabels {
			labelNames = append(labelNames, labelName)
		}

		metrics[metricConfig.Name] = prometheus.NewDesc(
			metricConfig.Name,
			metricConfig.Help,
			labelNames,
			nil,
		)
	}

	logCollector := &LogCollector{
		config:  config,
		metrics: metrics,
	}

	prometheus.MustRegister(logCollector)

	http.Handle("/metrics", promhttp.Handler())
	log.Println("Beginning to serve on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
