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

package v2

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/signals"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
	networkclientset "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned"
	networkinformers "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/informers/externalversions"
	networklisters "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/listers/cloud/v1"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/grpclb"
	"google.golang.org/grpc"
	grpccredentials "google.golang.org/grpc/credentials"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	k8slistcorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	clientGoCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/config"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/internal/cache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/pkg/actions"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-custom-scheduler/pkg/metrics"
)

// IpScheduler k8s scheduler extender api for bcs cloud netservice
type IpScheduler struct {
	ctx    context.Context
	cancel context.CancelFunc

	Cluster           string
	KubeClient        kubernetes.Interface
	NetworkClient     networkclientset.Interface
	NodeNetworkLister networklisters.NodeNetworkLister
	CloudIpLister     networklisters.CloudIPLister
	NodeLister        k8slistcorev1.NodeLister
	CloudNetClient    pbcloudnet.CloudNetserviceClient
	CacheLock         sync.Mutex
	NodeIPCache       *cache.ResourceCache

	QuotaLock  sync.Mutex
	QuotaLimit int

	CniAnnotationKey       string
	CniAnnotationValue     string
	FixedIpAnnotationKey   string
	FixedIpAnnotationValue string
}

// DefaultIpScheduler default v2 IP scheduler
var DefaultIpScheduler *IpScheduler

// NewIpScheduler create a v2 IpScheduler
// nolint funlen
func NewIpScheduler(conf *config.CustomSchedulerConfig) (*IpScheduler, error) {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

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

	// create cloud netservice client
	var conn *grpc.ClientConn
	if conf.CloudNetserviceCert == nil {
		conn, err = grpc.Dial(
			"",
			grpc.WithInsecure(),
			// nolint grpc.WithBalancer is deprecated
			grpc.WithBalancer(grpc.RoundRobin(grpclb.NewPseudoResolver(conf.CloudNetserviceEndpoints))),
		)
	} else {
		tlsConfig, tlsErr := ssl.ClientTslConfVerity(
			conf.CloudNetserviceCert.CAFile,
			conf.CloudNetserviceCert.CertFile,
			conf.CloudNetserviceCert.CertPasswd,
			static.ClientCertPwd,
		)
		if tlsErr != nil {
			return nil, fmt.Errorf("failed to load tls files, certs %v, err %s", conf.CloudNetserviceCert, err.Error())
		}
		conn, err = grpc.Dial(
			"",
			grpc.WithTransportCredentials(grpccredentials.NewTLS(tlsConfig)),
			// nolint grpc.WithBalancer is deprecated
			grpc.WithBalancer(grpc.RoundRobin(grpclb.NewPseudoResolver(conf.CloudNetserviceEndpoints))),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud netserivce client connection, err %s", err.Error())
	}
	cloudNetClient := pbcloudnet.NewCloudNetserviceClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	ipScheduler := &IpScheduler{
		ctx:                    ctx,
		cancel:                 cancel,
		Cluster:                conf.Cluster,
		CniAnnotationKey:       conf.CniAnnotationKey,
		CniAnnotationValue:     conf.CniAnnotationValue,
		FixedIpAnnotationKey:   conf.FixedIpAnnotationKey,
		FixedIpAnnotationValue: conf.FixedIpAnnotationValue,
		CloudNetClient:         cloudNetClient,
		NodeIPCache:            cache.NewResourceCache(),
	}
	if conf.CniAnnotationKey != "" {
		ipScheduler.CniAnnotationKey = conf.CniAnnotationKey
	}
	if conf.FixedIpAnnotationKey != "" {
		ipScheduler.FixedIpAnnotationKey = conf.FixedIpAnnotationKey
	}

	networkClientSet, err := networkclientset.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("build bcs network clientset error %s", err.Error())
	}
	ipScheduler.NetworkClient = networkClientSet

	factory := networkinformers.NewSharedInformerFactory(networkClientSet, 0)
	nodeNetworkInformer := factory.Cloud().V1().NodeNetworks()
	ipScheduler.NodeNetworkLister = nodeNetworkInformer.Lister()
	cloudIPInformer := factory.Cloud().V1().CloudIPs()
	ipScheduler.CloudIpLister = cloudIPInformer.Lister()

	go factory.Start(stopCh)
	blog.Infof("Waiting for cloud ip informer caches to sync")
	if ok := clientGoCache.WaitForCacheSync(stopCh, nodeNetworkInformer.Informer().HasSynced,
		cloudIPInformer.Informer().HasSynced); !ok {
		return nil, fmt.Errorf("failed to wait for caches to sync")
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes clientset: %s", err.Error())
	}
	k8sFactory := informers.NewSharedInformerFactory(kubeClient, 0)
	nodeInformer := k8sFactory.Core().V1().Nodes()
	nodeLister := nodeInformer.Lister()
	podInformer := k8sFactory.Core().V1().Pods()
	ipScheduler.KubeClient = kubeClient
	ipScheduler.NodeLister = nodeLister
	go k8sFactory.Start(stopCh)
	blog.Infof("Waiting for k8s core informer caches to sync")
	if ok := clientGoCache.WaitForCacheSync(stopCh,
		nodeInformer.Informer().HasSynced, podInformer.Informer().HasSynced); !ok {
		return nil, fmt.Errorf("failed to wait for caches to sync")
	}
	podInformer.Informer().AddEventHandler(ipScheduler)

	if err := ipScheduler.initCache(); err != nil {
		return nil, fmt.Errorf("init cache failed, err %s", err.Error())
	}

	go ipScheduler.startGetQuota(ctx)

	return ipScheduler, nil
}

