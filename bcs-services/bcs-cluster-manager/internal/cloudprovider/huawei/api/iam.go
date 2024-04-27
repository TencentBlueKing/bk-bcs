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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3/region"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// NewIamClient get iam client from common option
func NewIamClient(opt *cloudprovider.CommonOption) (*IamClient, error) {
	if opt == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}

	// global auth
	auth, err := getGlobalAuth(opt.Account.SecretID, opt.Account.SecretKey)
	if err != nil {
		return nil, err
	}

	// region
	defaultRegion := "cn-north-1"
	if opt.Region != "" {
		defaultRegion = opt.Region
	}
	rn, err := region.SafeValueOf(defaultRegion)
	if err != nil {
		return nil, err
	}

	// hc client
	hcClient, err := iam.IamClientBuilder().
		WithRegion(rn).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}
	iamClient := iam.NewIamClient(hcClient)

	// 创建IAM client
	return &IamClient{iam: iamClient}, nil
}

// IamClient iam client
type IamClient struct {
	iam *iam.IamClient
}

// ListCloudRegions get cloud regions
func (cli *IamClient) ListCloudRegions() ([]model.Region, error) {
	// 导入账号时 检验账号的有效性时使用错误的AKSK会panic
	defer utils.RecoverPrintStack("ListCloudRegions")

	keystoneListRegionsResponse, err := cli.iam.KeystoneListRegions(&model.KeystoneListRegionsRequest{})
	if err != nil {
		return nil, err
	}

	return *keystoneListRegionsResponse.Regions, nil
}

// ShowCloudRegion show cloud region
func (cli *IamClient) ShowCloudRegion(region string) (*model.Region, error) {
	// 区域详情
	keystoneShowRegionRequest := &model.KeystoneShowRegionRequest{}
	keystoneShowRegionRequest.RegionId = region
	keystoneShowRegionResponse, err := cli.iam.KeystoneShowRegion(keystoneShowRegionRequest)
	if err != nil {
		return nil, err
	}

	return keystoneShowRegionResponse.Region, nil
}

// ListAuthDomains 查询IAM用户可以访问的账号详情
func (cli *IamClient) ListAuthDomains() ([]model.Domains, error) {
	keystoneListAuthDomainsRequest := &model.KeystoneListAuthDomainsRequest{}
	keystoneListAuthDomainsResponse, err := cli.iam.KeystoneListAuthDomains(keystoneListAuthDomainsRequest)
	if err != nil {
		return nil, err
	}

	return *keystoneListAuthDomainsResponse.Domains, nil
}

// ListAuthProjects 查询IAM用户可以访问的项目列表
func (cli *IamClient) ListAuthProjects() ([]model.AuthProjectResult, error) {
	keystoneListAuthProjectsRequest := &model.KeystoneListAuthProjectsRequest{}
	keystoneListAuthProjectsResponse, err := cli.iam.KeystoneListAuthProjects(keystoneListAuthProjectsRequest)
	if err != nil {
		return nil, err
	}

	return *keystoneListAuthProjectsResponse.Projects, nil
}

// ListProjects 查询指定条件下的项目列表 (name 项目名称 - 即 regionId)
func (cli *IamClient) ListProjects(name string) ([]model.ProjectResult, error) {
	keystoneListProjectsRequest := &model.KeystoneListProjectsRequest{}

	var (
		defaultPage    int32 = 1
		defaultPerPage int32 = 5000
	)

	if name != "" {
		keystoneListProjectsRequest.Name = &name
	}
	keystoneListProjectsRequest.Page = &defaultPage
	keystoneListProjectsRequest.PerPage = &defaultPerPage

	keystoneListProjectsResponse, err := cli.iam.KeystoneListProjects(keystoneListProjectsRequest)
	if err != nil {
		return nil, err
	}

	return *keystoneListProjectsResponse.Projects, nil
}

// ShowProject 查询项目详情
func (cli *IamClient) ShowProject(projectId string) (*model.ProjectResult, error) {
	keystoneShowProjectRequest := &model.KeystoneShowProjectRequest{}
	keystoneShowProjectRequest.ProjectId = projectId

	keystoneShowProjectResponse, err := cli.iam.KeystoneShowProject(keystoneShowProjectRequest)
	if err != nil {
		return nil, err
	}

	return keystoneShowProjectResponse.Project, nil
}

// ShowProjectDetailsAndStatus 查询项目详情与状态
func (cli *IamClient) ShowProjectDetailsAndStatus(projectId string) (*model.ProjectDetailsAndStatusResult, error) {
	showProjectDetailsAndStatusRequest := &model.ShowProjectDetailsAndStatusRequest{}

	/*
		项目的状态信息，参数的值为"suspended"或"normal"。
		status值为"suspended"时，会将项目设置为冻结状态。
		status值为"normal"时，会将项目设置为正常（解冻）状态。
	*/
	showProjectDetailsAndStatusRequest.ProjectId = projectId
	showProjectDetailsAndStatusResponse, err := cli.iam.ShowProjectDetailsAndStatus(showProjectDetailsAndStatusRequest)
	if err != nil {
		return nil, err
	}

	return showProjectDetailsAndStatusResponse.Project, nil
}
