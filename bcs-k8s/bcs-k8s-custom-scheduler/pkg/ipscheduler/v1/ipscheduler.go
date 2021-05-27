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

package v1

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	types "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/config"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/actions"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/metrics"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

// IpScheduler ip scheduler
type IpScheduler struct {
	Cluster      string
	UpdatePeriod uint
	netClient    bcsapi.Netservice
	NetPools     []*types.NetPool
	KubeClient   *kubernetes.Clientset
}

// DefaultIpScheduler default ip scheduler
var DefaultIpScheduler *IpScheduler

// NewIpScheduler create ip scheduler
func NewIpScheduler(conf *config.CustomSchedulerConfig) *IpScheduler {
	netClit, err := createNetSvcClient(conf)
	if err != nil {
		fmt.Printf("error create default ipScheduler: %v\n", err)
		os.Exit(1)
	}

	cfg, err := clientcmd.BuildConfigFromFlags(conf.KubeMaster, conf.KubeConfig)
	if err != nil {
		fmt.Printf("error building kube config: %v\n", err)
		os.Exit(1)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		fmt.Printf("error building kube client: %v\n", err)
		os.Exit(1)
	}

	return &IpScheduler{
		UpdatePeriod: conf.UpdatePeriod,
		Cluster:      conf.Cluster,
		netClient:    netClit,
		KubeClient:   kubeClient,
	}
}

// UpdateNetPoolsPeriodically update netpools periodically
func (i *IpScheduler) UpdateNetPoolsPeriodically() {
	updatePeriod := i.UpdatePeriod
	ticker := time.NewTicker(time.Duration(updatePeriod) * time.Second)
	defer ticker.Stop()

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

// HandleIpSchedulerPredicate handle ip scheduler predicate
func HandleIpSchedulerPredicate(extenderArgs schedulerapi.ExtenderArgs) (*schedulerapi.ExtenderFilterResult, error) {
	if DefaultIpScheduler == nil {
		return nil, fmt.Errorf("invalid type of custom scheduler, please check the custome scheduler config")
	}
	canSchedule := make([]v1.Node, 0, len(extenderArgs.Nodes.Items))
	canNotSchedule := make(map[string]string)
	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV1, actions.TotalNodeNumKey, float64(len(extenderArgs.Nodes.Items)))

	if extenderArgs.Pod.Spec.HostNetwork == true {
		blog.Infof("hostNetwork pod %s, skip to interact with netService", extenderArgs.Pod.Name)
		for _, node := range extenderArgs.Nodes.Items {
			canSchedule = append(canSchedule, node)
		}
	} else {
		blog.Infof("starting to predicate for pod %s", extenderArgs.Pod.Name)
		for _, node := range extenderArgs.Nodes.Items {
			err := DefaultIpScheduler.checkSchedulable(node)
			if err != nil {
				canNotSchedule[node.Name] = err.Error()
			} else {
				canSchedule = append(canSchedule, node)
			}
		}
	}

	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV1, actions.CanSchedulerNodeNumKey, float64(len(canSchedule)))
	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV1, actions.CanNotSchedulerNodeNumKey, float64(len(canNotSchedule)))
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

// HandleIpSchedulerBinding handle ip scheduler binding
func HandleIpSchedulerBinding(extenderBindingArgs schedulerapi.ExtenderBindingArgs) error {
	if DefaultIpScheduler == nil {
		return fmt.Errorf("invalid type of custom scheduler, please check the custome scheduler config")
	}
	pod, err := DefaultIpScheduler.KubeClient.CoreV1().Pods(extenderBindingArgs.PodNamespace).Get(context.Background(), extenderBindingArgs.PodName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error when getting pod from cluster: %s", err.Error())
	}
	if pod.Spec.HostNetwork != true {
		blog.Infof("starting to bind pod %s, update netService data in cache", extenderBindingArgs.PodName)

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
					blog.Info("%d", len(DefaultIpScheduler.NetPools[i].Available))
				} else {
					blog.Warnf("no available ip in node %s for pod %s, delete it and reschedule... ", netPool.Net, extenderBindingArgs.PodName)
					err := DefaultIpScheduler.KubeClient.CoreV1().Pods(extenderBindingArgs.PodNamespace).Delete(context.Background(), extenderBindingArgs.PodName, metav1.DeleteOptions{})
					if err != nil {
						return fmt.Errorf("error when deleting pod %s, failed to reschedule: %s", extenderBindingArgs.PodName, err.Error())
					}
					return nil
				}

				break
			}
		}
	} else {
		blog.Infof("starting to bind pod %s, skip to update netService data in cache", extenderBindingArgs.PodName)
	}

	bind := &v1.Binding{
		ObjectMeta: metav1.ObjectMeta{Namespace: extenderBindingArgs.PodNamespace, Name: extenderBindingArgs.PodName, UID: extenderBindingArgs.PodUID},
		Target: v1.ObjectReference{
			Kind: "Node",
			Name: extenderBindingArgs.Node,
		},
	}

	err = DefaultIpScheduler.KubeClient.CoreV1().Pods(bind.Namespace).Bind(context.Background(), bind, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("error when binding pod to node: %s", err.Error())
	}

	return nil
}

func (i *IpScheduler) checkSchedulable(node v1.Node) error {
	var nodeIp string
	for _, nodeAddress := range node.Status.Addresses {
		if nodeAddress.Type == "InternalIP" {
			nodeIp = nodeAddress.Address
			break
		}
	}
	if nodeIp == "" {
		return fmt.Errorf("cant't find node ip")
	}

	var matchedNetPool *types.NetPool
	for _, netPool := range i.NetPools {
		found := false
		for _, hostIp := range netPool.Hosts {
			if hostIp == nodeIp {
				found = true
				break
			}
		}
		if found && netPool.Cluster == i.Cluster {
			matchedNetPool = netPool
			break
		}
	}
	if matchedNetPool == nil {
		return fmt.Errorf("can't find netPool, cluster: %s, node: %s", i.Cluster, nodeIp)
	}

	if len(matchedNetPool.Available) == 0 {
		return fmt.Errorf("no available ip address anymore")
	} else {
		return nil
	}
}
