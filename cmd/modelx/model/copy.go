/*
Copyright 2025 The KubeService-Stack Authors.

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

	"github.com/spf13/cobra"
	"kubegems.io/modelx/cmd/modelx/repo"
)

func NewCopyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy",
		Short: "copy a model to a modelx remote repository",
		Example: `
	# Copy myrepo/project/from@latest to repo myrepo/project/demo@latest

		modelx copy myrepo/project/demo myrepo/project/from 

	# Copy myrepo/project/from@v1.0.0 to repo myrepo/project/demo@latest

		modelx copy myrepo/project/demo myrepo/project/from@v1.0.0
		`,
		SilenceUsage: true,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 2 {
				return repo.CompleteRegistryRepositoryVersion(toComplete)
			}

			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := BaseContext()
			defer cancel()
			if len(args) != 2 {
				return errors.New("at two argument is required")
			}
			if err := CopyModel(ctx, args[0], args[1]); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func CopyModel(ctx context.Context, refTo string, refFrom string) error {
	referenceTo, err := ParseReference(refTo)
	if err != nil {
		return err
	}

	referenceFrom, err := ParseReference(refFrom)
	if err != nil {
		return err
	}

	if referenceTo.Registry != referenceFrom.Registry {
		return fmt.Errorf("Registry from %s must equal registry to %s", referenceFrom.Registry, referenceTo.Registry)
	}

	// TODO: Add new annotations from model config

	fmt.Printf("Copying %s to %s \n", referenceTo.String(), referenceFrom.String())
	return referenceFrom.Client().Copy(ctx, referenceTo.Repository, referenceTo.Version, referenceFrom.Repository, referenceFrom.Version)
}
