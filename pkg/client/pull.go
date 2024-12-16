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
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kubeservice-stack/common/pkg/utils"
	"github.com/opencontainers/go-digest"
	"golang.org/x/sync/errgroup"

	"kubegems.io/modelx/pkg/progress"
	"kubegems.io/modelx/pkg/response"
	"kubegems.io/modelx/pkg/util"
)

func (c *Client) Pull(ctx context.Context, repo string, version string, into string, force bool) error {
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
	if force {
		dirlists, err := utils.ListDir(into)
		if err != nil {
			return fmt.Errorf("force clean %s model fail, Please use pull model to other dirctoty.", into)
		}

		for _, dirlist := range dirlists {
			if dirlist == ModelCacheDir || dirlist == ModelConfigFileName || dirlist == ReadmeFileName {
				continue
			}
			flag := false
			for _, blob := range blobs {
				if dirlist == blob.Name {
					flag = true
				}
			}
			if flag == false {
				_ = utils.RemoveDir(filepath.Join(into, dirlist))
				_ = utils.RemoveFile(filepath.Join(into, dirlist))
			}
		}
	}

	return c.PullBlobs(ctx, repo, into, blobs)
}

func (c *Client) PullBlobs(ctx context.Context, repo string, basedir string, blobs []util.Descriptor) error {
	mb, ctx := progress.NewMuiltiBarContext(ctx, os.Stdout, 60, DefaultPullPushConcurrency)
	for _, blob := range blobs {
		mb.Go(blob.Name, "pending", func(b *progress.Bar) error {
			return c.pullBlobProgress(ctx, repo, blob, basedir, b)
		})
	}
	return mb.Wait()
}

func (c *Client) pullBlobProgress(ctx context.Context, repo string, desc util.Descriptor, basedir string, bar *progress.Bar) error {
	switch desc.MediaType {
	case MediaTypeModelDirectoryTarGz:
		if err := os.MkdirAll(filepath.Join(basedir, desc.Name), 0o755); err != nil {
			return fmt.Errorf("create directory %s: %v", filepath.Join(basedir, desc.Name), err)
		}
		return c.pullDirectory(ctx, repo, desc, basedir, bar, true)
	case MediaTypeModelFile:
		return c.pullFile(ctx, repo, desc, basedir, bar)
	case MediaTypeModelConfigYaml:
		return c.pullConfig(ctx, repo, desc, basedir, bar)
	default:
		return fmt.Errorf("unsupported media type %s", desc.MediaType)
	}
}

func OpenWriteFile(filename string, perm os.FileMode) (*os.File, error) {
	if perm == 0 {
		perm = 0o644
	}
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return nil, err
	}
	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm.Perm())
}

func WriteToFile(filename string, src io.Reader, perm os.FileMode) error {
	var f *os.File
	var err error

	if perm == 0 {
		perm = 0o644
	}

	f, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm.Perm())
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
			return err
		}
		f, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm.Perm())
		if err != nil {
			return err
		}
	}

	defer f.Close()

	if closer, ok := src.(io.Closer); ok {
		defer closer.Close()
	}

	_, err = io.Copy(f, src)
	return err
}

func (c Client) pullConfig(ctx context.Context, repo string, desc util.Descriptor, basedir string, bar *progress.Bar) error {
	return c.pullFile(ctx, repo, desc, basedir, bar)
}

func (c Client) pullFile(ctx context.Context, repo string, desc util.Descriptor, basedir string, bar *progress.Bar) error {
	// check hash
	bar.SetNameStatus(desc.Name, "checking", false)
	filename := filepath.Join(basedir, desc.Name)
	if f, err := os.Open(filename); err == nil {
		digest, err := digest.FromReader(f)
		if err != nil {
			return err
		}
		if digest.String() == desc.Digest.String() {
			bar.SetNameStatus(desc.Digest.Hex()[:8], "already exists", true)
			return nil
		}
		_ = f.Close()
	} else if !os.IsNotExist(err) {
		return err
	}

	f, err := OpenWriteFile(filename, desc.Mode.Perm())
	if err != nil {
		return err
	}
	defer f.Close()
	if desc.Digest == EmptyFileDigiest {
		return nil
	}
	w := bar.WrapWriter(f, desc.Digest.Hex()[:8], desc.Size, "downloading")
	if err := c.PullBlob(ctx, repo, desc, w); err != nil {
		return err
	}
	bar.SetStatus("done", true)
	return nil
}

func (c Client) pullDirectory(ctx context.Context, repo string, desc util.Descriptor, basedir string, bar *progress.Bar, useCache bool) error {
	// check hash
	bar.SetNameStatus(desc.Name, "checking", false)
	digest, err := TGZ(ctx, filepath.Join(basedir, desc.Name), "")
	if err != nil {
		return err
	}
	if digest.String() == desc.Digest.String() {
		bar.SetNameStatus(desc.Digest.Hex()[:8], "already exists", true)
		return nil
	}

	// pull to cache
	if useCache {
		cache := filepath.Join(basedir, ".modelx", desc.Name+".tar.gz")
		wf, err := OpenWriteFile(cache, desc.Mode)
		if err != nil {
			return err
		}
		defer wf.Close()

		w := bar.WrapWriter(wf, desc.Digest.Hex()[:8], desc.Size, "downloading")
		if err := c.PullBlob(ctx, repo, desc, w); err != nil {
			return err
		}
		_ = wf.Close()

		// extract
		rf, err := os.Open(cache)

		if err != nil {
			return err
		}
		r := bar.WrapReader(rf, desc.Digest.Hex()[:8], desc.Size, "extracting")
		if err := UnTGZ(ctx, filepath.Join(basedir, desc.Name), r); err != nil {
			return err
		}
		bar.SetStatus("done", true)
		return nil
	} else {
		// download and extract at same time
		piper, pipew := io.Pipe()
		var src io.Reader = piper

		eg, ctx := errgroup.WithContext(ctx)
		// download
		eg.Go(func() error {
			w := bar.WrapWriter(pipew, desc.Digest.Hex()[:8], desc.Size, "downloading")
			return c.PullBlob(ctx, repo, desc, w)
		})
		// extract
		eg.Go(func() error {
			if err := UnTGZ(ctx, filepath.Join(basedir, desc.Name), src); err != nil {
				return err
			}
			bar.SetStatus("done", true)
			return nil
		})
		return eg.Wait()
	}
}

func (c Client) PullBlob(ctx context.Context, repo string, desc util.Descriptor, into io.Writer) error {
	location, err := c.Remote.GetBlobLocation(ctx, repo, desc, util.BlobLocationPurposeDownload)
	if err != nil {
		if !IsServerUnsupportError(err) {
			return err
		}
		return c.Remote.GetBlobContent(ctx, repo, desc.Digest, into)
	}
	return c.Extension.Download(ctx, desc, *location, into)
}

func IsServerUnsupportError(err error) bool {
	info := response.ErrorInfo{}
	if stderrors.As(err, &info) {
		return info.Code == response.ErrCodeUnsupported || info.HttpStatus == http.StatusNotFound
	}
	return false
}
