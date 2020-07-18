/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package inspector

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
	clientset "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned"
	cloudclient "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned/typed/cloud/v1"
	factory "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/informers/externalversions"
	cloudinformer "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/informers/externalversions/cloud/v1"
	cloudlister "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/listers/cloud/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netagent/internal/networkutil"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netagent/internal/options"
)

var (
	FINALIZER_NAME = "nodeagent.cloud.bkbcs.tencent.com"
)

// NodeNetworkInspector inspector who watches apiserver NodeNetwork, and set up network interface on node
type NodeNetworkInspector struct {
	address string

	nodeNetwork     *cloudv1.NodeNetwork
	nodeNetworkLock sync.Mutex

	option *options.NetAgentOption

	kubeconfig           string
	kubeResyncPeriod     int
	kubeCacheSyncTimeout int

	factory  factory.SharedInformerFactory
	client   cloudclient.CloudV1Interface
	lister   cloudlister.NodeNetworkLister
	informer cloudinformer.NodeNetworkInformer

	netUtil networkutil.Interface

	stopCh           chan struct{}
	readyForAllocate bool
}

// New create new node network inspector
func New(option *options.NetAgentOption) *NodeNetworkInspector {
	return &NodeNetworkInspector{
		option:               option,
		kubeconfig:           option.Kubeconfig,
		kubeResyncPeriod:     option.KubeResyncPeriod,
		kubeCacheSyncTimeout: option.KubeCacheSyncTimeout,
		stopCh:               make(chan struct{}),
		netUtil:              new(networkutil.NetUtil),
	}
}

// Init init node network inspector
func (nni *NodeNetworkInspector) Init() error {
	var config *rest.Config
	var err error
	// when out-of-cluster, kubeconfig must be
	if len(nni.kubeconfig) != 0 {
		config, err = clientcmd.BuildConfigFromFlags("", nni.kubeconfig)
		if err != nil {
			blog.Errorf("build config from kubeconfig %s failed, err %s", nni.kubeconfig, err.Error())
			return err
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			blog.Errorf("build incluster config failed, err %s", err.Error())
			return err
		}
	}
	clientset, err := clientset.NewForConfig(config)
	if err != nil {
		blog.Errorf("build clientset failed, err %s", err.Error())
		return err
	}
	nni.factory = factory.NewSharedInformerFactory(clientset, time.Duration(nni.kubeResyncPeriod)*time.Second)
	nni.informer = nni.factory.Cloud().V1().NodeNetworks()
	nni.informer.Informer().AddEventHandlerWithResyncPeriod(nni, time.Duration(nni.kubeCacheSyncTimeout)*time.Second)
	nni.lister = nni.factory.Cloud().V1().NodeNetworks().Lister()
	nni.client = clientset.CloudV1()

	// start informers
	nni.factory.Start(nni.stopCh)
	syncFlag := make(chan struct{})
	go func() {
		blog.Infof("wait for informer factory cache sync")
		nni.factory.WaitForCacheSync(nni.stopCh)
		close(syncFlag)
	}()
	select {
	case <-time.After(time.Duration(nni.kubeCacheSyncTimeout) * time.Second):
		return fmt.Errorf("wait for cache sync timeout after %s seconds", nni.kubeCacheSyncTimeout)
	case <-syncFlag:
		break
	}
	blog.Infof("wait informer factory cache sync done")

	// get node address
	ifacesStr := strings.Replace(nni.option.Ifaces, ";", ",", -1)
	ifaces := strings.Split(ifacesStr, ",")
	instanceIP, _, err := nni.netUtil.GetAvailableHostIP(ifaces)
	if err != nil {
		blog.Errorf("get node ip failed, err %s", err.Error())
		return fmt.Errorf("get node ip failed, err %s", err.Error())
	}
	nni.address = instanceIP

	return nil
}

// OnAdd add event
func (nni *NodeNetworkInspector) OnAdd(obj interface{}) {
	nodenetwork, ok := obj.(*cloudv1.NodeNetwork)
	if !ok {
		blog.Warnf("received invalid add obj")
		return
	}

	if nodenetwork.Spec.NodeAddress != nni.address {
		return
	}

	blog.Infof("node network add: %+v", nodenetwork)
	err := nni.reconcileNodeNetwork(nodenetwork)
	if err != nil {
		blog.Errorf("reconcile NodeNetwork failed, err %s", err.Error())
	}

	nni.nodeNetworkLock.Lock()
	nni.nodeNetwork = nodenetwork
	nni.readyForAllocate = true
	nni.nodeNetworkLock.Unlock()
}

