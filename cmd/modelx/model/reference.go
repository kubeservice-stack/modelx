package model

import (
	"fmt"
	"net/url"
	"strings"

	"kubegems.io/modelx/cmd/modelx/repo"
	"kubegems.io/modelx/pkg/client"
)

type Reference struct {
	Registry      string
	Repository    string
	Version       string
	Authorization string
}

func (r Reference) String() string {
	if r.Version == "" {
		return fmt.Sprintf("%s/%s", r.Registry, r.Repository)
	}
	return fmt.Sprintf("%s/%s@%s", r.Registry, r.Repository, r.Version)
}

func (r Reference) Client() *client.Client {
	return client.NewClient(r.Registry, r.Authorization)
}

func ParseReference(raw string) (Reference, error) {
	auth := ""
	if !strings.Contains(raw, "://") {
		splits := strings.SplitN(raw, ":", 2)
		details, err := repo.DefaultRepoManager.Get(splits[0])
		if err != nil {
			return Reference{}, err
		}

		auth = "Bearer " + details.Token

		if len(splits) == 2 {
			raw = details.URL + "/" + splits[1]
		} else {
			raw = details.URL
		}
	}

	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return Reference{}, fmt.Errorf("invalid reference: %s", err)
	}
	if u.Host == "" {
		return Reference{}, fmt.Errorf("invalid reference: missing host")
	}
	repository, version := "", ""
	splits := strings.SplitN(u.Path, "@", 2)
	if len(splits) != 2 || splits[1] == "" {
		version = ""
	} else {
		version = splits[1]
	}
	if sp0 := splits[0]; sp0 != "" {
		repository = sp0[1:]
	}
	ref := Reference{
		Registry:      u.Scheme + "://" + u.Host,
		Repository:    repository,
		Version:       version,
		Authorization: auth,
	}
	return ref, nil
}