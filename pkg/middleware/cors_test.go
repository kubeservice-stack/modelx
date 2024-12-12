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
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_CORS(t *testing.T) {
	assert := assert.New(t)
	router := gin.New()
	router.Use(Cors())
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "dongjiang")
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal(http.Header{
		"Access-Control-Allow-Headers":  []string{"X-Requested-With, Content-Type, Origin, Allow"},
		"Access-Control-Allow-Methods":  []string{"GET, HEAD, POST, PUT, OPTIONS"},
		"Access-Control-Allow-Origin":   []string{"*"},
		"Access-Control-Expose-Headers": []string{"Content-Length, Etag"},
		"Content-Type":                  []string{"text/plain; charset=utf-8"},
	}, w.Header())
	assert.Equal("dongjiang", w.Body.String())
}
