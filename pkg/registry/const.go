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
	logging "github.com/kubeservice-stack/common/pkg/logger"
)

var registryLogger = logging.GetLogger("pkg/registry", "fs")

const (
	RegistryIndexFileName = "index.json"

	DefaultFileMode = 0o644
	DefaultDirMode  = 0o755

	MediaTypeModelIndexJson = "application/vnd.modelx.model.index.v1.json"

	DefaultMaxBytesRead = int64(1 << 20) // 1MB
)

const (
	NameRegexp      = `[a-zA-Z0-9]+(?:[._-][a-zA-Z0-9]+)*/(?:[a-zA-Z0-9]+(?:[._-][a-zA-Z0-9]+)*)`
	ReferenceRegexp = `[a-zA-Z0-9_][a-zA-Z0-9._-]{0,127}`
	DigestRegexp    = `[A-Za-z][A-Za-z0-9]*(?:[-_+.][A-Za-z][A-Za-z0-9]*)*[:][[:xdigit:]]{32,}`
)
