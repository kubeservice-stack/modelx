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

package metrics

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/kubeservice-stack/common/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/uber-go/tally"
	promreporter "github.com/uber-go/tally/prometheus"
)

var onceEnable sync.Once

var (
	ErrMetricsInitRegistryError = fmt.Errorf("metrics: can not init metrics registry")
	ErrMetricsDoNotExist        = fmt.Errorf("metrics: metrics do not exists")
	ErrMetricsIsDuplicated      = fmt.Errorf("metrics: metric is duplicated")
)

// 默认Registerer
var metricsRegisterer = prometheus.DefaultRegisterer

// 默认prometheus.DefaultRegisterer, 没有特殊需要，不用使用NewRegisterer
func DefaultRegistry() prometheus.Registerer {
	return metricsRegisterer
}

// metrics 发布器
type TallyScope struct {
	Reporter promreporter.Reporter
	Scope    tally.Scope
	Closer   io.Closer
}

// TODO 默认metrics Scope
var DefaultTallyScope *TallyScope

// 创建
func NewTallyScope(cfg *config.Metrics) *TallyScope {
	if cfg.EnableGoRuntimeMetrics {
		onceEnable.Do(func() {
			DefaultRegistry().Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
			DefaultRegistry().Register(collectors.NewGoCollector())
		})
	} else {
		onceEnable.Do(func() {
			DefaultRegistry().Unregister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
			DefaultRegistry().Unregister(collectors.NewGoCollector())
		})
	}

	r := promreporter.NewReporter(promreporter.Options{})
	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Prefix:         cfg.MetricsPrefix, // cfg.MetricsPrefix
		Tags:           cfg.MetricsTags,   // cfg.MetricsTags
		CachedReporter: r,
		Separator:      promreporter.DefaultSeparator,
	}, cfg.FlushInterval*time.Second) // cfg.FlushInterval * time.Second

	return &TallyScope{Scope: scope, Closer: closer, Reporter: r}
}

// 进程失败销毁
func (t *TallyScope) Destroy() error {
	return t.Closer.Close()
}

func init() {
	DefaultTallyScope = NewTallyScope(&config.GlobalCfg.Metrics) // 默认填充 defaultTallyScope
}
