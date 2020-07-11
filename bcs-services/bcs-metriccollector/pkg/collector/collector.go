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

package collector

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/app/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/pkg/output"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metriccollector/pkg/role"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-metricservice/pkg/types"
)

type cfgMgr struct {
	version string
	cancel  context.CancelFunc
	cfgs    []types.CollectorCfg
}

// collector sync the collector config from remote server
type collector struct {
	output       output.Output
	role         role.RoleInterface
	clusterID    string
	clusterType  string
	namespace    string
	name         string
	lock         sync.RWMutex
	cfgs         cfgMgr
	metricClient *httpclient.HttpClient
	cfg          *config.Config
}

func (cli *collector) syncConfigHandle() error {

	host, hostErr := cli.cfg.Rd.GetMetricServer()
	if nil != hostErr {
		blog.Error("failed to get metric server address, error info is %s ", hostErr.Error())
		return hostErr
	}

	address := fmt.Sprintf("%s/api/v1", host)
	uri := fmt.Sprintf("%s/metric/collector/%s/%s/%s/%s", address, cli.clusterType, cli.clusterID, cli.namespace, cli.name)

	// blog.V(3).Infof("fetch config data %s", uri)
	rsp, rspErr := cli.metricClient.GET(uri, nil, nil)
	if nil != rspErr {
		blog.Errorf("fetch config data failed: %v", rspErr)
		return rspErr
	}

	//blog.Debug("collector config :%s", string(rsp))

	var rst rspRst
	if jserr := json.Unmarshal(rsp, &rst); nil != jserr {
		blog.Errorf("unmarshal failed: %v", jserr)
		return jserr
	}

	// 提取配置，客户端收到配置后必须更新，是否给前段发送配置由后端服务决定
	cfgs := make([]types.CollectorCfg, 0)
	for _, items := range rst.Data.Cfgs {
		for _, subItem := range items {
			cfgs = append(cfgs, subItem)
		}
	}
	cli.lock.Lock()
	cli.cfgs.version = rst.Data.Version
	cli.cfgs.cfgs = cfgs
	cli.lock.Unlock()
	blog.V(5).Infof("after sync version: %+v", cli.cfgs.version)

	return nil
}
func (cli *collector) syncConfig(ctx context.Context, frequency time.Duration) {

	for {
		select {
		case <-ctx.Done():
			blog.Error("exit from the function syncconfig")
			return
		case <-time.After(frequency):

			if err := cli.syncConfigHandle(); nil != err {
				blog.Error("failed to sync config, error info is %s", err.Error())
				// TODO: 不需要退出
			}

		}
	}
}

