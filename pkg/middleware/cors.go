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

	"github.com/gin-gonic/gin"
)

const (
	CORS       = "CORS"
	CORSWEIGHT = 50
)

// 处理跨域请求,支持options访问.
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, OPTIONS")
		if len(c.Request.Header["Access-Control-Request-Headers"]) > 0 {
			c.Header("Access-Control-Allow-Headers", c.Request.Header["Access-Control-Request-Headers"][0])
		} else {
			c.Header("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Origin, Allow")
		}
		c.Header("Access-Control-Expose-Headers", "Content-Length, Etag")

		// 放行所有OPTIONS方法
		if method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
		}
		// 处理请求
		c.Next()
	}
}

func init() {
	Register(&Instance{Name: CORS, F: Cors, Weight: CORSWEIGHT})
}
