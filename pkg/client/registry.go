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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/opencontainers/go-digest"
	"kubegems.io/modelx/pkg/response"
	"kubegems.io/modelx/pkg/util"
)

func NewRegistryClient(addr string, auth string) *RegistryClient {
	return &RegistryClient{
		Registry:      addr,
		Authorization: auth,
	}
}

type RegistryClient struct {
	Registry      string
	Authorization string
}

func (t *RegistryClient) GetManifest(ctx context.Context, repository string, version string) (*util.Manifest, error) {
	if version == "" {
		version = "latest"
	}
	manifest := &util.Manifest{}
	path := "/" + repository + "/manifests/" + version
	if err := t.simplerequest(ctx, "GET", path, manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}

func (t *RegistryClient) PutManifest(ctx context.Context, repository string, version string, manifest util.Manifest) error {
	if version == "" {
		version = "latest"
	}
	path := "/" + repository + "/manifests/" + version
	return t.simpleuploadrequest(ctx, "PUT", path, manifest, nil)
}

func (t *RegistryClient) GetIndex(ctx context.Context, repository string, search string) (*util.Index, error) {
	index := &util.Index{}
	path := "/" + repository + "/index" + "?search=" + search
	if err := t.simplerequest(ctx, "GET", path, index); err != nil {
		return nil, err
	}
	return index, nil
}

func (t *RegistryClient) GetGlobalIndex(ctx context.Context, search string) (*util.Index, error) {
	query := url.Values{}
	if search != "" {
		query.Add("search", search)
	}
	path := "/"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	index := &util.Index{}
	if err := t.simplerequest(ctx, "GET", path, index); err != nil {
		return nil, err
	}
	return index, nil
}

func (t *RegistryClient) HeadBlob(ctx context.Context, repository string, digest digest.Digest) (bool, error) {
	path := "/" + repository + "/blobs/" + digest.String()
	resp, err := t.request(ctx, "HEAD", path, nil, nil, nil)
	if err != nil {
		return false, err
	}
	return resp.StatusCode == http.StatusOK, nil
}

func (t *RegistryClient) GetBlobContent(ctx context.Context, repository string, digest digest.Digest, into io.Writer) error {
	path := "/" + repository + "/blobs/" + digest.String()
	return t.simplerequest(ctx, "GET", path, into)
}

func (t *RegistryClient) GetBlobLocation(ctx context.Context, repository string, desc util.Descriptor, purpose string) (*util.BlobLocation, error) {
	reqpath := "/" + path.Join(repository, "blobs", desc.Digest.String(), "locations", purpose)
	query := url.Values{}
	query.Set("size", strconv.FormatInt(desc.Size, 10))
	query.Set("name", desc.Name)
	query.Set("media-type", desc.MediaType)
	if desc.Annotations != nil {
		query.Set("annotations", desc.Annotations.String())
	}
	reqpath += "?" + query.Encode()
	into := &util.BlobLocation{}
	if err := t.simplerequest(ctx, "GET", reqpath, into); err != nil {
		return nil, err
	}
	return into, nil
}

func (t *RegistryClient) UploadBlobContent(ctx context.Context, repository string, blob DescriptorWithContent) error {
	header := map[string]string{
		"Content-Type": "application/octet-stream",
	}
	path := "/" + repository + "/blobs/" + blob.Digest.String()
	content, err := blob.GetContent()
	if err != nil {
		return err
	}
	_, err = t.request(ctx, "PUT", path, header, content, nil)
	return err
}

type GetContentFunc func() (io.ReadSeekCloser, error)

type RqeuestBody struct {
	ContentLength int64
	ContentBody   func() (io.ReadSeekCloser, error)
}

func (t *RegistryClient) simplerequest(ctx context.Context, method, url string, into any) error {
	_, err := t.request(ctx, method, url, nil, nil, into)
	return err
}

func (t *RegistryClient) simpleuploadrequest(ctx context.Context, method, url string, body any, into any) error {
	header := map[string]string{
		"Content-Type": "application/json",
	}
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	_, err = t.request(ctx, method, url, header, bytes.NewReader(data), into)
	return err
}

func (t *RegistryClient) request(ctx context.Context, method, url string,
	header map[string]string, reqbody io.Reader, into any,
) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, t.Registry+url, reqbody)
	if err != nil {
		return nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", t.Authorization)
	req.Header.Set("User-Agent", UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 && req.Method != "HEAD" {
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
	if into != nil {
		defer resp.Body.Close()
		switch dest := into.(type) {
		case io.Writer:
			_, err := io.Copy(dest, resp.Body)
			if err != nil {
				return nil, err
			}
		default:
			if err := json.NewDecoder(resp.Body).Decode(into); err != nil {
				return nil, err
			}
		}
	}
	return resp, nil
}
