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

package inspector

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netagent/internal/deviceplugin"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netagent/internal/ipcache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netagent/internal/networkutil"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netagent/internal/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"
	clientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/clientset/versioned"
	cloudclient "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/clientset/versioned/typed/cloud/v1"
	factory "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/informers/externalversions"
	cloudinformer "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/informers/externalversions/cloud/v1"
	cloudlister "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/listers/cloud/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NodeNetworkInspector inspector who watches apiserver NodeNetwork, and set up network interface on node
type NodeNetworkInspector struct {
	address string

	nodeNetwork     *cloudv1.NodeNetwork
	nodeNetworkLock sync.Mutex
	// lock for concurrent Alloc actions
	allocLock sync.Mutex

	option *options.NetAgentOption

	kubeconfig           string
	kubeResyncPeriod     int
	kubeCacheSyncTimeout int

	factory  factory.SharedInformerFactory
	client   cloudclient.CloudV1Interface
	lister   cloudlister.NodeNetworkLister
	informer cloudinformer.NodeNetworkInformer

	ipLister   cloudlister.CloudIPLister
	ipInformer cloudinformer.CloudIPInformer

	netUtil networkutil.Interface

	cloudNetClient pbcloudnet.CloudNetserviceClient

	devicePluginOp *deviceplugin.DevicePluginOp

	ipCache *ipcache.Cache

	stopCh           chan struct{}
	readyForAllocate bool
}

// New create new node network inspector
func New(
	option *options.NetAgentOption,
	cloudNetClient pbcloudnet.CloudNetserviceClient,
	devicePluginOp *deviceplugin.DevicePluginOp) *NodeNetworkInspector {
	return &NodeNetworkInspector{
		option:               option,
		kubeconfig:           option.Kubeconfig,
		kubeResyncPeriod:     option.KubeResyncPeriod,
		kubeCacheSyncTimeout: option.KubeCacheSyncTimeout,
		cloudNetClient:       cloudNetClient,
		devicePluginOp:       devicePluginOp,
		ipCache:              ipcache.NewCache(),
		stopCh:               make(chan struct{}),
		netUtil:              new(networkutil.NetUtil),
	}
}

// Init init node network inspector
func (nni *NodeNetworkInspector) Init() error {
	// get node address
	ifacesStr := strings.Replace(nni.option.Ifaces, ";", ",", -1)
	ifaces := strings.Split(ifacesStr, ",")
	instanceIP, _, err := nni.netUtil.GetAvailableHostIP(ifaces)
	if err != nil {
		blog.Errorf("get node ip failed, err %s", err.Error())
		return fmt.Errorf("get node ip failed, err %s", err.Error())
	}
	nni.address = instanceIP

	if err = nni.initIPCache(nni.address); err != nil {
		return err
	}

	// start reconcile loop for ready eni
	go nni.reconcileLoop()
	blog.Infof("start reconcile loop for ready eni")

	var config *rest.Config
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

	nni.client = clientset.CloudV1()
	nni.factory = factory.NewSharedInformerFactory(clientset, time.Duration(nni.kubeResyncPeriod)*time.Second)
	nni.informer = nni.factory.Cloud().V1().NodeNetworks()
	nni.informer.Informer().AddEventHandler(nni)
	nni.lister = nni.factory.Cloud().V1().NodeNetworks().Lister()

	nni.ipInformer = nni.factory.Cloud().V1().CloudIPs()
	nni.ipLister = nni.factory.Cloud().V1().CloudIPs().Lister()

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
		return fmt.Errorf("wait for cache sync timeout after %d seconds", nni.kubeCacheSyncTimeout)
	case <-syncFlag:
		break
	}
	blog.Infof("wait informer factory cache sync done")

	if nni.nodeNetwork != nil {
		// do dirty ip check
		if err := nni.DirtyCheck(nni.nodeNetwork); err != nil {
			blog.Warnf("do dirty ip check failed, err %s", err.Error())
		}
	}

	return nil
}

