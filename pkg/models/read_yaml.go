package models

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigStruct struct {
	HostName       string `yaml:"host_name"`
	Port           string `yaml:"port"`
	UrlLifeTime    string `yaml:"url_life_time"`
	TelegramApiKey string `yaml:"tg_key"`
	DatabasebUrl   string `yaml:"db_url"`
	DatabaseApiKey string `yaml:"db_key"`
}

func ReadConfig() {
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Error read file: %v", err)
	}

	var yamlConfig ConfigStruct

	err = yaml.Unmarshal(yamlFile, &yamlConfig)
	if err != nil {
		log.Fatalf("Error in decode yaml %v", err)
	}
	Config = yamlConfig
}
