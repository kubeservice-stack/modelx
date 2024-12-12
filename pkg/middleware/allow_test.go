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
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAllow(t *testing.T) {
	assert := assert.New(t)
	router := gin.New()
	router.Use(Allowz())
	router.GET("/", func(c *gin.Context) {
		allow := IsAllow(c)
		c.String(http.StatusOK, strconv.FormatBool(allow))
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("false", w.Body.String())
}

func TestAllowWithHost(t *testing.T) {
	assert := assert.New(t)
	router := gin.New()
	router.Use(Allowz())
	router.GET("/metrics", func(c *gin.Context) {
		allow := IsAllow(c)
		c.String(http.StatusOK, strconv.FormatBool(allow))
	})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Host = "127.0.0.1:9445"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("true", w.Body.String())
}
