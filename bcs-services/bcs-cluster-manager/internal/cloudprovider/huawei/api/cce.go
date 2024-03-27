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
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	cce "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/region"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// CceClient cce client
type CceClient struct {
	*cce.CceClient
}

// NewCceClient init cce client
func NewCceClient(opt *cloudprovider.CommonOption) (*CceClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}

	projectID, err := GetProjectIDByRegion(opt)
	if err != nil {
		return nil, err
	}

	auth, err := basic.NewCredentialsBuilder().WithAk(opt.Account.SecretID).WithSk(opt.Account.SecretKey).
		WithProjectId(projectID).SafeBuild()
	if err != nil {
		return nil, err
	}

	rn, err := region.SafeValueOf(opt.Region)
	if err != nil {
		return nil, err
	}

	// 创建CCE client
	hcClient, err := cce.CceClientBuilder().WithCredential(auth).WithRegion(rn).SafeBuild()
	if err != nil {
		return nil, err
	}

	return &CceClient{
		CceClient: cce.NewCceClient(hcClient),
	}, nil
}

// ListCceCluster get cce cluster list, region parameter init tke client
func (cli *CceClient) ListCceCluster() (*model.ListClustersResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := model.ListClustersRequest{}
	rsp, err := cli.ListClusters(&req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

// GetCceCluster get cce cluster
func (cli *CceClient) GetCceCluster(clusterID string) (*model.ShowClusterResponse, error) {
	if cli == nil {
		return nil, cloudprovider.ErrServerIsNil
	}

	req := model.ShowClusterRequest{
		ClusterId: clusterID,
	}
	rsp, err := cli.ShowCluster(&req)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}
