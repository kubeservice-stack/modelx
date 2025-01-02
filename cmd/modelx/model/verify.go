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
	"os"
	"path/filepath"

	"github.com/kubeservice-stack/common/pkg/dag"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"kubegems.io/modelx/cmd/modelx/repo"
	"kubegems.io/modelx/pkg/client"
)

func NewVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "verify dependencies have expected content",
		Example: `
    Verify checks that the dependencies of the current model.
    If all the modules are unmodified,verify prints "all modules verified." 
    
	# verify current model in current directory

		modelx verify .

	# verify abc model with directory abc
			
		modelx verify abc

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
			if err := VerifyModel(ctx, args[0]); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func VerifyModel(ctx context.Context, dir string) error {
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

	// Only verify up to GOMAXPROCS zips at once.
	//type token struct{}
	//sem := make(chan token, runtime.GOMAXPROCS(0))
	Dag := dag.NewDAG()

	err = CheckGraphMainfestContent(ctx, config.Dependencies, Dag)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		return nil
	}

	fmt.Printf("all models verified\n")
	return nil
}

func CheckGraphMainfestContent(ctx context.Context, refs []string, Dag *dag.DAG) error {
	for _, ref := range refs {
		refn := dag.NewVertex(ref, ref)
		if err := Dag.AddVertex(refn); err != nil {
			return err
		}

		reference, err := ParseReference(ref)
		if err != nil {
			return err
		}

		cli := reference.Client()
		manfiest, err := cli.GetManifest(ctx, reference.Repository, reference.Version)
		if err != nil {
			return fmt.Errorf("dependencies reference %s not found, err: %s", ref, err.Error())
		}

		into := bytes.NewBuffer(nil)
		if err := cli.Remote.GetBlobContent(ctx, reference.Repository, manfiest.Config.Digest, into); err != nil {
			return err
		}

		var config ModelConfig
		if err := yaml.Unmarshal(into.Bytes(), &config); err != nil {
			return fmt.Errorf("parse model config:%s %w", client.ModelConfigFileName, err)
		}
		for _, d := range config.Dependencies {
			dn := dag.NewVertex(d, d)
			if err := Dag.AddVertex(dag.NewVertex(d, d)); err != nil {
				return err
			}

			if err := Dag.AddEdge(refn, dn); err != nil {
				return err
			}
		}
		return CheckGraphMainfestContent(ctx, config.Dependencies, Dag)
	}
	return nil
}
