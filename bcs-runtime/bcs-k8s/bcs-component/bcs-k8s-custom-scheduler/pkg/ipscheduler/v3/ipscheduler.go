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

package v3

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/config"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/pkg/actions"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/pkg/metrics"
)

// IpScheduler k8s scheduler extender api for bcs netservice
type IpScheduler struct {
	DynamicClient        dynamic.Interface
	FixedIpAnnotationKey string
}

type netResource struct {
	netClaim *BCSNetIPClaim
	netIP    *BCSNetIP
	netPool  *BCSNetPool
}

// DefaultIpScheduler default v3 IP scheduler
var DefaultIpScheduler *IpScheduler

// NewIpScheduler create a v3 IpScheduler
func NewIpScheduler(conf *config.CustomSchedulerConfig) (*IpScheduler, error) {
	var cfg *rest.Config
	var err error
	if len(conf.KubeConfig) == 0 {
		cfg, err = rest.InClusterConfig()
	} else {
		cfg, err = clientcmd.BuildConfigFromFlags(conf.KubeMaster, conf.KubeConfig)
	}
	if err != nil {
		blog.Errorf("error building kube config: %v", err)
		return nil, fmt.Errorf("error building kube config: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes dynamic client: %s", err.Error())
	}
	ipScheduler := &IpScheduler{
		DynamicClient: dynamicClient,
	}
	if conf.FixedIpAnnotationKey != "" {
		ipScheduler.FixedIpAnnotationKey = conf.FixedIpAnnotationKey
	}

	return ipScheduler, nil
}

// HandleIpSchedulerPredicate handle v3 IpScheduler predicate
func HandleIpSchedulerPredicate(extenderArgs schedulerapi.ExtenderArgs) (*schedulerapi.ExtenderFilterResult, error) {
	// invalid type of custom scheduler, it should be IpSchedulerV3
	if DefaultIpScheduler == nil {
		return nil, fmt.Errorf("invalid type of custom scheduler, please check the custome scheduler config")
	}
	canSchedule := make([]v1.Node, 0, len(extenderArgs.Nodes.Items))
	canNotSchedule := make(map[string]string)
	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV3, actions.TotalNodeNumKey,
		float64(len(extenderArgs.Nodes.Items)))

	fixedIpAnnotationKey, ok := extenderArgs.Pod.ObjectMeta.Annotations[DefaultIpScheduler.FixedIpAnnotationKey]
	if ok {
		resource := &netResource{}
		// get claim from pod annotation
		if err := getBCSNetCR(extenderArgs.Pod.Namespace, fixedIpAnnotationKey, "bcsnetipclaims", resource); err != nil {
			return nil, err
		}
		if resource.netClaim.Status.Phase != "Bound" || resource.netClaim.Status.BoundedIP == "" {
			blog.Errorf("claim %s/%s hasn't bound an IP", extenderArgs.Pod.Namespace, resource.netClaim.Name)
			return nil, fmt.Errorf("claim %s/%s hasn't bound an IP", extenderArgs.Pod.Namespace, resource.netClaim.Name)
		}

		// get IP from claim
		if err := getBCSNetCR("", resource.netClaim.Status.BoundedIP, "bcsnetips", resource); err != nil {
			return nil, err
		}

		// get pool
		poolName, ok := resource.netIP.Labels["pool"]
		if !ok {
			return nil, fmt.Errorf("can't find pool name from IP %s lables", resource.netIP.Name)
		}
		if err := getBCSNetCR("", poolName, "bcsnetpools", resource); err != nil {
			return nil, err
		}

		for _, node := range extenderArgs.Nodes.Items {
			err := checkSchedulable(node, resource.netPool.Spec.Hosts)
			if err != nil {
				canNotSchedule[node.Name] = err.Error()
			} else {
				canSchedule = append(canSchedule, node)
			}
		}
		blog.V(5).Infof("pod %s/%s using bcsnetipclaim %s can not schedule on nodes %v",
			extenderArgs.Pod.Namespace, extenderArgs.Pod.Name, resource.netClaim.Name, canNotSchedule)
	} else {
		// unmatched annotation, skip to schedule with IpSchedulerV3
		blog.Infof("pod %s/%s without fixed annotation, skip to schedule with IpSchedulerV3 ",
			extenderArgs.Pod.Namespace, extenderArgs.Pod.Name)
		for _, node := range extenderArgs.Nodes.Items {
			canSchedule = append(canSchedule, node)
		}
	}
	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV3, actions.CanSchedulerNodeNumKey,
		float64(len(canSchedule)))
	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV3, actions.CanNotSchedulerNodeNumKey,
		float64(len(canNotSchedule)))
	scheduleResult := &schedulerapi.ExtenderFilterResult{
		Nodes: &v1.NodeList{
			Items: canSchedule,
		},
		FailedNodes: canNotSchedule,
		Error:       "",
	}

	return scheduleResult, nil
}

func getBCSNetCR(namespace, name, types string, resource *netResource) error {
	unstructured, err := DefaultIpScheduler.DynamicClient.Resource(schema.GroupVersionResource{
		Group:    "netservice.bkbcs.tencent.com",
		Version:  "v1",
		Resource: types,
	}).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	unstructuredByte, err := json.Marshal(unstructured.Object)
	if err != nil {
		return err
	}

	switch types {
	case "bcsnetpools":
		pool := &BCSNetPool{}
		err = json.Unmarshal(unstructuredByte, pool)
		if err != nil {
			return err
		}
		resource.netPool = pool
	case "bcsnetips":
		ip := &BCSNetIP{}
		err = json.Unmarshal(unstructuredByte, ip)
		if err != nil {
			return err
		}
		resource.netIP = ip
	case "bcsnetipclaims":
		claim := &BCSNetIPClaim{}
		err = json.Unmarshal(unstructuredByte, claim)
		if err != nil {
			return err
		}
		resource.netClaim = claim
	default:
		return fmt.Errorf("invalid bcs netservice resource type")
	}

	return nil
}

// HandleIpSchedulerBinding handle ip scheduler binding
func HandleIpSchedulerBinding(extenderBindingArgs schedulerapi.ExtenderBindingArgs) error {
	return nil
}

// checkSchedulable check whether a node is schedulable
func checkSchedulable(node v1.Node, hosts []string) error {
	// get the node ip
	var nodeIP string
	for _, nodeAddress := range node.Status.Addresses {
		if nodeAddress.Type == "InternalIP" {
			nodeIP = nodeAddress.Address
			if stringInSlice(hosts, nodeIP) {
				return nil
			}
			return fmt.Errorf("node %s is not in netPool hosts list", nodeIP)
		}
	}
	return fmt.Errorf("node %s is not in netPool hosts list", node.Name)
}

func stringInSlice(strs []string, str string) bool {
	for _, item := range strs {
		if str == item {
			return true
		}
	}
	return false
}
