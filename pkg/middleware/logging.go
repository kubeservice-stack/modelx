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
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kubeservice-stack/common/pkg/logger"
)

var ginAccessLogger = logger.GetLogger(logger.HTTPModule, "access")
var ginCrashLogger = logger.GetLogger(logger.CrashModule, "crash")
var ginLogger = logger.GetLogger("pkg/middleware", "gin")

const (
	ACCESSLOG       = "ACCESSLOG"
	ACCESSLOGWEIGHT = 101
	STACKLOG        = "STACKLOG"
	STACKLOGWEIGHT  = 102
)

// access logging.
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		reqSz := computeApproximateRequestSize(c.Request)
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start) //ns
		respSz := c.Writer.Size()

		// remoteip
		remoteIP := c.RemoteIP()
		// host
		host := c.Request.Host
		if host == "" {
			host = c.Request.URL.Host
		}
		if strings.Contains(host, ":") {
			host, _, _ = net.SplitHostPort(host)
		}

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			for _, e := range c.Errors.Errors() {
				ginLogger.Error(e)
			}
		} else {
			ginAccessLogger.Info("access",
				logger.Any("status", c.Writer.Status()),
				logger.Any("method", c.Request.Method),
				logger.Any("path", path),
				logger.Any("query", rawQuery),
				logger.Any("client-ip", c.ClientIP()),
				logger.Any("remote-ip", remoteIP),
				logger.Any("user-agent", c.Request.UserAgent()),
				logger.Any("latency", latency),
				logger.Any("host", host),
				logger.Any("request-content-size", reqSz),
				logger.Any("response-content-size", respSz),
			)
		}

	}
}

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					ginLogger.Error(c.Request.URL.Path,
						logger.Any("error", err),
						logger.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				ginCrashLogger.Error("Recovery from panic",
					logger.Any("time", time.Now()),
					logger.Any("error", err),
					logger.String("request", string(httpRequest)),
					logger.String("stack", string(debug.Stack())),
				)

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func init() {
	Register(&Instance{Name: ACCESSLOG, F: Logging, Weight: ACCESSLOGWEIGHT})
	Register(&Instance{Name: STACKLOG, F: Recovery, Weight: STACKLOGWEIGHT})
}
