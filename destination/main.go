package destination

import (
	"errors"
	"reflectory/configuration"
	"reflectory/source"
)

type Gitea struct {
	config configuration.RepoDestinationConfiguration
}

func New(config configuration.RepoDestinationConfiguration) (*Gitea, error) {
	if len(config.Url) == 0 {
		return nil, errors.New("url is not defined")
	}
	if len(config.ApiKey) == 0 {
		return nil, errors.New("api key is not defined")
	}
	return &Gitea{config: config}, nil
}

func (g *Gitea) Mirror(repos []source.Repository) error {
	if len(repos) == 0 {
		return nil
	}

	return nil //TODO
}
