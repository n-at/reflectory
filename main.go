package main

import (
	"os"
	"reflectory/configuration"
	"reflectory/destination"
	"reflectory/source"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)

	config, err := configuration.Read()
	if err != nil {
		log.Fatalf("unable to read configuration: %s", err)
	}

	if config.Verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if len(config.Sources) == 0 {
		log.Fatalf("no sources provided")
	}

	var repositories []source.Repository

	for _, sourceConfig := range config.Sources {
		var sourceRepositories []source.Repository

		switch sourceConfig.Type {
		case "github":
			g, err := source.NewGitHub(sourceConfig)
			if err != nil {
				log.Errorf("unable to init GitHub source: %s", err)
				break
			}
			sourceRepositories, err = g.Export()
			if err != nil {
				log.Errorf("unable to export GitHub repositories: %s", err)
			}

		case "gitlab":
			g, err := source.NewGitLab(sourceConfig)
			if err != nil {
				log.Errorf("unable to init GitLab source: %s", err)
				break
			}
			sourceRepositories, err = g.Export()
			if err != nil {
				log.Errorf("unable to export GitLab repositories: %s", err)
			}

		default:
			log.Warnf("unknown source type: %s", sourceConfig.Type)
		}

		repositories = append(repositories, sourceRepositories...)
	}

	if len(repositories) == 0 {
		log.Fatalf("no repositories found")
	}

	g, err := destination.New(config.Destination)
	if err != nil {
		log.Fatalf("unable to init Gitea destination: %s", err)
	}
	for _, repo := range repositories {
		if err := g.Mirror(repo); err != nil {
			log.Errorf("unable to mirror repository %s/%s: %s", repo.DestinationOwner, repo.DestinationName, err)
		}
	}

	log.Infof("DONE")
}
