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
	"kubegems.io/modelx/pkg/version"
)

func NewDiffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "diff manifests",
		Long:  "list <repo>/[project]/[name]@[version] .",
		Example: `
	# Diff projects of repo and local directory abc

  		modelx diff  myrepo/project/demo abc

	# Diff projects of repo and local directory .

		modelx diff  myrepo/project/demo .
		
		modelx diff  myrepo/project/demo
		`,
		Version: version.Get().String(),
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
			if err := DiffModel(ctx, args[0], args[1]); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func DiffModel(ctx context.Context, ref string, dir string) error {
	reference, err := ParseReference(ref)
	if err != nil {
		return err
	}
	if reference.Repository == "" {
		return errors.New("repository is not specified")
	}
	if dir == "" {
		dir = path.Base(reference.Repository)
	}
	fmt.Printf("Diff with %s and %s \n", reference.String(), dir)
	return reference.Client().Diff(ctx, reference.Repository, reference.Version, dir)
}
