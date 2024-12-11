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
	"fmt"
)

type RateLimit struct {
	QPS   int `toml:"qps" json:"qps" env:"GIN_RATELIMIT_QPS"`       // qps
	Burst int `toml:"burst" json:"burst" env:"GIN_RATELIMIT_BURST"` // 并发数
}

func (rl RateLimit) TOML() string {
	return fmt.Sprintf(`
# 访问频率限制
[ratelimit]
  ## qps
  qps = %d
  ## 并发数
  burst = %d`,
		rl.QPS, rl.Burst,
	)
}

func (rl RateLimit) DefaultConfig() RateLimit {
	return RateLimit{
		QPS:   100,
		Burst: 20,
	}
}
