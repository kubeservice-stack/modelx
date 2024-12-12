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

package router

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kubeservice-stack/common/pkg/logger"
	"kubegems.io/modelx/pkg/middleware"
)

var log = logger.GetLogger("pkg/router", "router")

// handername - hander func.
var handlerAdapter = make(map[string]HandlerInfo)

func Register(name, group, path, method string, h gin.HandlerFunc) {
	if handlerAdapter == nil {
		panic("gin.Handler: Register adapter is nil")
	}
	info := HandlerInfo{
		Name:   name,
		Group:  group,
		Path:   path,
		Method: method,
	}
	if _, ok := handlerAdapter[info.String()]; ok {
		panic("gin.Handler: Register called twice for adapter :" + name)
	}
	info.H = h
	handlerAdapter[info.String()] = info
}

func FullRegisters() map[string]HandlerInfo {
	return handlerAdapter
}

// 配置信息.
type HandlerInfo struct {
	Name   string // handle name
	Group  string // default group "/"
	Path   string // domain path
	Method string // http.Method
	H      gin.HandlerFunc
}

func (info HandlerInfo) String() string {
	return fmt.Sprintf("%s:%s:%s:%s", info.Name, info.Group, info.Path, info.Method)
}

// Router 路由规则.
func Router(r *gin.Engine) {
	for _, mid := range middleware.AllMiddlewarePlugins() {
		log.Info("use gin middleware", logger.String("name", mid.Name))
		r.Use(mid.F())
	}

	for _, info := range handlerAdapter {
		v := r.Group(info.Group)
		v.Handle(info.Method, info.Path, info.H)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "404 Not Found"})
	})
}
