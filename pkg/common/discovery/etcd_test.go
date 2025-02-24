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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.etcd.io/etcd/integration"
)

type ETCDMockCluster struct {
	cluster   *integration.ClusterV3
	Endpoints []string
}

func StartEtcdMockCluster(t *testing.T) *ETCDMockCluster {
	cluster := integration.NewClusterV3(t, &integration.ClusterConfig{Size: 1})
	return &ETCDMockCluster{
		cluster:   cluster,
		Endpoints: []string{cluster.Members[0].GRPCAddr()},
	}
}

func (etcd *ETCDMockCluster) Terminate(t *testing.T) {
	etcd.cluster.Terminate(t)
}

type EtcdClusterTestSuite struct {
	suite.Suite
	Cluster *ETCDMockCluster
}

func (s *EtcdClusterTestSuite) SetupTest() {
	s.Cluster = StartEtcdMockCluster(s.T())
}

func (s *EtcdClusterTestSuite) TearDownTest() {
	s.Cluster.Terminate(s.T())
}

func (s *EtcdClusterTestSuite) TestWriteRead() {
	ed, err := newEtedDiscovery(Options{
		Endpoints: s.Cluster.Endpoints,
	}, "nobody")

	s.Nil(err)

	err = ed.Put(context.TODO(), "/test/key1", []byte("dongjiang"))
	s.Nil(err)

	d1, err1 := ed.Get(context.TODO(), "/test/key1")
	s.Nil(err1)
	s.Equal(string(d1), "dongjiang")

	err2 := ed.Delete(context.TODO(), "/test/key1")
	s.Nil(err2)

	err3 := ed.Close()
	s.Nil(err3)
}
func (s *EtcdClusterTestSuite) TestWriteReadWithPrefix() {
	ed, err := newEtedDiscovery(Options{
		Endpoints: s.Cluster.Endpoints,
		Prefix:    "test-",
	}, "nobody")

	s.Nil(err)

	err = ed.Put(context.TODO(), "/test/key1", []byte("dongjiang"))
	s.Nil(err)

	d1, err1 := ed.Get(context.TODO(), "/test/key1")
	s.Nil(err1)
	s.Equal(string(d1), "dongjiang")

	err2 := ed.Delete(context.TODO(), "/test/key1")
	s.Nil(err2)

	err3 := ed.Close()
	s.Nil(err3)
}

func (s *EtcdClusterTestSuite) TestList() {
	ed, err := newEtedDiscovery(Options{
		Namespace: "/test/list",
		Endpoints: s.Cluster.Endpoints,
	}, "nobody")

	s.Nil(err)

	err = ed.Put(context.TODO(), "/test/key1", []byte("dongjiang"))
	s.Nil(err)

	err = ed.Put(context.TODO(), "/test/key2", []byte("dongjiang"))
	s.Nil(err)

	// put 空
	err = ed.Put(context.TODO(), "/test/key3", []byte{})
	s.Nil(err)

	list, err := ed.List(context.TODO(), "/test")
	s.Nil(err)

	s.Equal([]KeyValue{{Key: "/test/key1", Value: []byte("dongjiang")}, {Key: "/test/key2", Value: []byte("dongjiang")}}, list)
	s.Equal(2, len(list))

	err3 := ed.Close()
	s.Nil(err3)
}

func (s *EtcdClusterTestSuite) TestNewDiscovery() {
	_, err := newEtedDiscovery(Options{}, "nobody")
	s.NotNil(err)
}

func (s *EtcdClusterTestSuite) TestHeartBeat() {
	ed, err := newEtedDiscovery(Options{
		Endpoints: s.Cluster.Endpoints,
	}, "nobody")
	s.Nil(err)

	heartbeat := fmt.Sprintf("/cluster1/storage/heartbeat/%s:%d", "127.0.0.1", 2918)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var ch <-chan Closed
	ch, err = ed.Heartbeat(ctx, heartbeat, []byte("dongjiang"), 1)
	s.Nil(err)

	_, err = ed.Get(ctx, heartbeat)
	s.Nil(err)

	cancel()

	time.Sleep(time.Second)

	_, err = ed.Get(ctx, heartbeat)
	s.NotNil(err, "heartbeat should be deleted automatically")

	select {
	case <-ch:
	case <-time.After(500 * time.Millisecond):
		s.Nil(fmt.Errorf("heartbeat should be deleted automatically"))
	}

	err3 := ed.Close()
	s.Nil(err3)
}

