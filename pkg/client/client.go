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

package client

import (
	"context"

	"kubegems.io/modelx/pkg/util"
)

type Client struct {
	Remote    *RegistryClient
	Extension ExtensionInterface
}

func NewClient(registry string, auth string) *Client {
	return &Client{
		Remote:    NewRegistryClient(registry, auth),
		Extension: NewDelegateExtension(),
	}
}

func (c *Client) Ping(ctx context.Context) error {
	if _, err := c.Remote.GetGlobalIndex(ctx, ""); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetManifest(ctx context.Context, repo, version string) (*util.Manifest, error) {
	return c.Remote.GetManifest(ctx, repo, version)
}

func (c *Client) PutManifest(ctx context.Context, repo, version string, manifest util.Manifest) error {
	return c.Remote.PutManifest(ctx, repo, version, manifest)
}

func (c *Client) CopyBlobs(ctx context.Context, repoTo, repoFrom, versionTo, versionFrom string) error {
	return c.Remote.CopyBlobs(ctx, repoTo, repoFrom, versionTo, versionFrom)
}

func (c *Client) GetIndex(ctx context.Context, repo string, search string) (*util.Index, error) {
	return c.Remote.GetIndex(ctx, repo, search)
}

func (c *Client) GetGlobalIndex(ctx context.Context, search string) (*util.Index, error) {
	return c.Remote.GetGlobalIndex(ctx, search)
}
