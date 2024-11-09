package main

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"path/filepath"
)

// MetricConfig defines the structure of each metric in the config file
type MetricConfig struct {
	Name          string            `yaml:"name"`
	Help          string            `yaml:"help"`
	Index         string            `yaml:"index"`
	Query         string            `yaml:"query"`
	StaticLabels  map[string]string `yaml:"static_labels"`
	DynamicLabels map[string]string `yaml:"dynamic_labels"`
}

// Config defines the structure of the config file
type Config struct {
	Elasticsearch struct {
		Addresses []string `yaml:"addresses"`
		Username  string   `yaml:"username"`
		Password  string   `yaml:"password"`
	} `yaml:"elasticsearch"`
	Metrics []MetricConfig `yaml:"metrics"`
}

func loadConfig(configDir string) (*Config, error) {
	config := &Config{}

	err := filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".yaml" {
			log.Printf("Loading config file: %s", path)
			fileContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			var fileConfig Config
			if err := yaml.Unmarshal(fileContent, &fileConfig); err != nil {
				return err
			}

			// Merge Elasticsearch config (use the first one found)
			if config.Elasticsearch.Addresses == nil && config.Elasticsearch.Username == "" && config.Elasticsearch.Password == "" {
				config.Elasticsearch = fileConfig.Elasticsearch
			}

			// Merge metrics
			config.Metrics = append(config.Metrics, fileConfig.Metrics...)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return config, nil
}
