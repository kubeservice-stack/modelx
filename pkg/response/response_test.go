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
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kubeservice-stack/common/pkg/errno"
	"github.com/stretchr/testify/assert"
)

func TestResponse(t *testing.T) {
	assert := assert.New(t)
	router := gin.New()
	router.GET("/aaa", func(c *gin.Context) {
		JSON(c, errno.BadRequest, map[string]string{"bb": "aa"})
	})
	router.GET("/bbb", func(c *gin.Context) {
		JSON(c, nil, map[string]string{"bb": "bb"})
	})

	router.GET("/ccc", func(c *gin.Context) {
		JSON(c, errors.New("aaa"), map[string]string{"bb": "cc"})
	})

	req := httptest.NewRequest(http.MethodGet, "/aaa", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(http.StatusBadRequest, w.Code)
	assert.Len(w.Body.String(), 56)

	req1 := httptest.NewRequest(http.MethodGet, "/bbb", nil)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(http.StatusOK, w1.Code)
	assert.Len(w1.Body.String(), 53)

	req2 := httptest.NewRequest(http.MethodGet, "/ccc", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(http.StatusInternalServerError, w2.Code)
	assert.Len(w2.Body.String(), 65)
}
