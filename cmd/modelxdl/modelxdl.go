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

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"kubegems.io/modelx/cmd/modelx/model"
	types "kubegems.io/modelx/pkg/util"
	"kubegems.io/modelx/pkg/version"
)

const ErrExitCode = 1

func main() {
	if err := NewDLCmd().Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(ErrExitCode)
	}
}

func NewDLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "modelxdl",
		Short:   "modelx storage initalizer for seldon",
		Version: version.Get().String(),
		Example: `
		# Just for NO-Auth environment
		modelxdl http://127.0.0.1:8080/library/model@v1 /mnt/model
		
		# Authorizations config from environment variable MODELX_AUTH
		modelxdl http://127.0.0.1:8080/library/model@v1?token=<token> /mnt/model
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("requires two arguments")
			}
			ctx, cancel := model.BaseContext()
			defer cancel()

			// Seldon Storage Initializer accept two arguments: modelUri and modelPath
			// Authorizations config from environment variable MODELX_AUTH
			return Run(ctx, args[0], args[1])
		},
	}
	return cmd
}

func Run(ctx context.Context, uri string, dest string) error {
	ref, err := model.ParseReference(uri)
	if err != nil {
		return err
	}
	fmt.Printf("Pulling %s into %s \n", ref.String(), dest)
	cli := ref.Client()

	manifest, err := cli.Remote.GetManifest(ctx, ref.Repository, ref.Version)
	if err != nil {
		return err
	}
	into := bytes.NewBuffer(nil)

	if err := cli.Remote.GetBlobContent(ctx, ref.Repository, manifest.Config.Digest, into); err != nil {
		return err
	}

	config := &model.ModelConfig{}
	if err := yaml.Unmarshal(into.Bytes(), config); err != nil {
		return err
	}

	// filter modelfiles
	pullblobs := []types.Descriptor{}

	if len(config.ModelFiles) == 0 {
		pullblobs = append(pullblobs, manifest.Config)
		pullblobs = append(pullblobs, manifest.Blobs...)
	} else {
		for _, modelfile := range config.ModelFiles {
			// case:  a/models/b.bin
			// 	use:  a
			firstelem := filepath.SplitList(modelfile)[0]
			for _, manifestdesc := range manifest.Blobs {
				if manifestdesc.Name == firstelem {
					pullblobs = append(pullblobs, manifestdesc)
				}
			}
		}
	}

	files := []string{}
	for _, blob := range pullblobs {
		files = append(files, blob.Name)
	}
	fmt.Printf("Pulling files %v\n into %s", files, dest)
	return cli.PullBlobs(ctx, ref.Repository, dest, pullblobs)
}
