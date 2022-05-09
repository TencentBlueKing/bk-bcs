/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"os"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/cmd"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/google/uuid"
	"github.com/micro/go-micro/v2/sync"
	"github.com/micro/go-micro/v2/sync/etcd"
)

func main() {
	opts := cmd.NewDataManagerOptions()
	conf.Parse(opts)
	blog.InitLogs(opts.LogConfig)
	defer blog.CloseLogs()
	ctx, cancel := context.WithCancel(context.Background())
	dataManager := cmd.NewServer(ctx, cancel, opts)
	if err := dataManager.Init(); err != nil {
		blog.Fatalf("init data manager failed, err %s", err.Error())
	}
	dataManager.Run()
	go workAsConsumer(ctx, dataManager)
	leader := becomeLeader(ctx, opts)
	go workAsProducer(ctx, dataManager, leader, opts)
	<-ctx.Done()
}

func becomeLeader(ctx context.Context, opt *cmd.DataManagerOptions) sync.Leader {
	// ready to campaign
	etcdTLS, err := opt.GetEtcdRegistryTLS()
	if err != nil {
		blog.Errorf("etcd configuration err: %s, service exit.", err.Error())
		os.Exit(-1)
	}

	etcdSync := etcd.NewSync(
		sync.WithTLS(etcdTLS),
		sync.WithContext(ctx),
		sync.Nodes(strings.Split(opt.Etcd.EtcdEndpoints, ",")...),
		sync.Prefix(common.ServiceDomain),
	)
	id := uuid.New().String()
	blog.Infof("Node %s waiting to become a leader...", id)
	leader, err := etcdSync.Leader(id, sync.LeaderContext(ctx))
	if err != nil {
		blog.Errorf("create etcd leader election failed: %s, service exit.", err.Error())
		os.Exit(-1)
	}
	blog.Infof("I'm leader!")
	return leader
}

func workAsProducer(ctx context.Context, server *cmd.Server, leader sync.Leader, opt *cmd.DataManagerOptions) {
	server.RunAsProducer()
	blog.Infof("listen leader status")
	lost := leader.Status()
	blog.Infof("init ticker")
	select {
	case <-lost:
		blog.Infof("lost leader")
		server.StopProducer()
		leader = becomeLeader(ctx, opt)
		go workAsProducer(ctx, server, leader, opt)
	case <-ctx.Done():
		err := leader.Resign()
		if err != nil {
			blog.Errorf("ctx done, leader resign error :%v", err)
		} else {
			blog.Infof("ctx done, leader resign")
		}
	}
}

func workAsConsumer(ctx context.Context, server *cmd.Server) {
	server.RunAsConsumer()
	<-ctx.Done()
	blog.Infof("ctx done, exit consumer")
	server.StopConsumer()
}
