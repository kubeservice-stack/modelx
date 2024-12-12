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

package registry

import (
	"context"

	"github.com/opencontainers/go-digest"

	"kubegems.io/modelx/pkg/util"
)

type FSProvider interface {
	Put(ctx context.Context, path string, content BlobContent) error
	Get(ctx context.Context, path string) (*BlobContent, error)
	Stat(ctx context.Context, path string) (FsObjectMeta, error)
	Remove(ctx context.Context, path string, recursive bool) error
	Exists(ctx context.Context, path string) (bool, error)
	List(ctx context.Context, path string, recursive bool) ([]FsObjectMeta, error)
}

type RegistryInterface interface {
	GetGlobalIndex(ctx context.Context, search string) (util.Index, error)

	GetIndex(ctx context.Context, repository string, search string) (util.Index, error)
	RemoveIndex(ctx context.Context, repository string) error

	ExistsManifest(ctx context.Context, repository string, reference string) (bool, error)
	GetManifest(ctx context.Context, repository string, reference string) (*util.Manifest, error)
	PutManifest(ctx context.Context, repository string, reference string, contentType string, manifest util.Manifest) error
	DeleteManifest(ctx context.Context, repository string, reference string) error

	ListBlobs(ctx context.Context, repository string) ([]digest.Digest, error)
	GetBlob(ctx context.Context, repository string, digest digest.Digest) (*BlobContent, error)
	DeleteBlob(ctx context.Context, repository string, digest digest.Digest) error
	PutBlob(ctx context.Context, repository string, digest digest.Digest, content BlobContent) error
	ExistsBlob(ctx context.Context, repository string, digest digest.Digest) (bool, error)
	GetBlobMeta(ctx context.Context, repository string, digest digest.Digest) (BlobMeta, error)

	GetBlobLocation(ctx context.Context, repository string, digest digest.Digest,
		purpose string, properties map[string]string) (*BlobLocation, error)
}
