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

package favicon

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"kubegems.io/modelx/pkg/routers"
)

// @BasePath /

// Favicon godoc
// @Summary Favicon
// @Schemes
// @Description Favicon
// @Tags favicon
// @Accept json
// @Produce json
// @Success 200 {string} Favicon
// @Router /favicon.ico [get]
func Favicon(c *gin.Context) {
	c.JSON(http.StatusOK, "")
}

func init() {
	router.Register("favicon", "/", "favicon.ico", http.MethodGet, Favicon)
}
