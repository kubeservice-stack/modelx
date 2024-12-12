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

	"go.uber.org/zap"

	"github.com/opencontainers/go-digest"
)

func GCBlobsAll(ctx context.Context, store RegistryInterface) error {
	globalindex, err := store.GetGlobalIndex(ctx, "")
	if err != nil {
		return err
	}
	for _, repository := range globalindex.Manifests {
		if _, err := GCBlobs(ctx, store, repository.Name); err != nil {
			return err
		}
	}
	return nil
}

func GCBlobs(ctx context.Context, store RegistryInterface, repository string) (map[digest.Digest]string, error) {

	registryLogger.Info("star blobs garbage collect", zap.Any("repository", repository))
	defer registryLogger.Info("stop blobs garbage collect")

	manifests, err := store.GetIndex(ctx, repository, "")
	if err != nil {
		return nil, err
	}
	all, err := store.ListBlobs(ctx, repository)
	if err != nil {
		return nil, err
	}

	inuse := map[digest.Digest]struct{}{}
	for _, version := range manifests.Manifests {
		manifest, err := store.GetManifest(ctx, repository, version.Name)
		if err != nil {
			return nil, err
		}
		for _, blob := range append(manifest.Blobs, manifest.Config) {
			inuse[blob.Digest] = struct{}{}
		}
	}

	toremove := map[digest.Digest]string{}
	for _, blobdigest := range all {
		if _, ok := inuse[blobdigest]; !ok {
			registryLogger.Info("mark blob unused", zap.Any("digest", blobdigest.String()))
			toremove[blobdigest] = ""
		}
	}

	for digest := range toremove {
		if err := store.DeleteBlob(ctx, repository, digest); err != nil {
			registryLogger.Error("remove unused blob", zap.Any("digest", digest.String()), zap.Error(err))
			toremove[digest] = err.Error()
			return nil, err
		} else {
			registryLogger.Error("removed unused blob", zap.Any("digest", digest.String()), zap.Error(err))
			toremove[digest] = "removed"
		}
	}
	return toremove, nil
}
