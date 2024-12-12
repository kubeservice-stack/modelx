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

	"github.com/spf13/cobra"
	"kubegems.io/modelx/cmd/modelx/repo"
	"kubegems.io/modelx/pkg/client"
)

func NewLoginCmd() *cobra.Command {
	token := ""
	cmd := &cobra.Command{
		Use:   "login",
		Short: "login to a modelx repository",
		Example: `
	1. Add a repo

  		modelx repo add myrepo http://modelx.example.com

	2. Login to myrepo with token

  		modelx login myrepo --token <token>

		`,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return repo.CompleteRegistry(toComplete)
			}
			return nil, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := BaseContext()
			defer cancel()
			if len(args) == 0 {
				return errors.New("at least one argument is required")
			}
			if token == "" {
				fmt.Print("Token: ")
				fmt.Scanln(&token)
			}
			return LoginModelx(ctx, args[0], token)
		},
	}
	cmd.Flags().StringVarP(&token, "token", "t", "", "token")
	return cmd
}

func LoginModelx(ctx context.Context, reponame string, token string) error {
	repoDetails, err := repo.DefaultRepoManager.Get(reponame)
	if err != nil {
		return err
	}
	repoDetails.Token = token
	if err := repoDetails.Client().Ping(ctx); err != nil {
		return err
	}
	fmt.Printf("Login successful for %s\n", reponame)
	return repo.DefaultRepoManager.Set(repoDetails)
}

func Ping(ctx context.Context, repo string, token string) error {
	token = "Bearer " + token
	return client.NewClient(repo, token).Ping(ctx)
}
