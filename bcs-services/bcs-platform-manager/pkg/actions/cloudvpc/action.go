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

// Package cloudvpc cloudvpc operate
package cloudvpc

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"github.com/golang/protobuf/ptypes/wrappers"

	clustermgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// CloudVPCAction cloud vpc action interface
type CloudVPCAction interface { // nolint
	// CreateCloudVPC create cloud vpc
	CreateCloudVPC(ctx context.Context, req *types.CreateCloudVPCReq) (bool, error)
	UpdateCloudVPC(ctx context.Context, req *types.UpdateCloudVPCReq) (bool, error)
	GetCloudVPCRecommendCIDR(ctx context.Context, req *types.GetCloudVPCRecommendCIDRReq) (*types.GetCloudVPCRecommendCIDRResp, error)
}

// Action action for cloud vpc
type Action struct{}

// NewCloudVPCAction new cloud vpc action
func NewCloudVPCAction() CloudVPCAction {
	return &Action{}
}

// CreateCloudVPC create cloud vpc
func (a *Action) CreateCloudVPC(ctx context.Context, req *types.CreateCloudVPCReq) (bool, error) {
	cmReq := &clustermanager.CreateCloudVPCRequest{
		CloudID:       req.CloudID,
		NetworkType:   req.NetworkType,
		Region:        req.Region,
		RegionName:    req.RegionName,
		VpcName:       req.VpcName,
		VpcID:         req.VpcID,
		Available:     req.Available,
		Extra:         req.Extra,
		Creator:       req.Creator,
		ReservedIPNum: req.ReservedIPNum,
		BusinessID:    req.BusinessID,
		Overlay: func() *clustermanager.Cidr {
			if req.Overlay == nil {
				return nil
			}
			cidrs := make([]*clustermanager.CidrState, 0)
			for _, cidr := range req.Overlay.Cidrs {
				cidrs = append(cidrs, &clustermanager.CidrState{
					Cidr:  cidr.Cidr,
					Block: cidr.Block,
				})
			}
			return &clustermanager.Cidr{
				Cidrs:         cidrs,
				ReservedIPNum: req.Overlay.ReservedIPNum,
				ReservedCidrs: req.Overlay.ReservedCidrs,
			}
		}(),
		Underlay: func() *clustermanager.Cidr {
			if req.Underlay == nil {
				return nil
			}
			cidrs := make([]*clustermanager.CidrState, 0)
			for _, cidr := range req.Underlay.Cidrs {
				cidrs = append(cidrs, &clustermanager.CidrState{
					Cidr:  cidr.Cidr,
					Block: cidr.Block,
				})
			}
			return &clustermanager.Cidr{
				Cidrs:         cidrs,
				ReservedIPNum: req.Underlay.ReservedIPNum,
				ReservedCidrs: req.Underlay.ReservedCidrs,
			}
		}(),
	}

	result, err := clustermgr.CreateCloudVPC(ctx, cmReq)
	if err != nil {
		return false, err
	}

	return result, nil
}

// UpdateCloudVPC update cloud vpc
func (a *Action) UpdateCloudVPC(ctx context.Context, req *types.UpdateCloudVPCReq) (bool, error) {
	result, err := clustermgr.UpdateCloudVPC(ctx, &clustermanager.UpdateCloudVPCRequest{
		CloudID:     req.CloudID,
		NetworkType: req.NetworkType,
		Region:      req.Region,
		RegionName:  req.RegionName,
		VpcName:     req.VpcName,
		VpcID:       req.VpcID,
		Available:   req.Available,
		Updater:     req.Updater,
		ReservedIPNum: func() *wrappers.UInt32Value {
			if req.ReservedIPNum != nil {
				return nil
			}
			return &wrappers.UInt32Value{Value: *req.ReservedIPNum}
		}(),
		BusinessID: func() *wrappers.StringValue {
			if req.BusinessID != nil {
				return nil
			}
			return &wrappers.StringValue{Value: *req.BusinessID}
		}(),
		Overlay: func() *clustermanager.Cidr {
			if req.Overlay != nil {
				return nil
			}
			cidrs := make([]*clustermanager.CidrState, 0)
			for _, cidr := range req.Overlay.Cidrs {
				cidrs = append(cidrs, &clustermanager.CidrState{
					Cidr:  cidr.Cidr,
					Block: cidr.Block,
				})
			}
			return &clustermanager.Cidr{
				Cidrs:         cidrs,
				ReservedIPNum: req.Overlay.ReservedIPNum,
				ReservedCidrs: req.Overlay.ReservedCidrs,
			}
		}(),
		Underlay: func() *clustermanager.Cidr {
			if req.Underlay != nil {
				return nil
			}
			cidrs := make([]*clustermanager.CidrState, 0)
			for _, cidr := range req.Underlay.Cidrs {
				cidrs = append(cidrs, &clustermanager.CidrState{
					Cidr:  cidr.Cidr,
					Block: cidr.Block,
				})
			}
			return &clustermanager.Cidr{
				Cidrs:         cidrs,
				ReservedIPNum: req.Underlay.ReservedIPNum,
				ReservedCidrs: req.Underlay.ReservedCidrs,
			}
		}(),
	})
	if err != nil {
		return false, err
	}

	return result, nil
}

// GetCloudVPCRecommendCIDR get cloud vpc recommend cidr
func (a *Action) GetCloudVPCRecommendCIDR(ctx context.Context, req *types.GetCloudVPCRecommendCIDRReq) (*types.GetCloudVPCRecommendCIDRResp, error) {
	/*cidrs, err := clustermgr.ListRecommendCloudVpcCidr(ctx, &clustermanager.ListRecommendCloudVpcCidrRequest{
		CloudID:     req.CloudID,
		Region:      req.Region,
		AccountID:   req.AccountID,
		VpcID:       req.VpcID,
		NetworkType: req.NetworkType,
		Mask:        req.Mask,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetCloudVPCRecommendCIDRResp{
		Cidrs: cidrs.Cidrs,
	}, nil*/
	return nil, nil
}
