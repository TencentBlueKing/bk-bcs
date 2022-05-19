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

package bcs

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager"
)

const (
	maxRecordsReturnedByAPI = 100
)

var _ cloudprovider.NodeGroup = &NodeGroup{}

// NodeGroup implements NodeGroup interface.
type NodeGroup struct {
	InstanceRef
	scalingType  string
	instanceType string
	nodeGroupID  string
	minSize      int
	maxSize      int
	closedSize   int
	soldout      bool
	lk           sync.Mutex
	// nodeCache take node from api as source of true
	nodeCache map[string]string
	client    clustermanager.NodePoolClientInterface
}

// TimeRange defines crontab regular
type TimeRange struct {
	Name       string
	Schedule   string
	Zone       string
	DesiredNum int
}
type nodeTemplate struct {
	InstanceType string
	Region       string
	Resources    map[apiv1.ResourceName]resource.Quantity
	Label        map[string]string
}

// MaxSize returns maximum size of the node group.
func (group *NodeGroup) MaxSize() int {
	defer group.lk.Unlock()
	group.lk.Lock()
	return group.maxSize
}

// MinSize returns minimum size of the node group.
func (group *NodeGroup) MinSize() int {
	defer group.lk.Unlock()
	group.lk.Lock()
	return group.minSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kuberentes.
func (group *NodeGroup) TargetSize() (int, error) {
	pc, err := group.client.GetPoolConfig(group.nodeGroupID)
	if err != nil {
		return -1, err
	}
	return int(pc.DesiredSize), err
}

// IncreaseSize increases NodeGroup size
func (group *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := group.TargetSize()
	if err != nil {
		return err
	}
	if size+delta > group.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, group.MaxSize())
	}
	if group.IsSoldOut() {
		if group.scalingType != ScalingTypeWakeUpStopped {
			return fmt.Errorf("available instance type in selected-zone are sold out - group: %v", group.nodeGroupID)
		}
		if group.scalingType == ScalingTypeWakeUpStopped {
			if group.closedSize < delta {
				err := group.client.UpdateDesiredNode(group.nodeGroupID, size)
				if err != nil {
					return fmt.Errorf("available instance type in selected-zone are sold out,"+
						" starting up %v closed instances meet error %v - group: %v",
						group.closedSize, err.Error(), group.nodeGroupID)
				}
				return fmt.Errorf("available instance type in selected-zone are sold out,"+
					" starting up %v closed instances - group: %v",
					group.closedSize, group.nodeGroupID)
			}
		}
	}
	return group.client.UpdateDesiredNode(group.nodeGroupID, size+delta)
}

