package configuration

import (
	"encoding/json"
	"fmt"
	"os"
)

type Configuration struct {
	Verbose     bool                         `json:"verbose"`
	Sources     []RepoSourceConfiguration    `json:"sources"`
	Destination RepoDestinationConfiguration `json:"destination"`
}

type RepoSourceConfiguration struct {
	Type                  string `json:"type"`
	Url                   string `json:"url"`
	ApiKey                string `json:"api-key"`
	User                  string `json:"user"`
	DestinationOwner      string `json:"dest-owner"`
	DestinationNamePrefix string `json:"dest-name-prefix"`
}

type RepoDestinationConfiguration struct {
	Url      string `json:"url"`
	ApiKey   string `json:"api-key"`
	DataPath string `json:"data-path"`
}

func Read() (*Configuration, error) {
	contents, err := os.ReadFile("config.json")
	if err != nil {
		return nil, fmt.Errorf("unable to open config file: %s", err)
	}

	var config Configuration
	if err := json.Unmarshal(contents, &config); err != nil {
		return nil, fmt.Errorf("unable to parse config file: %s", err)
	}

	return &config, nil
}
