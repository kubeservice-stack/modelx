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

func Test_Logging(t *testing.T) {
	assert := assert.New(t)
	router := gin.New()
	router.Use(Logging())  //logging
	router.Use(Recovery()) //logging
	router.GET("/test1", func(c *gin.Context) {
		c.String(http.StatusOK, "dongjiang")
	})

	router.GET("/test2", func(_ *gin.Context) {
		panic("bug")
	})

	req := httptest.NewRequest(http.MethodGet, "/test1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("dongjiang", w.Body.String())

	req1 := httptest.NewRequest(http.MethodGet, "/test2", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(http.StatusInternalServerError, w1.Code)
}
