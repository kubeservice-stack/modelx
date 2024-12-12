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

package repo

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRepoRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "remove a repository",
		Long:  "Remove a repository",
		Example: `
	# Remove a repository

		modelx repo remove my-repo`,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return CompleteRegistry(toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("repo remove requires at least one argument")
			}
			for _, name := range args {
				if err := DefaultRepoManager.Remove(name); err != nil {
					return err
				}
			}
			return nil
		},
	}
	return cmd
}
