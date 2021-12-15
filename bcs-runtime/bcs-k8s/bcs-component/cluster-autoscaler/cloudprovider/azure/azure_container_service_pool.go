/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package azure

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2018-03-31/containerservice"
	"k8s.io/klog"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

//ContainerServiceAgentPool implements NodeGroup interface for agent pool deployed in ACS/AKS
type ContainerServiceAgentPool struct {
	azureRef
	manager *AzureManager
	util    *AzUtil

	minSize           int
	maxSize           int
	serviceType       string
	clusterName       string
	resourceGroup     string
	nodeResourceGroup string

	curSize     int
	lastRefresh time.Time
	mutex       sync.Mutex
}

//NewContainerServiceAgentPool constructs ContainerServiceAgentPool from the --node param
//and azure manager
func NewContainerServiceAgentPool(spec *dynamic.NodeGroupSpec, am *AzureManager) (*ContainerServiceAgentPool, error) {
	asg := &ContainerServiceAgentPool{
		azureRef: azureRef{
			Name: spec.Name,
		},
		minSize: spec.MinSize,
		maxSize: spec.MaxSize,
		manager: am,
		curSize: -1,
	}

	asg.util = &AzUtil{
		manager: am,
	}
	asg.serviceType = am.config.VMType
	asg.clusterName = am.config.ClusterName
	asg.resourceGroup = am.config.ResourceGroup

	// In case of AKS there is a different resource group for the worker nodes, where as for
	// ACS the vms are in the same group as that of the service.
	if am.config.VMType == vmTypeAKS {
		asg.nodeResourceGroup = am.config.NodeResourceGroup
	} else {
		asg.nodeResourceGroup = am.config.ResourceGroup
	}
	return asg, nil
}

//GetAKSAgentPool is an internal function which figures out ManagedClusterAgentPoolProfile from the list based on the pool name provided in the --node parameter passed
//to the autoscaler main
func (agentPool *ContainerServiceAgentPool) GetAKSAgentPool(agentProfiles *[]containerservice.ManagedClusterAgentPoolProfile) (ret *containerservice.ManagedClusterAgentPoolProfile) {
	for _, value := range *agentProfiles {
		profileName := *value.Name
		klog.V(5).Infof("AKS AgentPool profile name: %s", profileName)
		if strings.EqualFold(profileName, agentPool.azureRef.Name) {
			return &value
		}
	}

	return nil
}

//GetACSAgentPool is an internal function which figures out AgentPoolProfile from the list based on the pool name provided in the --node parameter passed
//to the autoscaler main
func (agentPool *ContainerServiceAgentPool) GetACSAgentPool(agentProfiles *[]containerservice.AgentPoolProfile) (ret *containerservice.AgentPoolProfile) {
	for _, value := range *agentProfiles {
		profileName := *value.Name
		klog.V(5).Infof("ACS AgentPool profile name: %s", profileName)
		if strings.EqualFold(profileName, agentPool.azureRef.Name) {
			return &value
		}
	}

	// Note: In some older ACS clusters, the name of the agentProfile can be different from kubernetes
	// label and vm pool tag. It can be have a "pool0" appended.
	// In the above loop we would check the normal case and if not returned yet, we will try matching
	// the node pool name with "pool0" and try a match as a workaround.
	for _, value := range *agentProfiles {
		profileName := *value.Name
		poolName := agentPool.azureRef.Name + "pool0"
		klog.V(5).Infof("Workaround match check - ACS AgentPool Profile: %s <=> Poolname: %s", profileName, poolName)
		if strings.EqualFold(profileName, poolName) {
			return &value
		}
	}

	return nil
}

// getAKSNodeCount gets node count for AKS agent pool.
func (agentPool *ContainerServiceAgentPool) getAKSNodeCount() (count int, err error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	managedCluster, err := agentPool.manager.azClient.managedContainerServicesClient.Get(ctx,
		agentPool.resourceGroup,
		agentPool.clusterName)
	if err != nil {
		klog.Errorf("Failed to get AKS cluster (name:%q): %v", agentPool.clusterName, err)
		return -1, err
	}

	pool := agentPool.GetAKSAgentPool(managedCluster.AgentPoolProfiles)
	if pool == nil {
		return -1, fmt.Errorf("could not find pool with name: %s", agentPool.azureRef)
	}

	if pool.Count != nil {
		return int(*pool.Count), nil
	}

	return 0, nil
}

