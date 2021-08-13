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

package kube

import (
	"context"
	"fmt"
	"strconv"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientgocache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
	bcsclientset "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned"
	cloudv1set "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned/typed/cloud/v1"
	bcsinformers "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/informers/externalversions"
	listercloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/listers/cloud/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/utils"
)

const (
	// CrdVersionV1 crd version v1
	CrdVersionV1 = "v1"
	// CrdNameCloudSubnet crd name for cloud subnet
	CrdNameCloudSubnet = "CloudSubnet"
	// CrdNameCloudIP crd name for cloud ip
	CrdNameCloudIP = "CloudIP"

	// CrdNameLabelsVpcID crd labels name for vpc id
	CrdNameLabelsVpcID = "vpc.cloud.bkbcs.tencent.com"
	// CrdNameLabelsRegion crd labels name for region
	CrdNameLabelsRegion = "region.cloud.bkbcs.tencent.com"
	// CrdNameLabelsZone crd labels name for zone
	CrdNameLabelsZone = "zone.cloud.bkbcs.tencent.com"
	// CrdNameLabelsSubnetID crd labels name for subent id
	CrdNameLabelsSubnetID = "subnet.cloud.bkbcs.tencent.com"
	// CrdNameLabelsCluster crd labels name for cluster
	CrdNameLabelsCluster = "cluster.cloud.bkbcs.tencent.com"
	// CrdNameLabelsNamespace crd labels name for namespaces
	CrdNameLabelsNamespace = "namespace.cloud.bkbcs.tencent.com"
	// CrdNameLabelsWorkloadKind crd labels name for workload king
	CrdNameLabelsWorkloadKind = "workloadkind.cloud.bkbcs.tencent.com"
	// CrdNameLabelsWorkloadName crd labels name for workload name
	CrdNameLabelsWorkloadName = "workloadname.cloud.bkbcs.tencent.com"
	// CrdNameLabelsStatus crd labels name for status
	CrdNameLabelsStatus = "status.cloud.bkbcs.tencent.com"
	// CrdNameLabelsIsFixed crd labels name for fixed
	CrdNameLabelsIsFixed = "fixed.cloud.bkbcs.tencent.com"
	// CrdNameLabelsEni  crd labels name for eni
	CrdNameLabelsEni = "eni.cloud.bkbcs.tencent.com"
	// CrdNameLabelsHost crd labels name for host
	CrdNameLabelsHost = "host.cloud.bkbcs.tencent.com"
)

// Client client for kube
type Client struct {
	cloudv1Client cloudv1set.CloudV1Interface
	subnetLister  listercloudv1.CloudSubnetLister
	ipLister      listercloudv1.CloudIPLister
	ipInformer    clientgocache.SharedIndexInformer
	k8sClientSet  kubernetes.Interface
	stopCh        chan struct{}
}

// EventHandler handler for informer event callback
type EventHandler struct{}

// NewEventHandler create event handler
func NewEventHandler() *EventHandler {
	return &EventHandler{}
}

// OnAdd add event
func (handler *EventHandler) OnAdd(obj interface{}) {}

// OnUpdate update event
func (handler *EventHandler) OnUpdate(objOld, objNew interface{}) {}

// OnDelete delete event
func (handler *EventHandler) OnDelete(obj interface{}) {}