// OnUpdate update event
func (nni *NodeNetworkInspector) OnUpdate(oldObj, newObj interface{}) {
	oldNode, okOld := oldObj.(*cloudv1.NodeNetwork)
	if !okOld {
		blog.Warnf("received invalid old obj")
		return
	}
	newNode, okNew := newObj.(*cloudv1.NodeNetwork)
	if !okNew {
		blog.Warnf("received invalid new obj")
		return
	}

	if newNode.Spec.NodeAddress != nni.address {
		return
	}

	if reflect.DeepEqual(oldNode, newNode) {
		blog.Warnf("old and new nodenetwork are the same, no need update")
		return
	}

	if !newNode.DeletionTimestamp.IsZero() {
		nni.nodeNetworkLock.Lock()
		nni.nodeNetwork = nil
		nni.readyForAllocate = false
		nni.nodeNetworkLock.Unlock()

		if err := nni.cleanNodeNetwork(newNode); err != nil {
			blog.Warnf("clean node network failed, err %s", err.Error())
			return
		}
	}
}

// OnDelete delete event
func (nni *NodeNetworkInspector) OnDelete(obj interface{}) {

}

// GetNodeNetwork get node network
func (nni *NodeNetworkInspector) GetNodeNetwork() *cloudv1.NodeNetwork {
	nni.nodeNetworkLock.Lock()
	nodeNetwork := nni.nodeNetwork
	nni.nodeNetworkLock.Unlock()
	return nodeNetwork
}

// reconcile node network, set up eni, set route table
func (nni *NodeNetworkInspector) reconcileNodeNetwork(nodenetwork *cloudv1.NodeNetwork) error {
	// get all ip rules
	rules, err := nni.netUtil.RuleList()
	if err != nil {
		blog.Errorf("list rule failed, err %s", err.Error())
		return fmt.Errorf("list rule failed, err %s", err.Error())
	}
	// set up eni
	if nodenetwork.Status.FloatingIPEni != nil {
		netiface := nodenetwork.Status.FloatingIPEni.Eni
		err := nni.netUtil.SetUpNetworkInterface(
			netiface.Address.IP,
			netiface.EniSubnetCidr,
			netiface.MacAddress,
			netiface.EniIfaceName,
			netiface.RouteTableID,
			nni.option.EniMTU,
			rules,
		)
		if err != nil {
			blog.Errorf("sync network interface failed, err %s", err.Error())
			return err
		}
	}

	nodenetwork.Finalizers = append(nodenetwork.Finalizers, FINALIZER_NAME)
	_, err = nni.client.NodeNetworks(nodenetwork.GetNamespace()).Update(context.TODO(), nodenetwork, metav1.UpdateOptions{})
	if err != nil {
		blog.Errorf("add finalizer to nodenetwork failed, err %s", err.Error())
	}

	return nil
}

func (nni *NodeNetworkInspector) cleanNodeNetwork(nodenetwork *cloudv1.NodeNetwork) error {

	if containsString(nodenetwork.Finalizers, FINALIZER_NAME) {
		rules, err := nni.netUtil.RuleList()
		if err != nil {
			blog.Errorf("list rule failed, err %s", err.Error())
			return fmt.Errorf("list rule failed, err %s", err.Error())
		}
		// set down eni
		if nodenetwork.Status.FloatingIPEni != nil {
			netiface := nodenetwork.Status.FloatingIPEni.Eni
			err := nni.netUtil.SetDownNetworkInterface(
				netiface.Address.IP,
				netiface.EniSubnetCidr,
				netiface.MacAddress,
				netiface.EniIfaceName,
				netiface.RouteTableID,
				rules,
			)
			if err != nil {
				blog.Errorf("set down network interface failed, err %s", err.Error())
				return err
			}
		}

		nodenetwork.Finalizers = removeString(nodenetwork.Finalizers, FINALIZER_NAME)
		_, err = nni.client.NodeNetworks(nodenetwork.GetNamespace()).Update(context.TODO(), nodenetwork, metav1.UpdateOptions{})
		if err != nil {
			blog.Errorf("add finalizer to nodenetwork failed, err %s", err.Error())
		}
	}

	return nil
}
