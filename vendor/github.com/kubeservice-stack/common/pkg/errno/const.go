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

package errno

import "net/http"

var (
	// 2xx

	// 4xx
	BadRequest          = &Errno{http.StatusBadRequest, "BadRequest"}           // "错误的请求"
	InvalidPath         = &Errno{http.StatusForbidden, "BadRequest"}            // "url path不合法"
	InvalidParams       = &Errno{http.StatusBadRequest, "InvalidParams"}        // "参数不合法"
	ParseParamsFail     = &Errno{http.StatusBadRequest, "ParseParamsFail"}      // "解析请求参数失败"
	PageParamInvalid    = &Errno{http.StatusBadRequest, "PageParamInvalid"}     // "page 参数不合法"
	NotFound            = &Errno{http.StatusNotFound, "NotFound"}               // "资源不存在"
	InvalidAccessKey    = &Errno{http.StatusForbidden, "InvalidAccessKey"}      // "无效的 AccessKey"
	SignatureExpires    = &Errno{http.StatusForbidden, "SignatureExpires"}      // "签名过期"
	SignatureNotMatch   = &Errno{http.StatusForbidden, "SignatureNotMatch"}     // "签名错误"
	PermissionForbidden = &Errno{http.StatusForbidden, "PermissionForbidden"}   // "没有操作权限"
	TooManyRequests     = &Errno{http.StatusTooManyRequests, "TooManyRequests"} //"请求过多"

	// 5xx
	InternalServerError = &Errno{http.StatusInternalServerError, "InternalServerError"} //"内部服务出错"
	OperationFailed     = &Errno{http.StatusInternalServerError, "OperationFailed"}     // "操作失败"
)
