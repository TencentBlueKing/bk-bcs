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

// NOCC:tosa/comment_ratio(none)

// Package v3 scheduler for bcs-netservice-controller
package v3

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/dynamic/dynamiclister"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/config"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/pkg/actions"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/pkg/metrics"
)

const (
	groupNetserviceController   = "netservice.bkbcs.tencent.com"
	versionNetserviceController = "v1"
	resourceBCSNetPool          = "bcsnetpools"
	resourceBCSNetIP            = "bcsnetips"
	resourceBCsNetIPClaim       = "bcsnetipclaims"
)

var (
	bcsNetPoolGVR = schema.GroupVersionResource{
		Group:    groupNetserviceController,
		Version:  versionNetserviceController,
		Resource: resourceBCSNetPool,
	}
	bcsNetIPGVR = schema.GroupVersionResource{
		Group:    groupNetserviceController,
		Version:  versionNetserviceController,
		Resource: resourceBCSNetIP,
	}
	bcsNetIPClaimGVR = schema.GroupVersionResource{
		Group:    groupNetserviceController,
		Version:  versionNetserviceController,
		Resource: resourceBCsNetIPClaim,
	}
	podGVR = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
	nodeGVR = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "nodes",
	}
)

// IpScheduler k8s scheduler extender api for bcs netservice
type IpScheduler struct {
	DynamicClient        dynamic.Interface
	InformerFactory      dynamicinformer.DynamicSharedInformerFactory
	PoolLister           dynamiclister.Lister
	IPInformer           cache.SharedInformer
	IPLister             dynamiclister.Lister
	ClaimLister          dynamiclister.Lister
	PodLister            dynamiclister.Lister
	NodeLister           dynamiclister.Lister
	FixedIpAnnotationKey string

	cache  *PoolCache
	StopCh chan struct{}
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
	informerFactory := dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, 0)
	poolInformer := informerFactory.ForResource(bcsNetPoolGVR).Informer()
	ipInformer := informerFactory.ForResource(bcsNetIPGVR).Informer()
	claimInformer := informerFactory.ForResource(bcsNetIPClaimGVR).Informer()
	podInformer := informerFactory.ForResource(podGVR).Informer()
	nodeInformer := informerFactory.ForResource(nodeGVR).Informer()
	poolLister := dynamiclister.New(poolInformer.GetIndexer(), bcsNetPoolGVR)
	ipLister := dynamiclister.New(ipInformer.GetIndexer(), bcsNetIPGVR)
	claimLister := dynamiclister.New(claimInformer.GetIndexer(), bcsNetIPClaimGVR)
	podLister := dynamiclister.New(podInformer.GetIndexer(), podGVR)
	nodeLister := dynamiclister.New(nodeInformer.GetIndexer(), nodeGVR)

	cache := NewPoolCache()
	ipScheduler := &IpScheduler{
		InformerFactory: informerFactory,
		DynamicClient:   dynamicClient,
		PoolLister:      poolLister,
		IPLister:        ipLister,
		IPInformer:      ipInformer,
		ClaimLister:     claimLister,
		PodLister:       podLister,
		NodeLister:      nodeLister,
		cache:           cache,
		StopCh:          make(chan struct{}),
	}
	if conf.FixedIpAnnotationKey != "" {
		ipScheduler.FixedIpAnnotationKey = conf.FixedIpAnnotationKey
	}

	return ipScheduler, nil
}

// StartInformers start informers
func StartInformers() error {
	if DefaultIpScheduler == nil {
		return fmt.Errorf("default scheduler is nil")
	}
	ih := newIPHander(DefaultIpScheduler.cache)

	DefaultIpScheduler.IPInformer.AddEventHandler(ih)
	blog.Infof("start informers factory")
	DefaultIpScheduler.InformerFactory.Start(DefaultIpScheduler.StopCh)
	DefaultIpScheduler.InformerFactory.WaitForCacheSync(DefaultIpScheduler.StopCh)
	blog.Infof("informers caches are synced")
	return nil
}

