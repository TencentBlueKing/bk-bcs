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
 */

package sync

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ipmasq-cidrsync/internal/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
	cloudcliset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/clientset/versioned"
	cloudfactory "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/informers/externalversions"
	cloudlister "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/listers/cloud/v1"
	"gopkg.in/yaml.v2"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8scorecliset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Synchronizer synchronizer of ip masq configmap
type Synchronizer struct {
	opt        *options.SyncOption
	bcsLister  cloudlister.NodeNetworkLister
	coreClient k8scorecliset.Interface

	isUpdate           bool
	syncIntervalSecond int
	syncChan           chan struct{}
	stopChan           chan struct{}

	syncMutex sync.Mutex
}

// New create synchronizer
func New(opt *options.SyncOption) (*Synchronizer, error) {
	if opt == nil {
		return nil, fmt.Errorf("options cannot be empty")
	}
	return &Synchronizer{
		opt:                opt,
		syncIntervalSecond: opt.SyncIntervalSecond,
		syncChan:           make(chan struct{}, 2),
		stopChan:           make(chan struct{}),
	}, nil
}

// initRestConfig return rest config
func (s *Synchronizer) initRestConfig() (*rest.Config, error) {
	var restConfig *rest.Config
	var err error
	if len(s.opt.Kubeconfig) == 0 {
		blog.Infof("use in-cluster kubeconfig")
		restConfig, err = rest.InClusterConfig()
	} else {
		restConfig, err = clientcmd.BuildConfigFromFlags("", s.opt.Kubeconfig)
	}
	if err != nil {
		blog.Errorf("get rest config failed, err %s", err.Error())
		return nil, fmt.Errorf("get rest config failed, err %s", err.Error())
	}
	return restConfig, nil
}

// initKubeCoreClient init kube core object client
func (s *Synchronizer) initKubeCoreClient(restConfig *rest.Config) error {
	cliSet, err := k8scorecliset.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("build k8s core client set failed, err %s", err.Error())
		return fmt.Errorf("build k8s core client set failed, err %s", err.Error())
	}
	s.coreClient = cliSet
	return nil
}

// initBcsClient init bcs object client
func (s *Synchronizer) initBcsClient(restConfig *rest.Config) error {
	bcsCliSet, err := cloudcliset.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("build bcs client set failed, err %s", err.Error())
		return fmt.Errorf("build bcs client set failed, err %s", err.Error())
	}
	blog.Infof("create bcs cloud informer factory")
	factory := cloudfactory.NewSharedInformerFactory(bcsCliSet, 0)
	nodeNetworkInformer := factory.Cloud().V1().NodeNetworks()
	nodeNetworkLister := nodeNetworkInformer.Lister()
	nodeNetworkInformer.Informer().AddEventHandler(s)
	s.bcsLister = nodeNetworkLister
	factory.Start(s.stopChan)

	// wait data sync
	syncFlag := make(chan struct{})
	go func() {
		blog.Infof("wait for bcs cloud informer factory cache sync")
		factory.WaitForCacheSync(s.stopChan)
		close(syncFlag)
	}()
	select {
	case <-syncFlag:
		blog.Infof("bcs cloud informer synced")
		break
	}
	return nil
}

// Init init k8s client
func (s *Synchronizer) Init() error {
	restConfig, err := s.initRestConfig()
	if err != nil {
		return err
	}
	if err = s.initKubeCoreClient(restConfig); err != nil {
		return err
	}
	if err = s.initBcsClient(restConfig); err != nil {
		return err
	}
	return nil
}

// get eni cidr from nodenetwork
func (s *Synchronizer) getNodeNetworkCIDR() ([]string, error) {
	selector := k8slabels.Everything()
	nodeNetList, err := s.bcsLister.NodeNetworks(constant.CloudCrdNamespaceBcsSystem).List(selector)
	if err != nil {
		blog.Errorf("get nodenetwork list failed, err %s", err.Error())
		return nil, fmt.Errorf("get nodenetwork list failed, err %s", err.Error())
	}
	var ipList []string
	for _, nodeNet := range nodeNetList {
		for _, eni := range nodeNet.Status.Enis {
			ipList = append(ipList, eni.EniSubnetCidr)
		}
	}
	return ipList, nil
}

