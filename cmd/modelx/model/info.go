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
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"kubegems.io/modelx/cmd/modelx/repo"
)

func NewInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "show config of model",
		Example: `
	# Show modelx.yaml of a remote model.

  		modex info  myrepo/project/demo@version

		`,
		SilenceUsage: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return repo.CompleteRegistryRepositoryVersion(toComplete)
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := BaseContext()
			defer cancel()
			if len(args) == 0 {
				return errors.New("at least one argument is required")
			}
			config, err := GetConfig(ctx, args[0])
			if err != nil {
				return err
			}
			fmt.Print(string(config))
			return nil
		},
	}
	return cmd
}

func GetConfig(ctx context.Context, ref string) ([]byte, error) {
	reference, err := ParseReference(ref)
	if err != nil {
		return nil, err
	}
	if reference.Repository == "" {
		return nil, errors.New("repository is not specified")
	}
	cli := reference.Client()
	manfiest, err := cli.GetManifest(ctx, reference.Repository, reference.Version)
	if err != nil {
		return nil, err
	}
	into := bytes.NewBuffer(nil)
	if err := cli.Remote.GetBlobContent(ctx, reference.Repository, manfiest.Config.Digest, into); err != nil {
		return nil, err
	}
	return into.Bytes(), nil
}
