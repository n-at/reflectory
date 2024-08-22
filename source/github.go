package source

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflectory/configuration"

	log "github.com/sirupsen/logrus"
)

type GitHub struct {
	config configuration.RepoSourceConfiguration
}

type githubRepository struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Url      string `json:"clone_url"`
}

func NewGitHub(config configuration.RepoSourceConfiguration) (*GitHub, error) {
	if len(config.Url) == 0 {
		return nil, errors.New("url is not defined")
	}
	if len(config.User) == 0 {
		return nil, errors.New("user is not defined")
	}
	if len(config.ApiKey) == 0 {
		return nil, errors.New("api key is not defined")
	}

	return &GitHub{config: config}, nil
}

func (g *GitHub) Export() ([]Repository, error) {
	var repositories []Repository
	page := 1

	for {
		repos, err := g.getRepositories(page)
		if err != nil {
			return nil, err
		}
		if len(repos) == 0 {
			break
		}
		for _, r := range repos {
			repositories = append(repositories, Repository{
				Name:             r.Name,
				CloneUrl:         r.Url,
				CloneUsername:    g.config.User,
				ClonePassword:    g.config.ApiKey,
				DestinationOwner: g.config.DestinationOwner,
				DestinationName:  r.Name,
			})
		}
		page++
	}

	return repositories, nil
}

func (g *GitHub) getRepositories(page int) ([]githubRepository, error) {
	url := fmt.Sprintf("%s/user/repos?visibility=all&affiliation=owner&page=%d&per_page=100", g.config.Url, page)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.config.ApiKey))
	request.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	log.Debugf("Querying GitHub page %d...", page)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var repos []githubRepository
	if err := json.Unmarshal(body, &repos); err != nil {
		return nil, err
	}

	log.Debugf("Found GitHub repos: %d", len(repos))

	return repos, nil
}