// update config in configmap of ip masq agent
func (s *Synchronizer) updateIPMasqCidrConfig() error {
	ns := s.opt.IPMasqConfigmapNamespace
	name := s.opt.IPMasqConfigmapName
	cm, err := s.coreClient.CoreV1().ConfigMaps(ns).Get(
		context.Background(), name, k8smetav1.GetOptions{})
	if err != nil {
		blog.Errorf("get configmap %s/%s failed, err %s", ns, name, err.Error())
		return fmt.Errorf("get configmap %s/%s failed, err %s", ns, name, err.Error())
	}

	configData := cm.Data["config"]
	masqConfig := &IPMasqConfig{}
	if err = yaml.UnmarshalStrict([]byte(configData), masqConfig); err != nil {
		blog.Errorf("unmarshal config of ip masq agent failed, err %s", err.Error())
		return fmt.Errorf("unmarshal config of ip masq agent failed, err %s", err.Error())
	}

	cidrList, err := s.getNodeNetworkCIDR()
	if err != nil {
		return err
	}

	isUpdate := false
	for _, cidr := range cidrList {
		isFound1, isFound2 := false, false
		for _, nonMasqCidr := range masqConfig.NonMasqueradeCIDRs {
			if cidr == nonMasqCidr {
				isFound1 = true
			}
		}
		for _, nonMasqSrcCidr := range masqConfig.NonMasqueradeSrcCIDRs {
			if cidr == nonMasqSrcCidr {
				isFound2 = true
			}
		}
		if !isFound1 {
			masqConfig.NonMasqueradeCIDRs = append(masqConfig.NonMasqueradeCIDRs, cidr)
			isUpdate = true
		}
		if !isFound2 {
			masqConfig.NonMasqueradeSrcCIDRs = append(masqConfig.NonMasqueradeSrcCIDRs, cidr)
			isUpdate = true
		}
	}
	if !isUpdate {
		blog.Infof("cidr list no change for %v", cidrList)
		return nil
	}
	mashalData, err := yaml.Marshal(masqConfig)
	if err != nil {
		blog.Errorf("marshal config data failed, err %s", err.Error())
		return fmt.Errorf("marshal config data failed, err %s", err.Error())
	}

	// update content of ip-masq-agent configmap
	cm.Data["config"] = string(mashalData)
	_, err = s.coreClient.CoreV1().ConfigMaps(ns).Update(context.Background(), cm, k8smetav1.UpdateOptions{})
	if err != nil {
		blog.Errorf("update configmap %s/%s failed, err %s", cm.GetName(), cm.GetNamespace(), err.Error())
		return fmt.Errorf("update configmap %s/%s failed, err %s", cm.GetName(), cm.GetNamespace(), err.Error())
	}
	blog.Infof("update configmap %s/%s to data config %s successfully",
		cm.GetName(), cm.GetNamespace(), cm.Data["config"])
	return nil
}

// Start start get nodenetwork and sync configmap
func (s *Synchronizer) Start() {
	ticker := time.NewTicker(time.Duration(s.syncIntervalSecond) * time.Second)
	for {
		select {
		case <-ticker.C:
			s.syncMutex.Lock()
			if s.isUpdate {
				s.isUpdate = false
				if err := s.updateIPMasqCidrConfig(); err != nil {
					blog.Errorf("do sync ip masq configmap failed, err %s", err.Error())
				}
			}
			s.syncMutex.Unlock()
		case <-s.stopChan:
			blog.Warnf("sync loop ask to stop")
			return
		}
	}
}

// Stop stop the sync loop
func (s *Synchronizer) Stop() {
	blog.Warnf("close stop channel")
	close(s.stopChan)
	// sleep to wait sync loop exit
	time.Sleep(1 * time.Second)
}

// OnAdd get add event for node network, implements event handler for informer
func (s *Synchronizer) OnAdd(add interface{}) {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()
	s.isUpdate = true
}

// OnUpdate get update event for node network, implements event handler for informer
func (s *Synchronizer) OnUpdate(old, new interface{}) {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()
	s.isUpdate = true
}

// OnDelete get delete event for node network, implements event handler for informer
func (s *Synchronizer) OnDelete(del interface{}) {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()
	s.isUpdate = true
}
