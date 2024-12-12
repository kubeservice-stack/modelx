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

	"github.com/mholt/archiver/v4"
	"github.com/opencontainers/go-digest"
	"kubegems.io/modelx/pkg/util"
)

type DescriptorWithContent struct {
	util.Descriptor
	GetContent func() (io.ReadSeekCloser, error)
}

var tgz = archiver.Archive{
	Archival:    archiver.Tar{},
	Compression: archiver.Gz{},
	Extraction:  archiver.Tar{},
}

func TGZ(ctx context.Context, dir string, intofile string) (digest.Digest, error) {
	files, err := archiver.FilesFromDisk(
		&archiver.FromDiskOptions{ClearAttributes: true},
		map[string]string{dir + string(os.PathSeparator): ""},
	)
	if err != nil {
		return "", err
	}

	writers := []io.Writer{}
	if intofile != "" {
		if err := os.MkdirAll(filepath.Dir(intofile), 0o755); err != nil {
			return "", err
		}
		f, err := os.Create(intofile)
		if err != nil {
			return "", err
		}
		defer f.Close()

		writers = append(writers, f)
	}
	d := digest.Canonical.Digester()
	writers = append(writers, d.Hash())

	if err := tgz.Archive(ctx, io.MultiWriter(writers...), files); err != nil {
		return "", err
	}
	return d.Digest(), nil
}

func UnTGZ(ctx context.Context, intodir string, readercloser io.Reader) error {
	return tgz.Extract(ctx, readercloser, func(ctx context.Context, f archiver.FileInfo) error {
		nameinlocal := filepath.Join(intodir, f.NameInArchive)
		if f.IsDir() {
			return os.MkdirAll(nameinlocal, f.Mode())
		}
		srcfile, err := f.Open()
		if err != nil {
			return err
		}
		defer srcfile.Close()

		intofile, err := os.OpenFile(nameinlocal, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer intofile.Close()

		_, err = io.Copy(intofile, srcfile)
		if err != nil {
			return err
		}
		return nil
	})
}
