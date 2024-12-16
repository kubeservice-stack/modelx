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
	logging "github.com/kubeservice-stack/common/pkg/logger"
	"github.com/opencontainers/go-digest"

	"kubegems.io/modelx/pkg/version"
)

var (
	// global register extensions
	GlobalExtensions = map[string]ExtensionInterface{}

	extensionLogger = logging.GetLogger("pkg/client", "extension")

	UserAgent = "modelx/" + version.Get().GitVersion

	EmptyFileDigiest = digest.Canonical.FromBytes(nil)
)

const (
	MediaTypeModelIndexJson      = "application/vnd.modelx.model.index.v1.json"
	MediaTypeModelManifestJson   = "application/vnd.modelx.model.manifest.v1.json"
	MediaTypeModelConfigYaml     = "application/vnd.modelx.model.config.v1.yaml"
	MediaTypeModelFile           = "application/vnd.modelx.model.file.v1"
	MediaTypeModelDirectoryTarGz = "application/vnd.modelx.model.directory.v1.tar+gz"

	// default retry count
	DefaultPullPushConcurrency = 5

	ModelConfigFileName = "modelx.yaml"
	ReadmeFileName      = "README.md"
	ModelCacheDir       = ".modelx"
)
