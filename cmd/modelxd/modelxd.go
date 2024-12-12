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

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	logging "github.com/kubeservice-stack/common/pkg/logger"
	"github.com/oklog/run"
	"github.com/spf13/cobra"

	"kubegems.io/modelx/internal/goruntime"
	"kubegems.io/modelx/pkg/config"
	_ "kubegems.io/modelx/pkg/favicon"
	_ "kubegems.io/modelx/pkg/health"
	_ "kubegems.io/modelx/pkg/metrics"
	"kubegems.io/modelx/pkg/model"
	"kubegems.io/modelx/pkg/registry"
	"kubegems.io/modelx/pkg/routers"
	"kubegems.io/modelx/pkg/version"
)

const (
	ErrExitCode          = 1
	ServerName           = "modelxd"
	DefaultMemlimitRatio = 0.85
)

var (
	mainLogger = logging.GetLogger("cmd", ServerName)
	printVer   bool
	printShort bool
)

func main() {
	if err := NewRegistryCmd().Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(ErrExitCode)
	}
}

func NewRegistryCmd() *cobra.Command {
	config.GlobalModelxdOptions = config.DefaultOptions()
	cmd := &cobra.Command{
		Use:     "modelxd",
		Short:   "modelxd",
		Version: version.Get().String(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var g run.Group
			term := make(chan os.Signal, 1)
			signal.Notify(term, os.Interrupt, syscall.SIGTERM)
			g.Add(func() error {
				select {
				case <-term:
					mainLogger.Info("Received SIGTERM, exiting gracefully...")
				case <-ctx.Done():
				}

				return nil
			}, func(error) {})

			return Run(ctx, g, config.GlobalModelxdOptions)
		},
	}

	flags := cmd.Flags()
	flags.StringVar(&config.GlobalModelxdOptions.Listen, "listen", config.GlobalModelxdOptions.Listen, "listen address")
	flags.StringVar(&config.GlobalModelxdOptions.TLS.CAFile, "tls-ca", config.GlobalModelxdOptions.TLS.CAFile, "tls ca file")
	flags.StringVar(&config.GlobalModelxdOptions.TLS.CertFile, "tls-cert", config.GlobalModelxdOptions.TLS.CertFile, "tls cert file")
	flags.StringVar(&config.GlobalModelxdOptions.TLS.KeyFile, "tls-key", config.GlobalModelxdOptions.TLS.KeyFile, "tls key file")
	flags.StringVar(&config.GlobalModelxdOptions.S3.Buket, "s3-bucket", config.GlobalModelxdOptions.S3.Buket, "s3 bucket")
	flags.StringVar(&config.GlobalModelxdOptions.S3.URL, "s3-url", config.GlobalModelxdOptions.S3.URL, "s3 url")
	flags.StringVar(&config.GlobalModelxdOptions.S3.AccessKey, "s3-access-key", config.GlobalModelxdOptions.S3.AccessKey, "s3 access key")
	flags.StringVar(&config.GlobalModelxdOptions.S3.SecretKey, "s3-secret-key", config.GlobalModelxdOptions.S3.SecretKey, "s3 secret key")
	flags.DurationVar(&config.GlobalModelxdOptions.S3.PresignExpire, "s3-presign-expire", config.GlobalModelxdOptions.S3.PresignExpire, "s3 presign expire")
	flags.StringVar(&config.GlobalModelxdOptions.S3.Region, "s3-region", config.GlobalModelxdOptions.S3.Region, "s3 region")
	flags.StringVar(&config.GlobalModelxdOptions.OIDC.Issuer, "oidc-issuer", config.GlobalModelxdOptions.OIDC.Issuer, "oidc issuer")
	flags.BoolVar(&config.GlobalModelxdOptions.EnableRedirect, "enable-redirect", config.GlobalModelxdOptions.EnableRedirect, "enable blob storage redirect")
	flags.BoolVar(&config.GlobalModelxdOptions.EnableMetrics, "enable-metrics", true, "enable metrics api")

	return cmd
}

func Run(ctx context.Context, g run.Group, opts *config.Options) error {
	var err error
	model.GlobalRegistry, err = NewRegistryConfig(ctx, opts)
	if err != nil {
		return err
	}

	mainLogger.Info("Starting server")

	goruntime.SetMaxProcs(mainLogger)
	goruntime.SetMemLimit(mainLogger, DefaultMemlimitRatio)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	router.Router(r)
	srv := http.Server{
		Addr:              config.GlobalModelxdOptions.Listen,
		WriteTimeout:      time.Second * 1500,
		ReadHeaderTimeout: time.Second * 60,
		ReadTimeout:       time.Second * 1500,
		IdleTimeout:       time.Second * 60,
		Handler:           r,
		MaxHeaderBytes:    1 << 20,
	}

	g.Add(func() error {
		mainLogger.Info("Starting web server", logging.Any("listenAddress", config.GlobalModelxdOptions.Listen))
		if config.GlobalModelxdOptions.TLS.CertFile != "" && config.GlobalModelxdOptions.TLS.KeyFile != "" {
			mainLogger.Info("registry listening", logging.Any("https", config.GlobalModelxdOptions.Listen))
			return srv.ListenAndServeTLS(config.GlobalModelxdOptions.TLS.CertFile, config.GlobalModelxdOptions.TLS.KeyFile)
		} else {
			mainLogger.Info("registry listening", logging.Any("http", opts.Listen))
			return srv.ListenAndServe()
		}
	}, func(error) {
		srv.Close()
	})

	if err := g.Run(); err != nil {
		mainLogger.Error("Failed to run", logging.Error(err))
		os.Exit(1)
	}

	return nil
}

func NewRegistryConfig(ctx context.Context, opt *config.Options) (*model.Registry, error) {
	mainLogger.Info("prepare registry", logging.Any("options", opt))
	var registryStore registry.RegistryInterface
	if registryStore == nil && opt.S3 != nil && opt.S3.URL != "" {
		mainLogger.Info("start modelx registry with S3 type", logging.Any("options", opt))

		s3store, err := registry.NewS3RegistryStore(ctx, opt)
		if err != nil {
			return nil, err
		}
		registryStore = s3store
	}
	if registryStore == nil {
		mainLogger.Info("start modelx registry with LocalFS type")
		fsstore, err := registry.NewFSRegistryStore(ctx, opt)
		if err != nil {
			return nil, err
		}
		registryStore = fsstore
	}
	if registryStore == nil {
		return nil, fmt.Errorf("no storage backend set")
	}
	return &model.Registry{Store: registryStore}, nil
}
