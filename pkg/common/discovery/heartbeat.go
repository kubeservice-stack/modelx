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
	"context"
	"fmt"
	"time"

	etcd "go.etcd.io/etcd/clientv3"
)

const defaultTTL = 10 // 默认heartbeat 时间间隔10s

var errKeepaliveStopped = fmt.Errorf("heartbeat keepalive stopped")

// etcd的heartbeat, 在后台goroutine keepalive执行
type heartbeat struct {
	client *etcd.Client
	key    string
	value  []byte

	keepaliveCh <-chan *etcd.LeaseKeepAliveResponse
	isElect     bool

	ttl int64
}

func newHeartbeat(client *etcd.Client, key string, value []byte, ttl int64, isElect bool) *heartbeat {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &heartbeat{
		client:  client,
		isElect: isElect,
		key:     key,
		value:   value,
		ttl:     ttl,
	}
}

func (h *heartbeat) grantKeepAliveLease(ctx context.Context) (bool, error) {
	resp, err := h.client.Grant(ctx, h.ttl)
	if err != nil {
		return false, err
	}
	var ops []etcd.Cmp
	if h.isElect {
		ops = append(ops, etcd.Compare(etcd.CreateRevision(h.key), "=", 0))
	}
	txn := h.client.Txn(ctx).If(ops...)
	txn = txn.Then(etcd.OpPut(h.key, string(h.value), etcd.WithLease(resp.ID)))
	txn = txn.Else(etcd.OpGet(h.key))
	response, err := txn.Commit()
	if err != nil {
		return false, err
	}
	response.Responses[0].GetResponse()
	if response.Succeeded {
		h.keepaliveCh, err = h.client.KeepAlive(ctx, resp.ID)
	}
	return response.Succeeded, err
}

func (h *heartbeat) keepAlive(ctx context.Context) {
	var (
		err error
		gap = 100 * time.Millisecond // 时间间隔
	)
	for {
		if err != nil {
			time.Sleep(gap)
			if h.isElect {
				isSuccess, e := h.grantKeepAliveLease(ctx)
				err = e
				if !isSuccess {
					// 写入失败， 关闭心跳
					return
				}
			} else {
				_, err = h.grantKeepAliveLease(ctx)
			}
			// ctx出错，停止keepAlive
			if ctx.Err() != nil {
				return
			}
		} else {
			err = h.handleAliveResp(ctx)
			if err != nil && err == errKeepaliveStopped {
				return
			}
		}
	}
}

func (h *heartbeat) handleAliveResp(ctx context.Context) error {
	select {
	case aliveResp := <-h.keepaliveCh:
		if aliveResp == nil {
			return errKeepaliveStopped
		}
	case <-ctx.Done():
		return errKeepaliveStopped
	}
	return nil
}