// NewClient create new client for kube-apiserver
func NewClient(kubeconfig string) (*Client, error) {

	var restConfig *rest.Config
	var err error
	if len(kubeconfig) == 0 {
		blog.Infof("access kube-apiserver using incluster mod")
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			blog.Errorf("get incluster config failed, err %s", err.Error())
			return nil, fmt.Errorf("get incluster config failed, err %s", err.Error())
		}
	} else {
		blog.Infof("access kube-apiserver using kubeconfig %s", kubeconfig)
		//parse configuration
		restConfig, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			blog.Errorf("create internal client with kubeconfig %s failed, err %s", kubeconfig, err.Error())
			return nil, err
		}
	}
	restConfig.QPS = 1e6
	restConfig.Burst = 2e6

	clientset, err := bcsclientset.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("NewForConfig failed, err %s", err.Error())
		return nil, fmt.Errorf("NewForConfig failed, err %s", err.Error())
	}

	k8sClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("k8s NewForConfig failed err %s", err.Error())
		return nil, fmt.Errorf("k8s NewForConfig failed err %s", err.Error())
	}

	eventHandler := NewEventHandler()
	factory := bcsinformers.NewSharedInformerFactory(clientset, time.Duration(120)*time.Second)
	cloudSubnetInformer := factory.Cloud().V1().CloudSubnets()
	cloudSubnetInformer.Informer().AddEventHandler(eventHandler)
	cloudSubnetLister := factory.Cloud().V1().CloudSubnets().Lister()
	cloudIPInformer := factory.Cloud().V1().CloudIPs()
	ipInformer := cloudIPInformer.Informer()
	indexFuncContainerID := func(obj interface{}) ([]string, error) {
		cloudIP, ok := obj.(*cloudv1.CloudIP)
		if !ok {
			return nil, fmt.Errorf("%v is not CloudIP", obj)
		}
		vals := []string{utils.KeyToNamespacedKey(cloudIP.GetNamespace(), cloudIP.Spec.ContainerID)}
		return vals, nil
	}
	indexFuncPodName := func(obj interface{}) ([]string, error) {
		cloudIP, ok := obj.(*cloudv1.CloudIP)
		if !ok {
			return nil, fmt.Errorf("%v is not CloudIP", obj)
		}
		vals := []string{utils.KeyToNamespacedKey(cloudIP.GetNamespace(), cloudIP.Spec.PodName)}
		return vals, nil
	}
	ipInformer.AddIndexers(clientgocache.Indexers{utils.FieldIndexName("spec.containerID"): indexFuncContainerID})
	ipInformer.AddIndexers(clientgocache.Indexers{utils.FieldIndexName("spec.podName"): indexFuncPodName})
	ipInformer.AddEventHandler(eventHandler)
	cloudIPLister := factory.Cloud().V1().CloudIPs().Lister()

	cloudv1Client := clientset.CloudV1()

	stopCh := make(chan struct{})

	factory.Start(stopCh)
	blog.Infof("start cloud subnet informers factory")

	factory.WaitForCacheSync(stopCh)
	blog.Infof("wait for cloud subnet cache synced")

	return &Client{
		cloudv1Client: cloudv1Client,
		subnetLister:  cloudSubnetLister,
		ipLister:      cloudIPLister,
		ipInformer:    ipInformer,
		//eniLister:     cloudEniLister,
		k8sClientSet: k8sClientSet,
		stopCh:       stopCh,
	}, nil
}

// ensureNamespace create namespace when it's not existed
func (c *Client) ensureNamespace(ns string) error {
	_, err := c.k8sClientSet.CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			newNs := &corev1.Namespace{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Namespace",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: ns,
				},
			}
			_, err := c.k8sClientSet.CoreV1().Namespaces().Create(context.Background(), newNs, metav1.CreateOptions{})
			if err != nil {
				blog.Errorf("create ns %+v failed, err %s", err.Error())
				return fmt.Errorf("create ns %+v failed, err %s", newNs, err.Error())
			}
		}
		return fmt.Errorf("get kubernetes namespace %s failed, err %s", ns, err.Error())
	}
	return nil
}

