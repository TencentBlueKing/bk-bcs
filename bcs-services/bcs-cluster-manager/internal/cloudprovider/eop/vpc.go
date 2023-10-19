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

package eop

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/eop/api"
	"strconv"
	"sync"
)

var vpcMgr sync.Once

func init() {
	vpcMgr.Do(func() {
		// init VPC manager
		cloudprovider.InitVPCManager(cloudName, &VPCClient{})
	})
}

// VPCClient is the client for VPC
type VPCClient struct {
}

// ListVPCs list vpc
func (V VPCClient) ListVPCs(vpcID string, opt *cloudprovider.CommonOption) ([]*cmproto.CloudVPC, error) {
	cli, err := api.NewCTClient(opt)
	if err != nil {
		blog.Errorf("ListVPCs create eck client when failed: %v", err)
		return nil, err
	}

	result, err := cli.ListVpcs(opt.Region)
	if err != nil {
		blog.Errorf("ListVPCs failed: %v", err)
		return nil, err
	}

	vpcs := make([]*cmproto.CloudVPC, 0)

	if vpcID != "" {
		for _, v := range result {
			if v.Name == vpcID {
				vpcs = append(vpcs, &cmproto.CloudVPC{
					CloudID: cloudName,
					Region:  opt.Region,
					VpcID:   strconv.Itoa(int(v.VpcId)),
					VpcName: v.Name,
				})
				break
			}
		}
		return vpcs, nil
	}

	for _, v := range result {
		vpcs = append(vpcs, &cmproto.CloudVPC{
			CloudID: cloudName,
			Region:  opt.Region,
			VpcID:   strconv.Itoa(int(v.VpcId)),
			VpcName: v.Name,
		})
	}

	return vpcs, nil
}

// ListSubnets list vpc subnets
func (V VPCClient) ListSubnets(vpcID string, opt *cloudprovider.CommonOption) ([]*cmproto.Subnet, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}

// ListSecurityGroups list security groups
func (V VPCClient) ListSecurityGroups(opt *cloudprovider.CommonOption) ([]*cmproto.SecurityGroup, error) {
	return nil, cloudprovider.ErrCloudNotImplemented
}
