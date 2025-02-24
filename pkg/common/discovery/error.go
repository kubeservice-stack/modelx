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

package discovery

import (
	"fmt"

	etcdcliv3 "go.etcd.io/etcd/clientv3"
)

var (
	ErrWatchFailed = fmt.Errorf("etcd watch returns a nil chan")
	ErrNoKey       = fmt.Errorf("etcd has no such key")
	ErrTxnFailed   = fmt.Errorf("role changed or target revision mismatch")
	ErrTxnConvert  = fmt.Errorf("cannot covert etcd transaction")
)

// 将txn响应和错误转化为一个错误
func TxnErr(resp *etcdcliv3.TxnResponse, err error) error {
	if err != nil {
		return err
	}
	if !resp.Succeeded {
		return ErrTxnFailed
	}
	return nil
}
