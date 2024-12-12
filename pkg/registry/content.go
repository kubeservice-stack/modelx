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
	stderrors "errors"
	"io"
	"path"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
	"kubegems.io/modelx/pkg/util"
)

var ErrRegistryStoreNotFound = stderrors.New("not found")

type BlobLocation util.BlobLocation

var (
	BlobLocationPurposeUpload   = util.BlobLocationPurposeUpload
	BlobLocationPurposeDownload = util.BlobLocationPurposeDownload
)

type BlobContent struct {
	ContentType   string
	ContentLength int64
	Content       io.ReadCloser
}

func (s *BlobContent) Close() error {
	if s.Content != nil {
		return s.Content.Close()
	}
	return nil
}

func (s *BlobContent) Read(p []byte) (int, error) {
	return s.Content.Read(p)
}

type BlobMeta struct {
	ContentType   string
	ContentLength int64
}

type FsObjectMeta struct {
	Name         string
	Size         int64
	LastModified time.Time
	ContentType  string
}

func BlobDigestPath(repository string, d digest.Digest) string {
	if d == "" {
		d = ":"
	}
	return path.Join(repository, "blobs", d.Algorithm().String(), d.Hex())
}

func IndexPath(repository string) string {
	return path.Join(repository, RegistryIndexFileName)
}

func ManifestPath(repository string, reference string) string {
	return path.Join(repository, "manifests", reference)
}

func SplitManifestPath(in string) (string, string) {
	in = strings.TrimPrefix(in, "manifests")
	return path.Split(in)
}

func IsRegistryStoreNotNotFound(err error) bool {
	return stderrors.Is(err, ErrRegistryStoreNotFound)
}
