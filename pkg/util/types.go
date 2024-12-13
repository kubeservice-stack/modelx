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

package util

import (
	"cmp"
	"os"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"
)

const (
	AnnotationFileMode   = "filemode"
	DefaultschemaVersion = 1 // default version v1
)

const (
	BlobLocationPurposeUpload   string = "upload"
	BlobLocationPurposeDownload string = "download"
)

type BlobLocation struct {
	Provider   string     `json:"provider,omitempty"`
	Purpose    string     `json:"purpose,omitempty"` // 上传 还是 下载
	Properties Properties `json:"properties,omitempty"`
}

type Properties map[string]any

type Descriptor struct {
	Name        string        `json:"name"`
	MediaType   string        `json:"mediaType,omitempty"`
	Digest      digest.Digest `json:"digest,omitempty"`
	Size        int64         `json:"size,omitempty"`
	Mode        os.FileMode   `json:"mode,omitempty"`
	URLs        []string      `json:"urls,omitempty"`
	Modified    time.Time     `json:"modified,omitempty"`
	Annotations Annotations   `json:"annotations,omitempty"`
}

type Annotations map[string]string

func (a Annotations) String() string {
	var result []string
	for k, v := range a {
		result = append(result, k+"="+v)
	}
	return strings.Join(result, ",")
}

func SortDescriptorName(a, b Descriptor) int {
	return cmp.Compare(a.Name, b.Name)
}

type Index struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType,omitempty"`
	Manifests     []Descriptor      `json:"manifests"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

type Manifest struct {
	SchemaVersion int               `json:"schemaVersion"`
	MediaType     string            `json:"mediaType,omitempty"`
	Config        Descriptor        `json:"config"`
	Blobs         []Descriptor      `json:"blobs"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}