func (cli *collector) fetchData(ctx context.Context, cfg types.CollectorCfg) {
	blog.Infof("enter a fetchData goroutine, address: %s, version: %s", cfg.Address, cfg.Version)
	header := http.Header{}
	for key, val := range cfg.Head {
		header.Set(key, val)
	}

	parameters := url.Values{}
	for key, val := range cfg.Parameters {
		if 0 != len(key) {
			parameters.Add(key, val)
		}
	}

	address := cfg.Address
	if 0 != len(parameters) {
		address = fmt.Sprintf("%s/?%s", cfg.Address, parameters.Encode())
	}

	tlsCfg, err := cfg.TLSConfig.GetTLSConfig()
	if err != nil {
		blog.Errorf("get metric addr[%s] with dataid[%s] and name[%s] failed, exist now. err: %v", address, cfg.DataID, cfg.Meta.Name, err)
		return
	}

	client := httpclient.NewHttpClient()
	client.SetTlsVerityConfig(tlsCfg)

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30
	}
	client.SetTimeOut(time.Duration(timeout) * time.Second)

	duration := time.Duration(cfg.Frequency) * time.Second
	// 启动
	for {
		select {
		case <-ctx.Done():
			blog.Warn("exit from a fetchData goroutine, address: %s, version: %s", cfg.Address, cfg.Version)
			return
		case <-time.After(duration):

			blog.Infof("start to fetch data, address[%s], name[%s], ns[%s] dataid[%d] ",
				address, cfg.Meta.Name, cfg.Meta.NameSpace, cfg.DataID)

			textData, dataErr := client.Request(address, cfg.Method, header, nil)
			if nil != dataErr {
				blog.Error("failed to fetch data, address[%s], name[%s], ns[%s] dataid[%d] error info is %s",
					address, cfg.Meta.Name, cfg.Meta.NameSpace, cfg.DataID, dataErr.Error())
				continue
			}

			msg := &output.InputMessage{
				DataID: cfg.DataID,
				ObjMeta: output.MessageMeta{
					ObjectMeta: cfg.Meta,
					IP:         cfg.IP,
				}}

			if cfg.MetricType == types.MetricPrometheus {
				data, err := ParsePromTextToOldVersion(bytes.NewReader(textData), cfg.PrometheusConstLabels)
				if err != nil {
					blog.Errorf("parse prometheus metric failed, addr[%s], err: %v ", address, err)
					continue
				}
				msg.Prometheus = data
			} else {
				msg.Data = string(textData)
			}

			if inputErr := cli.output.Input(msg); nil != inputErr {
				blog.Error("failed to send data success, address[%s], name[%s], ns[%s] dataid[%d], err: %v ",
					address, cfg.Meta.Name, cfg.Meta.NameSpace, cfg.DataID, inputErr)
				continue
			}
			blog.V(5).Infof("fetch and send data success, address[%s], name[%s], ns[%s] dataid[%d]",
				address, cfg.Meta.Name, cfg.Meta.NameSpace, cfg.DataID)
		}
	}
}

func (cli *collector) Run(ctx context.Context, cfg *config.Config) error {
	blog.Infof("collector init and sync data first time")
	cli.syncConfigHandle()

	duration := 10
	blog.Infof("start sync data per %d seconds", duration)
	go cli.syncConfig(ctx, time.Duration(duration)*time.Second)

	pool := &agentPool{}
	for {
		select {
		case <-ctx.Done():
			blog.Error("exit from the function fetchdata")
			return nil

		case <-time.After(1 * time.Second):
			// only collector master can fetch data from metric.
			if !cli.role.IsMaster() {
				pool.clean()
				continue
			}

			pool.handle(cli, ctx)
		}
	}
}

type agentPool struct {
	pool   map[string]*agent
	ctx    context.Context
	cancel context.CancelFunc
}

func (ap *agentPool) clean() {
	ap.pool = nil
	if ap.cancel != nil {
		ap.cancel()
		ap.cancel = nil
	}
}

func (ap *agentPool) handle(cli *collector, pCtx context.Context) {
	if ap.pool == nil {
		ap.ctx, ap.cancel = context.WithCancel(pCtx)
		ap.pool = make(map[string]*agent)
	}
	mark := make(map[string]bool, 0)
	for k := range ap.pool {
		mark[k] = true
	}

	cli.lock.RLock()
	for _, cfg := range cli.cfgs.cfgs {
		delete(mark, cfg.CfgKey)

		a := ap.pool[cfg.CfgKey]
		if a == nil {
			a = &agent{}
			ap.pool[cfg.CfgKey] = a
		}
		if a.version == cfg.Version {
			continue
		}
		a.start(cli, ap.ctx, cfg)
	}
	cli.lock.RUnlock()

	for k := range mark {
		a := ap.pool[k]
		if a != nil {
			a.stop()
		}
		delete(ap.pool, k)
	}
}

type agent struct {
	version string
	cancel  context.CancelFunc
}

func (a *agent) start(cli *collector, pCtx context.Context, cfg types.CollectorCfg) {
	a.stop()
	ctx, cancel := context.WithCancel(pCtx)
	a.version = cfg.Version
	a.cancel = cancel
	go cli.fetchData(ctx, cfg)
}

func (a *agent) stop() {
	if a.cancel != nil {
		a.cancel()
	}
}