// getACSNodeCount gets node count for ACS agent pool.
func (agentPool *ContainerServiceAgentPool) getACSNodeCount() (count int, err error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	acsCluster, err := agentPool.manager.azClient.containerServicesClient.Get(ctx,
		agentPool.resourceGroup,
		agentPool.clusterName)
	if err != nil {
		klog.Errorf("Failed to get ACS cluster (name:%q): %v", agentPool.clusterName, err)
		return -1, err
	}

	pool := agentPool.GetACSAgentPool(acsCluster.AgentPoolProfiles)
	if pool == nil {
		return -1, fmt.Errorf("could not find pool with name: %s", agentPool.azureRef)
	}

	if pool.Count != nil {
		return int(*pool.Count), nil
	}

	return 0, nil
}

// setAKSNodeCount sets node count for AKS agent pool.
func (agentPool *ContainerServiceAgentPool) setAKSNodeCount(count int) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	managedCluster, err := agentPool.manager.azClient.managedContainerServicesClient.Get(ctx,
		agentPool.resourceGroup,
		agentPool.clusterName)
	if err != nil {
		klog.Errorf("Failed to get AKS cluster (name:%q): %v", agentPool.clusterName, err)
		return err
	}

	pool := agentPool.GetAKSAgentPool(managedCluster.AgentPoolProfiles)
	if pool == nil {
		return fmt.Errorf("could not find pool with name: %s", agentPool.azureRef)
	}

	klog.Infof("Current size: %d, Target size requested: %d", *pool.Count, count)

	updateCtx, updateCancel := getContextWithCancel()
	defer updateCancel()
	*pool.Count = int32(count)
	aksClient := agentPool.manager.azClient.managedContainerServicesClient
	future, err := aksClient.CreateOrUpdate(updateCtx, agentPool.resourceGroup,
		agentPool.clusterName, managedCluster)
	if err != nil {
		klog.Errorf("Failed to update AKS cluster (%q): %v", agentPool.clusterName, err)
		return err
	}

	err = future.WaitForCompletionRef(updateCtx, aksClient.Client)
	isSuccess, realError := isSuccessHTTPResponse(future.Response(), err)
	if isSuccess {
		klog.V(3).Infof("aksClient.CreateOrUpdate for aks cluster %q success", agentPool.clusterName)
		return nil
	}

	klog.Errorf("aksClient.CreateOrUpdate for aks cluster %q failed: %v", agentPool.clusterName, realError)
	return realError
}

// setACSNodeCount sets node count for ACS agent pool.
func (agentPool *ContainerServiceAgentPool) setACSNodeCount(count int) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	acsCluster, err := agentPool.manager.azClient.containerServicesClient.Get(ctx,
		agentPool.resourceGroup,
		agentPool.clusterName)
	if err != nil {
		klog.Errorf("Failed to get ACS cluster (name:%q): %v", agentPool.clusterName, err)
		return err
	}

	pool := agentPool.GetACSAgentPool(acsCluster.AgentPoolProfiles)
	if pool == nil {
		return fmt.Errorf("could not find pool with name: %s", agentPool.azureRef)
	}

	klog.Infof("Current size: %d, Target size requested: %d", *pool.Count, count)

	updateCtx, updateCancel := getContextWithCancel()
	defer updateCancel()
	*pool.Count = int32(count)
	acsClient := agentPool.manager.azClient.containerServicesClient
	future, err := acsClient.CreateOrUpdate(updateCtx, agentPool.resourceGroup,
		agentPool.clusterName, acsCluster)
	if err != nil {
		klog.Errorf("Failed to update ACS cluster (%q): %v", agentPool.clusterName, err)
		return err
	}

	err = future.WaitForCompletionRef(updateCtx, acsClient.Client)
	isSuccess, realError := isSuccessHTTPResponse(future.Response(), err)
	if isSuccess {
		klog.V(3).Infof("acsClient.CreateOrUpdate for acs cluster %q success", agentPool.clusterName)
		return nil
	}

	klog.Errorf("acsClient.CreateOrUpdate for acs cluster %q failed: %v", agentPool.clusterName, realError)
	return realError
}

//GetNodeCount returns the count of nodes from the managed agent pool profile
func (agentPool *ContainerServiceAgentPool) GetNodeCount() (count int, err error) {
	if agentPool.serviceType == vmTypeAKS {
		return agentPool.getAKSNodeCount()
	}

	return agentPool.getACSNodeCount()
}

//SetNodeCount sets the count of nodes in the in memory pool profile
func (agentPool *ContainerServiceAgentPool) SetNodeCount(count int) (err error) {
	if agentPool.serviceType == vmTypeAKS {
		return agentPool.setAKSNodeCount(count)
	}

	return agentPool.setACSNodeCount(count)
}

//GetProviderID converts the name of a node into the form that kubernetes cloud
//provider id is presented in.
func (agentPool *ContainerServiceAgentPool) GetProviderID(name string) string {
	//TODO: come with a generic way to make it work with provider id formats
	// in different version of k8s.
	return "azure://" + name
}

