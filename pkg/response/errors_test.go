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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorInfo(t *testing.T) {
	assert := assert.New(t)
	e := ErrorInfo{
		HttpStatus: 200,
		Code:       ErrCode("200"),
		Message:    "ok",
		Detail:     "this is ok",
	}

	assert.Equal("200: ok", e.Error())
}

func TestNewUnauthorizedError(t *testing.T) {
	assert := assert.New(t)
	e := NewUnauthorizedError("this is error")
	assert.Equal("UNAUTHORIZED: this is error", e.Error())
}
