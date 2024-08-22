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

type GitLab struct {
	config configuration.RepoSourceConfiguration
}

type project struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"http_url_to_repo"`
}

func NewGitLab(config configuration.RepoSourceConfiguration) (*GitLab, error) {
	if len(config.Url) == 0 {
		return nil, errors.New("url is not defined")
	}
	if len(config.User) == 0 {
		return nil, errors.New("user is not defined")
	}
	if len(config.ApiKey) == 0 {
		return nil, errors.New("api key is not defined")
	}

	return &GitLab{config: config}, nil
}

func (g *GitLab) Export() ([]Repository, error) {
	var repos []Repository
	page := 1

	for {
		projects, err := g.getProjects(page)
		if err != nil {
			return nil, err
		}
		if len(projects) == 0 {
			break
		}
		for _, project := range projects {
			repos = append(repos, Repository{
				Name:             project.Name,
				CloneUrl:         project.Url,
				CloneUsername:    g.config.User,
				ClonePassword:    g.config.ApiKey,
				DestinationOwner: g.config.DestinationOwner,
				DestinationName:  project.Name,
			})
		}
		page++
	}

	return repos, nil
}

func (g *GitLab) getProjects(page int) ([]project, error) {
	url := fmt.Sprintf("%s/api/v4/projects?private_token=%s&page=%d&per_page=100&owned=1", g.config.Url, g.config.ApiKey, page)

	log.Debugf("Querying GitLab page %d...", page)

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var projects []project
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, err
	}

	log.Debugf("Found GitLab projects: %d", len(projects))

	return projects, nil
}
