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
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ALLOW       = "ALLOW"
	ALLOWWEIGHT = 90
)

var defaultAllowPath = map[string]AllowConfig{
	"favicon.ico": {
		HostAllowDomain: true,
		HostAllowIP:     true,
	},
	"metrics": {
		HostAllowDomain: false,
		HostAllowIP:     true,
	},
	"healthz": {
		HostAllowDomain: false,
		HostAllowIP:     true,
	},
}

type AllowConfig struct {
	HostAllowDomain bool
	HostAllowIP     bool
}

func Allowz() gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			ocf AllowConfig
			ok  bool
		)
		if ocf, ok = defaultAllowPath[strings.Trim(c.Request.URL.Path, "/")]; !ok {
			return
		}
		var (
			host string
			ip   net.IP
		)
		host, _, _ = net.SplitHostPort(c.Request.Host)
		ip = net.ParseIP(host)
		if (ocf.HostAllowDomain && ip == nil) ||
			(ocf.HostAllowIP && ip != nil) {
			c.Set("ALLOW", true)
		}
	}
}

func IsAllow(c *gin.Context) bool {
	var (
		opsInterface interface{}
		ops          bool
		ok           bool
	)
	if opsInterface, ok = c.Get("ALLOW"); !ok {
		return false
	}
	if ops, ok = opsInterface.(bool); !ok {
		return false
	}
	return ops
}

func init() {
	Register(&Instance{Name: ALLOW, F: Allowz, Weight: ALLOWWEIGHT})
}
