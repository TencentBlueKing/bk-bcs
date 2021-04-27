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

package v2

import (
	"fmt"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/signals"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/config"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/actions"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/metrics"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
	networkclientset "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned"
	networkinformers "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/informers/externalversions"
	networklisters "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/listers/cloud/v1"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientGoCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

type IpScheduler struct {
	Cluster              string
	KubeClient           kubernetes.Interface
	NetworkClient        networkclientset.Interface
	NodeNetworkLister    networklisters.NodeNetworkLister
	CloudIpLister        networklisters.CloudIPLister
	CniAnnotationKey     string
	FixedIpAnnotationKey string
}

// DefaultIpScheduler default v2 IP scheduler
var DefaultIpScheduler *IpScheduler

// NewIpScheduler create a v2 IpScheduler
func NewIpScheduler(conf *config.CustomSchedulerConfig) (*IpScheduler, error) {
	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(conf.KubeMaster, conf.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error building kube config: %s", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error building kubernetes clientset: %s", err.Error())
	}

	ipScheduler := &IpScheduler{
		Cluster:              conf.Cluster,
		KubeClient:           kubeClient,
		CniAnnotationKey:     CniAnnotationKey,
		FixedIpAnnotationKey: FixedIpAnnotationKey,
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
	blog.Infof("Waiting for informer caches to sync")
	if ok := clientGoCache.WaitForCacheSync(stopCh, nodeNetworkInformer.Informer().HasSynced, cloudIPInformer.Informer().HasSynced); !ok {
		return nil, fmt.Errorf("failed to wait for caches to sync")
	}

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
	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV2, actions.TotalNodeNumKey, float64(len(extenderArgs.Nodes.Items)))

	cniAnnotationValue, ok := extenderArgs.Pod.ObjectMeta.Annotations[DefaultIpScheduler.CniAnnotationKey]
	if ok && cniAnnotationValue == CniAnnotationValue {
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
		for _, node := range extenderArgs.Nodes.Items {
			canSchedule = append(canSchedule, node)
		}
	}

	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV2, actions.CanSchedulerNodeNumKey, float64(len(canSchedule)))
	metrics.ReportK8sCustomSchedulerNodeNum(actions.IpSchedulerV2, actions.CanNotSchedulerNodeNumKey, float64(len(canNotSchedule)))
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

// checkSchedulable check whether a node is schedulable
func (i *IpScheduler) checkSchedulable(pod *v1.Pod, node v1.Node) error {
	// get the node ip
	var nodeIP string
	for _, nodeAddress := range node.Status.Addresses {
		if nodeAddress.Type == "InternalIP" {
			nodeIP = nodeAddress.Address
			break
		}
	}

	// get NodeNetwork
	nodeNetwork, err := i.NodeNetworkLister.NodeNetworks(BcsSystem).Get(nodeIP)
	if err != nil {
		return fmt.Errorf("failed to get NodeNetwork from cluster: %s", err.Error())
	}

	// get all CloudIp on this node
	cloudIpsOnThisNode, err := i.getCloudIpsOnNode(nodeIP)
	if err != nil {
		return err
	}

	// get existed CloudIp
	found, existedCloudIp, err := i.getExistedFixedCloudIp(pod)
	if err != nil {
		return err
	}

	fixedIpAnnotationValue, ok := pod.ObjectMeta.Annotations[i.FixedIpAnnotationKey]
	if ok && fixedIpAnnotationValue == FixedIpAnnotationValue && found {
		// pod with fixed ip annotation
		subnetIdMatched := nodeNetwork.Status.FloatingIPEni.Eni.EniSubnetID == existedCloudIp.Spec.SubnetID
		if subnetIdMatched {
			if existedCloudIp.Spec.Host == nodeNetwork.Name {
				return nil
			} else if existedCloudIp.Spec.Host != nodeNetwork.Name && nodeNetwork.Status.FloatingIPEni.IPLimit-len(cloudIpsOnThisNode) > 0 {
				return nil
			} else {
				return fmt.Errorf("no available eni ip anymore")
			}
		} else {
			return fmt.Errorf("subnetId unmatched for fixed ip request, pod: %s, node: %s", pod.Name, nodeNetwork.Name)
		}
	} else {
		// ordinary pod, without fixed ip
		if nodeNetwork.Status.FloatingIPEni.IPLimit-len(cloudIpsOnThisNode) > 0 {
			return nil
		} else {
			return fmt.Errorf("no available eni ip anymore")
		}
	}
}

// getExistedFixedCloudIp get existed CloudIp to a pod which need fixed ip
func (i *IpScheduler) getExistedFixedCloudIp(pod *v1.Pod) (bool, *cloudv1.CloudIP, error) {
	// get matched CloudIp to this Pod
	var found bool
	var existedCloudIp *cloudv1.CloudIP
	labelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{IP_LABEL_KEY_FOR_IS_FIXED: strconv.FormatBool(true)},
	}
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return found, existedCloudIp, fmt.Errorf("failed to build label selector: %s", err.Error())
	}
	namespacedFixedCloudIps, err := i.CloudIpLister.CloudIPs(pod.Namespace).List(selector)
	if err != nil {
		return found, existedCloudIp, fmt.Errorf("failed to get all fixed CloudIp in namespace %s of pod %s", pod.Namespace, pod.Name)
	}
	for _, cloudIP := range namespacedFixedCloudIps {
		if cloudIP.Spec.PodName == pod.Name && cloudIP.Spec.IsFixed {
			found = true
			existedCloudIp = cloudIP
			break
		}
	}

	return found, existedCloudIp, nil
}

func (i *IpScheduler) getCloudIpsOnNode(nodeIP string) ([]*cloudv1.CloudIP, error) {
	var cloudIPs []*cloudv1.CloudIP
	labelSelector := &metav1.LabelSelector{
		MatchLabels: map[string]string{IP_LABEL_KEY_FOR_HOST: nodeIP},
	}
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return cloudIPs, fmt.Errorf("failed to build label selector: %s", err.Error())
	}
	cloudIPs, err = i.CloudIpLister.List(selector)
	if err != nil {
		return cloudIPs, fmt.Errorf("failed to get CloudIp on node: %s", nodeIP)
	}

	return cloudIPs, nil
}
