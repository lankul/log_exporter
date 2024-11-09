package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"log"
)

func queryElasticsearch(index string, query string) (int, map[string]interface{}, error) {
	var queryBody map[string]interface{}
	err := json.Unmarshal([]byte(query), &queryBody)
	if err != nil {
		log.Printf("Error unmarshaling query: %s", err)
		return 0, nil, err
	}

	queryBodyWrapper := map[string]interface{}{
		"query": queryBody,
		"sort": []interface{}{
			map[string]interface{}{"@timestamp": map[string]interface{}{"order": "asc"}},
		},
		"size": 1,
	}

	queryJSON, err := json.Marshal(queryBodyWrapper)
	//log.Printf("queryBodyWrapper: %s", queryJSON)
	if err != nil {
		log.Printf("Error marshaling query body: %s", err)
		return 0, nil, err
	}

	//log.Printf("Querying Elasticsearch: index=%s, query=%s", index, string(queryJSON))

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		log.Printf("Error executing search request: %s", err)
		return 0, nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Printf("Error response from Elasticsearch: %s", res.String())
		return 0, nil, fmt.Errorf("error response: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response body: %s", err)
		return 0, nil, err
	}

	//log.Printf("Elasticsearch response: %+v", result)

	hits := result["hits"].(map[string]interface{})["hits"].([]interface{})
	count := int(result["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))
	if len(hits) > 0 {
		record := hits[0].(map[string]interface{})["_source"].(map[string]interface{})
		//log.Printf("First hit: %+v", record)
		return count, record, nil
	}

	//log.Printf("No hits found")
	return count, nil, nil
}
