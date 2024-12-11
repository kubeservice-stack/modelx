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

	"github.com/caarlos0/env/v10"

	"github.com/kubeservice-stack/common/pkg/config/ltoml"
)

var GlobalCfg Global

func init() {
	GlobalCfg.Logging = GlobalCfg.Logging.DefaultConfig()
	GlobalCfg.Metrics = GlobalCfg.Metrics.DefaultConfig()
	GlobalCfg.Discovery = GlobalCfg.Discovery.DefaultConfig()
	GlobalCfg.GinConfig = GlobalCfg.GinConfig.DefaultConfig()
	GlobalCfg.RateLimit = GlobalCfg.RateLimit.DefaultConfig()
	GlobalCfg.Temporary = GlobalCfg.Temporary.DefaultConfig()
	GlobalCfg.DBConfig = GlobalCfg.DBConfig.DefaultConfig()

}

func LoadGlobalConfig(cfgPath string) (err error) {
	// 文件配置覆盖默认值
	if err = ltoml.LoadConfig(cfgPath, cfgPath, &GlobalCfg); err != nil {
		return
	}
	// 环境变量覆盖配置的值
	if err = env.Parse(&GlobalCfg); err != nil {
		return
	}
	return
}

type Global struct {
	Logging   `toml:"logging"`
	Metrics   `toml:"metrics"`
	Discovery `toml:"discovery"`
	GinConfig `toml:"gin"`
	RateLimit `toml:"ratelimit"`
	Temporary `toml:"temporary"`
	DBConfig  `toml:"database"`
}

func (g Global) TOML() string {
	return fmt.Sprint(
		g.Logging.TOML(),
		g.Metrics.TOML(),
		g.Discovery.TOML(),
		g.GinConfig.TOML(),
		g.RateLimit.TOML(),
		g.Temporary.TOML(),
		g.DBConfig.TOML(),
	)
}