// CreateSubnet create subnet
func (c *Client) CreateSubnet(ctx context.Context, subnet *types.CloudSubnet) error {
	timeNowStr := time.Now().UTC().String()
	newCloudSubnet := &cloudv1.CloudSubnet{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdNameCloudSubnet,
			APIVersion: CrdVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      subnet.SubnetID,
			Namespace: "bcs-system",
			Labels: map[string]string{
				CrdNameLabelsVpcID:  subnet.VpcID,
				CrdNameLabelsRegion: subnet.Region,
				CrdNameLabelsZone:   subnet.Zone,
			},
		},
		Spec: cloudv1.CloudSubnetSpec{
			SubnetID:   subnet.SubnetID,
			SubnetCidr: subnet.SubnetCidr,
			VpcID:      subnet.VpcID,
			Region:     subnet.Region,
			Zone:       subnet.Zone,
		},
		Status: cloudv1.CloudSubnetStatus{
			AvailableIPNum: subnet.AvailableIPNum,
			MinIPNumPerEni: subnet.MinIPNumPerEni,
			State:          subnet.State,
			CreateTime:     timeNowStr,
			UpdateTime:     timeNowStr,
		},
	}

	err := c.ensureNamespace("bcs-system")
	if err != nil {
		return err
	}
	_, err = c.cloudv1Client.CloudSubnets("bcs-system").Create(ctx, newCloudSubnet, metav1.CreateOptions{})
	if err != nil {
		blog.Infof("create crd %+v failed, err %s", newCloudSubnet, err.Error())
		return fmt.Errorf("create crd %+v failed, err %s", newCloudSubnet, err.Error())
	}

	return nil
}

// DeleteSubnet delete subnet
func (c *Client) DeleteSubnet(ctx context.Context, subnetID string) error {

	err := c.cloudv1Client.CloudSubnets("bcs-system").Delete(ctx, subnetID, metav1.DeleteOptions{})
	if err != nil {
		blog.Errorf("delete crd %s failed, err %s", subnetID, err.Error())
		return fmt.Errorf("delete crd %s failed, err %s", subnetID, err.Error())
	}

	return nil
}

// UpdateSubnetState update subnet state
func (c *Client) UpdateSubnetState(ctx context.Context, subnetID string, state, minIPNumPerEni int32) error {

	subnet, err := c.cloudv1Client.CloudSubnets("bcs-system").Get(ctx, subnetID, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("get subnet %s failed, err %s", subnetID, err.Error())
		return fmt.Errorf("get subnet %s failed, err %s", subnetID, err.Error())
	}

	timeNowStr := time.Now().UTC().String()
	updatedSubnet := &cloudv1.CloudSubnet{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdNameCloudSubnet,
			APIVersion: CrdVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            subnet.Name,
			Namespace:       subnet.Namespace,
			Labels:          subnet.Labels,
			ResourceVersion: subnet.ResourceVersion,
		},
		Spec: cloudv1.CloudSubnetSpec{
			SubnetID:   subnet.Spec.SubnetID,
			SubnetCidr: subnet.Spec.SubnetCidr,
			VpcID:      subnet.Spec.VpcID,
			Region:     subnet.Spec.Region,
			Zone:       subnet.Spec.Zone,
		},
		Status: cloudv1.CloudSubnetStatus{
			State:          state,
			AvailableIPNum: subnet.Status.AvailableIPNum,
			MinIPNumPerEni: minIPNumPerEni,
			CreateTime:     subnet.Status.CreateTime,
			UpdateTime:     timeNowStr,
		},
	}
	_, err = c.cloudv1Client.CloudSubnets("bcs-system").Update(ctx, updatedSubnet, metav1.UpdateOptions{})
	if err != nil {
		blog.Errorf("update subent failed, err %s", err.Error())
		return fmt.Errorf("update subent failed, err %s", err.Error())
	}

	return nil
}

