/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/service"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/app"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/utils"

	"github.com/google/uuid"
	"github.com/micro/go-micro/v2/sync"
	"github.com/micro/go-micro/v2/sync/etcd"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//discovery now is designed for stage that bkbcs routes http/https traffics.
// so layer 4 traffic forwarding to bkbcs backends are temporarily out of our consideration.

func main() {
	//init command line option
	opt := app.NewServerOptions()
	//option parsing
	conf.Parse(opt)
	//init logger
	blog.InitLogs(opt.LogConfig)
	defer blog.CloseLogs()

	cxt := service.SetupSignalContext()
	leader := becomeLeader(cxt, opt)
	go work(cxt, leader, opt)
	<-cxt.Done()
	time.Sleep(time.Second * 3)
}

func becomeLeader(cxt context.Context, opt *app.ServerOptions) sync.Leader {
	etcdTLS, err := opt.GetEtcdRegistryTLS()
	if err != nil {
		fmt.Printf("etcd configuration err: %s, service exit.\n", err.Error())
		os.Exit(-1)
	}
	//ready to campaign
	synch := etcd.NewSync(
		sync.Nodes(strings.Split(opt.Etcd.Address, ",")...),
		sync.WithTLS(etcdTLS),
		sync.Prefix("gatewaydiscovery.bkbcs.tencent.com"),
		sync.WithContext(cxt),
	)
	id := uuid.New().String()
	blog.Infof("Node %s waiting to become a leader...", id)
	leader, err := synch.Leader(id, sync.LeaderContext(cxt))
	if err != nil {
		fmt.Printf("create etcd leader election failed: %s, service exit.\n", err.Error())
		os.Exit(-1)
	}
	blog.Infof("I'm leader!")
	return leader
}

func work(cxt context.Context, leader sync.Leader, opt *app.ServerOptions) {
	//create app, init
	svc := app.New(cxt)
	if err := svc.Init(opt); err != nil {
		fmt.Printf("bcs-gateway-discovery init failed, %s. service exit\n", err.Error())
		os.Exit(-1)
	}
	// run prometheus server
	runPrometheusServer(opt)
	go leaderTracing(cxt, leader, svc, opt)
	//start running
	if err := svc.Run(); err != nil {
		fmt.Printf("bcs-gateway-discovery enter running loop failed, %s\n", err.Error())
		os.Exit(-1)
	}
}

func leaderTracing(cxt context.Context, leader sync.Leader, svc *app.DiscoveryServer, opt *app.ServerOptions) {
	lost := leader.Status()
	select {
	case <-cxt.Done():
		leader.Resign()
		blog.Infof("catch signal, ready to exit leader tracing")
		return
	case <-lost:
		blog.Infof("I lost leader, clean all works and exit")
		svc.Stop()
		//try to campaign leader & re-Init
		leader := becomeLeader(cxt, opt)
		go work(cxt, leader, opt)
		return
	}
}

// runPrometheusServer run prometheus metrics server
func runPrometheusServer(opt *app.ServerOptions) {
	http.Handle("/metrics", promhttp.Handler())
	addr := opt.Address + ":" + strconv.Itoa(int(opt.MetricPort))
	go http.ListenAndServe(addr, nil)
}
