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

package ipscheduler

import (
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/config"
	"bk-bcs/bcs-services/bcs-netservice/pkg/netservice"
	"bk-bcs/bcs-services/bcs-netservice/pkg/netservice/types"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"

	"fmt"
	"os"
	"strings"
	"time"
)

type Ipscheduler struct {
	Cluster    string
	netClient  netservice.Client
	NetPools   []*types.NetPool
	KubeClient *kubernetes.Clientset
}

var DefaultIpScheduler *Ipscheduler

func NewIpscheduler() *Ipscheduler {
	netClit, err := createNetSvcClient()
	if err != nil {
		fmt.Printf("error create default ipScheduler: %v\n", err)
		os.Exit(1)
	}

	cfg, err := clientcmd.BuildConfigFromFlags(config.KubeMaster, config.Kubeconfig)
	if err != nil {
		fmt.Printf("error building kube config: %v\n", err)
		os.Exit(1)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("error building kube client: %v\n", err)
		os.Exit(1)
	}

	return &Ipscheduler{
		Cluster:    config.Cluster,
		netClient:  netClit,
		KubeClient: kubeClient,
	}
}

func (i *Ipscheduler) UpdateNetPoolsPeriodically() {
	updatePeriod := config.UpdatePeriod
	ticker := time.NewTicker(time.Duration(updatePeriod) * time.Second)
	for {
		blog.Info("starting to update netpool...")
		netPools, err := i.netClient.ListAllPoolWithCluster(i.Cluster)
		if err != nil {
			blog.Errorf("err calling netservice ListAllPool: %s", err.Error())
			continue
		}

		i.NetPools = netPools

		select {
		case <-ticker.C:
		}
	}
}

func HandleIpschedulerPredicate(extenderArgs schedulerapi.ExtenderArgs) (*schedulerapi.ExtenderFilterResult, error) {
	canSchedule := make([]v1.Node, 0, len(extenderArgs.Nodes.Items))
	canNotSchedule := make(map[string]string)

	for _, node := range extenderArgs.Nodes.Items {
		result, err := DefaultIpScheduler.checkSchedulable(node)
		if err != nil || !result {
			canNotSchedule[node.Name] = err.Error()
		} else {
			if result {
				canSchedule = append(canSchedule, node)
			}
		}
	}

	blog.Info("%v", canNotSchedule)
	scheduleResult := schedulerapi.ExtenderFilterResult{
		Nodes: &v1.NodeList{
			Items: canSchedule,
		},
		FailedNodes: canNotSchedule,
		Error:       "",
	}

	return &scheduleResult, nil
}

func HandleIpschedulerBinding(extenderBindingArgs schedulerapi.ExtenderBindingArgs) error {

	strArray := strings.Split(extenderBindingArgs.Node, "-")
	ipArray := strArray[1:5]
	ipAddress := strings.Join(ipArray, ".")

	for i, netPool := range DefaultIpScheduler.NetPools {
		if netPool.Cluster == DefaultIpScheduler.Cluster && netPool.Net == ipAddress {
			length := len(netPool.Available)
			blog.Info(netPool.Net)
			blog.Info("%d", length)
			if length > 0 {
				DefaultIpScheduler.NetPools[i].Available = netPool.Available[:length-1]
			}
			blog.Info("%d", len(DefaultIpScheduler.NetPools[i].Available))
			break
		}
	}

	bind := &v1.Binding{
		ObjectMeta: metav1.ObjectMeta{Namespace: extenderBindingArgs.PodNamespace, Name: extenderBindingArgs.PodName, UID: extenderBindingArgs.PodUID},
		Target: v1.ObjectReference{
			Kind: "Node",
			Name: extenderBindingArgs.Node,
		},
	}

	err := DefaultIpScheduler.KubeClient.CoreV1().Pods(bind.Namespace).Bind(bind)
	if err != nil {
		return fmt.Errorf("error when binding pod to node: %s", err.Error())
	}

	return nil
}

func (i *Ipscheduler) checkSchedulable(node v1.Node) (bool, error) {
	var nodeIp string
	for _, nodeAddress := range node.Status.Addresses {
		if nodeAddress.Type == "InternalIP" {
			nodeIp = nodeAddress.Address
			break
		}
	}
	if nodeIp == "" {
		return false, fmt.Errorf("cant't find node ip")
	}

	var matchedNetPool *types.NetPool
	for _, netPool := range i.NetPools {
		if netPool.Cluster == i.Cluster && netPool.Net == nodeIp {
			matchedNetPool = netPool
			break
		}
	}
	if matchedNetPool == nil {
		return false, fmt.Errorf("can't find netPool, cluster: %s, node: %s", i.Cluster, nodeIp)
	}

	if len(matchedNetPool.Available) == 0 {
		return false, fmt.Errorf("no available ip address anymore")
	} else {
		return true, nil
	}
}
