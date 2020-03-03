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
	"fmt"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/clbingress"
	clbIngressClient "bk-bcs/bcs-services/bcs-clb-controller/pkg/clbingress/kube"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/client/informers"
	clientSet "bk-bcs/bcs-services/bcs-clb-controller/pkg/client/internalclientset"
	listenerclient "bk-bcs/bcs-services/bcs-clb-controller/pkg/listenerclient"
	"bk-bcs/bcs-services/bcs-clb-controller/pkg/model"
	svcclient "bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient"
	svccadapter "bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient/adapter"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Processor processor for ingresses and services
type Processor struct {
	opt             *Option
	serviceClient   svcclient.Client
	ingressRegistry clbingress.Registry
	listenerClient  listenerclient.Interface
	updater         *Updater
	updateFlag      *model.AtomicBool
	doingFlag       *model.AtomicBool
	stopCh          chan struct{}
}

// Option processor options
type Option struct {
	Port            int
	ServiceRegistry string
	ClbName         string
	NetType         string
	BackendIPType   string
	Namespace       string
	Cluster         string
	Kubeconfig      string
	UpdatePeriod    int
	NodeSyncPeriod  int
	SyncPeriod      int
}

// NewProcessor create processor for clb controller
func NewProcessor(opt *Option) (*Processor, error) {

	stopCh := make(chan struct{})
	proc := &Processor{
		stopCh: stopCh,
	}
	updateFlag := model.NewAtomicBool()
	updateFlag.Set(true)
	doingFlag := model.NewAtomicBool()
	doingFlag.Set(false)
	proc.opt = opt
	proc.updateFlag = updateFlag
	proc.doingFlag = doingFlag
	// create service discovery client
	svcHandler := NewAppServiceHandler()
	svcHandler.RegisterProcessor(proc)
	svcClient, err := svccadapter.NewClient(opt.ServiceRegistry, opt.Kubeconfig, svcHandler, opt.SyncPeriod)
	if err != nil {
		return nil, err
	}
	blog.Infof("success to create app service client")

	//parse config
	var restConfig *rest.Config
	if len(opt.Kubeconfig) == 0 {
		blog.Infof("use in-cluster kube config")
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			blog.Errorf("get incluster config failed, err %s", err.Error())
			return nil, err
		}
	} else {
		//parse configuration
		restConfig, err = clientcmd.BuildConfigFromFlags("", opt.Kubeconfig)
		if err != nil {
			blog.Errorf("create internal client with kubeconfig %s failed, err %s", opt.Kubeconfig, err.Error())
			return nil, err
		}
	}
	//initialize specified client implementation
	cliset, err := clientSet.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("create clientset failed, with rest config %v, err %s", restConfig, err.Error())
		return nil, err
	}
	ingressHandler := NewIngressHandler()
	ingressHandler.RegisterProcessor(proc)
	factory := informers.NewSharedInformerFactory(cliset, time.Duration(opt.SyncPeriod)*time.Second)
	ingressInformer := factory.Clb().V1().ClbIngresses()
	ingressInformer.Informer().AddEventHandler(ingressHandler)
	ingressLister := factory.Clb().V1().ClbIngresses().Lister()
	ingressInterface := cliset.ClbV1()

	listenerHandler := NewListenerHandler()
	listenerInformer := factory.Network().V1().CloudListeners()
	listenerInformer.Informer().AddEventHandler(listenerHandler)
	listenerLister := factory.Network().V1().CloudListeners().Lister()
	listenerInterface := cliset.NetworkV1()
	factory.Start(stopCh)
	blog.Infof("success to clbIngress and cloudListener informer factory")
	factory.WaitForCacheSync(stopCh)
	blog.Infof("success to wait clbIngress and cloudListener synced")

	// create ingress registry
	ingressRegistry, err := clbIngressClient.NewKubeRegistry(opt.ClbName, ingressInformer, ingressLister, ingressInterface)
	if err != nil {
		blog.Errorf("create ingress registry  failed, err %s", err.Error())
		return nil, fmt.Errorf("create ingress registry  failed, err %s", err.Error())
	}
	blog.Infof("success to create ingress registry")

	// create listener client
	listenerClient, err := listenerclient.NewListenerClient(opt.ClbName, listenerInterface, listenerLister)
	if err != nil {
		blog.Errorf("create listener client failed, err %s", err.Error())
		return nil, fmt.Errorf("create listener client failed, err %s", err.Error())
	}
	blog.Infof("success to create listener client")

	// create updater
	updater, err := NewUpdater(opt, svcClient, ingressRegistry, listenerClient)
	if err != nil {
		blog.Errorf("create updater with opt %v failed, err %s", opt, err.Error())
		return nil, fmt.Errorf("create updater with opt %v failed, err %s", opt, err.Error())
	}
	blog.Infof("success to create updater")

	proc.serviceClient = svcClient
	proc.ingressRegistry = ingressRegistry
	proc.listenerClient = listenerClient
	proc.updater = updater
	return proc, nil
}

func (p *Processor) Init() error {
	return p.updater.EnsureLoadBalancer()
}

func (p *Processor) Run() {

	updateTick := time.NewTicker(time.Second * time.Duration(p.opt.UpdatePeriod))
	for {
		select {
		case <-p.stopCh:
			blog.Infof("Processor get close event, exit")
			return
		case <-updateTick.C:
			blog.V(3).Infof("update tick rings")
			if !p.updateFlag.Value() {
				blog.V(3).Infof("no update event happend, continue")
				continue
			}

			if !p.doingFlag.Value() {
				blog.V(3).Infof("get update event, going to do some small things...")
				p.doingFlag.Set(true)
				p.updateFlag.Set(false)
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

func (p *Processor) SetUpdated() {
	p.updateFlag.Set(true)
}

func (p *Processor) Handle() {
	err := p.updater.Update()
	if err != nil {
		blog.Errorf("updater do updater failed, err %s", err.Error())
	}
}

func (p *Processor) Stop() {
	close(p.stopCh)
}