// HandleIpSchedulerPredicate handle v2 IpScheduler predicate
func HandleIpSchedulerPredicate(extenderArgs schedulerapi.ExtenderArgs) (*schedulerapi.ExtenderFilterResult, error) {
	// invalid type of custom scheduler, it should be IpSchedulerV2
	if DefaultIpScheduler == nil {
		return nil, fmt.Errorf("invalid type of custom scheduler, please check the custome scheduler config")
	}
	canSchedule := make([]v1.Node, 0, len(extenderArgs.Nodes.Items))
	canNotSchedule := make(map[string]string)
	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV2, actions.TotalNodeNumKey,
		float64(len(extenderArgs.Nodes.Items)))

	cniAnnotationValue, ok := extenderArgs.Pod.ObjectMeta.Annotations[DefaultIpScheduler.CniAnnotationKey]
	if ok && cniAnnotationValue == DefaultIpScheduler.CniAnnotationValue {
		// matched annotation, schedule with IpSchedulerV2
		blog.Infof("starting to predicate for pod %s", extenderArgs.Pod.Name)
		for _, node := range extenderArgs.Nodes.Items {
			err := DefaultIpScheduler.checkSchedulable(extenderArgs.Pod, node)
			if err != nil {
				canNotSchedule[node.Name] = err.Error()
			} else {
				canSchedule = append(canSchedule, node)
			}
		}
	} else {
		// unmatched annotation, skip to schedule with IpSchedulerV2
		blog.Infof("pod %s without cni annotation, skip to schedule with IpSchedulerV2 ", extenderArgs.Pod.Name)
		canSchedule = append(canSchedule, extenderArgs.Nodes.Items...)
	}

	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV2, actions.CanSchedulerNodeNumKey,
		float64(len(canSchedule)))
	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV2, actions.CanNotSchedulerNodeNumKey,
		float64(len(canNotSchedule)))
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
	// invalid type of custom scheduler, it should be IpSchedulerV2
	if DefaultIpScheduler == nil {
		return fmt.Errorf("invalid type of custom scheduler, please check the custome scheduler config")
	}
	node, err := DefaultIpScheduler.NodeLister.Get(extenderBindingArgs.Node)
	if err != nil {
		blog.Errorf("get node info by node name %s failed, err %s", extenderBindingArgs.Node, err.Error())
		return fmt.Errorf("get node info by node name %s failed, err %s", extenderBindingArgs.Node, err.Error())
	}
	nodeAddr := ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			nodeAddr = addr.Address
			break
		}
	}
	if nodeAddr == "" {
		blog.Errorf("node %s has no internal ip", extenderBindingArgs.Node)
		return fmt.Errorf("node %s has no internal ip", extenderBindingArgs.Node)
	}

	// get NodeNetwork
	nodeNetwork, err := DefaultIpScheduler.NodeNetworkLister.NodeNetworks(BcsSystem).Get(extenderBindingArgs.Node)
	if err != nil {
		return fmt.Errorf("failed to get NodeNetwork from cluster: %s", err.Error())
	}

	DefaultIpScheduler.CacheLock.Lock()
	nodeResources := DefaultIpScheduler.NodeIPCache.GetNodeResources(nodeAddr)
	if nodeNetwork.Spec.IPNumPerENI*getNodeReadyEniNum(nodeNetwork)-len(nodeResources) <= 0 {
		DefaultIpScheduler.CacheLock.Unlock()
		return fmt.Errorf("no enough resource to bind to node %s/%s", extenderBindingArgs.Node, nodeAddr)
	}
	// nolint
	DefaultIpScheduler.NodeIPCache.UpdateResource(&cache.Resource{
		PodName:      extenderBindingArgs.PodName,
		PodNamespace: extenderBindingArgs.PodNamespace,
		Node:         nodeAddr,
		ResourceKind: "CloudIP",
		Value:        1,
	})
	DefaultIpScheduler.CacheLock.Unlock()

	bind := &v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: extenderBindingArgs.PodNamespace,
			Name:      extenderBindingArgs.PodName,
			UID:       extenderBindingArgs.PodUID},
		Target: v1.ObjectReference{
			Kind: "Node",
			Name: extenderBindingArgs.Node,
		},
	}

	err = DefaultIpScheduler.KubeClient.CoreV1().Pods(bind.Namespace).Bind(
		context.Background(), bind, metav1.CreateOptions{})
	if err != nil {
		DefaultIpScheduler.CacheLock.Lock()
		// nolint
		DefaultIpScheduler.NodeIPCache.DeleteResource(
			cache.GetMetaKey(extenderBindingArgs.PodName, extenderBindingArgs.PodNamespace))
		DefaultIpScheduler.CacheLock.Unlock()
		return fmt.Errorf("error when binding pod to node: %s", err.Error())
	}

	return nil
}

