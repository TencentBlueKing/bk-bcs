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

package processor

import (
	rdiscover "bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/ssl"
	"bk-bcs/bcs-common/common/static"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/clbingress"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/model"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/processor"
	svcclient "bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient"
	svccadapter "bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient/adapter"
	"bk-bcs/bcs-services/bcs-gw-controller/pkg/gw"
	"context"
	"fmt"
	"time"
)

// Processor gw controller core processor
type Processor struct {
	opt             *Option
	serviceClient   svcclient.Client
	ingressRegistry clbingress.Registry
	updater         Updater
	doingFlag       *model.AtomicBool
	rootCtx         context.Context
	rootCancel      context.CancelFunc
}

// TLSOption tls option
type TLSOption struct {
	CaFile         string
	ServerCertFile string
	ServerKeyFile  string
	ClientCertFile string
	ClientKeyFile  string
}

// Option options
type Option struct {
	TLSOption
	Port              int
	ServiceRegistry   string
	Cluster           string
	BackendIPType     string
	Kubeconfig        string
	GwZkHosts         string
	GwZkPath          string
	GwBizID           string
	ServiceLabel      map[string]string
	DomainLabelKey    string
	ProxyPortLabelKey string
	PortLabelKey      string
	PathLabelKey      string
	UpdatePeriod      int
	NodeSyncPeriod    int
	SyncPeriod        int
}

// NewProcessor new processor
func NewProcessor(opt *Option) (*Processor, error) {
	// context to control processor run loop and gwClient goroutine(rdiscover goroutine will be close in gwClient)
	ctx, cancel := context.WithCancel(context.Background())
	proc := &Processor{
		rootCtx:    ctx,
		rootCancel: cancel,
	}
	doingFlag := model.NewAtomicBool()
	doingFlag.Set(false)
	proc.opt = opt
	proc.doingFlag = doingFlag
	// service handler change the update flag
	svcHandler := processor.NewAppServiceHandler()
	svcHandler.RegisterProcessor(proc)
	svcClient, err := svccadapter.NewClient(opt.ServiceRegistry, opt.Kubeconfig, svcHandler, opt.SyncPeriod)
	if err != nil {
		return nil, err
	}
	blog.Infof("success to create app service client")

	updater := NewGWUpdater()
	updater.SetOption(opt)
	updater.SetServiceClient(svcClient)

	blog.Infof("start start discovery for gw servers")
	rd := rdiscover.NewRegDiscover(opt.GwZkHosts)
	err = rd.Start()
	if err != nil {
		blog.Errorf("start discovery failed, err %s", err.Error())
		return nil, fmt.Errorf("start discovery failed, err %s", err.Error())
	}

	// create gw client
	var gwClient gw.Interface
	if len(opt.CaFile) != 0 && len(opt.ClientCertFile) != 0 && len(opt.ClientKeyFile) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(opt.CaFile, opt.ClientCertFile, opt.ClientKeyFile, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("load client tls config for gw client failed, err %s", err.Error())
			return nil, fmt.Errorf("load client tls config for gw client failed, err %s", err.Error())
		}
		gwClient = gw.NewClientWithTLS(ctx, opt.Cluster, rd, opt.GwZkPath, tlsConfig)
	} else {
		gwClient = gw.NewClient(ctx, opt.Cluster, rd, opt.GwZkPath)
	}

	blog.Infof("start gw client")
	go gwClient.Run()

	updater.SetGWClient(gwClient)
	proc.updater = updater

	return proc, nil
}

// SetUpdated set update flag
func (p *Processor) SetUpdated() {}

// Run run processor loop
func (p *Processor) Run() {
	updateTick := time.NewTicker(time.Second * time.Duration(p.opt.UpdatePeriod))
	for {
		select {
		case <-p.rootCtx.Done():
			blog.Warnf("stop processor")
			return
		case <-updateTick.C:
			blog.V(3).Infof("update tick rings")

			if !p.doingFlag.Value() {
				blog.V(3).Infof("get update event, going to do some small things...")
				p.doingFlag.Set(true)
				go func() {
					p.Handle()
					p.doingFlag.Set(false)
				}()
				continue
			}
			blog.V(3).Infof("processor is doing, continue")
		}
	}
}

// Handle handle update
func (p *Processor) Handle() {
	err := p.updater.Update()
	if err != nil {
		blog.Errorf("updater do update failed, err %s", err.Error())
	}
}

// Stop stop processor
func (p *Processor) Stop() {
	p.rootCancel()
}
