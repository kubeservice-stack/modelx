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
)

var ErrNotExist = fmt.Errorf("discovery is not exist")

type Options struct {
	Namespace   string        `json:"namespace"`   // 命名空间
	Endpoints   []string      `json:"endpoints"`   // 连接端点
	DialTimeout time.Duration `json:"dialTimeout"` // 连接超时时间
	Prefix      string        `json:"prefix"`      // 前缀key
}

type DiscoveryFactory interface {
	CreateDiscovery(cfg Options) (Discovery, error)
}

type Discovery interface {
	// 从service center中获得数据
	Get(ctx context.Context, key string) ([]byte, error)                                         // 根据key获得value
	List(ctx context.Context, prefix string) ([]KeyValue, error)                                 // 根据前缀获得数据
	Put(ctx context.Context, key string, val []byte) error                                       // key-value写入， 为了控制面使用
	Delete(ctx context.Context, key string) error                                                // 删除key，为了控制面使用
	Heartbeat(ctx context.Context, key string, value []byte, ttl int64) (<-chan Closed, error)   // endponit 健康检查
	Elect(ctx context.Context, key string, value []byte, ttl int64) (bool, <-chan Closed, error) // 选举写入： key 不存在，写入成功，返回成功；key存在，写入失败，返回error
	Watch(ctx context.Context, key string, fetchVal bool) WatchEventChan                         // watch key
	WatchPrefix(ctx context.Context, prefixKey string, fetchVal bool) WatchEventChan             // watch 前缀key
	Batch(ctx context.Context, batch Batch) (bool, error)                                        // 批写入
	NewTransaction() Transaction                                                                 // 新transaction
	Commit(ctx context.Context, txn Transaction) error                                           // commit
	Close() error                                                                                // close discovery
}

type EventType int

// Event类型
const (
	EventTypeModify EventType = iota
	EventTypeDelete
	EventTypeAll
)

func (e EventType) String() string {
	switch e {
	case EventTypeModify:
		return "modify"
	case EventTypeDelete:
		return "delete"
	case EventTypeAll:
		return "all"
	default:
		return "unknown"
	}
}

type KeyValue struct {
	Key   string
	Value []byte
}

// 批写入数据
type Batch struct {
	KVs []KeyValue
}

type EventKeyValue struct {
	Key   string
	Value []byte
	Rev   int64
}

// 定义 discovery watch 健值或者前缀的event
type Event struct {
	Type      EventType
	KeyValues []EventKeyValue

	Err error
}

type Closed struct{}

// Watch Event Chan
type WatchEventChan <-chan *Event

// default discovery factory
type discoveryFactory struct {
	owner string
}

func NewDiscoveryFactory(owner string) DiscoveryFactory {
	return &discoveryFactory{owner: owner}
}

func (df *discoveryFactory) CreateDiscovery(cfg Options) (Discovery, error) {
	// 默认etcd discovery
	return newEtedDiscovery(cfg, df.owner)
}

type Transaction interface {
	ModRevisionCmp(key, op string, v interface{})
	Put(key string, value []byte)
	Delete(key string)
}