// UpdateSubnetAvailableIP update subnet available
func (c *Client) UpdateSubnetAvailableIP(ctx context.Context, subnetID string, availableIP int64) error {
	subnet, err := c.cloudv1Client.CloudSubnets("bcs-system").Get(ctx, subnetID, metav1.GetOptions{})
	if err != nil {
		blog.Errorf("get subnet %s failed, err %s", subnetID, err.Error())
		return fmt.Errorf("get subnet %s failed, err %s", subnetID, err.Error())
	}

	timeNowStr := time.Now().UTC().String()
	updatedSubnet := &cloudv1.CloudSubnet{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdNameCloudSubnet,
			APIVersion: CrdVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            subnet.Name,
			Namespace:       subnet.Namespace,
			Labels:          subnet.Labels,
			ResourceVersion: subnet.ResourceVersion,
		},
		Spec: cloudv1.CloudSubnetSpec{
			SubnetID:   subnet.Spec.SubnetID,
			SubnetCidr: subnet.Spec.SubnetCidr,
			VpcID:      subnet.Spec.VpcID,
			Region:     subnet.Spec.Region,
			Zone:       subnet.Spec.Zone,
		},
		Status: cloudv1.CloudSubnetStatus{
			State:          subnet.Status.State,
			AvailableIPNum: availableIP,
			MinIPNumPerEni: subnet.Status.MinIPNumPerEni,
			CreateTime:     subnet.Status.CreateTime,
			UpdateTime:     timeNowStr,
		},
	}
	_, err = c.cloudv1Client.CloudSubnets("bcs-system").Update(ctx, updatedSubnet, metav1.UpdateOptions{})
	if err != nil {
		blog.Errorf("update subent failed, err %s", err.Error())
		return fmt.Errorf("update subent failed, err %s", err.Error())
	}
	return nil
}

// ListSubnet list subnet
func (c *Client) ListSubnet(ctx context.Context, labelsMap map[string]string) ([]*types.CloudSubnet, error) {
	var err error
	var selector labels.Selector
	if len(labelsMap) == 0 {
		selector = labels.Everything()
	} else {
		selector = labels.NewSelector()
		for k, v := range labelsMap {
			requirement, err := labels.NewRequirement(k, selection.Equals, []string{v})
			if err != nil {
				return nil, fmt.Errorf("create requirement failed, err %s", err.Error())
			}
			selector = selector.Add(*requirement)
		}
	}

	subnets, err := c.subnetLister.CloudSubnets("bcs-system").List(selector)
	if err != nil {
		blog.Errorf("list crd subnets failed, err %s", err.Error())
	}

	var retSubnets []*types.CloudSubnet
	if subnets != nil {
		for _, sn := range subnets {
			retSubnets = append(retSubnets, &types.CloudSubnet{
				SubnetID:       sn.Spec.SubnetID,
				VpcID:          sn.Spec.VpcID,
				Region:         sn.Spec.Region,
				Zone:           sn.Spec.Zone,
				SubnetCidr:     sn.Spec.SubnetCidr,
				State:          sn.Status.State,
				AvailableIPNum: sn.Status.AvailableIPNum,
				MinIPNumPerEni: sn.Status.MinIPNumPerEni,
				CreateTime:     sn.Status.CreateTime,
				UpdateTime:     sn.Status.UpdateTime,
			})
		}
	}

	return retSubnets, nil
}

// GetSubnet get subnet by name
func (c *Client) GetSubnet(ctx context.Context, subnetID string) (*types.CloudSubnet, error) {
	sn, err := c.subnetLister.CloudSubnets("bcs-system").Get(subnetID)
	if err != nil {
		blog.Errorf("get subnet from store failed, err %s", err.Error())
		return nil, fmt.Errorf("get subnet from store failed, err %s", err.Error())
	}
	return &types.CloudSubnet{
		SubnetID:       sn.Spec.SubnetID,
		VpcID:          sn.Spec.VpcID,
		Region:         sn.Spec.Region,
		Zone:           sn.Spec.Zone,
		SubnetCidr:     sn.Spec.SubnetCidr,
		State:          sn.Status.State,
		AvailableIPNum: sn.Status.AvailableIPNum,
		CreateTime:     sn.Status.CreateTime,
		UpdateTime:     sn.Status.UpdateTime,
	}, nil
}