// IsSoldOut returns whether the instances of the node group are sold out
func (group *NodeGroup) IsSoldOut() bool {
	return group.soldout
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (group *NodeGroup) DecreaseTargetSize(delta int) error {
	// delta canbe positive, cause that scale down may failed.
	// if delta >= 0 {
	// 	return fmt.Errorf("size decrease size must be negative")
	// }
	size, err := group.TargetSize()
	if err != nil {
		return err
	}
	nodes, err := group.getGroupNodes()
	if err != nil {
		return err
	}
	if size+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return group.client.UpdateDesiredSize(group.nodeGroupID, size+delta)
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (group *NodeGroup) Belongs(node *apiv1.Node) (bool, error) {
	ip := getIP(node)
	if len(ip) == 0 {
		qcloudref, err := InstanceRefFromProviderID(node.Spec.ProviderID)
		if err != nil {
			return false, err
		}
		ip = group.nodeCache[qcloudref.Name]
		if len(ip) == 0 {
			return false, fmt.Errorf("cannot ensure the node")
		}
	}
	apiNode, err := group.client.GetNode(ip)
	if err != nil {
		return false, err
	}
	if apiNode.NodeGroupID != group.nodeGroupID {
		return false, fmt.Errorf("node %v pool %v not equal group pool %v",
			apiNode.InnerIP, apiNode.NodeGroupID, group.nodeGroupID)
	}
	return true, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (group *NodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (group *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (group *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (group *NodeGroup) Autoprovisioned() bool {
	return true
}

// DeleteNodes deletes the nodes from the group.
func (group *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	size, err := group.TargetSize()
	if err != nil {
		return err
	}

	if size <= group.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	ips := make([]string, 0, len(nodes))
	for _, node := range nodes {
		belongs, err := group.Belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("%s,%s belongs to a different group than %s", node.Name, node.Spec.ProviderID, group.Id())
		}
		ip := getIP(node)
		if len(ip) == 0 {
			qcloudref, err := InstanceRefFromProviderID(node.Spec.ProviderID)
			if err != nil {
				continue
			}
			ip = group.nodeCache[qcloudref.Name]
		}
		ips = append(ips, ip)
	}

	//TODO： max support 100, separates to multi requests
	if len(ips) < maxRecordsReturnedByAPI {
		klog.Infof("DeleteInstances len(%d)", len(ips))
		return group.deleteInstances(ips)
	}
	for i := 0; i < len(ips); i = i + maxRecordsReturnedByAPI {
		klog.Infof("page DeleteInstances i %d, len(%d)", i, len(ips))
		idx := math.Min(float64(i+maxRecordsReturnedByAPI), float64(len(ips)))
		err := group.deleteInstances(ips[i:int(idx)])
		if err != nil {
			return err
		}
		time.Sleep(intervalTimeDetach)
		klog.Infof("page DeleteInstances i %d, len(%d) done", i, len(ips))
	}
	return nil

}

// Id returns node group id.
func (group *NodeGroup) Id() string {
	return group.nodeGroupID
}

// Debug returns a debug string for the NodeGroup.
func (group *NodeGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", group.Id(), group.MinSize(), group.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (group *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	result := make([]cloudprovider.Instance, 0)
	instances, err := group.client.GetNodes(group.nodeGroupID)
	if err != nil {
		klog.Error(err)
		return []cloudprovider.Instance{}, err
	}
	cache := map[string]string{}
	for _, instance := range instances {
		if instance.Status == "DELETING" {
			continue
		}
		if instance.NodeGroupID != group.nodeGroupID {
			continue
		}

		i := cloudprovider.Instance{
			Id:     fmt.Sprintf("qcloud:///%v/%s", instance.Zone, instance.NodeID),
			Status: &cloudprovider.InstanceStatus{},
		}
		cache[instance.NodeID] = instance.InnerIP
		switch instance.Status {
		case "creating":
			i.Status.State = cloudprovider.InstanceCreating
		case "running":
			// check more node status
			i.Status.State = cloudprovider.InstanceRunning
		default:
			i.Status.State = cloudprovider.InstanceDeleting
		}
		result = append(result, i)
	}
	group.nodeCache = cache
	return result, nil
}

// TemplateNodeInfo returns template node info of the node group
func (group *NodeGroup) TemplateNodeInfo() (*nodeinfo.NodeInfo, error) {
	template, err := group.getNodeTemplate()
	if err != nil {
		return nil, err
	}

	node, err := group.buildNodeFromTemplate(template)
	if err != nil {
		return nil, err
	}

	nodeInfo := nodeinfo.NewNodeInfo()
	err = nodeInfo.SetNode(node)
	if err != nil {
		return nil, err
	}
	return nodeInfo, nil
}

// GetNodeGroup returns the information of node group
func (group *NodeGroup) GetNodeGroup() (*clustermanager.NodeGroup, error) {
	return group.client.GetPool(group.nodeGroupID)
}

// GetGroupNodes returns NodeGroup nodes.
func (group *NodeGroup) getGroupNodes() ([]string, error) {
	nodes, err := group.client.GetNodes(group.nodeGroupID)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0)
	for _, node := range nodes {
		if node.NodeGroupID != group.nodeGroupID {
			continue
		}
		if node.Status == "running" || node.Status == "creating" {
			result = append(result, node.NodeID)
		}
	}
	return result, nil
}

// deleteInstances deletes the given instances. All instances must be controlled by the same ASG.
func (group *NodeGroup) deleteInstances(ips []string) error {
	groupID := group.nodeGroupID
	klog.V(4).Infof("Start remove nodes %v", ips)
	return group.client.RemoveNodes(groupID, ips)
}

func (group *NodeGroup) getNodeTemplate() (*nodeTemplate, error) {
	nodeGroup, err := group.client.GetPool(group.nodeGroupID)
	if err != nil {
		return nil, err
	}
	if nodeGroup.AutoScaling == nil || nodeGroup.LaunchTemplate == nil {
		return nil, fmt.Errorf("node group scaling info is not set")
	}
	resources := convertResource(nodeGroup.LaunchTemplate)
	return &nodeTemplate{
		InstanceType: nodeGroup.LaunchTemplate.InstanceType,
		Region:       nodeGroup.Region,
		Resources:    resources,
		Label:        nodeGroup.Labels,
	}, nil
}

func (group *NodeGroup) buildNodeFromTemplate(template *nodeTemplate) (*apiv1.Node, error) {
	node := apiv1.Node{}
	nodeName := fmt.Sprintf("%s-%d", group.nodeGroupID, rand.Int63())

	node.ObjectMeta = metav1.ObjectMeta{
		Name:     nodeName,
		SelfLink: fmt.Sprintf("/api/v1/nodes/%s", nodeName),
		Labels:   map[string]string{},
	}

	node.Status = apiv1.NodeStatus{
		Capacity: apiv1.ResourceList{},
	}
	node.Status.Capacity = template.Resources
	// TODO: get a real value.
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	if _, ok := node.Status.Capacity[gpu.ResourceNvidiaGPU]; !ok {
		node.Status.Capacity[gpu.ResourceNvidiaGPU] = template.Resources["gpu"]
	}

	// TODO: use proper allocatable!!
	node.Status.Allocatable = node.Status.Capacity

	node.Labels = cloudprovider.JoinStringMaps(node.Labels, template.Label)
	// GenericLabels
	node.Labels = cloudprovider.JoinStringMaps(node.Labels, buildGenericLabels(template, nodeName))

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	return &node, nil
}

func getIP(node *apiv1.Node) string {
	for _, address := range node.Status.Addresses {
		if address.Type != apiv1.NodeInternalIP {
			continue
		}
		return address.Address
	}
	return ""
}

// TimeRanges returns the crontab regulars of the node group
func (group *NodeGroup) TimeRanges() ([]*TimeRange, error) {
	result := make([]*TimeRange, 0)
	pc, err := group.client.GetPoolConfig(group.nodeGroupID)
	if err != nil {
		return result, err
	}
	for _, t := range pc.TimeRanges {
		result = append(result, &TimeRange{
			Name:       t.Name,
			Schedule:   t.Schedule,
			Zone:       t.Zone,
			DesiredNum: int(t.DesiredNum),
		})
	}
	return result, err
}
