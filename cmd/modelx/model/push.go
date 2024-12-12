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
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"kubegems.io/modelx/cmd/modelx/repo"
)

func NewPushCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push",
		Short: "push a model to a modelx repository",
		Example: `
	# Push current directory to repo myrepo

		modelx push myrepo/project/demo

	# Push current directory to repo myrepo as v1.0.0
			
		modelx push myrepo/project/demo@v1.0.0

	# Push directory abc to repo myrepo
			
		modelx push myrepo/project/demo@v1.0.0 abc

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
			if err := PushModel(ctx, args[0], args[1]); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func PushModel(ctx context.Context, ref string, dir string) error {
	reference, err := ParseReference(ref)
	if err != nil {
		return err
	}
	if dir == "" {
		dir = "."
	}
	// parse annotations from model config
	configcontent, err := os.ReadFile(filepath.Join(dir, ModelConfigFileName))
	if err != nil {
		return fmt.Errorf("read model config:%s %w", ModelConfigFileName, err)
	}
	var config ModelConfig
	if err := yaml.Unmarshal(configcontent, &config); err != nil {
		return fmt.Errorf("parse model config:%s %w", ModelConfigFileName, err)
	}
	fmt.Printf("Pushing to %s \n", reference.String())
	return reference.Client().Push(ctx, reference.Repository, reference.Version, ModelConfigFileName, dir)
}
