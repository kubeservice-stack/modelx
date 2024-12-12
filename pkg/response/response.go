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

package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubeservice-stack/common/pkg/errno"
)

// Response is commonly used to return JSON format response.
type Response struct {
	Status  int         `json:"status,omitempty" xml:"status,omitempty"`
	Message string      `json:"message,omitempty" xml:"message,omitempty"`
	Data    interface{} `json:"data,omitempty" xml:"data,omitempty"`
}

// JSON response Json format for media-server.
func JSON(c *gin.Context, err error, data interface{}) {
	c.Header("X-Request-Id", c.GetString("requestId"))
	if err == nil {
		c.JSON(http.StatusOK, &Response{http.StatusOK, "Success", data})
	} else {
		var nerr *errno.Errno
		if errors.As(err, &nerr) {
			c.JSON(nerr.Status(), &Response{nerr.Status(), nerr.Message(), data})
		} else {
			c.JSON(http.StatusInternalServerError, &Response{
				errno.InternalServerError.Status(),
				errno.InternalServerError.Message(),
				data},
			)
		}
	}
}
