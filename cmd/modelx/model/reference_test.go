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
	"reflect"
	"testing"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    Reference
		wantErr bool
	}{
		{
			name: "valid",
			raw:  "https://registry.example.com/repository@sha256:abcdef",
			want: Reference{
				Registry:   "https://registry.example.com",
				Repository: "library/repository",
				Version:    "sha256:abcdef",
			},
		},
		{
			raw: "https://registry.example.com:8443/repository/name@v1",
			want: Reference{
				Registry:   "https://registry.example.com:8443",
				Repository: "repository/name",
				Version:    "v1",
			},
		},
		{
			raw: "https://registry.example.com/repo/name",
			want: Reference{
				Registry:   "https://registry.example.com",
				Repository: "repo/name",
				Version:    "",
			},
		},
		{
			raw: "https://registry.example.com/repo/name@latest",
			want: Reference{
				Registry:   "https://registry.example.com",
				Repository: "repo/name",
				Version:    "latest",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReference(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseReference() = %v, want %v", got, tt.want)
			}
		})
	}
}
