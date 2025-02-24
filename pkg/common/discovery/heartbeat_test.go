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
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	etcdcliv3 "go.etcd.io/etcd/clientv3"
)

type EtcdHeartbeatTestSuite struct {
	suite.Suite
	Cluster *ETCDMockCluster
}

func (s *EtcdHeartbeatTestSuite) SetupTest() {
	s.Cluster = StartEtcdMockCluster(s.T())
}

func (s *EtcdHeartbeatTestSuite) TearDownTest() {
	s.Cluster.Terminate(s.T())
}

func (s *EtcdHeartbeatTestSuite) TestHeartBeatKeepaliveStop() {
	cfg := etcdcliv3.Config{
		Endpoints: s.Cluster.Endpoints,
	}
	cli, err := etcdcliv3.New(cfg)
	s.Nil(err)

	heartbeat := newHeartbeat(cli, "/test/heartbeat", []byte("dongjiang"), 0, false)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ok, err := heartbeat.grantKeepAliveLease(ctx)
	s.Nil(err)
	s.Equal(true, ok)

	go func() {
		heartbeat.keepAlive(ctx)
	}()

	val, err := cli.Get(ctx, "/test/heartbeat")
	s.Nil(err)
	s.Equal(1, len(val.Kvs))

	err = cli.Close()
	s.Nil(err)
	time.Sleep(time.Second)
}

// go test 入口
func TestEtcdHeartbeatTestSuite(t *testing.T) {
	suite.Run(t, new(EtcdHeartbeatTestSuite))
}