//GetName extracts the name of the node (a format which underlying cloud service understands)
//from the cloud providerID (format which kubernetes understands)
func (agentPool *ContainerServiceAgentPool) GetName(providerID string) (string, error) {
	// Remove the "azure://" string from it
	providerID = strings.TrimPrefix(providerID, "azure://")
	ctx, cancel := getContextWithCancel()
	defer cancel()
	vms, err := agentPool.manager.azClient.virtualMachinesClient.List(ctx, agentPool.nodeResourceGroup)
	if err != nil {
		return "", err
	}
	for _, vm := range vms {
		if strings.EqualFold(*vm.ID, providerID) {
			return *vm.Name, nil
		}
	}
	return "", fmt.Errorf("VM list empty")
}

//MaxSize returns the maximum size scale limit provided by --node
//parameter to the autoscaler main
func (agentPool *ContainerServiceAgentPool) MaxSize() int {
	return agentPool.maxSize
}

//MinSize returns the minimum size the cluster is allowed to scaled down
//to as provided by the node spec in --node parameter.
func (agentPool *ContainerServiceAgentPool) MinSize() int {
	return agentPool.minSize
}

//TargetSize gathers the target node count set for the cluster by
//querying the underlying service.
func (agentPool *ContainerServiceAgentPool) TargetSize() (int, error) {
	agentPool.mutex.Lock()
	defer agentPool.mutex.Unlock()

	if agentPool.lastRefresh.Add(15 * time.Second).After(time.Now()) {
		return agentPool.curSize, nil
	}

	count, err := agentPool.GetNodeCount()
	if err != nil {
		return -1, err
	}
	klog.V(5).Infof("Got new size %d for agent pool (%q)", count, agentPool.Name)

	agentPool.curSize = count
	agentPool.lastRefresh = time.Now()
	return agentPool.curSize, nil
}

//SetSize contacts the underlying service and sets the size of the pool.
//This will be called when a scale up occurs and will be called just after
//a delete is performed from a scale down.
func (agentPool *ContainerServiceAgentPool) SetSize(targetSize int, isScalingDown bool) (err error) {
	agentPool.mutex.Lock()
	defer agentPool.mutex.Unlock()

	return agentPool.setSizeInternal(targetSize, isScalingDown)
}

// setSizeInternal contacts the underlying service and sets the size of the pool.
// It should be called under lock protected.
func (agentPool *ContainerServiceAgentPool) setSizeInternal(targetSize int, isScalingDown bool) (err error) {
	if isScalingDown && targetSize < agentPool.MinSize() {
		klog.Errorf("size-decreasing request of %d is smaller than min size %d", targetSize, agentPool.MinSize())
		return fmt.Errorf("size-decreasing request of %d is smaller than min size %d", targetSize, agentPool.MinSize())
	}

	klog.V(2).Infof("Setting size for cluster (%q) with new count (%d)", agentPool.clusterName, targetSize)
	if agentPool.serviceType == vmTypeAKS {
		err = agentPool.setAKSNodeCount(targetSize)
	} else {
		err = agentPool.setACSNodeCount(targetSize)
	}
	if err != nil {
		return err
	}

	agentPool.curSize = targetSize
	agentPool.lastRefresh = time.Now()
	return nil
}

//IncreaseSize calls in the underlying SetSize to increase the size in response
//to a scale up. It calculates the expected size based on a delta provided as
//parameter
func (agentPool *ContainerServiceAgentPool) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be +ve")
	}
	currentSize, err := agentPool.TargetSize()
	if err != nil {
		return err
	}
	targetSize := int(currentSize) + delta
	if targetSize > agentPool.MaxSize() {
		return fmt.Errorf("size-increasing request of %d is bigger than max size %d", targetSize, agentPool.MaxSize())
	}
	return agentPool.SetSize(targetSize, false)
}

// deleteNodesInternal calls the underlying vm service to delete the node.
// It should be called within lock protected.
func (agentPool *ContainerServiceAgentPool) deleteNodesInternal(providerIDs []string) (deleted int, err error) {
	for _, providerID := range providerIDs {
		klog.Infof("ProviderID got to delete: %s", providerID)
		nodeName, err := agentPool.GetName(providerID)
		if err != nil {
			return deleted, err
		}
		klog.Infof("VM name got to delete: %s", nodeName)

		err = agentPool.util.DeleteVirtualMachine(agentPool.nodeResourceGroup, nodeName)
		if err != nil {
			klog.Errorf("Failed to delete virtual machine %q with error: %v", nodeName, err)
			return deleted, err
		}

		// increase the deleted count after delete VM succeed.
		deleted++
	}

	return deleted, nil
}

