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
	"context"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
)

type String string

const (
	REQUESTINFO       = "REQUESTINFO"
	REQUESTINFOWEIGHT = 120
)

func RequestInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqCxt := c.Request.Context()
		//来源请求ID
		forwardRequestID := c.Request.Header.Get("uniqID")
		reqCxt = context.WithValue(reqCxt, String("forwardRequestID"), forwardRequestID)
		//请求ID
		requestID := c.Request.Header.Get("requestID")

		if requestID == "" {
			requestID = uuid.NewV4().String()
		}

		reqCxt = context.WithValue(reqCxt, String("requestID"), requestID)
		reqCxt = context.WithValue(reqCxt, String("clientAddress"), c.Request.RemoteAddr)
		if http.LocalAddrContextKey != nil && reqCxt.Value(http.LocalAddrContextKey) != nil {
			reqCxt = context.WithValue(reqCxt, String("serverAddress"), reqCxt.Value(http.LocalAddrContextKey).(*net.TCPAddr).String())
		}
		c.Request = c.Request.WithContext(reqCxt)

		// 处理请求
		c.Next()
	}
}

func init() {
	Register(&Instance{F: RequestInfo, Weight: REQUESTINFOWEIGHT, Name: REQUESTINFO})
}