// HandleIpSchedulerPredicate handle v3 IpScheduler predicate
func HandleIpSchedulerPredicate(extenderArgs schedulerapi.ExtenderArgs) (*schedulerapi.ExtenderFilterResult, error) {
	// invalid type of custom scheduler, it should be IpSchedulerV3
	if DefaultIpScheduler == nil {
		return nil, fmt.Errorf("invalid type of custom scheduler, please check the custom scheduler config")
	}
	canSchedule := make([]v1.Node, 0, len(extenderArgs.Nodes.Items))
	canNotSchedule := make(map[string]string)
	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV3, actions.TotalNodeNumKey,
		float64(len(extenderArgs.Nodes.Items)))

	var availableHosts []string
	claimName, ok := extenderArgs.Pod.ObjectMeta.Annotations[DefaultIpScheduler.FixedIpAnnotationKey]
	if ok {
		netClaim, err := getIPClaim(extenderArgs.Pod.Namespace, claimName)
		if err != nil {
			return nil, fmt.Errorf("get claim failed")
		}
		if netClaim.Status.Phase == BCSNetIPClaimExpiredStatus {
			blog.Errorf("claim %s/%s was expired", extenderArgs.Pod.Namespace, claimName)
			return nil, fmt.Errorf("claim has expired")
		}
		// if claim has bound ip, just schedule to the pool of the bounded ip
		if netClaim.Status.Phase == BCSNetIPClaimBoundedStatus && netClaim.Status.BoundedIP != "" {
			netIP, err := getIP(netClaim.Status.BoundedIP)
			if err != nil {
				return nil, fmt.Errorf("get ip of claim failed")
			}
			// get pool
			poolName, ok := netIP.Labels["pool"]
			if !ok {
				return nil, fmt.Errorf("can't find pool name from IP %s lables", netIP.Name)
			}
			netPool, err := getPool(poolName)
			if err != nil {
				return nil, fmt.Errorf("get pool of claim failed")
			}
			availableHosts = append(availableHosts, netPool.Spec.Hosts...)
		}
	}
	blog.Infof("pod %s/%s without claim or bounded fixed ip", extenderArgs.Pod.Namespace, extenderArgs.Pod.Name)
	availablePools := DefaultIpScheduler.cache.GetAvailablePoolNameList()
	for _, poolName := range availablePools {
		netPool, err := getPool(poolName)
		if err != nil {
			blog.Warnf("get net pool %s failed, err %s", poolName, err.Error())
			continue
		}
		blog.Infof("find available hosts in pool %s", poolName)
		availableHosts = append(availableHosts, netPool.Spec.Hosts...)
	}

	for _, node := range extenderArgs.Nodes.Items {
		err := checkNodeInHosts(node, availableHosts)
		if err != nil {
			canNotSchedule[node.Name] = err.Error()
		} else {
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

// get IP Claim
func getIPClaim(namespace, claimKey string) (*BCSNetIPClaim, error) {
	claimUnstruct, err := DefaultIpScheduler.ClaimLister.Namespace(namespace).Get(claimKey)
	if err != nil {
		blog.Warnf("get BCSNetIPClaim %s/%s failed, err %s", namespace, claimKey, err.Error())
		return nil, fmt.Errorf("get BCSNetIPClaim %s/%s failed, err %s", namespace, claimKey, err.Error())
	}
	claim := &BCSNetIPClaim{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		claimUnstruct.UnstructuredContent(), claim); err != nil {
		blog.Warnf("failed to convert unstructured claim %s/%s", namespace, claimKey)
		return nil, fmt.Errorf("failed to convert unstructured claim %s/%s", namespace, claimKey)
	}
	return claim, nil
}

// get IP
func getIP(ipName string) (*BCSNetIP, error) {
	// Get retrieves a resource from the indexer with the given name
	ipUnstruct, err := DefaultIpScheduler.IPLister.Get(ipName)
	if err != nil {
		blog.Warnf("get BCSNetIP %s failed, err %s", ipName, err.Error())
		return nil, fmt.Errorf("get BCSNetIP %s failed, err %s", ipName, err.Error())
	}
	ip := &BCSNetIP{}
	// FromUnstructured converts an object from map[string]interface{} representation into a concrete type.
	// It uses encoding/json/Unmarshaler if object implements it or reflection if not.
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		ipUnstruct.UnstructuredContent(), ip); err != nil {
		blog.Warnf("failed to convert unstructured ip %s", ipName)
		return nil, fmt.Errorf("failed to convert unstructured ip %s", ipName)
	}
	return ip, nil
}

// get Pool
func getPool(poolName string) (*BCSNetPool, error) {
	poolUnstruct, err := DefaultIpScheduler.PoolLister.Get(poolName)
	if err != nil {
		blog.Warnf("get BCSNetPool %s failed, err %s", poolName, err.Error())
		return nil, fmt.Errorf("get BCSNetPool %s failed, err %s", poolName, err.Error())
	}
	pool := &BCSNetPool{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		poolUnstruct.UnstructuredContent(), pool); err != nil {
		blog.Warnf("failed to convert unstructured net pool %s", poolName)
		return nil, fmt.Errorf("failed to convert unstructured net pool %s", poolName)
	}
	return pool, nil
}

// get Pool By Hostname
func getPoolByHostname(hostName string) (*BCSNetPool, error) {
	node, err := getNode(hostName)
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, fmt.Errorf("node is empty")
	}
	poolListUnstruct, err := DefaultIpScheduler.PoolLister.List(labels.Everything())
	if err != nil {
		blog.Warnf("get all BCSNetPoolList failed, err %s", err.Error())
		return nil, fmt.Errorf("get all BCSNetPoolList failed, err %s", err.Error())
	}
	for _, poolUnstruct := range poolListUnstruct {
		pool := &BCSNetPool{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
			poolUnstruct.UnstructuredContent(), pool); err != nil {
			blog.Warnf("failed to convert unstructured net pool %s", pool.Name)
			return nil, fmt.Errorf("failed to convert unstructured net pool %s", pool.Name)
		}
		if err := checkNodeInHosts(*node, pool.Spec.Hosts); err == nil {
			return pool, nil
		}
	}
	return nil, fmt.Errorf("host %s is not in any net pool", hostName)
}

