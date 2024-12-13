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
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
	"github.com/opencontainers/go-digest"

	registry "kubegems.io/modelx/pkg/registry"
	errors "kubegems.io/modelx/pkg/response"
	"kubegems.io/modelx/pkg/routers"
	types "kubegems.io/modelx/pkg/util"
)

type Registry struct {
	Store registry.RegistryInterface
}

func HeadManifest(c *gin.Context) {
	name, reference := GetRepositoryReference(c)
	exist, err := GlobalRegistry.Store.ExistsManifest(c.Request.Context(), name, reference)
	if err != nil {
		if registry.IsRegistryStoreNotNotFound(err) {
			c.Writer.WriteHeader(http.StatusNotFound)
		} else {
			errors.ResponseError(c.Writer, err)
		}
		return
	}
	if exist {
		c.Writer.WriteHeader(http.StatusOK)
	} else {
		c.Writer.WriteHeader(http.StatusNotFound)
	}
}

func GetGlobalIndex(c *gin.Context) {
	index, err := GlobalRegistry.Store.GetGlobalIndex(c.Request.Context(), c.Request.URL.Query().Get("search"))
	if err != nil {
		if registry.IsRegistryStoreNotNotFound(err) {
			errors.ResponseOK(c.Writer, types.Index{})
		} else {
			errors.ResponseError(c.Writer, err)
		}
		return
	}
	errors.ResponseOK(c.Writer, index)
}

func GetIndex(c *gin.Context) {
	name, _ := GetRepositoryReference(c)
	index, err := GlobalRegistry.Store.GetIndex(c.Request.Context(), name, c.Request.URL.Query().Get("search"))
	if err != nil {
		if registry.IsRegistryStoreNotNotFound(err) {
			errors.ResponseError(c.Writer, errors.NewIndexUnknownError(name))
		} else {
			errors.ResponseError(c.Writer, err)
		}
		return
	}
	errors.ResponseOK(c.Writer, index)
}

func DeleteIndex(c *gin.Context) {
	name, _ := GetRepositoryReference(c)
	if err := GlobalRegistry.Store.RemoveIndex(c.Request.Context(), name); err != nil {
		if registry.IsRegistryStoreNotNotFound(err) {
			errors.ResponseError(c.Writer, errors.NewIndexUnknownError(name))
		} else {
			errors.ResponseError(c.Writer, err)
		}
		return
	}
	errors.ResponseOK(c.Writer, "ok")
}

func GetManifest(c *gin.Context) {
	name, reference := GetRepositoryReference(c)
	manifest, err := GlobalRegistry.Store.GetManifest(c.Request.Context(), name, reference)
	if err != nil {
		if registry.IsRegistryStoreNotNotFound(err) {
			errors.ResponseError(c.Writer, errors.NewManifestUnknownError(reference))
		} else {
			errors.ResponseError(c.Writer, err)
		}
		return
	}
	errors.ResponseOK(c.Writer, manifest)
}

func PutManifest(c *gin.Context) {
	name, reference := GetRepositoryReference(c)
	var manifest types.Manifest
	if err := json.NewDecoder(c.Request.Body).Decode(&manifest); err != nil {
		errors.ResponseError(c.Writer, errors.NewManifestInvalidError(err))
		return
	}
	contenttype := c.Request.Header.Get("Content-Type")
	if err := GlobalRegistry.Store.PutManifest(c.Request.Context(), name, reference, contenttype, manifest); err != nil {
		errors.ResponseError(c.Writer, err)
		return
	}
	c.Writer.WriteHeader(http.StatusCreated)
}

func DeleteManifest(c *gin.Context) {
	name, reference := GetRepositoryReference(c)
	if err := GlobalRegistry.Store.DeleteManifest(c.Request.Context(), name, reference); err != nil {
		if registry.IsRegistryStoreNotNotFound(err) {
			errors.ResponseError(c.Writer, errors.NewManifestUnknownError(reference))
		} else {
			errors.ResponseError(c.Writer, err)
		}
		return
	}
	c.Writer.WriteHeader(http.StatusAccepted)
}

func GetRepositoryReference(c *gin.Context) (string, string) {
	return c.Param("repository") + "/" + c.Param("name"), c.Param("reference")
}

func HeadBlob(c *gin.Context) {
	BlobDigestFun(c, func(ctx context.Context, repository string, digest digest.Digest) {
		ok, err := GlobalRegistry.Store.ExistsBlob(c.Request.Context(), repository, digest)
		if err != nil {
			errors.ResponseError(c.Writer, err)
			return
		}
		if ok {
			c.Writer.WriteHeader(http.StatusOK)
		} else {
			c.Writer.WriteHeader(http.StatusNotFound)
		}
	})
}

// 如果客户端 包含 contentLength 则直接上传
// 如果客户端 不包含 contentLength 则返回一个 Location 后续上传至该地址
func PutBlob(c *gin.Context) {
	BlobDigestFun(c, func(ctx context.Context, repository string, digest digest.Digest) {
		contentType := c.Request.Header.Get("Content-Type")
		if contentType == "" {
			errors.ResponseError(c.Writer, errors.NewContentTypeInvalidError("empty"))
			return
		}
		content := registry.BlobContent{
			ContentLength: c.Request.ContentLength,
			ContentType:   contentType,
			Content:       c.Request.Body,
		}
		if err := GlobalRegistry.Store.PutBlob(c.Request.Context(), repository, digest, content); err != nil {
			modelLogger.Error("store put blob", zap.Error(err), zap.Any("action", "put-blob"), zap.Any("repository", repository), zap.Any("digest", digest.String()))
			errors.ResponseError(c.Writer, err)
			return
		}
		c.Writer.WriteHeader(http.StatusCreated)
	})
}

