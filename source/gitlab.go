package source

import (
	"errors"
	"reflectory/configuration"
)

type GitLab struct {
	config configuration.RepoSourceConfiguration
}

func NewGitLab(config configuration.RepoSourceConfiguration) (*GitLab, error) {
	if len(config.Url) == 0 {
		return nil, errors.New("url is not defined")
	}
	if len(config.ApiKey) == 0 {
		return nil, errors.New("api key is not defined")
	}

	return &GitLab{config: config}, nil
}

func (g *GitLab) Export() ([]Repository, error) {
	return nil, nil //TODO
}
