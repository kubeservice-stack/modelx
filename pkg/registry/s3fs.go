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

package registry

import (
	"context"
	"errors"
	"os"
	"path"
	"strings"
	"time"

	"k8s.io/utils/ptr"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/transport/http"

	modelxconfig "kubegems.io/modelx/pkg/config"
	modelxerrors "kubegems.io/modelx/pkg/response"
)

var _ FSProvider = &S3StorageProvider{}

type S3StorageProvider struct {
	Bucket  string
	Client  *s3.Client
	PreSign *s3.PresignClient
	Expire  time.Duration
	Prefix  string
}

func NewS3FSProvider(ctx context.Context, options *modelxconfig.S3Options) (*S3StorageProvider, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(options.AccessKey, options.SecretKey, ""),
		),
		config.WithRegion(options.Region),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(
				func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
					return aws.Endpoint{URL: options.URL}, nil
				},
			),
		),
	)
	if err != nil {
		return nil, err
	}
	s3cli := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = options.PathStyle
	})
	return &S3StorageProvider{
		Bucket:  options.Buket,
		Client:  s3cli,
		Expire:  options.PresignExpire,
		Prefix:  "registry",
		PreSign: s3.NewPresignClient(s3cli),
	}, nil
}

func (m *S3StorageProvider) Put(ctx context.Context, path string, content BlobContent) error {
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

func (m *S3StorageProvider) Remove(ctx context.Context, path string, recursive bool) error {
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

func (m *S3StorageProvider) Get(ctx context.Context, path string) (*BlobContent, error) {
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

func (m *S3StorageProvider) Exists(ctx context.Context, path string) (bool, error) {
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

func (m *S3StorageProvider) Stat(ctx context.Context, path string) (FsObjectMeta, error) {
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

func (m *S3StorageProvider) List(ctx context.Context, path string, recursive bool) ([]FsObjectMeta, error) {
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

func (m *S3StorageProvider) prefixedKey(key string) *string {
	return aws.String(path.Join(m.Prefix, key))
}
