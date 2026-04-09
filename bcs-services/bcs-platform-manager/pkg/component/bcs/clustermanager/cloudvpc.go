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

// Package clustermanager xxx
package clustermanager

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/constants"
)

// ListCloudVpc list cloud vpc from cluster manager
func ListCloudVpc(ctx context.Context,
	req *clustermanager.ListCloudVPCRequest) (*clustermanager.ListCloudVPCResponse, error) {
	cli, close, err := clustermanager.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.ListCloudVPC(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListCloudVpc error: %s", err)
	}
	if p.Code != 0 {
		return nil, fmt.Errorf("ListCloudVpc error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	return p, nil
}

// ListCloudVpcsPage list cloud vpcs page from cluster manager
func ListCloudVpcsPage(ctx context.Context,
	req *clustermanager.ListCloudVpcsPageRequest) (*clustermanager.ListCloudVpcsPageResponse, error) {
	cli, close, err := clustermanager.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.ListCloudVpcsPage(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListCloudVpcsPage error: %s", err)
	}
	if p.Code != 0 {
		return nil, fmt.Errorf("ListCloudVpcsPage error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	return p, nil
}

// ListCloudVpcCluster list cloud vpcs cluster from cluster manager
func ListCloudVpcCluster(ctx context.Context,
	req *clustermanager.ListCloudVpcClusterRequest) (*clustermanager.ListCloudVpcClusterResponse, error) {
	cli, close, err := clustermanager.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.ListCloudVpcCluster(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListCloudVpcCluster error: %s", err)
	}
	if p.Code != 0 {
		return nil, fmt.Errorf("ListCloudVpcCluster error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	return p, nil
}

// UpdateCloudVpcs update cloud vpcs cluster from cluster manager
func UpdateCloudVpcs(ctx context.Context,
	req *clustermanager.UpdateCloudVpcsRequest) (*clustermanager.UpdateCloudVpcsResponse, error) {
	cli, close, err := clustermanager.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.UpdateCloudVpcs(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UpdateCloudVpcs error: %s", err)
	}
	if p.Code != 0 {
		return nil, fmt.Errorf("UpdateCloudVpcs error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	return p, nil
}

// ListCloudSubnets list cloud subnets cluster from cluster manager
func ListCloudSubnets(ctx context.Context,
	req *clustermanager.ListCloudSubnetsRequest) (*clustermanager.ListCloudSubnetsResponse, error) {
	cli, close, err := clustermanager.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.ListCloudSubnets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListCloudSubnets error: %s", err)
	}
	if p.Code != 0 {
		return nil, fmt.Errorf("ListCloudSubnets error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	return p, nil
}

// CreateCloudSubnets create cloud subnets cluster from cluster manager
func CreateCloudSubnets(ctx context.Context,
	req *clustermanager.CreateCloudSubnetsRequest) (*clustermanager.CreateCloudSubnetsResponse, error) {
	cli, close, err := clustermanager.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.CreateCloudSubnets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("CreateCloudSubnets error: %s", err)
	}
	if p.Code != 0 || p.Data == nil {
		return nil, fmt.Errorf("CreateCloudSubnets error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	return p, nil
}

// UpdateCloudSubnets update cloud subnets cluster from cluster manager
func UpdateCloudSubnets(ctx context.Context,
	req *clustermanager.UpdateCloudSubnetsRequest) (*clustermanager.UpdateCloudSubnetsResponse, error) {
	cli, close, err := clustermanager.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.UpdateCloudSubnets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("UpdateCloudSubnets error: %s", err)
	}
	if p.Code != 0 {
		return nil, fmt.Errorf("UpdateCloudSubnets error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	return p, nil
}

// DeleteCloudSubnets delete cloud subnets cluster from cluster manager
func DeleteCloudSubnets(ctx context.Context,
	req *clustermanager.DeleteCloudSubnetsRequest) (*clustermanager.DeleteCloudSubnetsResponse, error) {
	cli, close, err := clustermanager.GetClient(constants.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return nil, err
	}
	p, err := cli.DeleteCloudSubnets(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("DeleteCloudSubnets error: %s", err)
	}
	if p.Code != 0 {
		return nil, fmt.Errorf("DeleteCloudSubnets error, code: %d, message: %s", p.Code, p.GetMessage())
	}
	return p, nil
}