// CreateIPObject create ip
func (c *Client) CreateIPObject(ctx context.Context, ip *types.IPObject) error {
	timeNow := time.Now()
	newIPObj := &cloudv1.CloudIP{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdNameCloudIP,
			APIVersion: CrdVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      ip.Address,
			Namespace: "bcs-system",
			Labels: map[string]string{
				CrdNameLabelsVpcID:     ip.VpcID,
				CrdNameLabelsRegion:    ip.Region,
				CrdNameLabelsSubnetID:  ip.SubnetID,
				CrdNameLabelsCluster:   ip.Cluster,
				CrdNameLabelsStatus:    ip.Status,
				CrdNameLabelsEni:       ip.EniID,
				CrdNameLabelsHost:      ip.Host,
				CrdNameLabelsNamespace: ip.Namespace,
				CrdNameLabelsIsFixed:   strconv.FormatBool(ip.IsFixed),
			},
		},
		Spec: cloudv1.CloudIPSpec{
			Address:      ip.Address,
			VpcID:        ip.VpcID,
			Region:       ip.Region,
			SubnetID:     ip.SubnetID,
			SubnetCidr:   ip.SubnetCidr,
			Cluster:      ip.Cluster,
			Namespace:    ip.Namespace,
			PodName:      ip.PodName,
			WorkloadName: ip.WorkloadName,
			WorkloadKind: ip.WorkloadKind,
			ContainerID:  ip.ContainerID,
			Host:         ip.Host,
			EniID:        ip.EniID,
			IsFixed:      ip.IsFixed,
			KeepDuration: ip.KeepDuration,
		},
		Status: cloudv1.CloudIPStatus{
			Status:     ip.Status,
			CreateTime: utils.FormatTime(timeNow),
			UpdateTime: utils.FormatTime(timeNow),
		},
	}

	_, err := c.cloudv1Client.CloudIPs("bcs-system").Create(ctx, newIPObj, metav1.CreateOptions{})
	if err != nil {
		blog.Errorf("create CloudIP to Store failed, err %s", err.Error())
		return fmt.Errorf("create CloudIP to Store failed, err %s", err.Error())
	}
	return nil
}

// UpdateIPObject update ip
func (c *Client) UpdateIPObject(ctx context.Context, ip *types.IPObject) (*types.IPObject, error) {
	if ip == nil {
		return nil, fmt.Errorf("ip object is nil")
	}
	timeNow := time.Now()
	newIPObj := &cloudv1.CloudIP{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdNameCloudIP,
			APIVersion: CrdVersionV1,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:            ip.Address,
			Namespace:       "bcs-system",
			ResourceVersion: ip.ResourceVersion,
			Labels: map[string]string{
				CrdNameLabelsVpcID:     ip.VpcID,
				CrdNameLabelsRegion:    ip.Region,
				CrdNameLabelsSubnetID:  ip.SubnetID,
				CrdNameLabelsCluster:   ip.Cluster,
				CrdNameLabelsStatus:    ip.Status,
				CrdNameLabelsEni:       ip.EniID,
				CrdNameLabelsHost:      ip.Host,
				CrdNameLabelsNamespace: ip.Namespace,
				CrdNameLabelsIsFixed:   strconv.FormatBool(ip.IsFixed),
			},
		},
		Spec: cloudv1.CloudIPSpec{
			Address:      ip.Address,
			VpcID:        ip.VpcID,
			Region:       ip.Region,
			SubnetID:     ip.SubnetID,
			SubnetCidr:   ip.SubnetCidr,
			Cluster:      ip.Cluster,
			Namespace:    ip.Namespace,
			PodName:      ip.PodName,
			WorkloadName: ip.WorkloadName,
			WorkloadKind: ip.WorkloadKind,
			ContainerID:  ip.ContainerID,
			Host:         ip.Host,
			EniID:        ip.EniID,
			IsFixed:      ip.IsFixed,
			KeepDuration: ip.KeepDuration,
		},
		Status: cloudv1.CloudIPStatus{
			Status:     ip.Status,
			CreateTime: utils.FormatTime(ip.CreateTime),
			UpdateTime: utils.FormatTime(timeNow),
		},
	}

	ipObj, err := c.cloudv1Client.CloudIPs("bcs-system").Update(ctx, newIPObj, metav1.UpdateOptions{})
	if err != nil {
		blog.Errorf("update CloudIP to store failed, err %s", err.Error())
		return nil, fmt.Errorf("update CloudIP to store failed, err %s", err.Error())
	}

	return &types.IPObject{
		Address:         ipObj.Spec.Address,
		VpcID:           ipObj.Spec.VpcID,
		Region:          ipObj.Spec.Region,
		SubnetID:        ipObj.Spec.SubnetID,
		SubnetCidr:      ipObj.Spec.SubnetCidr,
		Cluster:         ipObj.Spec.Cluster,
		Namespace:       ipObj.Spec.Namespace,
		PodName:         ipObj.Spec.PodName,
		WorkloadName:    ipObj.Spec.WorkloadName,
		WorkloadKind:    ipObj.Spec.WorkloadKind,
		ContainerID:     ipObj.Spec.ContainerID,
		Host:            ipObj.Spec.Host,
		EniID:           ipObj.Spec.EniID,
		IsFixed:         ipObj.Spec.IsFixed,
		Status:          ipObj.Status.Status,
		ResourceVersion: ipObj.ResourceVersion,
		CreateTime:      ip.CreateTime,
		UpdateTime:      timeNow,
	}, nil
}

