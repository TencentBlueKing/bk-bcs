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

// Package api xxx
package api

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/pkg/errors"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// AksServiceImpl azure服务
type AksServiceImpl struct {
	resourcesGroup       string
	netClient            *armnetwork.InterfacesClient
	resourceClient       *armcompute.ResourceSKUsClient
	poolClient           *armcontainerservice.AgentPoolsClient
	setClient            *armcompute.VirtualMachineScaleSetsClient
	vmClient             *armcompute.VirtualMachineScaleSetVMsClient
	vnetClient           *armnetwork.VirtualNetworksClient
	clustersClient       *armcontainerservice.ManagedClustersClient
	securityGroupsClient *armnetwork.SecurityGroupsClient
	resourceGroupsClient *armnetwork.SecurityGroupsClient
}

// NewAksServiceImplWithCommonOption 从 CommonOption 创建 AksService
func NewAksServiceImplWithCommonOption(opt *cloudprovider.CommonOption) (AksService, error) {
	if opt == nil || opt.Account == nil {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	account := opt.Account
	if len(account.SubscriptionID) == 0 || len(account.TenantID) == 0 ||
		len(account.ClientID) == 0 || len(account.ClientSecret) == 0 ||
		len(account.ResourceGroupName) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	return NewAKsServiceImpl(account.SubscriptionID, account.TenantID, account.ClientID, account.ClientSecret,
		account.ResourceGroupName)
}

// NewAksServiceImplWithAccount 从 Account 创建 AksService
func NewAksServiceImplWithAccount(account *proto.Account) (AksService, error) {
	if account == nil {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(account.SubscriptionID) == 0 || len(account.TenantID) == 0 ||
		len(account.ClientID) == 0 || len(account.ClientSecret) == 0 ||
		len(account.ResourceGroupName) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}

	return NewAKsServiceImpl(account.SubscriptionID, account.TenantID, account.ClientID, account.ClientSecret,
		account.ResourceGroupName)
}

// NewAKsServiceImpl 创建AksService
func NewAKsServiceImpl(subscriptionID, tenantID, clientID, clientSecret, resourceGroupName string) (AksService, error) {
	if len(subscriptionID) == 0 || len(tenantID) == 0 || len(clientID) == 0 || len(clientSecret) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	cred, err := getClientCredential(tenantID, clientID, clientSecret)
	if err != nil {
		return nil, errors.Wrapf(err, "get Azure Credential failed,TenantID:%s", tenantID)
	}
	poolClient, err := armcontainerservice.NewAgentPoolsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create pool client,SubscriptionID:%s,", subscriptionID)
	}
	clustersClient, err := armcontainerservice.NewManagedClustersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create cluster client,SubscriptionID:%s", subscriptionID)
	}
	setClient, err := armcompute.NewVirtualMachineScaleSetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create vmScaleSets client,SubscriptionID:%s", subscriptionID)
	}
	vmClient, err := armcompute.NewVirtualMachineScaleSetVMsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create vmScaleSetVMs client,SubscriptionID:%s", subscriptionID)
	}
	netClient, err := armnetwork.NewInterfacesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create networkInterfaces client,SubscriptionID:%s", subscriptionID)
	}
	resourceClient, err := armcompute.NewResourceSKUsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create resource sku client,SubscriptionID:%s", subscriptionID)
	}
	vnetClient, err := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create virtual networks client,SubscriptionID:%s",
			subscriptionID)
	}
	securityGroupsClient, err := armnetwork.NewSecurityGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create security groups client,SubscriptionID:%s",
			subscriptionID)
	}

	return &AksServiceImpl{
		resourcesGroup:       resourceGroupName,
		vmClient:             vmClient,
		setClient:            setClient,
		netClient:            netClient,
		poolClient:           poolClient,
		clustersClient:       clustersClient,
		resourceClient:       resourceClient,
		vnetClient:           vnetClient,
		securityGroupsClient: securityGroupsClient,
	}, nil
}

// GetCluster 查询集群
//
// resourceGroupName - 资源组名称(Account.resourceGroupName)
//
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) GetCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
	*armcontainerservice.ManagedCluster, error) {
	resourceGroupName := info.CmOption.Account.ResourceGroupName
	return aks.GetClusterWithName(ctx, resourceGroupName, info.Cluster.SystemID)
}

// GetClusterWithName 查询集群
//
// resourceGroupName - 资源组名称(Account.resourceGroupName)
//
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) GetClusterWithName(ctx context.Context, resourceGroupName, resourceName string) (
	*armcontainerservice.ManagedCluster, error) {
	resp, err := aks.clustersClient.Get(ctx, resourceGroupName, resourceName, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request")
	}
	return &resp.ManagedCluster, nil
}

