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
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/opencontainers/go-digest"

	"kubegems.io/modelx/pkg/progress"
	"kubegems.io/modelx/pkg/util"
)

func (c *Client) Push(ctx context.Context, repo, version string, configfile, basedir string, forcepush bool) error {
	manifest, err := ParseManifest(ctx, basedir, configfile, forcepush)
	if err != nil {
		return err
	}
	p, ctx := progress.NewMuiltiBarContext(ctx, os.Stdout, 60, DefaultPullPushConcurrency)
	// push blobs
	for i := range manifest.Blobs {
		desc := &manifest.Blobs[i]
		p.Go(desc.Name, "pending", func(b *progress.Bar) error {
			switch desc.MediaType {
			case MediaTypeModelFile:
				return c.pushFile(ctx, filepath.Join(basedir, desc.Name), desc, repo, b)
			case MediaTypeModelDirectoryTarGz:
				return c.pushDirectory(ctx, basedir, filepath.Join(basedir, desc.Name), desc, repo, b)
			default:
				return nil
			}
		})
	}
	// push config
	p.Go(manifest.Config.Name, "pending", func(b *progress.Bar) error {
		return c.pushFile(ctx, filepath.Join(basedir, manifest.Config.Name), &manifest.Config, repo, b)
	})
	if err := p.Wait(); err != nil {
		return err
	}
	// push manifest
	p.Go("manifest", "pushing", func(b *progress.Bar) error {
		if err := c.PutManifest(ctx, repo, version, *manifest); err != nil {
			return err
		}
		b.SetNameStatus("manifest", "done", true)
		return nil
	})
	return p.Wait()
}

func ParseManifest(ctx context.Context, basedir string, configfile string, forcepush bool) (*util.Manifest, error) {
	manifest := &util.Manifest{
		MediaType: MediaTypeModelManifestJson,
	}
	ds, err := os.ReadDir(basedir)
	if err != nil {
		return nil, err
	}
	for _, entry := range ds {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if forcepush && entry.Name() == ModelVendorDir {
			continue
		}
		if entry.Name() == configfile {
			manifest.Config = util.Descriptor{
				Name:      entry.Name(),
				MediaType: MediaTypeModelConfigYaml,
			}
			continue
		}
		if entry.IsDir() {
			manifest.Blobs = append(manifest.Blobs, util.Descriptor{
				Name:      entry.Name(),
				MediaType: MediaTypeModelDirectoryTarGz,
			})
			continue
		}
		manifest.Blobs = append(manifest.Blobs, util.Descriptor{
			Name:      entry.Name(),
			MediaType: MediaTypeModelFile,
		})
	}
	slices.SortFunc(manifest.Blobs, util.SortDescriptorName)
	return manifest, nil
}

func (c *Client) pushDirectory(ctx context.Context, cachedir, blobdir string, desc *util.Descriptor, repo string, bar *progress.Bar) error {
	diri, err := os.Stat(blobdir)
	if err != nil {
		return err
	}
	desc.Mode = diri.Mode()
	desc.Modified = diri.ModTime()

	bar.SetNameStatus(desc.Name, "digesting", false)
	filename := filepath.Join(cachedir, ".modelx", desc.Name+".tar.gz")
	digest, err := TGZ(ctx, blobdir, filename)
	if err != nil {
		return err
	}
	desc.Digest = digest
	return c.pushFile(ctx, filename, desc, repo, bar)
}

func (c *Client) pushFile(ctx context.Context, blobfile string, desc *util.Descriptor, repo string, bar *progress.Bar) error {
	fi, err := os.Stat(blobfile)
	if err != nil {
		return err
	}
	if desc.Digest == "" {
		bar.SetNameStatus(desc.Name, "digesting", false)
		digest, err := c.digest(ctx, blobfile)
		if err != nil {
			return err
		}
		desc.Digest = digest
	}
	if desc.Size == 0 {
		desc.Size = fi.Size()
	}
	if desc.Mode == 0 {
		desc.Mode = fi.Mode()
	}
	if desc.Modified.IsZero() {
		desc.Modified = fi.ModTime()
	}
	getReader := func() (io.ReadSeekCloser, error) {
		return os.Open(blobfile)
	}
	bar.SetNameStatus(desc.Digest.Hex()[:8], "pending", false)
	return c.PushBlob(ctx, repo, DescriptorWithContent{Descriptor: *desc, GetContent: getReader}, bar)
}

func (c *Client) digest(ctx context.Context, blobfile string) (digest.Digest, error) {
	f, err := os.Open(blobfile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	go func() {
		<-ctx.Done()
		f.Close()
	}()
	return digest.FromReader(f)
}

func (c *Client) PushBlob(ctx context.Context, repo string, desc DescriptorWithContent, p *progress.Bar) error {
	if desc.Digest == EmptyFileDigiest {
		p.SetStatus("empty", true)
		return nil
	}
	exist, err := c.Remote.HeadBlob(ctx, repo, desc.Digest)
	if err != nil {
		return err
	}
	if exist {
		p.SetStatus("exists", true)
		return nil
	}
	wrappdesc := DescriptorWithContent{
		Descriptor: desc.Descriptor,
		GetContent: func() (io.ReadSeekCloser, error) {
			content, err := desc.GetContent()
			if err != nil {
				return nil, err
			}
			content = p.WrapReader(content, desc.Digest.Hex()[:8], desc.Size, "pushing")
			return content, nil
		},
	}
	if err := c.pushBlob(ctx, repo, wrappdesc); err != nil {
		return err
	}
	p.SetStatus("done", true)
	return nil
}

func (c *Client) pushBlob(ctx context.Context, repo string, desc DescriptorWithContent) error {
	location, err := c.Remote.GetBlobLocation(ctx, repo, desc.Descriptor, util.BlobLocationPurposeUpload)
	if err != nil {
		if !IsServerUnsupportError(err) {
			return err
		}

		return c.Remote.UploadBlobContent(ctx, repo, desc)
	}
	return c.Extension.Upload(ctx, desc, *location)
}
