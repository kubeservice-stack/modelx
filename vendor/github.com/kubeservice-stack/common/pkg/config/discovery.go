/*
Copyright 2022 The KubeService-Stack Authors.

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

package config

import (
	"encoding/json"
	"fmt"

	"github.com/kubeservice-stack/common/pkg/utils"
)

type Discovery struct {
	Namespace   string         `toml:"namespace" json:"namespace" env:"DISCOVERY_NAMESPACE"`         // 命名空间
	Endpoints   []string       `toml:"endpoints" json:"endpoints" env:"DISCOVERY_ENDPOINTS"`         // 连接端点
	DialTimeout utils.Duration `toml:"dial_timeout" json:"dial_timeout" env:"DISCOVERY_DIALTIMEOUT"` // 连接超时时间
	Prefix      string         `toml:"prefix" json:"prefix" env:"DISCOVERY_PREFIX"`                  // 前缀
}

func (ds Discovery) TOML() string {
	if len(ds.Endpoints) == 0 {
		ds.Endpoints = []string{}
	}
	endpoints, _ := json.Marshal(ds.Endpoints)
	return fmt.Sprintf(`
[discovery]
  ## etcd namespace
  namespace = "%s"
  ## etcd 集群配置
  endpoints = %s
  ## ETCD连接 timeout时间
  dial_timeout = "%s"
  ## ETCD前缀key
  prefix = "%s"`,
		ds.Namespace,
		endpoints,
		ds.DialTimeout.String(),
		ds.Prefix,
	)
}

func (ds Discovery) DefaultConfig() Discovery {
	ds = Discovery{
		Namespace: "application",
		Endpoints: []string{"http://127.0.0.1:2379"},
		Prefix:    "",
	}
	return ds
}
