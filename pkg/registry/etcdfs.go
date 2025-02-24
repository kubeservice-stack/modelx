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

package registry

import (
	"context"

	discovery "kubegems.io/modelx/pkg/common/discovery"
	modelxconfig "kubegems.io/modelx/pkg/config"
	// modelxerrors "kubegems.io/modelx/pkg/response"
)

//var _ FSProvider = &EtcdStorageProvider{}

type EtcdStorageProvider struct {
	ctx  context.Context
	etcd discovery.Discovery
}

func NewEtcdFSProvider(ctx context.Context, options *modelxconfig.EtcdOptions) (*EtcdStorageProvider, error) {
	var (
		factory discovery.DiscoveryFactory = discovery.NewDiscoveryFactory("modelx")
		etcd    discovery.Discovery
		err     error
	)

	opts := discovery.Options{
		Namespace:   options.Namespace,
		Endpoints:   options.Endpoints,
		DialTimeout: options.DialTimeout,
		Prefix:      options.Prefix,
	}

	if etcd, err = factory.CreateDiscovery(opts); err != nil {
		return nil, err
	}

	ed := &EtcdStorageProvider{
		ctx:  ctx,
		etcd: etcd,
	}

	// add etcd heartbeat
	go func() {
		var (
			closed <-chan discovery.Closed
			err    error
		)
		for {
			if closed, err = ed.etcd.Heartbeat(ed.ctx, "/heartbeat", []byte("heartbeat"), 30); err != nil {
				continue
			}
			<-closed
		}
	}(ed)

	return ed, nil
}

func (e *EtcdStorageProvider) Put(ctx context.Context, path string, content BlobContent) error {
	uploadobj := &s3.PutObjectInput{
		Bucket:        aws.String(m.Bucket),
		Key:           m.prefixedKey(path),
		Body:          content.Content,
		ContentLength: ptr.To(content.ContentLength),
		ContentType:   aws.String(content.ContentType),
	}

	if _, err := manager.NewUploader(m.Client).Upload(ctx, uploadobj); err != nil {
		return modelxerrors.NewInternalError(err)
	}

	return nil
}

func (m *EtcdStorageProvider) Remove(ctx context.Context, path string, recursive bool) error {
	if recursive {
		prefix := m.prefixedKey(path)
		if !strings.HasSuffix(*prefix, "/") {
			*prefix += "/"
		}
		output, err := m.Client.ListObjects(ctx, &s3.ListObjectsInput{
			Bucket: aws.String(m.Bucket),
			Prefix: prefix,
		})
		if err != nil {
			return err
		}
		if len(output.Contents) == 0 {
			return nil
		}
		objectsids := make([]types.ObjectIdentifier, 0, len(output.Contents))
		for _, object := range output.Contents {
			objectsids = append(objectsids, types.ObjectIdentifier{Key: object.Key})
		}

		deleteOutput, err := m.Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(m.Bucket),
			Delete: &types.Delete{Objects: objectsids},
		})
		if err != nil {
			return err
		}
		_ = deleteOutput
		return nil
	} else {
		_, err := m.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(m.Bucket),
			Key:    m.prefixedKey(path),
		})
		return err
	}
}

func (m *EtcdStorageProvider) Get(ctx context.Context, path string) (*BlobContent, error) {
	getobjout, err := m.Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(m.Bucket),
		Key:    m.prefixedKey(path),
	})
	if err != nil {
		return nil, err
	}
	return &BlobContent{
		Content:       getobjout.Body,
		ContentType:   StringDeref(getobjout.ContentType, ""),
		ContentLength: *getobjout.ContentLength,
	}, nil
}

func (m *EtcdStorageProvider) Exists(ctx context.Context, path string) (bool, error) {
	_, err := m.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(m.Bucket),
		Key:    m.prefixedKey(path),
	})
	if err != nil {
		if IsS3StorageNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *EtcdStorageProvider) Stat(ctx context.Context, path string) (FsObjectMeta, error) {
	headobjout, err := m.Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(m.Bucket),
		Key:    m.prefixedKey(path),
	})
	if err != nil {
		if IsS3StorageNotFound(err) {
			return FsObjectMeta{}, os.ErrNotExist
		}
		return FsObjectMeta{}, err
	}
	return FsObjectMeta{
		Name:         path,
		Size:         *headobjout.ContentLength,
		LastModified: TimeDeref(headobjout.LastModified, time.Time{}),
		ContentType:  StringDeref(headobjout.ContentType, ""),
	}, nil
}

func (m *EtcdStorageProvider) List(ctx context.Context, path string, recursive bool) ([]FsObjectMeta, error) {
	prefix := *m.prefixedKey(path)
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	listinput := &s3.ListObjectsInput{
		Bucket: aws.String(m.Bucket),
		Prefix: aws.String(prefix),
	}
	if !recursive {
		listinput.Delimiter = aws.String("/")
	}
	var result []FsObjectMeta
	listobjout, err := m.Client.ListObjects(ctx, listinput)
	if err != nil {
		return nil, err
	}
	for _, obj := range listobjout.Contents {
		result = append(result, FsObjectMeta{
			Name:         strings.TrimPrefix(*obj.Key, prefix),
			Size:         *obj.Size,
			LastModified: *obj.LastModified,
		})
	}
	for *listobjout.IsTruncated {
		listinput.Marker = listobjout.NextMarker
		listobjout, err = m.Client.ListObjects(ctx, listinput)
		if err != nil {
			return nil, err
		}
		for _, obj := range listobjout.Contents {
			result = append(result, FsObjectMeta{
				Name:         strings.TrimPrefix(*obj.Key, prefix),
				Size:         *obj.Size,
				LastModified: *obj.LastModified,
			})
		}
	}
	return result, nil
}

func IsS3StorageNotFound(err error) bool {
	var apie *http.ResponseError
	if errors.As(err, &apie) {
		return apie.HTTPStatusCode() == 404
	}
	return false
}

func (m *EtcdStorageProvider) prefixedKey(key string) *string {
	return aws.String(path.Join(m.Prefix, key))
}