func (s *EtcdClusterTestSuite) TestWatch() {
	ed, err := newEtedDiscovery(Options{
		Endpoints: s.Cluster.Endpoints,
	}, "nobody")
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())

	// 不存在数据
	ch := ed.Watch(ctx, "/cluster1/data/1", true)
	s.NotNil(ch)

	var wg sync.WaitGroup
	var mutex sync.RWMutex

	val := make(map[string]string)

	// 同步数据闭包
	syncKVs := func(ch WatchEventChan) {
		for event := range ch {
			if event.Err != nil {
				continue
			}
			mutex.Lock()
			for _, kv := range event.KeyValues {
				val[kv.Key] = string(kv.Value)
			}
			mutex.Unlock()
		}
	}
	wg.Add(1)

	go func() {
		defer wg.Done()
		syncKVs(ch)
	}()
	s.Equal(0, len(val))

	err = ed.Put(ctx, "/cluster1/data/1", []byte("dongjiang1"))
	s.Nil(err)

	// 测试存在的数据
	err = ed.Put(ctx, "/cluster1/data/2", []byte("dongjiang2"))
	s.Nil(err)
	ch2 := ed.Watch(ctx, "/cluster1/data/2", true)

	wg.Add(1)
	go func() {
		defer wg.Done()
		syncKVs(ch2)
	}()

	err = ed.Put(ctx, "/cluster1/controller/2", []byte("222"))
	s.Nil(err)

	time.Sleep(200 * time.Millisecond)

	cancel()
	wg.Wait()

	s.Equal(2, len(val))
	s.Equal("dongjiang1", val["/cluster1/data/1"])
	s.Equal("dongjiang2", val["/cluster1/data/2"])

	err3 := ed.Close()
	s.Nil(err3)
}

func (s *EtcdClusterTestSuite) TestGetWatchPrefix() {
	ed, err := newEtedDiscovery(Options{
		Endpoints: s.Cluster.Endpoints,
	}, "nobody")
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = ed.Put(context.TODO(), "/test/data/1", []byte("dongjiang1"))
	s.Nil(err)

	err = ed.Put(context.TODO(), "/test/data/2", []byte("dongjiang2"))
	s.Nil(err)

	ch := ed.WatchPrefix(ctx, "/test/data", true)
	s.NotNil(ch)
	time.Sleep(100 * time.Millisecond)

	err = ed.Put(context.TODO(), "/test/data/3", []byte("dongjiang3"))
	s.Nil(err)

	bytes1, err := ed.Get(context.TODO(), "/test/data/3")
	s.Equal("dongjiang3", string(bytes1))
	s.Nil(err)

	err = ed.Delete(context.TODO(), "/test/data/3")
	s.Nil(err)

	time.Sleep(time.Second)

	var allEvt, modifyEvt, deleteEvt bool
	for event := range ch {
		if event.Err != nil {
			continue
		}
		kvs := map[string]string{}
		for _, kv := range event.KeyValues {
			kvs[kv.Key] = string(kv.Value)
		}
		switch event.Type {
		case EventTypeAll:
			s.Equal(false, allEvt)
			s.Equal(false, modifyEvt)
			s.Equal(false, deleteEvt)
			s.Equal(2, len(kvs))
			s.Equal("dongjiang1", kvs["/test/data/1"])
			s.Equal("dongjiang2", kvs["/test/data/2"])

			allEvt = true

		case EventTypeModify:
			s.Equal(true, allEvt)
			s.Equal(false, modifyEvt)
			s.Equal(false, deleteEvt)
			s.Equal(1, len(kvs))
			s.Equal("dongjiang3", kvs["/test/data/3"])

			modifyEvt = true

		case EventTypeDelete:
			s.Equal(true, allEvt)
			s.Equal(true, modifyEvt)
			s.Equal(false, deleteEvt)
			s.Equal(1, len(kvs))
			s.Equal("", kvs["/test/data/3"])

			deleteEvt = true
			cancel()
		}
	}
	s.Equal(true, deleteEvt)

	err3 := ed.Close()
	s.Nil(err3)
}

