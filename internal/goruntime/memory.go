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

package goruntime

import (
	"fmt"
	"runtime/debug"

	"github.com/KimMachineGun/automemlimit/memlimit"
	"github.com/kubeservice-stack/common/pkg/logger"
	"go.uber.org/zap"
)

func SetMemLimit(logger *logger.Logger, memlimitRatio float64) {
	if memlimitRatio >= 1.0 {
		memlimitRatio = 1.0
	} else if memlimitRatio <= 0.0 {
		memlimitRatio = 0.0
	}

	// the memlimitRatio argument to 0, effectively disabling auto memory limit for all users.
	if memlimitRatio == 0.0 {
		return
	}

	if _, err := memlimit.SetGoMemLimitWithOpts(
		memlimit.WithRatio(memlimitRatio),
		memlimit.WithProvider(
			memlimit.ApplyFallback(
				memlimit.FromCgroup,
				memlimit.FromSystem,
			),
		),
	); err != nil {
		logger.Warn("Failed to set GOMEMLIMIT automatically", zap.String("component", "automemlimit"), zap.Any("err", err))
	}

	logger.Info(fmt.Sprintf("GOMEMLIMIT set to %d", debug.SetMemoryLimit(-1)))
}
