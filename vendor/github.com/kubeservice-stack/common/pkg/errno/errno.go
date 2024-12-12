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

package errno

import (
	"fmt"
)

type Errno struct {
	status  int    `errno:"errno"`
	message string `json:"errmsg"`
}

func (e *Errno) Status() int {
	return e.status
}

func (e *Errno) Message() string {
	return e.message
}

func (e *Errno) Error() string {
	return fmt.Sprintf("Error - errno: %d, errmsg: %s", e.status, e.message)
}

func New(status int, message string) *Errno {
	return &Errno{status: status, message: message}
}

func NewCode(status int) *Errno {
	return &Errno{status: status}
}