// DeleteIPObject delete ip
func (c *Client) DeleteIPObject(ctx context.Context, ip string) error {
	err := c.cloudv1Client.CloudIPs("bcs-system").Delete(ctx, ip, metav1.DeleteOptions{})
	if err != nil {
		blog.Errorf("delete CloudIP from store failed, err %s", err.Error())
		return fmt.Errorf("delete CloudIP from store failed, err %s", err.Error())
	}
	return nil
}

// GetIPObject get ip
func (c *Client) GetIPObject(ctx context.Context, ip string) (*types.IPObject, error) {
	ipObj, err := c.ipLister.CloudIPs("bcs-system").Get(ip)
	if err != nil {
		blog.Errorf("get ip %s from store faile, err %s", ip, err.Error())
		// just return err here, caller can use errors.IsNotFound() to check the err
		return nil, err
	}

	createTime, err := utils.ParseTimeString(ipObj.Status.CreateTime)
	if err != nil {
		return nil, fmt.Errorf("parse create time failed, err %s", err.Error())
	}
	updateTime, err := utils.ParseTimeString(ipObj.Status.UpdateTime)
	if err != nil {
		return nil, fmt.Errorf("parse update time failed, err %s", err.Error())
	}

	return &types.IPObject{
		Address:         ipObj.Spec.Address,
		VpcID:           ipObj.Spec.VpcID,
		Region:          ipObj.Spec.Region,
		SubnetID:        ipObj.Spec.SubnetID,
		SubnetCidr:      ipObj.Spec.SubnetCidr,
		Cluster:         ipObj.Spec.Cluster,
		Namespace:       ipObj.Spec.Namespace,
		PodName:         ipObj.Spec.PodName,
		WorkloadName:    ipObj.Spec.WorkloadName,
		WorkloadKind:    ipObj.Spec.WorkloadKind,
		ContainerID:     ipObj.Spec.ContainerID,
		Host:            ipObj.Spec.Host,
		EniID:           ipObj.Spec.EniID,
		IsFixed:         ipObj.Spec.IsFixed,
		Status:          ipObj.Status.Status,
		ResourceVersion: ipObj.ResourceVersion,
		KeepDuration:    ipObj.Spec.KeepDuration,
		CreateTime:      createTime,
		UpdateTime:      updateTime,
	}, nil
}

