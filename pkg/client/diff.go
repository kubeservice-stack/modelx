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

package client

import (
	"context"
	"fmt"
	"os"

	"github.com/sergi/go-diff/diffmatchpatch"
)

func (c *Client) Diff(ctx context.Context, repo string, version string, into string) error {
	// check if the directory exists and is empty
	if dirInfo, err := os.Stat(into); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(into, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %v", into, err)
		}
	} else {
		if !dirInfo.IsDir() {
			return fmt.Errorf("%s is not a directory", into)
		}
	}

	manifest, err := c.GetManifest(ctx, repo, version)
	if err != nil {
		return err
	}

	blobs := append(manifest.Blobs, manifest.Config)

	lists, err := os.ReadDir(into)
	if err != nil {
		return fmt.Errorf("Something error %s, Please use pull model to other dirctoty.", err.Error())
	}

	var ret []any

	for _, list := range lists {

		if list.IsDir() && list.Name() == ModelCacheDir {
			continue
		}

		if !list.IsDir() {

			flag := 1
			for _, blob := range blobs {
				if list.Name() == blob.Name {
					dig, _ := c.digest(ctx, list.Name())
					if dig == blob.Digest {
						flag = 3
					} else {
						flag = 2
					}
				}
			}
			fmt.Println(flag)
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain("aaa", "bb", true)
			diffs = dmp.DiffCleanupSemantic(diffs)
			diffs = dmp.DiffCleanupEfficiency(diffs)
			ret = append(ret, dmp.DiffPrettyText(diffs))

		} else if list.IsDir() {

			// something
		}
	}
	fmt.Println(ret...)

	return nil
}
