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
	"kubegems.io/modelx/pkg/client"
)

func NewVendorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vendor",
		Short: "vendor a model depend on other models",
		Example: `
	# vendor current model in current directory

		modelx vendor .

	# vendor abc model with directory abc
			
		modelx vendor abc

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
			if err := VendorModel(ctx, args[0]); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func VendorModel(ctx context.Context, dir string) error {
	if dir == "" {
		dir = "."
	}
	// parse annotations from model config
	configcontent, err := os.ReadFile(filepath.Join(dir, client.ModelConfigFileName))
	if err != nil {
		return fmt.Errorf("read model config:%s %w", client.ModelConfigFileName, err)
	}
	var config ModelConfig
	if err := yaml.Unmarshal(configcontent, &config); err != nil {
		return fmt.Errorf("parse model config:%s %w", client.ModelConfigFileName, err)
	}

	if len(config.Dependencies) > 0 {
		fmt.Printf("Pushing vendor to %s/%s \n", dir, client.ModelVendorDir)
		for _, depend := range config.Dependencies {
			reference, err := ParseReference(depend)
			if err != nil {
				return err
			}
			err = reference.Client().Pull(ctx, reference.Repository, reference.Version, dir+"/"+client.ModelVendorDir+"/"+reference.Name(), true)
			if err != nil {
				return err
			}
		}

	} else {
		fmt.Printf("No vendor need to %s/%s \n", dir, client.ModelVendorDir)
		return nil
	}

	return nil
}
