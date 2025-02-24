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
	"path/filepath"
	"strings"
	"time"

	etcdcliv3 "go.etcd.io/etcd/clientv3"
	etcdcliv3namespace "go.etcd.io/etcd/clientv3/namespace"
)

type etcdDiscovery struct {
	namespace string
	client    *etcdcliv3.Client
	prefix    string
}

func newEtedDiscovery(cfg Options, owner string) (Discovery, error) {
	cf := etcdcliv3.Config{
		Endpoints:          cfg.Endpoints,
		MaxCallSendMsgSize: 100 * 1024 * 1024, // 100MiB
		MaxCallRecvMsgSize: 0,                 // math.MaxInt32
		DialTimeout:        cfg.DialTimeout * time.Second,
	}
	cli, err := etcdcliv3.New(cf)
	if err != nil {
		return nil, fmt.Errorf("create etc client error:%s", err)
	}
	// prefix string
	if cfg.Prefix != "" {
		cli.KV = etcdcliv3namespace.NewKV(cli.KV, cfg.Prefix)
		cli.Watcher = etcdcliv3namespace.NewWatcher(cli.Watcher, cfg.Prefix)
		cli.Lease = etcdcliv3namespace.NewLease(cli.Lease, cfg.Prefix)
	}

	ed := etcdDiscovery{
		namespace: cfg.Namespace,
		client:    cli,
		prefix:    cfg.Prefix,
	}

	return &ed, nil
}

func (ed *etcdDiscovery) Get(ctx context.Context, key string) ([]byte, error) {
	resp, err := ed.get(ctx, key)
	if err != nil {
		return nil, err
	}
	return ed.getValue(key, resp)
}

func (ed *etcdDiscovery) get(ctx context.Context, key string) (*etcdcliv3.GetResponse, error) {
	resp, err := ed.client.Get(ctx, ed.keyPath(key))
	if err != nil {
		return nil, fmt.Errorf("get value failure for key[%s], error:%s", key, err)
	}
	return resp, nil
}

// keyPath return new key path with namespace prefix
func (ed *etcdDiscovery) keyPath(key string) string {
	if len(ed.namespace) > 0 {
		return filepath.Join(ed.namespace, key)
	}
	return key
}

func (ed *etcdDiscovery) getValue(key string, resp *etcdcliv3.GetResponse) ([]byte, error) {
	if len(resp.Kvs) == 0 {
		return nil, ErrNotExist
	}

	firstKV := resp.Kvs[0]
	if len(firstKV.Value) == 0 {
		return nil, fmt.Errorf("key[%s]'s value is empty", key)
	}
	return firstKV.Value, nil
}

func (ed *etcdDiscovery) List(ctx context.Context, prefix string) ([]KeyValue, error) {
	resp, err := ed.client.Get(ctx, ed.keyPath(prefix), etcdcliv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var result []KeyValue

	if len(resp.Kvs) > 0 {
		for _, kv := range resp.Kvs {
			if len(kv.Value) > 0 {
				result = append(result, KeyValue{Key: ed.parseKey(string(kv.Key)), Value: kv.Value})
			}
		}
	}
	return result, nil
}

// parseKey parses the key, removes the namespace
func (ed *etcdDiscovery) parseKey(key string) string {
	if len(ed.namespace) == 0 {
		return key
	}
	return strings.Replace(key, ed.namespace, "", 1)
}

func (ed *etcdDiscovery) Put(ctx context.Context, key string, val []byte) error {
	_, err := ed.client.Put(ctx, ed.keyPath(key), string(val))
	return err
}

func (ed *etcdDiscovery) Delete(ctx context.Context, key string) error {
	_, err := ed.client.Delete(ctx, ed.keyPath(key))
	return err
}

func (ed *etcdDiscovery) Close() error {
	return ed.client.Close()
}

func (ed *etcdDiscovery) Heartbeat(ctx context.Context, key string, value []byte, ttl int64) (<-chan Closed, error) {
	h := newHeartbeat(ed.client, ed.keyPath(key), value, ttl, false)
	_, err := h.grantKeepAliveLease(ctx)
	if err != nil {
		return nil, err
	}
	ch := make(chan Closed)
	// 后台gorounte keepalive
	go func() {
		// 关闭channel
		defer close(ch)
		h.keepAlive(ctx)
	}()
	return ch, nil
}

func (ed *etcdDiscovery) Elect(ctx context.Context, key string, value []byte, ttl int64) (bool, <-chan Closed, error) {
	h := newHeartbeat(ed.client, ed.keyPath(key), value, ttl, true)
	success, err := h.grantKeepAliveLease(ctx)
	if err != nil {
		return false, nil, err
	}
	if success {
		ch := make(chan Closed)
		// 后台gorounte keepalive
		go func() {
			// 关闭channel
			defer func() {
				close(ch)
			}()
			h.keepAlive(ctx)
		}()
		return success, ch, nil
	}
	return success, nil, nil
}

func (ed *etcdDiscovery) Watch(ctx context.Context, key string, fetchVal bool) WatchEventChan {
	watcher := newWatcher(ctx, ed, ed.keyPath(key), fetchVal)
	return watcher.EventC
}

func (ed *etcdDiscovery) WatchPrefix(ctx context.Context, prefixKey string, fetchVal bool) WatchEventChan {
	watcher := newWatcher(ctx, ed, ed.keyPath(prefixKey), fetchVal, etcdcliv3.WithPrefix())
	return watcher.EventC
}

func (ed *etcdDiscovery) Batch(ctx context.Context, batch Batch) (bool, error) {
	var ops []etcdcliv3.Op
	for _, kv := range batch.KVs {
		ops = append(ops, etcdcliv3.OpPut(
			ed.keyPath(kv.Key),
			string(kv.Value),
		))
	}

	resp, err := ed.client.Txn(ctx).Then(ops...).Commit()
	if err != nil {
		return false, err
	}
	return resp.Succeeded, nil
}

func (ed *etcdDiscovery) NewTransaction() Transaction {
	return newTransaction(ed)
}

func (ed *etcdDiscovery) Commit(ctx context.Context, txn Transaction) error {
	t, ok := txn.(*transaction)
	if !ok {
		return ErrTxnConvert
	}
	resp, err := ed.client.Txn(ctx).If(t.cmps...).Then(t.ops...).Commit()
	return TxnErr(resp, err)
}

type transaction struct {
	ops  []etcdcliv3.Op
	cmps []etcdcliv3.Cmp
	ed   *etcdDiscovery
}

func newTransaction(ed *etcdDiscovery) Transaction {
	return &transaction{ed: ed}
}

func (t *transaction) ModRevisionCmp(key, op string, v interface{}) {
	t.cmps = append(t.cmps, etcdcliv3.Compare(etcdcliv3.ModRevision(t.ed.keyPath(key)), op, v))
}

func (t *transaction) Put(key string, value []byte) {
	t.ops = append(t.ops, etcdcliv3.OpPut(t.ed.keyPath(key), string(value)))
}

func (t *transaction) Delete(key string) {
	t.ops = append(t.ops, etcdcliv3.OpDelete(t.ed.keyPath(key)))
}
