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
	"strings"
	"time"
)

// Metrics represents a logging configuration
type Metrics struct {
	FlushInterval          time.Duration     `toml:"flush_interval" json:"flush_interval" env:"METRICS_FLUSHINTERVAL"`                              // flush时间周期
	EnableGoRuntimeMetrics bool              `toml:"enable_goruntime_metrics" json:"enable_goruntime_metrics" env:"METRICS_ENABLEGORUNTIMEMETRICS"` // 是否收集goroutine相关信息
	MetricsPrefix          string            `toml:"metrics_prefix" json:"metrics_prefix" env:"METRICS_METRICSPREFIX"`                              // 自定义metric指标前缀
	MetricsTags            map[string]string `toml:"metrics_tags" json:"metrics_tags" env:"METRICS_METRICSTAGS"`                                    // 自定义metric自动填充kv数据
}

func (l Metrics) TOML() string {
	var b []byte
	b, err := json.Marshal(l.MetricsTags)
	if err != nil {
		b = []byte("")
	}
	return fmt.Sprintf(`
[metrics]
  ## flush时间周期, 默认是5秒
  flush_interval = %d
  ## 是否收集goroutine相关信息, 默认 开启为true
  enable_goruntime_metrics = %v
  ## metrics_prefix, 默认前缀 application_server
  metrics_prefix = "%s"
  ## 自定义metric自动填充kv数据, 默认为{}
  metrics_tags = '%s'`,
		l.FlushInterval,
		l.EnableGoRuntimeMetrics,
		l.MetricsPrefix,
		string(b))
}

func (l Metrics) DefaultConfig() Metrics {
	r := strings.NewReplacer(".", "_", "-", "_") // prometheus不支持「-」和「.」
	l = Metrics{
		FlushInterval:          5,
		EnableGoRuntimeMetrics: true,
		MetricsPrefix:          r.Replace(DefaultAppName) + "_server",
		MetricsTags:            map[string]string{},
	}
	return l
}
