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
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	autoscalerconfig "k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/cloudprovider/bcs/clustermanager"
)

const (
	// ProviderName is the cloud Provider name for QCLOUD
	ProviderName = "bcs"
)

var (
	availableGPUTypes = map[string]struct{}{
		"Tesla-P4": {},
		"M40":      {},
		"P100":     {},
		"V100":     {},
	}
)

// var gpuName = "alpha.kubernetes.io/nvidia-gpu"

var _ cloudprovider.CloudProvider = &Provider{}

// Provider implements CloudProvider interface.
type Provider struct {
	NodeGroupCache  *NodeGroupCache
	resourceLimiter *cloudprovider.ResourceLimiter
}

// BuildCloudProvider builds a cloud Provider.
func BuildCloudProvider(opts autoscalerconfig.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var (
		cache      *NodeGroupCache
		client     clustermanager.NodePoolClientInterface
		cloudError error
	)

	if opts.CloudConfig != "" {
		config, fileErr := os.Open(opts.CloudConfig)
		if fileErr != nil {
			klog.Fatalf("Couldn't open cloud Provider configuration %s: %#v", opts.CloudConfig, fileErr)
		}
		defer config.Close()
		cache, client, cloudError = CreateNodeGroupCache(config)
		if cloudError != nil {
			klog.Fatalf("Failed to create node group cache: %v", cloudError)
		}
	} else {
		cache, client, cloudError = CreateNodeGroupCache(nil)
		if cloudError != nil {
			klog.Fatalf("Failed to create node group cache: %v", cloudError)
		}
	}

	cloudProvider, err := BuildBcsCloudProvider(cache, client, do, rl)
	if err != nil {
		klog.Fatalf("Failed to create bcs cloud Provider: %v", err)
	}
	return cloudProvider
}

// BuildBcsCloudProvider builds CloudProvider implementation for Bcs.
func BuildBcsCloudProvider(cache *NodeGroupCache, client clustermanager.NodePoolClientInterface,
	discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, resourceLimiter *cloudprovider.ResourceLimiter) (
	cloudprovider.CloudProvider, error) {
	if discoveryOpts.StaticDiscoverySpecified() {
		cloud := &Provider{
			NodeGroupCache:  cache,
			resourceLimiter: resourceLimiter,
		}
		for _, spec := range discoveryOpts.NodeGroupSpecs {
			if err := cloud.addNodeGroup(spec, client); err != nil {
				return nil, err
			}
		}
		return cloud, nil
	}

	return nil, fmt.Errorf("Failed to build an BCS cloud Provider: Either node group specs or " +
		"node group auto discovery spec must be specified")
}

// GPULabel returns default gpu type
func (cloud *Provider) GPULabel() string {
	return gpu.DefaultGPUType
}

// GetAvailableGPUTypes returns available gpu types
func (cloud *Provider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

func (cloud *Provider) Cleanup() error {
	return nil
}

// Name returns name of the cloud Provider.
func (cloud *Provider) Name() string {
	return ProviderName
}

// NodeGroups returns all node groups configured for this cloud Provider.
func (cloud *Provider) NodeGroups() []cloudprovider.NodeGroup {
	groups := cloud.NodeGroupCache.GetRegisteredNodeGroups()
	result := make([]cloudprovider.NodeGroup, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node.
func (cloud *Provider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	ref, err := InstanceRefFromProviderID(node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}
	group, err := cloud.NodeGroupCache.FindForInstance(ref)
	if err != nil {
		klog.Errorf("Instance %v, node(%s) to found group err:%s ", ref, node.Name, err.Error())
		return group, err
	}

	if group == nil {
		klog.V(4).Infof("Instance %v, node(%s) is not found in any group", ref, node.Name)
		return group, err
	}
	return group, nil
}

// Pricing returns pricing model for this cloud Provider or error if not available.
func (cloud *Provider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud Provider.
func (cloud *Provider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided.
// The node group is not automatically created on the cloud Provider side.
// The node group is not returned by NodeGroups() until it is created.
func (cloud *Provider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (cloud *Provider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return cloud.resourceLimiter, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud Provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (cloud *Provider) Refresh() error {
	klog.V(4).Infof("Refresh loop")
	if cloud.NodeGroupCache == nil {
		klog.Errorf("Refresh cloud manager is nil")
		return fmt.Errorf("Refresh cloud manager is nil")
	}

	return cloud.NodeGroupCache.regenerateCache()
}

// addNodeGroup adds node group defined in string spec. Format:
// minNodes:maxNodes:groupName
func (cloud *Provider) addNodeGroup(spec string, client clustermanager.NodePoolClientInterface) error {
	group, err := buildNodeGroupFromSpec(spec, client)
	if err != nil {
		return err
	}

	cloud.NodeGroupCache.Register(group)
	return nil
}