func (s *EtcdClusterTestSuite) TestTransaction() {
	ed, err := newEtedDiscovery(Options{
		Namespace: "/test/batch",
		Endpoints: s.Cluster.Endpoints,
	}, "nobody")
	s.Nil(err)

	txn := ed.NewTransaction()
	txn.Put("test", []byte("dongjiang"))
	err = ed.Commit(context.TODO(), txn)
	s.Nil(err)

	v, _ := ed.Get(context.TODO(), "test")
	s.Equal([]byte("dongjiang"), v)

	txn = ed.NewTransaction()
	txn.ModRevisionCmp("key", "=", 0)
	txn.Put("test", []byte("dongjiang-new"))
	err = ed.Commit(context.TODO(), txn)
	s.Nil(err)

	v, err = ed.Get(context.TODO(), "test")
	s.Nil(err)
	s.Equal("dongjiang-new", string(v))

	txn = ed.NewTransaction()
	txn.ModRevisionCmp("key", "=", 33)
	txn.Delete("test")
	err = ed.Commit(context.TODO(), txn)
	s.NotNil(err)

	v, err = ed.Get(context.TODO(), "test")
	s.Nil(err)
	s.Equal([]byte("dongjiang-new"), v)

	txn = ed.NewTransaction()
	txn.ModRevisionCmp("key", "=", 0)
	txn.Delete("test")
	err = ed.Commit(context.TODO(), txn)
	s.Nil(err)

	_, err = ed.Get(context.TODO(), "test")
	s.NotNil(err)

	s.NotNil(TxnErr(nil, fmt.Errorf("err")))

	err3 := ed.Close()
	s.Nil(err3)
}

func (s *EtcdClusterTestSuite) TestBatch() {
	ed, err := newEtedDiscovery(Options{
		Namespace: "/test/batch",
		Endpoints: s.Cluster.Endpoints,
	}, "nobody")
	s.Nil(err)

	batch := Batch{
		KVs: []KeyValue{
			{"key1", []byte("dongjiang1")},
			{"key2", []byte("dongjiang2")},
			{"key3", []byte("dongjiang3")},
		},
	}
	success, err := ed.Batch(context.TODO(), batch)
	s.Nil(err)
	s.Equal(true, success)

	list, err := ed.List(context.TODO(), "key")
	s.Nil(err)
	s.Equal(3, len(list))

	err3 := ed.Close()
	s.Nil(err3)
}

func (s *EtcdClusterTestSuite) TestElect() {
	ed, err := newEtedDiscovery(Options{
		Namespace: "/test/batch",
		Endpoints: s.Cluster.Endpoints,
	}, "nobody")
	s.Nil(err)

	ctx, cancel := context.WithCancel(context.Background())

	success, ch, err := ed.Elect(ctx, "/test/data/1", []byte("dongjiang"), 1)
	s.Nil(err)
	s.NotNil(ch)
	s.Equal(true, success)

	time.Sleep(2 * time.Second)

	bytes, err := ed.Get(context.TODO(), "/test/data/1")
	s.Nil(err)
	s.Equal("dongjiang", string(bytes))

	ctx2, cancel2 := context.WithCancel(context.Background())

	shouldFalse, _, err := ed.Elect(ctx2, "/test/data/1", []byte("dongjiang-new"), 1)
	s.Equal(false, shouldFalse)
	s.Nil(err)

	if cancel2 != nil {
		cancel2()
	}
	cancel()
	select {
	case <-ch:
	case <-time.After(500 * time.Millisecond):
		s.Nil(fmt.Errorf("cancel heartbeat timeout"))
	}
	time.Sleep(2 * time.Second)

	val, err := ed.Get(context.TODO(), "/test/data/1")
	s.NotNil(err)
	s.Equal("", string(val))

	ctx3, cancel3 := context.WithCancel(context.Background())
	shouldSuccess, cch, err := ed.Elect(ctx3, "/test/data/1", []byte("dongjiang-new-new"), 1)
	s.Equal(true, shouldSuccess)
	s.Nil(err)
	s.NotNil(cch)

	bytes3, err := ed.Get(context.TODO(), "/test/data/1")
	s.Nil(err)
	s.Equal("dongjiang-new-new", string(bytes3))

	cancel3()

	err3 := ed.Close()
	s.Nil(err3)
}

// go test 入口
func TestEtcdClusterTestSuite(t *testing.T) {
	suite.Run(t, new(EtcdClusterTestSuite))
}
