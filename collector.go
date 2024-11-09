package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"strings"
)

type LogCollector struct {
	config  *Config
	metrics map[string]*prometheus.Desc
}

func (lc *LogCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range lc.metrics {
		ch <- metric
	}
}

// getNestedValue is a helper function to get nested values from a map using dot notation keys
func getNestedValue(data map[string]interface{}, key string) (interface{}, bool) {
	keys := strings.Split(key, ".")
	var value interface{} = data
	for _, k := range keys {
		if m, ok := value.(map[string]interface{}); ok {
			value = m[k]
		} else {
			return nil, false
		}
	}
	return value, true
}

func (lc *LogCollector) Collect(ch chan<- prometheus.Metric) {
	for _, metricConfig := range lc.config.Metrics {
		count, record, err := queryElasticsearch(metricConfig.Index, metricConfig.Query)
		if err != nil {
			log.Printf("Error querying Elasticsearch: %s", err)
			continue
		}

		if record != nil {
			labelNames := make([]string, 0, len(metricConfig.StaticLabels)+len(metricConfig.DynamicLabels))
			labelValues := make([]string, 0, len(metricConfig.StaticLabels)+len(metricConfig.DynamicLabels))

			// 处理静态标签
			for labelName, labelValue := range metricConfig.StaticLabels {
				labelNames = append(labelNames, labelName)
				labelValues = append(labelValues, labelValue)
			}

			// 处理动态标签
			for labelName, fieldName := range metricConfig.DynamicLabels {
				labelNames = append(labelNames, labelName)
				if nestedValue, ok := getNestedValue(record, fieldName); ok {
					labelValues = append(labelValues, fmt.Sprintf("%v", nestedValue))
				} else {
					labelValues = append(labelValues, "")
				}
			}

			desc := prometheus.NewDesc(
				metricConfig.Name,
				metricConfig.Help,
				labelNames,
				nil,
			)

			ch <- prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue,
				float64(count),
				labelValues...,
			)
		} else {
			desc := prometheus.NewDesc(
				metricConfig.Name,
				metricConfig.Help,
				nil,
				nil,
			)

			ch <- prometheus.MustNewConstMetric(
				desc,
				prometheus.GaugeValue,
				float64(count),
			)
		}
	}
}
