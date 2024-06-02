package source

import (
	"errors"
	"reflectory/configuration"
)

type GitHub struct {
	config configuration.RepoSourceConfiguration
}

func NewGitHub(config configuration.RepoSourceConfiguration) (*GitHub, error) {
	if len(config.Url) == 0 {
		return nil, errors.New("url is not defined")
	}
	if len(config.ApiKey) == 0 {
		return nil, errors.New("api key is not defined")
	}

	return &GitHub{config: config}, nil
}

func (g *GitHub) Export() ([]Repository, error) {
	return nil, nil //TODO
}