func (nni *NodeNetworkInspector) initIPCache(host string) error {
	resp, err := nni.cloudNetClient.ListIP(context.Background(), &pbcloudnet.ListIPsReq{
		Cluster: nni.option.Cluster,
		Host:    host,
		Status:  constant.IPStatusActive,
	})
	if err != nil {
		return fmt.Errorf("list cloud netservice ip, cluster %s, host %s, status %s failed, err %s",
			nni.option.Cluster, host, constant.IPStatusActive, err.Error())
	}
	if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
		return fmt.Errorf("list cloud netservice ip, cluster %s, host %s, status %s failed, errCode %s, errMsg %s",
			nni.option.Cluster, host, constant.IPStatusActive, resp.ErrCode, resp.ErrMsg)
	}
	for _, ip := range resp.Ips {
		nni.ipCache.PutEniIP(ip.EniID, ip)
	}
	return nil
}

func (nni *NodeNetworkInspector) reconcileLoop() error {
	timer := time.NewTicker(time.Duration(nni.option.ReconcileInterval) * time.Second)
	for {
		select {
		case <-timer.C:
			nni.nodeNetworkLock.Lock()
			if nni.nodeNetwork == nil {
				nni.nodeNetworkLock.Unlock()
				continue
			}
			// get all ip rules
			rules, err := nni.netUtil.RuleList()
			if err != nil {
				blog.Warnf("list rule failed, err %s", err.Error())
				nni.nodeNetworkLock.Unlock()
				continue
			}
			for _, eniObj := range nni.nodeNetwork.Status.Enis {
				if eniObj.Status == constant.NodeNetworkEniStatusReady {
					if err := nni.netUtil.SetUpNetworkInterface(
						eniObj.Address.IP,
						eniObj.EniSubnetCidr,
						eniObj.MacAddress,
						eniObj.EniIfaceName,
						eniObj.RouteTableID,
						nni.option.EniMTU,
						rules,
					); err != nil {
						blog.Warnf("sync network interface failed, err %s", err.Error())
						nni.nodeNetworkLock.Unlock()
						continue
					}
					blog.Infof("reconcile eni %s/%s successfully", eniObj.EniID, eniObj.EniName)
				}
			}
			nni.nodeNetworkLock.Unlock()

		case <-nni.stopCh:
			blog.Warnf("stop chan recevied, exit reconcile loop")
			break
		}
	}
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
	nni.nodeNetworkLock.Lock()
	nni.nodeNetwork = nodenetwork
	nni.nodeNetworkLock.Unlock()

	err := nni.reconcileNodeNetwork(nodenetwork)
	if err != nil {
		blog.Errorf("reconcile NodeNetwork failed, err %s", err.Error())
		return
	}
	nni.setDevicePlugin()
}

// OnUpdate update event
func (nni *NodeNetworkInspector) OnUpdate(oldObj, newObj interface{}) {
	_, okOld := oldObj.(*cloudv1.NodeNetwork)
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

	nni.nodeNetworkLock.Lock()
	nni.nodeNetwork = newNode
	nni.nodeNetworkLock.Unlock()

	// periodically reconcile NodeNetwork
	err := nni.reconcileNodeNetwork(newNode)
	if err != nil {
		blog.Errorf("reconcile NodeNetwork failed, err %s", err.Error())
		return
	}
	nni.setDevicePlugin()
}

// OnDelete delete event
func (nni *NodeNetworkInspector) OnDelete(obj interface{}) {

}

// Lock lock nodenetwork
func (nni *NodeNetworkInspector) Lock() {
	nni.allocLock.Lock()
}

// Unlock unlock nodenetwork
func (nni *NodeNetworkInspector) Unlock() {
	nni.allocLock.Unlock()
}

// GetNodeNetwork get node network
func (nni *NodeNetworkInspector) GetNodeNetwork() *cloudv1.NodeNetwork {
	nni.nodeNetworkLock.Lock()
	nodeNetwork := nni.nodeNetwork
	nni.nodeNetworkLock.Unlock()
	return nodeNetwork
}