// ListIPObjectByField by field selector
func (c *Client) ListIPObjectByField(ctx context.Context, fieldKey string, fieldValue string) (
	[]*types.IPObject, error) {
	objs, err := c.ipInformer.GetIndexer().ByIndex(utils.FieldIndexName(fieldKey), fieldValue)
	if err != nil {
		return nil, err
	}
	var ipList []*types.IPObject
	for _, obj := range objs {
		ip, ok := obj.(*cloudv1.CloudIP)
		if !ok {
			blog.Warnf("obj %v is not CloudIP", obj)
			continue
		}
		createTime, err := utils.ParseTimeString(ip.Status.CreateTime)
		if err != nil {
			return nil, fmt.Errorf("parse create time failed, err %s", err.Error())
		}
		updateTime, err := utils.ParseTimeString(ip.Status.UpdateTime)
		if err != nil {
			return nil, fmt.Errorf("parse update time failed, err %s", err.Error())
		}
		ipList = append(ipList, &types.IPObject{
			Address:         ip.Spec.Address,
			VpcID:           ip.Spec.VpcID,
			Region:          ip.Spec.Region,
			SubnetID:        ip.Spec.SubnetID,
			SubnetCidr:      ip.Spec.SubnetCidr,
			Cluster:         ip.Spec.Cluster,
			Namespace:       ip.Spec.Namespace,
			PodName:         ip.Spec.PodName,
			WorkloadName:    ip.Spec.WorkloadName,
			WorkloadKind:    ip.Spec.WorkloadKind,
			ContainerID:     ip.Spec.ContainerID,
			Host:            ip.Spec.Host,
			EniID:           ip.Spec.EniID,
			IsFixed:         ip.Spec.IsFixed,
			Status:          ip.Status.Status,
			ResourceVersion: ip.ResourceVersion,
			KeepDuration:    ip.Spec.KeepDuration,
			CreateTime:      createTime,
			UpdateTime:      updateTime,
		})
	}
	return ipList, nil
}

// ListIPObject list ips
func (c *Client) ListIPObject(ctx context.Context, labelsMap map[string]string) ([]*types.IPObject, error) {
	var err error
	var selector labels.Selector
	if len(labelsMap) == 0 {
		selector = labels.Everything()
	} else {
		selector = labels.NewSelector()
		for k, v := range labelsMap {
			requirement, err := labels.NewRequirement(k, selection.Equals, []string{v})
			if err != nil {
				return nil, fmt.Errorf("create requirement failed, err %s", err.Error())
			}
			selector = selector.Add(*requirement)
		}
	}

	ips, err := c.ipLister.CloudIPs("bcs-system").List(selector)
	if err != nil {
		blog.Errorf("list crd subnets failed, err %s", err.Error())
	}

	var ipList []*types.IPObject
	for _, ip := range ips {
		createTime, err := utils.ParseTimeString(ip.Status.CreateTime)
		if err != nil {
			return nil, fmt.Errorf("parse create time failed, err %s", err.Error())
		}
		updateTime, err := utils.ParseTimeString(ip.Status.UpdateTime)
		if err != nil {
			return nil, fmt.Errorf("parse update time failed, err %s", err.Error())
		}
		ipList = append(ipList, &types.IPObject{
			Address:         ip.Spec.Address,
			VpcID:           ip.Spec.VpcID,
			Region:          ip.Spec.Region,
			SubnetID:        ip.Spec.SubnetID,
			SubnetCidr:      ip.Spec.SubnetCidr,
			Cluster:         ip.Spec.Cluster,
			Namespace:       ip.Spec.Namespace,
			PodName:         ip.Spec.PodName,
			WorkloadName:    ip.Spec.WorkloadName,
			WorkloadKind:    ip.Spec.WorkloadKind,
			ContainerID:     ip.Spec.ContainerID,
			Host:            ip.Spec.Host,
			EniID:           ip.Spec.EniID,
			IsFixed:         ip.Spec.IsFixed,
			Status:          ip.Status.Status,
			ResourceVersion: ip.ResourceVersion,
			KeepDuration:    ip.Spec.KeepDuration,
			CreateTime:      createTime,
			UpdateTime:      updateTime,
		})
	}

	return ipList, nil
}

// Stop stop client
func (c *Client) Stop() {
	c.stopCh <- struct{}{}
}
