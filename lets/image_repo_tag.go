package lets

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"shanhu.io/std/docker"
)

func parseRepoTag(repoTag string) (string, string) {
	repo, tag := docker.ParseImageTag(repoTag)
	if tag == "" {
		tag = "latest"
	}
	return repo, tag
}

func repoTag(repo, tag string) string {
	if tag == "" {
		return repo
	}
	return fmt.Sprintf("%s:%s", repo, tag)
}

func nameToRepoTag(name string) (string, error) {
	parts := strings.Split(name, "/")
	if len(parts) == 0 {
		return "", errors.New("empty name")
	}
	if len(parts) != 4 {
		return "", fmt.Errorf("invalid name %q", name)
	}

	domain := parts[0]
	project := parts[1]
	dockers := parts[2]
	base := parts[3]

	if dockers != "dockers" && !strings.HasSuffix(dockers, "-dockers") {
		return "", fmt.Errorf("not a docker image name: %q", name)
	}

	if domain == "shanhu.io" {
		domain = "cr.shanhu.io"
	}

	return path.Join(domain, project, base), nil
}
