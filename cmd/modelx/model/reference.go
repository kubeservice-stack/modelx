/*
Copyright 2024 The KubeService-Stack Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package model

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"kubegems.io/modelx/cmd/modelx/repo"
	"kubegems.io/modelx/pkg/client"
)

const ModelxAuthEnv = "MODELX_AUTH"

type Reference struct {
	Registry      string
	Repository    string
	Version       string
	Authorization string
}

func (r Reference) Name() string {
	re := strings.Split(strings.Trim(r.Repository, " "), "/")
	if len(re) == 1 {
		return re[0]
	} else if len(re) == 2 {
		return re[1]
	}

	return ""
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
	auth := os.Getenv(ModelxAuthEnv)
	if !strings.Contains(raw, "://") {
		splits := strings.SplitN(raw, repo.SplitorRepo, 2)
		details, err := repo.DefaultRepoManager.Get(splits[0])
		if err != nil {
			return Reference{}, err
		}
		if auth == "" {
			auth = "Bearer " + details.Token
		}
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
	if token := u.Query().Get("token"); token != "" {
		auth = "Bearer " + token
	}
	repository, version := "", ""
	splits := strings.SplitN(u.Path, repo.SplitorVersion, 2)
	if len(splits) != 2 || splits[1] == "" {
		version = ""
	} else {
		version = splits[1]
	}
	if sp0 := splits[0]; sp0 != "" {
		repository = sp0[1:]
	}

	if repository != "" && !strings.Contains(repository, "/") {
		repository = "library/" + repository
	}

	ref := Reference{
		Registry:      u.Scheme + "://" + u.Host,
		Repository:    repository,
		Version:       version,
		Authorization: auth,
	}
	return ref, nil
}
