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

package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubeservice-stack/common/pkg/metrics"

	"kubegems.io/modelx/pkg/config"
)

const (
	METRICS       = "METRICS"
	METRICSWEIGHT = 80
)

// From https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go
func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}

func MetricsFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !config.GlobalModelxdOptions.EnableMetrics {
			c.Next()
		} else {
			start := time.Now()
			reqSz := computeApproximateRequestSize(c.Request)

			c.Next()

			status := strconv.Itoa(c.Writer.Status())
			elapsed := float64(time.Since(start)) / float64(time.Second)
			resSz := float64(c.Writer.Size())

			if IsAllow(c) {
				metrics.DefaultTallyScope.Scope.Tagged(map[string]string{"code": status, "method": c.Request.Method, "host": c.Request.Host, "path": c.Request.URL.Path}).Histogram("request_duration_seconds", metrics.DefaultTallyBuckets).RecordValue(elapsed)
				metrics.DefaultTallyScope.Scope.Tagged(map[string]string{"code": status, "method": c.Request.Method, "host": c.Request.Host, "path": c.Request.URL.Path}).Counter("requests_total").Inc(1)
			} else {
				metrics.DefaultTallyScope.Scope.Tagged(map[string]string{"code": status, "method": c.Request.Method, "host": c.Request.Host, "path": "uri"}).Histogram("request_duration_seconds", metrics.DefaultTallyBuckets).RecordValue(elapsed)
				metrics.DefaultTallyScope.Scope.Tagged(map[string]string{"code": status, "method": c.Request.Method, "host": c.Request.Host, "path": "uri"}).Counter("requests_total").Inc(1)
			}
			if int64(reqSz) > 0 {
				metrics.DefaultTallyScope.Scope.Tagged(map[string]string{"code": status, "method": c.Request.Method, "host": c.Request.Host}).Counter("request_size_bytes").Inc(int64(reqSz))
			}
			if int64(resSz) > 0 {
				metrics.DefaultTallyScope.Scope.Tagged(map[string]string{"code": status, "method": c.Request.Method, "host": c.Request.Host}).Counter("response_size_bytes").Inc(int64(resSz))
			}

			metrics.DefaultTallyScope.Scope.Tagged(map[string]string{"code": status}).Counter("handler_requests_total").Inc(1)
		}
	}
}

func init() {
	Register(&Instance{Name: METRICS, F: MetricsFunc, Weight: METRICSWEIGHT})
}