func getNodeReadyEniNum(nodeNetwork *cloudv1.NodeNetwork) int {
	if nodeNetwork == nil {
		return 0
	}
	eniNum := 0
	for _, eni := range nodeNetwork.Status.Enis {
		if eni.Status == "Ready" {
			eniNum++
		}
	}
	return eniNum
}

// OnAdd implements EventHandler for informer
func (i *IpScheduler) OnAdd(add interface{}) {}

// OnUpdate implements EventHandler for informer
func (i *IpScheduler) OnUpdate(old, new interface{}) {}

// OnDelete implements EventHandler for informer, delete node resource
func (i *IpScheduler) OnDelete(del interface{}) {
	delPod, ok := del.(*v1.Pod)
	if !ok {
		return
	}
	if delPod.GetNamespace() == "bcs-system" {
		return
	}
	blog.Infof("pod %s/%s is deletd", delPod.GetName(), delPod.GetNamespace())
	i.CacheLock.Lock()
	// nolint
	i.NodeIPCache.DeleteResource(cache.GetMetaKey(delPod.GetName(), delPod.GetNamespace()))
	i.CacheLock.Unlock()
}

// Stop stop get quota
func (i *IpScheduler) Stop() {
	i.cancel()
}

// initCache init node resource cache
func (i *IpScheduler) initCache() error {
	cloudIPs, err := i.CloudIpLister.List(labels.Everything())
	if err != nil {
		blog.Errorf("list all cloudIPs failed, err %s", err.Error())
		return fmt.Errorf("list all cloudIPs failed, err %s", err.Error())
	}
	for _, ip := range cloudIPs {
		if ip.GetNamespace() != "bcs-system" {
			// nolint
			i.NodeIPCache.UpdateResource(&cache.Resource{
				PodName:      ip.Spec.PodName,
				PodNamespace: ip.Spec.Namespace,
				Node:         ip.Spec.Host,
				ResourceKind: "CloudIP",
				Value:        1,
			})
		}
	}
	for _, host := range i.NodeIPCache.GetNodes() {
		resList := i.NodeIPCache.GetNodeResources(host)
		blog.Infof("node %s has cloud ip %d", host, len(resList))
	}
	return nil
}

// startGetQuota start get quota loop
func (i *IpScheduler) startGetQuota(ctx context.Context) {
	i.setQuota()
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-ticker.C:
			i.setQuota()
		case <-ctx.Done():
			blog.Warnf("get quota loop end")
			return
		}
	}
}

// setQuota xxx
// set quota limit
func (i *IpScheduler) setQuota() {
	resp, err := i.CloudNetClient.GetQuota(context.Background(), &pbcloudnet.GetIPQuotaReq{
		Cluster: i.Cluster,
	})
	if err != nil {
		blog.Warnf("get quota of cluster %s", i.Cluster)
		return
	}
	if resp.ErrCode != 0 {
		blog.Warnf("get quota of cluster %s, errCode %d, errMsg %d", i.Cluster, resp.ErrCode, resp.ErrMsg)
		return
	}
	blog.Infof("get quota %d of cluster %s", int(resp.Quota.Limit), i.Cluster)
	i.QuotaLock.Lock()
	i.QuotaLimit = int(resp.Quota.Limit)
	i.QuotaLock.Unlock()
}

// checkQuota check whether it exceeds quota limit
// return true means can schedule
func (i *IpScheduler) checkQuota() bool {
	i.QuotaLock.Lock()
	defer i.QuotaLock.Unlock()
	resList := i.NodeIPCache.GetAllResources()
	return len(resList) < i.QuotaLimit
}