// GetCluster get clusterID
func (nni *NodeNetworkInspector) GetCluster() string {
	return nni.option.Cluster
}

// GetIPCache get ip cache
func (nni *NodeNetworkInspector) GetIPCache() *ipcache.Cache {
	return nni.ipCache
}

// setDevicePlugin set devices of device plugin
func (nni *NodeNetworkInspector) setDevicePlugin() {
	if nni.devicePluginOp != nil {
		if nni.nodeNetwork != nil {
			limitTotal := 0
			for _, eni := range nni.nodeNetwork.Status.Enis {
				if eni.Status == constant.NodeNetworkEniStatusReady {
					limitTotal += nni.nodeNetwork.Spec.IPNumPerENI
				}
			}
			nni.devicePluginOp.GetPlugin().SetDeviceLimit(limitTotal)
		} else {
			nni.devicePluginOp.GetPlugin().SetDeviceLimit(0)
		}
	}
}

// reconcile node network, set up eni, set route table
func (nni *NodeNetworkInspector) reconcileNodeNetwork(nodenetwork *cloudv1.NodeNetwork) error {
	if len(nodenetwork.Status.Enis) == 0 {
		blog.Infof("no enis on node %s", nodenetwork.GetName())
		return nil
	}
	lastIndex := len(nodenetwork.Status.Enis) - 1
	eniObj := nodenetwork.Status.Enis[lastIndex]
	switch eniObj.Status {
	case constant.NodeNetworkEniStatusNotReady:
		eniObj.Status = constant.NodeNetworkEniStatusInitializing
		_, err := nni.client.NodeNetworks(nodenetwork.GetNamespace()).
			Update(context.Background(), nodenetwork, metav1.UpdateOptions{})
		if err != nil {
			blog.Errorf("change eni %s to status %s failed, err %s", eniObj.EniName, eniObj.Status, err.Error())
			return nil
		}
		blog.Infof("change eni %s to status %s successfully", eniObj.EniName, eniObj.Status)
		return nil

	case constant.NodeNetworkEniStatusInitializing:
		if err := nni.reconcileENI(nodenetwork, lastIndex); err != nil {
			return err
		}

	case constant.NodeNetworkEniStatusReady, constant.NodeNetworkEniStatusCleaned,
		constant.NodeNetworkEniStatusDeleting:
		// do nothing

	case constant.NodeNetworkEniStatusCleaning:
		if err := nni.cleanENI(nodenetwork, lastIndex); err != nil {
			return err
		}
		return nil

	default:
		blog.Errorf("error status %s for eni name: %s, id: %s", eniObj.Status, eniObj.EniName, eniObj.EniID)
		return fmt.Errorf("error status %s for eni name: %s, id: %s", eniObj.Status, eniObj.EniName, eniObj.EniID)
	}
	return nil
}

// reconcileENI set up eni
func (nni *NodeNetworkInspector) reconcileENI(nodenetwork *cloudv1.NodeNetwork, index int) error {
	// get all ip rules
	rules, err := nni.netUtil.RuleList()
	if err != nil {
		blog.Errorf("list rule failed, err %s", err.Error())
		return fmt.Errorf("list rule failed, err %s", err.Error())
	}
	eniObj := nodenetwork.Status.Enis[index]
	if err = nni.netUtil.SetUpNetworkInterface(
		eniObj.Address.IP,
		eniObj.EniSubnetCidr,
		eniObj.MacAddress,
		eniObj.EniIfaceName,
		eniObj.RouteTableID,
		nni.option.EniMTU,
		rules,
	); err != nil {
		blog.Errorf("sync network interface failed, err %s", err.Error())
		return err
	}
	eniObj.Status = constant.NodeNetworkEniStatusReady
	_, err = nni.client.NodeNetworks(nodenetwork.GetNamespace()).
		Update(context.Background(), nodenetwork, metav1.UpdateOptions{})
	if err != nil {
		blog.Errorf("change eni %s to status %s failed, err %s", eniObj.EniName, eniObj.Status, err.Error())
		return nil
	}
	blog.Infof("change eni %s to status %s successfully", eniObj.EniName, eniObj.Status)
	return nil
}

