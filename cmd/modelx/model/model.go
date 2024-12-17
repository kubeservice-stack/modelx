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

package model

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"
	"kubegems.io/modelx/pkg/version"
)

const ModelxDebugEnv = "MODELX_DEBUG"

func NewModelxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "modelx",
		Short:   "modelx",
		Version: version.Get().String(),
	}
	cmd.AddCommand(NewInitCmd())
	cmd.AddCommand(NewLoginCmd())
	cmd.AddCommand(NewListCmd())
	cmd.AddCommand(NewInfoCmd())
	cmd.AddCommand(NewPushCmd())
	cmd.AddCommand(NewPullCmd())
	cmd.AddCommand(NewDiffCmd())
	return cmd
}

func BaseContext() (context.Context, context.CancelFunc) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	if os.Getenv(ModelxDebugEnv) == "1" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		ctx = logr.NewContext(ctx, stdr.NewWithOptions(log.Default(), stdr.Options{LogCaller: stdr.Error}))
	}
	return ctx, cancel
}