// get Pod
func getPod(ns, name string) (*v1.Pod, error) {
	podUnstruct, err := DefaultIpScheduler.PodLister.Namespace(ns).Get(name)
	if err != nil {
		blog.Warnf("get Pod %s/%s failed, err %s", ns, name, err.Error())
		return nil, fmt.Errorf("get Pod %s/%s failed, err %s", ns, name, err.Error())
	}
	pod := &v1.Pod{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		podUnstruct.UnstructuredContent(), pod); err != nil {
		blog.Warnf("failed to convert unstructured Pod %s/%s", ns, name)
		return nil, fmt.Errorf("failed to convert unstructured Pod %s/%s", ns, name)
	}
	return pod, nil
}

// get Node
func getNode(name string) (*v1.Node, error) {
	nodeUnstruct, err := DefaultIpScheduler.NodeLister.Get(name)
	if err != nil {
		blog.Warnf("get Node %s failed, err %s", name, err.Error())
		return nil, fmt.Errorf("get Node %s failed, err %s", name, err.Error())
	}
	node := &v1.Node{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		nodeUnstruct.UnstructuredContent(), node); err != nil {
		blog.Warnf("failed to convert unstructured Node %s", name)
		return nil, fmt.Errorf("failed to convert unstructured Node %s", name)
	}
	return node, nil
}

