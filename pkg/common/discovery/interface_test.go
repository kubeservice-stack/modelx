/*
Copyright 2022 The KubeService-Stack Authors.

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

package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRepo(t *testing.T) {
	assert := assert.New(t)
	cluster := StartEtcdMockCluster(t)
	defer cluster.Terminate(t)
	cfg := Options{
		Endpoints: cluster.Endpoints,
	}

	factory := NewDiscoveryFactory("nobody")
	ds, err := factory.CreateDiscovery(cfg)
	assert.Nil(err)
	assert.NotNil(ds)
}

func TestEventType_String(t *testing.T) {
	assert := assert.New(t)

	assert.Equal("delete", EventTypeDelete.String())
	assert.Equal("modify", EventTypeModify.String())
	assert.Equal("all", EventTypeAll.String())
	assert.Equal("unknown", EventType(111).String())
}