func GetBlob(c *gin.Context) {
	BlobDigestFun(c, func(ctx context.Context, repository string, digest digest.Digest) {
		result, err := GlobalRegistry.Store.GetBlob(c.Request.Context(), repository, digest)
		if err != nil {
			modelLogger.Error("store get blob", zap.Error(err), zap.Any("action", "get-blob"), zap.Any("repository", repository), zap.Any("digest", digest.String()))
			if registry.IsRegistryStoreNotNotFound(err) {
				errors.ResponseError(c.Writer, errors.NewBlobUnknownError(digest))
			}
			errors.ResponseError(c.Writer, err)
			return
		}
		defer result.Close()

		c.Writer.Header().Set("Content-Type", result.ContentType)
		c.Writer.WriteHeader(http.StatusOK)
		_, _ = io.Copy(c.Writer, result.Content)
	})
}

func GarbageCollect(c *gin.Context) {
	name, _ := GetRepositoryReference(c)
	result, err := registry.GCBlobs(c.Request.Context(), GlobalRegistry.Store, name)
	if err != nil {
		errors.ResponseError(c.Writer, errors.NewInternalError(err))
		return
	}
	errors.ResponseOK(c.Writer, result)
}

func GetBlobLocation(c *gin.Context) {
	BlobDigestFun(c, func(ctx context.Context, repository string, digest digest.Digest) {
		purpose := c.Param("purpose")
		properties := make(map[string]string)
		for k, v := range c.Request.URL.Query() {
			properties[k] = strings.Join(v, ",")
		}
		result, err := GlobalRegistry.Store.GetBlobLocation(c.Request.Context(), repository, digest, purpose, properties)
		if err != nil {
			if registry.IsRegistryStoreNotNotFound(err) {
				errors.ResponseError(c.Writer, errors.NewBlobUnknownError(digest))
			} else {
				errors.ResponseError(c.Writer, err)
			}
			return
		}
		errors.ResponseOK(c.Writer, result)
	})
}

func BlobDigestFun(c *gin.Context, fun func(ctx context.Context, repository string, digest digest.Digest)) {
	name, _ := GetRepositoryReference(c)
	digeststr := c.Param("digest")
	digest, err := digest.Parse(digeststr)
	if err != nil {
		errors.ResponseError(c.Writer, errors.NewDigestInvalidError(digeststr))
		return
	}
	fun(c.Request.Context(), name, digest)
}

func ParseDescriptor(c *gin.Context) (types.Descriptor, error) {
	digeststr := c.Param("digest")
	digest, err := digest.Parse(digeststr)
	if err != nil {
		return types.Descriptor{}, errors.NewDigestInvalidError(digeststr)
	}
	contentType := c.Request.Header.Get("Content-Type")
	if contentType == "" {
		return types.Descriptor{}, errors.NewContentTypeInvalidError("empty")
	}
	descriptor := types.Descriptor{
		Digest:    digest,
		MediaType: contentType,
	}
	return descriptor, nil
}

func ParseAndCheckContentRange(header http.Header) (int64, int64, error) {
	contentRange, contentLength := header.Get("Content-Range"), header.Get("Content-Length")
	ranges := strings.Split(contentRange, "-")
	if len(ranges) != 2 {
		return -1, -1, errors.NewContentRangeInvalidError("invalid format")
	}
	start, err := strconv.ParseInt(ranges[0], 10, 64)
	if err != nil {
		return -1, -1, errors.NewContentRangeInvalidError("invalid start")
	}
	end, err := strconv.ParseInt(ranges[1], 10, 64)
	if err != nil {
		return -1, -1, errors.NewContentRangeInvalidError("invalid end")
	}
	if start > end {
		return -1, -1, errors.NewContentRangeInvalidError("start > end")
	}
	contentLen, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return -1, -1, errors.NewContentRangeInvalidError("invalid content length")
	}
	if contentLen != (end-start)+1 {
		return -1, -1, errors.NewContentRangeInvalidError("content length != (end-start)+1")
	}
	return start, end, nil
}

func init() {
	// global index
	router.Register("GlobalIndex", "/", "/", http.MethodGet, GetGlobalIndex)

	// gc
	router.Register("GC", "/", ":repository/:name/garbage-collect", http.MethodPost, GarbageCollect)

	// index
	router.Register("Index", "/", ":repository/:name/index", http.MethodGet, GetIndex)
	router.Register("Index", "/", ":repository/:name/index", http.MethodDelete, DeleteIndex)

	// repository/manifests
	router.Register("Manifests", "/", ":repository/:name/manifests/:reference", http.MethodGet, GetManifest)
	router.Register("Manifests", "/", ":repository/:name/manifests/:reference", http.MethodPut, registry.MaxBytesReadHandler(PutManifest, registry.DefaultMaxBytesRead))
	router.Register("Manifests", "/", ":repository/:name/manifests/:reference", http.MethodDelete, DeleteManifest)

	// repository/blobs
	router.Register("Blobs", "/", ":repository/:name/blobs/:digest", http.MethodHead, HeadBlob)
	router.Register("Blobs", "/", ":repository/:name/blobs/:digest", http.MethodGet, GetBlob)
	router.Register("Blobs", "/", ":repository/:name/blobs/:digest", http.MethodPut, PutBlob)

	// repository/blobs/locations
	router.Register("BlobLocations", "/", ":repository/:name/blobs/:digest/locations/:purpose", http.MethodGet, GetBlobLocation)
}