// HandleIpSchedulerPreBind handle ip scheduler prebind
func HandleIpSchedulerPreBind(extenderBindingArgs schedulerapi.ExtenderBindingArgs) error {
	blog.Infof("prebind interface called, args %v", extenderBindingArgs)
	pod, err := getPod(extenderBindingArgs.PodNamespace, extenderBindingArgs.PodName)
	if err != nil {
		return err
	}
	claimName, ok := pod.Annotations[DefaultIpScheduler.FixedIpAnnotationKey]
	if ok {
		netClaim, gerr := getIPClaim(pod.Namespace, claimName)
		if gerr != nil {
			return fmt.Errorf("get claim failed")
		}
		if netClaim.Status.Phase == BCSNetIPClaimExpiredStatus {
			blog.Errorf("claim %s/%s was expired", pod.Namespace, claimName)
			return fmt.Errorf("claim has expired")
		}
		// if claim has bound ip, just schedule to the pool of the bounded ip
		if netClaim.Status.Phase == BCSNetIPClaimBoundedStatus && netClaim.Status.BoundedIP != "" {
			// no need to assume ip in cache
			return nil
		}
	}
	// assume ip in pool cache
	pool, err := getPoolByHostname(extenderBindingArgs.Node)
	if err != nil {
		return err
	}
	DefaultIpScheduler.cache.AssumeOne(pool.Name)
	return nil
}

// HandleIpSchedulerBinding handle ip scheduler binding
func HandleIpSchedulerBinding(extenderBindingArgs schedulerapi.ExtenderBindingArgs) error {
	return fmt.Errorf("not implements")
}

// checkNodeInHosts check whether a node is in hosts
func checkNodeInHosts(node v1.Node, hosts []string) error {
	// get the node ip
	var nodeIP string
	for _, nodeAddress := range node.Status.Addresses {
		if nodeAddress.Type == v1.NodeInternalIP {
			nodeIP = nodeAddress.Address
			if stringInSlice(hosts, nodeIP) {
				return nil
			}
			blog.Errorf("node %s is not in netPool hosts list", nodeIP)
			return fmt.Errorf("no available ip")
		}
	}
	blog.Errorf("node with name %s is not in netPool hosts list", node.Name)
	return fmt.Errorf("no available ip")
}

// sync Cached Pool By IP
func syncCachedPoolByIP(ip *BCSNetIP) {
	poolName, ok := ip.ObjectMeta.Labels[PodLabelKeyForPool]
	if !ok {
		blog.Warnf("ip %s/%s has no pool labels", ip.ObjectMeta.Namespace, ip.ObjectMeta.Name)
		return
	}
	syncCachedPoolIPNum(poolName)
}

// sync Cached Pool IP Num
func syncCachedPoolIPNum(poolName string) {
	if DefaultIpScheduler == nil {
		blog.Warnf("default scheduler is nil, wait for creation")
		return
	}
	ipNum := 0
	ipUnstructList, err := DefaultIpScheduler.IPLister.List(labels.SelectorFromSet(labels.Set(map[string]string{
		PodLabelKeyForPool: poolName,
	})))
	if err != nil {
		blog.Warnf("list ip list of pool %s failed, err %s", poolName, err.Error())
		return
	}
	for _, ipUnstruct := range ipUnstructList {
		ip := &BCSNetIP{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
			ipUnstruct.UnstructuredContent(), ip); err != nil {
			blog.Warnf("failed to convert unstructured, %v", ipUnstruct)
			return
		}
		if ip.Status.Phase == BCSNetIPAvailableStatus {
			ipNum++
		}
	}
	DefaultIpScheduler.cache.UpdatePool(poolName, ipNum)
	blog.Infof("update pool %s ipnum %d", poolName, ipNum)
}

// check if a string is in a string slice
func stringInSlice(strs []string, str string) bool {
	for _, item := range strs {
		if str == item {
			return true
		}
	}
	return false
}
