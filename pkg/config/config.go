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

package config

import (
	"time"
)

var GlobalModelxdOptions *Options = DefaultOptions()

type Options struct {
	Listen         string
	TLS            *TLSOptions
	S3             *S3Options
	Local          *LocalFSOptions
	EnableRedirect bool
	EnableMetrics  bool
	OIDC           *OIDCOptions
}

type OIDCOptions struct {
	Issuer string
}

func DefaultOptions() *Options {
	return &Options{
		Listen:         ":8080",
		TLS:            &TLSOptions{},
		S3:             NewDefaultS3Options(),
		OIDC:           &OIDCOptions{},
		Local:          NewDefaultLocalFSOptions(),
		EnableRedirect: false, // default to false
		EnableMetrics:  true,  // default to true
	}
}

type TLSOptions struct {
	CertFile string
	KeyFile  string
	CAFile   string
}

type S3Options struct {
	URL           string        `json:"url,omitempty"`
	Region        string        `json:"region,omitempty"`
	Bucket        string        `json:"bucket,omitempty"`
	AccessKey     string        `json:"accessKey,omitempty"`
	SecretKey     string        `json:"secretKey,omitempty"`
	PresignExpire time.Duration `json:"presignExpire,omitempty"`
	PathStyle     bool          `json:"pathStyle,omitempty"`
}

func NewDefaultS3Options() *S3Options {
	return &S3Options{
		Bucket:        "registry",
		URL:           "",
		AccessKey:     "",
		SecretKey:     "",
		PresignExpire: time.Hour,
		Region:        "",
		PathStyle:     true,
	}
}

type EtcdOptions struct {
	Namespace   string        `json:"namespace"`   // 命名空间
	Endpoints   []string      `json:"endpoints"`   // 连接端点
	DialTimeout time.Duration `json:"dialTimeout"` // 连接超时时间
	Prefix      string        `json:"prefix"`      // 前缀key
}

func NewDefaultEtcdOptions() *EtcdOptions {
	return &EtcdOptions{
		Namespace:   "default",
		Endpoints:   []string{},
		DialTimeout: 3 * time.Minute, // default 3min
		Prefix:      "modelx-",
	}
}

type LocalFSOptions struct {
	Basepath string
}

func NewDefaultLocalFSOptions() *LocalFSOptions {
	return &LocalFSOptions{
		Basepath: "data/registry",
	}
}
