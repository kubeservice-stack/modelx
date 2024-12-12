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
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"kubegems.io/modelx/pkg/response"
	"kubegems.io/modelx/pkg/util"
)

type ExtensionInterface interface {
	Download(ctx context.Context, blob util.Descriptor, location util.BlobLocation, into io.Writer) error
	Upload(ctx context.Context, blob DescriptorWithContent, location util.BlobLocation) error
}

func NewDelegateExtension() *DelegateExtension {
	return &DelegateExtension{
		Extensions: GlobalExtensions,
	}
}

type DelegateExtension struct {
	Extensions map[string]ExtensionInterface
}

func (e *DelegateExtension) Download(ctx context.Context, blob util.Descriptor, location util.BlobLocation, into io.Writer) error {
	extensionLogger.Debug("extend downloading blob", zap.Any("provider", location.Provider), zap.Any("properties", location.Properties))

	if ext, ok := e.Extensions[location.Provider]; ok {
		return ext.Download(ctx, blob, location, into)
	}
	return response.NewUnsupportedError("provider: " + location.Provider)
}

func (e DelegateExtension) Upload(ctx context.Context, blob DescriptorWithContent, location util.BlobLocation) error {
	extensionLogger.Debug("extend uploading blob", zap.Any("provider", location.Provider), zap.Any("properties", location.Properties))

	if ext, ok := e.Extensions[location.Provider]; ok {
		return ext.Upload(ctx, blob, location)
	}
	return response.NewUnsupportedError("provider: " + location.Provider)
}

type BlobContent struct {
	Content       io.ReadCloser
	ContentLength int64
}

func (t *RegistryClient) extrequest(ctx context.Context, method, url string, header map[string][]string, contentlen int64, content io.ReadCloser) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, t.Registry+url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header[k] = v
	}
	req.Header.Set("Authorization", t.Authorization)
	req.Header.Set("User-Agent", UserAgent)
	req.Body = content
	req.ContentLength = contentlen

	norediretccli := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // do not follow redirect
		},
	}

	resp, err := norediretccli.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest && req.Method != "HEAD" {
		var apierr response.ErrorInfo
		if resp.Header.Get("Content-Type") == "application/json" {
			if err := json.NewDecoder(resp.Body).Decode(&apierr); err != nil {
				return nil, err
			}
		} else {
			bodystr, _ := io.ReadAll(resp.Body)
			apierr.Message = string(bodystr)
		}
		apierr.HttpStatus = resp.StatusCode
		return nil, apierr
	}
	return resp, nil
}

func convertProperties(dest any, src any) error {
	raw, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dest)
}
