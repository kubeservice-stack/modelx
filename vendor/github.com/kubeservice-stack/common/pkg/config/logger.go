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

// 日志相关配置
package config

import (
	"fmt"
	"path/filepath"
)

var defaultParentDir = "/tmp/media"

// Logging represents a logging configuration
type Logging struct {
	IsTerminal bool   `toml:"isterminal" json:"isterminal" env:"LOGGING_ISTERMINAL"` // debug模式stdout输出
	Dir        string `toml:"dir" json:"dir" env:"LOGGING_DIR"`                      // 日志目录
	Filename   string `toml:"name" json:"name" env:"LOGGING_FILENAME"`               // 日志名称
	Level      string `toml:"level" json:"level" env:"LOGGING_LEVEL"`                // 打印日志等级
	MaxSize    uint16 `toml:"maxsize" json:"maxsize" env:"LOGGING_MAXSIZE"`          // 单日志尺寸
	MaxBackups uint16 `toml:"maxbackups" json:"maxbackups" env:"LOGGING_MAXBACKUPS"` // 日志备份数
	MaxAge     uint16 `toml:"maxage" json:"maxage" env:"LOGGING_MAXAGE"`             // 留旧日志文件的最大天数
}

func (l Logging) TOML() string {
	return fmt.Sprintf(`
[logging]
  ## debug模式: stdout输出
  isterminal = %v
  ## Dir是日志文件的输出目录
  dir = "%s"
  ## Name是日志名称
  name = "%s"
  ## 日志级别
  ## error, warn, info, 或者 debug
  level = "%s"
  ## 日志文件获取之前的最大大小（以兆字节为单位）. 默认 500MB
  maxsize = %d
  ## 要保留的最大旧日志文件数
  maxbackups = %d
  ## 根据以下情况保留旧日志文件的最大天数：时间戳编码在其文件名中； 一天定义为24小时
  maxage = %d`,
		l.IsTerminal,
		l.Dir,
		l.Filename,
		l.Level,
		l.MaxSize,
		l.MaxBackups,
		l.MaxAge)
}

func (l Logging) DefaultConfig() Logging {
	l = Logging{
		Dir:        filepath.Join(defaultParentDir, "log"),
		Filename:   "media.log",
		Level:      "info",
		MaxSize:    500,
		MaxBackups: 10,
		MaxAge:     30,
		IsTerminal: false,
	}
	return l
}
