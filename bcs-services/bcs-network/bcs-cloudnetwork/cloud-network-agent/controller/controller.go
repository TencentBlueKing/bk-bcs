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

package controller

import (
	"context"
	"strconv"
	"sync"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/cloud-network-agent/options"
	cloud "bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/apis/cloud/v1"
	"bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/eni"
	"bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/networkutil"
	"bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/nodenetwork"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CRD_NAMESPACES = "bcs-system"
	ENI_PREFIX     = "eni"
)

// NetworkController controller for cloud network
type NetworkController struct {
	hostip   string
	hostname string

	eniNum int
	ipNum  int

	options *options.NetworkOption

	nodeNetwork     *cloud.NodeNetwork
	nodeNetworkMutx sync.Mutex

	nodeNetClient nodenetwork.Interface

	eniClient eni.Interface
	netUtil   networkutil.Interface
}

// New create new network controller
func New(op *options.NetworkOption, nodeNetClient nodenetwork.Interface,
	eniClient eni.Interface, netUtil networkutil.Interface) *NetworkController {
	return &NetworkController{
		options:       op,
		nodeNetClient: nodeNetClient,
		eniClient:     eniClient,
		netUtil:       netUtil,
	}
}

// OnAdd add event
func (nc *NetworkController) OnAdd(obj interface{}) {
	// TODO:
}

// OnUpdate update event
func (nc *NetworkController) OnUpdate(oldObj, newObj interface{}) {
	// TODO:
}

// OnDelete delete event
func (nc *NetworkController) OnDelete(obj interface{}) {
	// TODO:
}

// GetNodeNetwork get node network config from etcd
func (nc *NetworkController) GetNodeNetwork() error {
	nodeNetwork, err := nc.nodeNetClient.Get(CRD_NAMESPACES, nc.hostname)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Warnf("node network not found in etcd")
			return nil
		}
		blog.Errorf("get node network failed, err %s", err.Error())
		return err
	}
	nc.nodeNetworkMutx.Lock()
	nc.nodeNetwork = nodeNetwork
	nc.nodeNetworkMutx.Unlock()
	return nil
}

func (nc *NetworkController) getEniQuota() error {
	eniNum, ipNum, err := nc.eniClient.GetENILimit()
	if err != nil {
		blog.Infof("get eni quota limit, err %s", err.Error())
	}

	if nc.options.EniNum == 0 || int(nc.options.EniNum) > eniNum {
		nc.eniNum = eniNum
	} else {
		nc.eniNum = int(nc.options.EniNum)
	}

	if nc.options.IPNumPerEni == 0 || int(nc.options.IPNumPerEni) > ipNum {
		nc.ipNum = ipNum
	} else {
		nc.ipNum = int(nc.options.IPNumPerEni)
	}
	return nil
}

func getEniName(instanceid string, index int) string {
	return instanceid + ENI_PREFIX + strconv.Itoa(index)
}

// Init node init
func (nc *NetworkController) Init() error {
	nc.nodeNetworkMutx.Lock()
	if nc.nodeNetwork != nil {
		blog.Infof("node work exists, no need to create eni")
		return nil
	}
	nc.nodeNetworkMutx.Unlock()

	if err := nc.eniClient.Init(); err != nil {
		blog.Errorf("aws client init failed, err %s", err.Error())
		return err
	}

	vmInfo, err := nc.eniClient.GetVMInfo()
	if err != nil {
		blog.Errorf("get vm info failed, err %s", err.Error())
		return err
	}

	newNode := &cloud.NodeNetwork{
		TypeMeta: k8smetav1.TypeMeta{
			APIVersion: cloud.SchemeGroupVersion.Version,
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      nc.hostip,
			Namespace: CRD_NAMESPACES,
		},
		Spec: cloud.NodeNetworkSpec{
			Cluster:     nc.options.Cluster,
			Hostname:    nc.hostname,
			NodeAddress: nc.hostip,
			VM:          vmInfo,
			ENINum:      nc.eniNum,
			IPNumPerENI: nc.ipNum,
		},
	}

	maxIndex, err := nc.netUtil.GetNetworkInterfaceMaxIndex()
	if err != nil {
		return err
	}

	if err := nc.getEniQuota(); err != nil {
		blog.Errorf("get eni quota failed, err %s", err.Error())
		return err
	}

	for i := 0; i < nc.eniNum; i++ {
		eniName := getEniName(vmInfo.InstanceID, i)

		// createENI
		newIf, err := nc.eniClient.CreateENI(eniName, nc.ipNum)
		if err != nil {
			blog.Errorf("create eni failed, err %s", err.Error())
			return err
		}
		blog.Infof("create eni %s done", eniName)

		// attachENI
		newIndex := maxIndex + i + 1
		attachment, err := nc.eniClient.AttachENI(newIndex, newIf.EniID, vmInfo.InstanceID)
		if err != nil {
			blog.Errorf("attach eni %s failed, err %s", err.Error())
			return err
		}
		newIf.Attachment = attachment
		blog.Infof("attach eni %s done", eniName)

		newNode.Status.Enis = append(newNode.Status.Enis, newIf)
	}

	// TODO: write new node network to apiserver
	err = nc.nodeNetClient.Create(newNode)
	if err != nil {
		blog.Errorf("write node network %+v to apiserver failed, err %s", newNode, err.Error())
	}

	// TODO: write new ip info to netservice

	return nil
}

// Run run controller
func (nc *NetworkController) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := nc.reconcileNodeNetwork(); err != nil {
		blog.Infof("first reconcile node network failed, err %s", err.Error())
		return
	}

	tick := time.NewTicker(time.Duration(nc.options.CheckInterval) * time.Second)
	for {
		select {
		case <-tick.C:
			blog.Infof("it's time to check node network and enis!")

			if err := nc.reconcileNodeNetwork(); err != nil {
				blog.Warnf("reconcile node network failed, err %s", err.Error())
			}

		case <-ctx.Done():
			blog.Infof("stop controller...")
			return
		}
	}
}

// reconcileNodeNetwork restore network interface on vm
func (nc *NetworkController) reconcileNodeNetwork() error {
	// eniMap := make(map[string]*cloud.ElasticNetworkInterface)
	// for _, netif := range nc.nodeNetwork.Status.Enis {
	// 	eniMap[netif.EniID] = netif
	// }

	// // query ENIs
	// remoteEnis, err := nc.eniClient.ListENIs(nc.nodeNetwork.Spec.VM.InstanceID)
	// if err != nil {
	// 	blog.Errorf("list enis failed, err %s", err.Error())
	// 	return err
	// }

	// compare difference

	return nil
}
