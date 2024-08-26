package destination

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflectory/configuration"
	"reflectory/source"
	"regexp"
	"strings"

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

		updated, err := g.mirrorUpdate(repo)
		if err != nil {
			return err
		}
		if !updated {
			return nil
		}
		if err := g.mirrorSync(repo); err != nil {
			return err
		}
	} else {
		log.Infof("repo %s/%s not exists, migrating", repo.DestinationOwner, repo.DestinationName)

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

func (g *Gitea) mirrorUpdate(repo source.Repository) (bool, error) {
	configPath := path.Join(g.config.DataPath, "git", "repositories", strings.ToLower(repo.DestinationOwner), strings.ToLower(repo.DestinationName+".git"), "config")
	contentsBytes, err := os.ReadFile(configPath)
	if err != nil {
		return false, err
	}

	contents := string(contentsBytes)

	re := regexp.MustCompile(`\[remote "origin"\][.\s]*url\s*=\s*(.*)\n`)
	matches := re.FindStringSubmatch(contents)
	if matches == nil || len(matches) < 2 {
		return false, errors.New("origin url not found")
	}
	originUrl := matches[1]

	u, err := url.Parse(originUrl)
	if err != nil {
		return false, err
	}
	u.User = url.UserPassword(repo.CloneUsername, repo.ClonePassword)
	newOriginUrl := u.String()

	if originUrl == newOriginUrl {
		return false, nil
	}

	contents = strings.Replace(contents, originUrl, newOriginUrl, 1)

	if err := os.WriteFile(configPath, []byte(contents), 0o666); err != nil {
		return false, err
	}

	return true, nil
}
