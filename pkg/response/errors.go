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

package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/opencontainers/go-digest"
)

const (
	ErrCodeBlobUnknown         ErrCode = "BLOB_UNKNOWN"
	ErrCodeBlobUploadInvalid   ErrCode = "BLOB_UPLOAD_INVALID"
	ErrCodeBlobUploadUnknown   ErrCode = "BLOB_UPLOAD_UNKNOWN"
	ErrCodeDigestInvalid       ErrCode = "DIGEST_INVALID"
	ErrCodeManifestBlobUnknown ErrCode = "MANIFEST_BLOB_UNKNOWN"
	ErrCodeManifestInvalid     ErrCode = "MANIFEST_INVALID"
	ErrCodeManifestUnknown     ErrCode = "MANIFEST_UNKNOWN"
	ErrCodeNameInvalid         ErrCode = "NAME_INVALID"
	ErrCodeNameUnknown         ErrCode = "NAME_UNKNOWN"
	ErrCodeSizeInvalid         ErrCode = "SIZE_INVALID"
	ErrCodeUnauthorized        ErrCode = "UNAUTHORIZED"
	ErrCodeDenied              ErrCode = "DENIED"
	ErrCodeUnsupported         ErrCode = "UNSUPPORTED"
	ErrCodeTooManyRequests     ErrCode = "TOOMANYREQUESTS"
	ErrCodeConfigInvalid       ErrCode = "CONFIG_INVALID"
	ErrCodeInvalidParameter    ErrCode = "INVALID_PARAMETER"
	ErrCodeIndexUnknown        ErrCode = "INDEX_UNKNOWN"
	ErrCodeUnknow              ErrCode = "UNKNOWN"
	ErrCodeInternal            ErrCode = "INTERNAL"
)

type ErrCode string

type ErrorInfo struct {
	HttpStatus int     `json:"-"`
	Code       ErrCode `json:"code"`
	Message    string  `json:"message"`
	Detail     string  `json:"detail"`
}

func (e ErrorInfo) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func IsErrCode(err error, code ErrCode) bool {
	if err == nil {
		return false
	}
	info := ErrorInfo{}
	if errors.As(err, &info) {
		return info.Code == code
	}
	return false
}

func NewUnauthorizedError(msg string) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusUnauthorized, Code: ErrCodeUnauthorized, Message: msg}
}

func NewUnsupportedError(msg string) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusNotImplemented, Code: ErrCodeUnsupported, Message: msg}
}

func NewInternalError(err error) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusInternalServerError, Code: ErrCodeInternal, Message: err.Error()}
}

func NewDigestInvalidError(got string) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusBadRequest, Code: ErrCodeDigestInvalid, Message: fmt.Sprintf("digest invalid: %s", got)}
}

func NewIndexUnknownError(repository string) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusNotFound, Code: ErrCodeIndexUnknown, Message: fmt.Sprintf("index: %s not found", repository)}
}

func NewBlobUnknownError(digest digest.Digest) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusNotFound, Code: ErrCodeBlobUnknown, Message: fmt.Sprintf("blob: %s not found", digest.String())}
}

func NewManifestUnknownError(reference string) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusNotFound, Code: ErrCodeManifestUnknown, Message: fmt.Sprintf("manifest: %s not found", reference)}
}

func NewManifestInvalidError(err error) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusBadRequest, Code: ErrCodeManifestInvalid, Message: err.Error()}
}

func NewContentTypeInvalidError(got string) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusBadRequest, Code: ErrCodeInvalidParameter, Message: fmt.Sprintf("content type invalid: %s", got)}
}

func NewContentRangeInvalidError(msg string) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusBadRequest, Code: ErrCodeSizeInvalid, Message: fmt.Sprintf("content range: %s", msg)}
}

func NewContentLengthInvalidError(msg string) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusBadRequest, Code: ErrCodeSizeInvalid, Message: fmt.Sprintf("content length: %s", msg)}
}

func NewConfigInvalidError(msg string) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusBadRequest, Code: ErrCodeConfigInvalid, Message: msg}
}

func NewParameterInvalidError(msg string) ErrorInfo {
	return ErrorInfo{HttpStatus: http.StatusBadRequest, Code: ErrCodeInvalidParameter, Message: msg}
}

func ResponseError(w http.ResponseWriter, err error) {
	info := ErrorInfo{}
	if !errors.As(err, &info) {
		info = ErrorInfo{
			HttpStatus: http.StatusBadRequest,
			Code:       ErrCodeUnknow,
			Message:    err.Error(),
			Detail:     err.Error(),
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(info.HttpStatus)
	_ = json.NewEncoder(w).Encode(info)
}

func ResponseOK(w http.ResponseWriter, data any) {
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}