// checkSchedulable check whether a node is schedulable
func (i *IpScheduler) checkSchedulable(pod *v1.Pod, node v1.Node) error {

	if !i.checkQuota() {
		return fmt.Errorf("quota %d is full", i.QuotaLimit)
	}

	// get the node ip
	var nodeIP string
	for _, nodeAddress := range node.Status.Addresses {
		if nodeAddress.Type == "InternalIP" {
			nodeIP = nodeAddress.Address
			break
		}
	}

	// get NodeNetwork
	nodeNetwork, err := i.NodeNetworkLister.NodeNetworks(BcsSystem).Get(node.GetName())
	if err != nil {
		return fmt.Errorf("failed to get NodeNetwork from cluster: %s", err.Error())
	}

	// get existed CloudIp
	foundExistedIP, existedCloudIP, err := i.getExistedFixedCloudIp(pod)
	if err != nil {
		return err
	}

	fixedIpAnnotationValue, ok := pod.ObjectMeta.Annotations[i.FixedIpAnnotationKey]

	// if pod request fixed ip and found existed ip
	if ok && fixedIpAnnotationValue == i.FixedIpAnnotationValue && foundExistedIP {
		for _, eni := range nodeNetwork.Status.Enis {
			// get all CloudIp on this node
			cloudIPsOnThisEni, err := i.getCloudIPsOnEni(eni.EniID)
			if err != nil {
				return err
			}
			if eni.Status == "Ready" &&
				nodeNetwork.Spec.IPNumPerENI-len(cloudIPsOnThisEni) > 0 &&
				eni.EniSubnetID == existedCloudIP.Spec.SubnetID {
				// pod with fixed ip annotation
				return nil
			}
		}
		return fmt.Errorf("no available eni for fixed ip %s request subnet %s",
			existedCloudIP.GetName(), existedCloudIP.Spec.SubnetID)
	}
	nodeResources := i.NodeIPCache.GetNodeResources(nodeIP)
	if nodeNetwork.Spec.IPNumPerENI*getNodeReadyEniNum(nodeNetwork)-len(nodeResources) > 0 {
		return nil
	}
	return fmt.Errorf("no available eni ip anymore")
}

// getExistedFixedCloudIp get existed CloudIp to a pod which need fixed ip
func (i *IpScheduler) getExistedFixedCloudIp(pod *v1.Pod) (bool, *cloudv1.CloudIP, error) {
	// get matched CloudIp to this Pod
	ipList, err := i.CloudNetClient.ListIP(context.Background(), &pbcloudnet.ListIPsReq{
		PodName:   pod.GetName(),
		Namespace: pod.GetNamespace(),
		Cluster:   i.Cluster,
	})
	if err != nil {
		return false, nil, err
	}
	if ipList.ErrCode != 0 {
		return false, nil, fmt.Errorf("list ip by podname %s and namespace %s failed, errCode %d, errMsg %s",
			pod.GetName(), pod.GetNamespace(), ipList.ErrCode, ipList.ErrMsg)
	}
	if len(ipList.Ips) == 0 {
		return false, nil, nil
	}
	ipObj := ipList.Ips[0]
	return true, &cloudv1.CloudIP{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ipObj.Address,
			Namespace: ipObj.Namespace,
		},
		Spec: cloudv1.CloudIPSpec{
			Address:    ipObj.Address,
			VpcID:      ipObj.VpcID,
			SubnetID:   ipObj.SubnetID,
			SubnetCidr: ipObj.SubnetCidr,
			Region:     ipObj.Region,
			Cluster:    ipObj.Cluster,
			Namespace:  ipObj.Namespace,
			PodName:    ipObj.PodName,
			IsFixed:    ipObj.IsFixed,
		},
		Status: cloudv1.CloudIPStatus{
			Status: ipObj.Status,
		},
	}, nil
}

func (i *IpScheduler) getCloudIPsOnEni(eniID string) ([]*cloudv1.CloudIP, error) {
	var retCloudIPs []*cloudv1.CloudIP
	labelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{IPLabelKeyForEni: eniID},
	}
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to build label selector: %s", err.Error())
	}
	cloudIPs, err := i.CloudIpLister.List(selector)
	if err != nil {
		return cloudIPs, fmt.Errorf("failed to get CloudIP on eni: %s", eniID)
	}
	for _, ip := range cloudIPs {
		// cloud netservice ips are stored in bcs-system
		// skip ip in bcs-system for the case that cloud netservice and cloud netagent are in the same cluster
		if ip.GetNamespace() == BcsSystem {
			continue
		}
		retCloudIPs = append(retCloudIPs, ip)
	}

	return retCloudIPs, nil
}