//DeleteNodes extracts the providerIDs from the node spec and calls into the internal
//delete method.
func (agentPool *ContainerServiceAgentPool) DeleteNodes(nodes []*apiv1.Node) error {
	agentPool.mutex.Lock()
	defer agentPool.mutex.Unlock()

	var providerIDs []string
	for _, node := range nodes {
		providerIDs = append(providerIDs, node.Spec.ProviderID)
	}

	deleted, deleteError := agentPool.deleteNodesInternal(providerIDs)
	// Update node count if there're some virtual machines got deleted.
	if deleted != 0 {
		targetSize := agentPool.curSize - deleted
		err := agentPool.setSizeInternal(targetSize, true)
		if err != nil {
			klog.Errorf("Failed to set size for agent pool %q with error: %v", agentPool.Name, err)
		} else {
			klog.V(3).Infof("Size for agent pool %q has been updated to %d", agentPool.Name, targetSize)
		}
	}
	return deleteError
}

//IsContainerServiceNode checks if the tag from the vm matches the agentPool name
func (agentPool *ContainerServiceAgentPool) IsContainerServiceNode(tags map[string]*string) bool {
	poolName := tags["poolName"]
	if poolName != nil {
		klog.V(5).Infof("Matching agentPool name: %s with tag name: %s", agentPool.azureRef.Name, *poolName)
		if strings.EqualFold(*poolName, agentPool.azureRef.Name) {
			return true
		}
	}
	return false
}

//GetNodes extracts the node list from the underlying vm service and returns back
//equivalent providerIDs  as list.
func (agentPool *ContainerServiceAgentPool) GetNodes() ([]string, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()
	vmList, err := agentPool.manager.azClient.virtualMachinesClient.List(ctx, agentPool.nodeResourceGroup)
	if err != nil {
		klog.Errorf("Azure client list vm error : %v", err)
		return nil, err
	}
	var nodeArray []string
	for _, node := range vmList {
		klog.V(5).Infof("Node Name: %s, ID: %s", *node.Name, *node.ID)
		if agentPool.IsContainerServiceNode(node.Tags) {
			providerID, err := convertResourceGroupNameToLower(agentPool.GetProviderID(*node.ID))
			if err != nil {
				// This shouldn't happen. Log a waring message for tracking.
				klog.Warningf("GetNodes.convertResourceGroupNameToLower failed with error: %v", err)
				continue
			}

			klog.V(5).Infof("Returning back the providerID: %s", providerID)
			nodeArray = append(nodeArray, providerID)
		}
	}
	return nodeArray, nil
}

//DecreaseTargetSize requests the underlying service to decrease the node count.
func (agentPool *ContainerServiceAgentPool) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		klog.Errorf("Size decrease error: %d", delta)
		return fmt.Errorf("size decrease must be negative")
	}
	currentSize, err := agentPool.TargetSize()
	if err != nil {
		klog.Error(err)
		return err
	}
	klog.V(5).Infof("DecreaseTargetSize get current size %d for agent pool %q", currentSize, agentPool.Name)

	// Get the current nodes in the list
	nodes, err := agentPool.GetNodes()
	if err != nil {
		klog.Error(err)
		return err
	}

	targetSize := currentSize + delta
	klog.V(5).Infof("DecreaseTargetSize get target size %d for agent pool %q", targetSize, agentPool.Name)
	if targetSize < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			currentSize, delta, len(nodes))
	}
	return agentPool.SetSize(targetSize, true)
}

//Id returns the name of the agentPool
func (agentPool *ContainerServiceAgentPool) Id() string {
	return agentPool.azureRef.Name
}

//Debug returns a string with basic details of the agentPool
func (agentPool *ContainerServiceAgentPool) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", agentPool.Id(), agentPool.MinSize(), agentPool.MaxSize())
}

//Nodes returns the list of nodes in the agentPool.
func (agentPool *ContainerServiceAgentPool) Nodes() ([]cloudprovider.Instance, error) {
	instanceNames, err := agentPool.GetNodes()
	if err != nil {
		return nil, err
	}
	instances := make([]cloudprovider.Instance, 0, len(instanceNames))
	for _, instanceName := range instanceNames {
		instances = append(instances, cloudprovider.Instance{Id: instanceName})
	}
	return instances, nil
}

//TemplateNodeInfo is not implemented.
func (agentPool *ContainerServiceAgentPool) TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

//Exist is always true since we are initialized with an existing agentpool
func (agentPool *ContainerServiceAgentPool) Exist() bool {
	return true
}

//Create is returns already exists since we don't support the
//agent pool creation.
func (agentPool *ContainerServiceAgentPool) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

//Delete is not implemented since we don't support agent pool
//deletion.
func (agentPool *ContainerServiceAgentPool) Delete() error {
	return cloudprovider.ErrNotImplemented
}

//Autoprovisioned is set to false to indicate that this code
//does not create agentPools by itself.
func (agentPool *ContainerServiceAgentPool) Autoprovisioned() bool {
	return false
}
