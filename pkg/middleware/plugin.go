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
	"fmt"
	"sort"

	"github.com/gin-gonic/gin"
)

var (
	ErrMiddlewareRegisterNil    = fmt.Errorf("middleware: Register adapter is nil")
	ErrMiddlewareDoubleRegister = fmt.Errorf("middleware: Register called twice for adapter: ")
)

type Instance struct {
	F      func() gin.HandlerFunc // middle instance
	Weight uint                   // plugin load weight 越大到越靠前load
	Name   string                 // plugin name
}

type Instances []*Instance

func (s Instances) Len() int           { return len(s) }
func (s Instances) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s Instances) Less(i, j int) bool { return s[i].Weight > s[j].Weight }

func (s Instances) In(n *Instance) bool {
	for _, i := range s {
		if i.Name == n.Name {
			return true
		}
	}
	return false
}

var plugins Instances

func Register(plugin *Instance) {
	if plugin == nil {
		panic(ErrMiddlewareRegisterNil.Error())
	}
	if plugins.In(plugin) {
		panic(ErrMiddlewareDoubleRegister.Error() + plugin.Name)
	}
	plugins = append(plugins, plugin)
}

// CLI 调用方法.
func ListPlugins() []string {
	var keys []string
	for _, p := range plugins {
		keys = append(keys, p.Name)
	}
	return keys
}

func AllMiddlewarePlugins() []*Instance {
	sort.Sort(plugins)
	return plugins
}
