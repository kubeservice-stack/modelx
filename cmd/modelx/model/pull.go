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
	"errors"
	"fmt"
	"path"

	"github.com/spf13/cobra"
	"kubegems.io/modelx/cmd/modelx/repo"
)

func NewPullCmd() *cobra.Command {
	IsForce := false
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "pull a model from a repository",
		Long:  "pull [--force-clean/-f] <repo>/[project]/[name]@[version] .",
		Example: `
	# Pull project/demo version latest to dirctory demo by default

  		modelx pull  myrepo/project/demo

	# Pull project/demo to current dirctoty

		modelx pull  myrepo/project/demo@version .
		
	# Pull project/demo to dirctoty abc

		modelx pull  myrepo/project/demo@version abc

	# Pull project/demo to dirctoty abc

		modelx pull -f myrepo/project/demo@version abc
		`,
		SilenceUsage: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return repo.CompleteRegistryRepositoryVersion(toComplete)
			}
			if len(args) == 1 {
				return nil, cobra.ShellCompDirectiveFilterDirs
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := BaseContext()
			defer cancel()
			if len(args) == 0 {
				return errors.New("at least one argument is required")
			}
			if len(args) == 1 {
				args = append(args, "")
			}
			return PullModelx(ctx, args[0], args[1], IsForce)
		},
	}
	cmd.Flags().BoolVarP(&IsForce, "force-clean", "f", false, "force pull clean local modelx file or directory")
	return cmd
}

func PullModelx(ctx context.Context, ref string, into string, force bool) error {
	reference, err := ParseReference(ref)
	if err != nil {
		return err
	}
	if reference.Repository == "" {
		return errors.New("repository is not specified")
	}
	if into == "" {
		into = path.Base(reference.Repository)
	}
	fmt.Printf("Pulling %s into %s \n", reference.String(), into)
	return reference.Client().Pull(ctx, reference.Repository, reference.Version, into, force)
}
