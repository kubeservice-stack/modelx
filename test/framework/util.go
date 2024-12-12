/*
Copyright 2022 The KubeService-Stack Authors.

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

package framework

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

func SourceToIOReader(source string) (io.Reader, error) {
	if strings.HasPrefix(source, "http") {
		return URLToIOReader(source)
	}
	return PathToOSFile(source)
}

func PathToOSFile(relativePath string) (*os.File, error) {
	path, err := filepath.Abs(relativePath)
	if err != nil {
		return nil, fmt.Errorf("failed generate absolute file path of %s: %w", relativePath, err)
	}

	manifest, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}

	return manifest, nil
}

func URLToIOReader(url string) (io.Reader, error) {
	var resp *http.Response

	err := wait.PollUntilContextTimeout(context.Background(), time.Second, time.Minute*5, false, func(_ context.Context) (bool, error) {
		var err error
		resp, err = http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return nil, fmt.Errorf(
			"waiting for %v to return a successful status code timed out. Last response from server was: %v: %w",
			url,
			resp,
			err,
		)
	}

	return resp.Body, nil
}
