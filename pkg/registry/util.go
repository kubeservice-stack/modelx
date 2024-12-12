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

package registry

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func StringDeref(ptr *string, def string) string {
	if ptr != nil {
		return *ptr
	}
	return def
}

func TimeDeref(ptr *time.Time, def time.Time) time.Time {
	if ptr != nil {
		return *ptr
	}
	return def
}

// MaxBytesReadHandler returns a Handler that runs h with its ResponseWriter and Request.Body wrapped by a MaxBytesReader.
func MaxBytesReadHandler(h gin.HandlerFunc, n int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用http.MaxBytesReader包装请求体，设置最大字节数限制
		reader := http.MaxBytesReader(c.Writer, c.Request.Body, n)
		c.Request.Body = reader
		defer func() {
			// 处理请求结束后关闭reader，防止资源泄漏
			_ = reader.Close()
		}()

		// 继续执行原本的HandlerFunc逻辑，这里假设传入的HandlerFunc是mainHandler
		h(c)
	}
}
