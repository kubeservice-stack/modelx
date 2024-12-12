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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnnotations(t *testing.T) {
	assert := assert.New(t)
	a := Annotations{"aa": "bb"}
	assert.Equal("aa=bb", a.String())

	a = Annotations{"aa": ""}
	assert.Equal("aa=", a.String())
}

func TestSortDescriptorName(t *testing.T) {
	assert := assert.New(t)
	a := SortDescriptorName(
		Descriptor{
			Name: "aa",
		},
		Descriptor{
			Name: "bb",
		},
	)
	assert.Equal(-1, a)

	a = SortDescriptorName(
		Descriptor{
			Name: "aa",
		},
		Descriptor{
			Name: "aa",
		},
	)
	assert.Equal(0, a)

	a = SortDescriptorName(
		Descriptor{
			Name: "bb",
		},
		Descriptor{
			Name: "aa",
		},
	)
	assert.Equal(1, a)
}