// ListClusterByLocation 根据位置查询该地区下的所有集群(不区分资源组)
//
// location - 位置
func (aks *AksServiceImpl) ListClusterByLocation(ctx context.Context, location string) (
	[]*armcontainerservice.ManagedCluster, error) {
	result := make([]*armcontainerservice.ManagedCluster, 0)
	pager := aks.clustersClient.NewListPager(nil)
	for pager.More() {
		next, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to advance page")
		}
		for i, v := range next.Value {
			if strings.ToLower(*v.Location) == strings.ToLower(location) { // nolint
				result = append(result, next.Value[i])
			}
		}
	}
	return result, nil
}

// ListClusterByResourceGroupName 查询集群列表
//
// location - 位置
//
// resourceGroupName - 资源组名称(Account.resourceGroupName)
func (aks *AksServiceImpl) ListClusterByResourceGroupName(ctx context.Context, location, resourceGroupName string) (
	[]*armcontainerservice.ManagedCluster, error) {
	result := make([]*armcontainerservice.ManagedCluster, 0)
	pager := aks.clustersClient.NewListByResourceGroupPager(resourceGroupName, nil)
	for pager.More() {
		next, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to advance page")
		}
		for i, v := range next.Value {
			if strings.ToLower(*v.Location) == strings.ToLower(location) { // nolint
				result = append(result, next.Value[i])
			}
		}
	}
	return result, nil
}

// DeleteCluster 删除集群
//
// resourceGroupName - 资源组名称(Account.resourceGroupName)
//
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) DeleteCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	resourceGroupName := info.CmOption.Account.ResourceGroupName
	return aks.DeleteClusterWithName(ctx, resourceGroupName, info.Cluster.SystemID)
}

// DeleteClusterWithName 删除集群
//
// resourceGroupName - 资源组名称(Account.resourceGroupName)
//
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) DeleteClusterWithName(ctx context.Context, resourceGroupName, resourceName string) error {
	poller, err := aks.clustersClient.BeginDelete(ctx, resourceGroupName, resourceName, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to finish the request, resourcesGroupName: %s, cluster name: %s",
			resourceGroupName, resourceName)
	}
	if _, err = poller.PollUntilDone(ctx, pollFrequency4); err != nil {
		return errors.Wrapf(err, "failed to finish the request, resourcesGroupName: %s, cluster name: %s",
			resourceGroupName, resourceName)
	}
	return nil
}

// GetClusterAdminCredentials 获取集群管理凭证
//
// resourceGroupName - 资源组名称(Account.resourceGroupName)
//
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) GetClusterAdminCredentials(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
	[]*armcontainerservice.CredentialResult, error) {
	resourceGroupName := info.CmOption.Account.ResourceGroupName
	return aks.GetClusterAdminCredentialsWithName(ctx, resourceGroupName, info.Cluster.SystemID)
}

// GetClusterAdminCredentialsWithName 获取集群管理凭证
//
// resourceGroupName - 资源组名称(Account.resourceGroupName)
//
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) GetClusterAdminCredentialsWithName(
	ctx context.Context, resourceGroupName, resourceName string) (
	[]*armcontainerservice.CredentialResult, error) {
	credentials, err := aks.clustersClient.ListClusterAdminCredentials(ctx, resourceGroupName, resourceName, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request")
	}
	return credentials.Kubeconfigs, nil
}

// GetClusterUserCredentials 获取集群用户凭证
//
// resourceGroupName - 资源组名称(Account.resourceGroupName)
//
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) GetClusterUserCredentials(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
	[]*armcontainerservice.CredentialResult, error) {
	resourceGroupName := info.CmOption.Account.ResourceGroupName
	return aks.GetClusterUserCredentialsWithName(ctx, resourceGroupName, info.Cluster.SystemID)
}

// GetClusterUserCredentialsWithName 获取集群用户凭证
//
// resourceGroupName - 资源组名称(Account.resourceGroupName)
//
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) GetClusterUserCredentialsWithName(
	ctx context.Context, resourceGroupName, resourceName string) (
	[]*armcontainerservice.CredentialResult, error) {
	credentials, err := aks.clustersClient.ListClusterUserCredentials(ctx, resourceGroupName, resourceName, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request")
	}
	return credentials.Kubeconfigs, nil
}

// GetClusterMonitoringUserCredentials 获取集群监控凭证
//
// resourceGroupName - 资源组名称(Account.resourceGroupName)
//
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) GetClusterMonitoringUserCredentials(
	ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (
	[]*armcontainerservice.CredentialResult, error) {
	resourceGroupName := info.CmOption.Account.ResourceGroupName
	return aks.GetClusterMonitorUserCredWithName(ctx, resourceGroupName, info.Cluster.SystemID)
}

// GetClusterMonitorUserCredWithName 获取集群监控凭证(GetClusterMonitoringUserCredentialsWithName)
//
// resourceGroupName - 资源组名称(Account.resourceGroupName)
//
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) GetClusterMonitorUserCredWithName(
	ctx context.Context, resourceGroupName, resourceName string) (
	[]*armcontainerservice.CredentialResult, error) {
	credentials, err := aks.clustersClient.ListClusterMonitoringUserCredentials(ctx, resourceGroupName, resourceName,
		nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request")
	}
	return credentials.Kubeconfigs, nil
}
