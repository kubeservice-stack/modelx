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

package client

import (
	"context"
	"fmt"
	"os"

	"kubegems.io/modelx/pkg/progress"
)

func (c *Client) Copy(ctx context.Context, repoTo, versionTo string, repoFrom, versionFrom string) error {
	if versionFrom == "" {
		versionFrom = "latest"
	}
	manifest, err := c.GetManifest(ctx, repoFrom, versionFrom)
	if err != nil {
		return fmt.Errorf("source reference %s/%s not found, err: %s", repoFrom, versionFrom, err.Error())
	}

	p, ctx := progress.NewMuiltiBarContext(ctx, os.Stdout, 60, DefaultPullPushConcurrency)

	// copy blobs
	p.Go(repoFrom+"/"+versionFrom, "copying", func(b *progress.Bar) error {
		err := c.CopyBlobs(ctx, repoTo, repoFrom, versionTo, versionFrom)
		if err != nil {
			b.SetStatus("fail", true)
			return err
		}
		b.SetStatus("done", true)
		return nil
	})

	// copy manifest
	p.Go("manifest", "copying", func(b *progress.Bar) error {
		if err := c.PutManifest(ctx, repoTo, versionTo, *manifest); err != nil {
			return err
		}
		b.SetNameStatus("manifest", "done", true)
		return nil
	})
	return p.Wait()
}
