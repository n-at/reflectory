package destination

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflectory/configuration"
	"reflectory/source"

	log "github.com/sirupsen/logrus"
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

func (g *Gitea) Mirror(repo source.Repository) error {
	exists, err := g.exists(repo)
	if err != nil {
		return err
	}

	if exists {
		log.Infof("repo %s/%s exists, syncing", repo.DestinationOwner, repo.DestinationName)

		//TODO update url
		log.Debugf("TODO update repo url")

		if err := g.mirrorSync(repo); err != nil {
			return err
		}
	} else {
		log.Infof("repo %s/%s not exist, migrating", repo.DestinationOwner, repo.DestinationName)

		if err := g.migrate(repo); err != nil {
			return err
		}
	}

	return nil
}

func (g *Gitea) exists(repo source.Repository) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s", g.config.Url, repo.DestinationOwner, repo.DestinationName)

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	request.Header.Add("Authorization", fmt.Sprintf("token %s", g.config.ApiKey))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false, err
	}

	defer response.Body.Close()

	return response.StatusCode == 200, nil
}

func (g *Gitea) migrate(repo source.Repository) error {
	url := fmt.Sprintf("%s/api/v1/repos/migrate", g.config.Url)

	body := make(map[string]any)
	body["clone_addr"] = repo.CloneUrl
	body["auth_username"] = repo.CloneUsername
	body["auth_password"] = repo.ClonePassword
	body["repo_owner"] = repo.DestinationOwner
	body["repo_name"] = repo.DestinationName
	body["service"] = "git"
	body["mirror"] = true
	body["private"] = true
	body["issues"] = true
	body["labels"] = true
	body["milestones"] = true
	body["releases"] = true
	body["wiki"] = true
	body["pull_requests"] = true

	bodyStr, err := json.Marshal(body)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyStr))
	if err != nil {
		return err
	}
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Authorization", fmt.Sprintf("token %s", g.config.ApiKey))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode == 201 {
		return nil
	}
	if response.StatusCode == 409 {
		return fmt.Errorf("repository with the same name already exists")
	}

	errorText, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	errorBody := make(map[string]string)
	if err := json.Unmarshal(errorText, &errorBody); err != nil {
		return err
	}

	message, ok := errorBody["message"]
	if ok {
		return errors.New(message)
	} else {
		return errors.New("unknown error")
	}
}

func (g *Gitea) mirrorSync(repo source.Repository) error {
	url := fmt.Sprintf("%s/api/v1/repos/%s/%s/mirror-sync", g.config.Url, repo.DestinationOwner, repo.DestinationName)

	request, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", fmt.Sprintf("token %s", g.config.ApiKey))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	return nil
}