func (nni *NodeNetworkInspector) checkENI(eniObj *cloudv1.ElasticNetworkInterface) (bool, error) {
	resp, err := nni.cloudNetClient.ListIP(context.Background(), &pbcloudnet.ListIPsReq{
		Seq:      common.TimeSequence(),
		Cluster:  nni.option.Cluster,
		EniID:    eniObj.EniID,
		SubnetID: eniObj.EniSubnetID,
		Status:   constant.IPStatusActive,
	})
	if err != nil {
		return false, fmt.Errorf("failed to list ip by cluster %s, eni %s, subnetid %s, status %s, err %s",
			nni.option.Cluster, eniObj.EniID, eniObj.EniSubnetID, constant.IPStatusActive, err.Error())
	}
	if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
		return false, fmt.Errorf(
			"failed to list ip by cluster %s, eni %s, subnetid %s, status %s, errCode %s, errMsg %s",
			nni.option.Cluster, eniObj.EniID, eniObj.EniSubnetID, constant.IPStatusActive, resp.ErrCode, resp.ErrMsg)
	}
	if len(resp.Ips) != 0 {
		return true, nil
	}
	resp, err = nni.cloudNetClient.ListIP(context.Background(), &pbcloudnet.ListIPsReq{
		Seq:      common.TimeSequence(),
		Cluster:  nni.option.Cluster,
		EniID:    eniObj.EniID,
		SubnetID: eniObj.EniSubnetID,
		Status:   constant.IPStatusDeleting,
	})
	if err != nil {
		return false, fmt.Errorf("failed to list ip by cluster %s, eni %s, subnetid %s, status %s, err %s",
			nni.option.Cluster, eniObj.EniID, eniObj.EniSubnetID, constant.IPStatusDeleting, err.Error())
	}
	if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
		return false, fmt.Errorf(
			"failed to list ip by cluster %s, eni %s, subnetid %s, status %s, errCode %s, errMsg %s",
			nni.option.Cluster, eniObj.EniID, eniObj.EniSubnetID, constant.IPStatusDeleting, resp.ErrCode, resp.ErrMsg)
	}
	if len(resp.Ips) != 0 {
		return true, nil
	}
	return false, nil
}

func (nni *NodeNetworkInspector) cleanENI(nodenetwork *cloudv1.NodeNetwork, index int) error {
	eniObj := nodenetwork.Status.Enis[index]
	foundIPs, err := nni.checkENI(eniObj)
	if err != nil {
		return err
	}
	if foundIPs {
		return fmt.Errorf("cannot clean eni %s/%s with active ip", eniObj.EniID, eniObj.EniName)
	}

	rules, err := nni.netUtil.RuleList()
	if err != nil {
		blog.Errorf("list rule failed, err %s", err.Error())
		return fmt.Errorf("list rule failed, err %s", err.Error())
	}
	// set down eni
	err = nni.netUtil.SetDownNetworkInterface(
		eniObj.Address.IP,
		eniObj.EniSubnetCidr,
		eniObj.MacAddress,
		eniObj.EniIfaceName,
		eniObj.RouteTableID,
		rules,
	)
	if err != nil {
		blog.Errorf("set down network interface failed, err %s", err.Error())
		return err
	}
	eniObj.Status = constant.NodeNetworkEniStatusCleaned
	_, err = nni.client.NodeNetworks(nodenetwork.GetNamespace()).
		Update(context.Background(), nodenetwork, metav1.UpdateOptions{})
	if err != nil {
		blog.Errorf("set eni %s of node %s to status %s failed, err %s",
			eniObj.EniName, nodenetwork.GetName(), eniObj.Status, err.Error())
	}
	blog.Infof("set eni %s of node %s to status %s successfully",
		eniObj.EniName, nodenetwork.GetName(), eniObj.Status)
	return nil
}
