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
	"time"

	"github.com/kubeservice-stack/common/pkg/utils"
)

const (
	DefaultAppName    = "application"
	DefaultServerType = "backend"
	DefaultPort       = 9445
)

type GinConfig struct {
	App                 string         `toml:"app" json:"app" env:"GIN_APP"`                                               // application
	ServerType          string         `toml:"server_type" json:"server_type" env:"GIN_SERVERTYPE"`                        // 服务类型
	EnablePprof         bool           `toml:"enable_pprof" json:"enable_pprof" env:"GIN_ENABLEPPROF"`                     // pprof开关
	EnableHealth        bool           `toml:"enable_health" json:"enable_health" env:"GIN_ENABLEHEALTH"`                  // health开关
	EnableDebug         bool           `toml:"enable_debug" json:"enable_debug" env:"GIN_ENABLEDEBUG"`                     // debug开关
	EnableMetrics       bool           `toml:"enable_metrics" json:"enable_metrics" env:"GIN_ENABLEMETRICS"`               // metrics开关
	EnableCache         bool           `toml:"enable_cache" json:"enable_cache" env:"GIN_ENABLECACHE"`                     // 缓存开关
	EnableAuth          bool           `toml:"enable_auth" json:"enable_auth" env:"GIN_ENABLEAUTH"`                        // 签名权限开关
	EnableVerifyCommand bool           `toml:"enable_verify_command" json:"enable_verify_command" env:"GIN_VERIFYCOMMAND"` // 指令验证开关
	Port                int            `toml:"port" json:"port" env:"GIN_PORT"`                                            // port
	Trace               string         `toml:"trace" json:"trace" env:"GIN_TRACE"`                                         // Trace
	GracefulTimeout     utils.Duration `toml:"graceful_timeout" json:"graceful_timeout" env:"GIN_GRACEFULTIMEOUT"`         // GracefulTimeout
}

func (gcf GinConfig) TOML() string {
	return fmt.Sprintf(`
[gin]
  ## APP name
  app = "%s"
  ## 服务类型，支持frontend/backend, 默认backend
  server_type = "%s"
  ## 是否打开pprof
  enable_pprof = %v
  ## 是否开启Health check
  enable_health = %v
  ## 是否开启debug 模式
  enable_debug = %v
  ## 是否开启metric接口
  enable_metrics = %v
  ## 缓存开关，默认false
  enable_cache = %v
  ## 是否开启签名权限验证
  enable_auth = %v
  ## 是否开启指令权限验证
  enable_verify_command = %v
  ## 服务启动端口
  port = %d
  ## Trace
  trace = "%s"
  ## GracefulTimeout
  graceful_timeout = "%s"`,
		gcf.App,
		gcf.ServerType,
		gcf.EnablePprof,
		gcf.EnableHealth,
		gcf.EnableDebug,
		gcf.EnableMetrics,
		gcf.EnableCache,
		gcf.EnableAuth,
		gcf.EnableVerifyCommand,
		gcf.Port,
		gcf.Trace,
		gcf.GracefulTimeout.String(),
	)
}

func (gcf GinConfig) DefaultConfig() GinConfig {
	gcf = GinConfig{
		App:                 DefaultAppName,
		ServerType:          DefaultServerType,
		EnablePprof:         false,
		EnableHealth:        true,
		EnableDebug:         false,
		EnableMetrics:       true,
		EnableCache:         false,
		EnableAuth:          false,
		EnableVerifyCommand: false,
		Port:                DefaultPort,
		Trace:               "",
		GracefulTimeout:     utils.Duration(time.Second * 3),
	}
	return gcf
}

func (gcf GinConfig) ListenAddr() string {
	if gcf.Port <= 0 {
		gcf.Port = DefaultPort
	}
	return fmt.Sprintf("0.0.0.0:%d", gcf.Port)
}
