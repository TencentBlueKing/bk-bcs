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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs"
)

// ListCloudVPCReq list cloud vpc request
type ListCloudVPCReq struct {
	CloudID     string `json:"cloudID" in:"query=cloudID"`
	Region      string `json:"region" in:"query=region"`
	VpcID       string `json:"vpcID" in:"query=vpcID"`
	NetworkType string `json:"networkType" in:"query=networkType"`
	BusinessID  string `json:"businessID" in:"query=businessID"`
}

// DeleteCloudVPCReq delete cloud vpc request
type DeleteCloudVPCReq struct {
	CloudID string `json:"cloudID" in:"query=cloudID"`
	VpcID   string `json:"vpcID" in:"query=vpcID"`
}

// ListCloudVPC 获取VPC列表
// @Summary 获取VPC列表
// @Tags    Logs
// @Produce json
// @Success 200 {array} k8sclient.Container
// @Router  /cloudvpc [get]
func ListCloudVPC(c context.Context, req *ListCloudVPCReq) (*[]*bcs.CloudVPC, error) {
	vpcs, err := bcs.ListCloudVPC(req.CloudID, req.Region, req.VpcID, req.NetworkType, req.BusinessID)
	if err != nil {
		return nil, err
	}

	return &vpcs, nil
}

// CreateCloudVPC 创建VPC
// @Summary 创建VPC
// @Tags    Logs
// @Produce json
// @Success 200 {array} k8sclient.Container
// @Router  /cloudvpc [post]
func CreateCloudVPC(c context.Context, req *bcs.CreateCloudVPCReq) (*bool, error) {
	result, err := bcs.CreateCloudVPC(req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateCloudVPC 更新VPC
// @Summary 更新VPC
// @Tags    Logs
// @Produce json
// @Success 200 {array} k8sclient.Container
// @Router  /cloudvpc [put]
func UpdateCloudVPC(c context.Context, req *bcs.UpdateCloudVPCReq) (*bcs.CloudVPC, error) {
	vpc, err := bcs.UpdateCloudVPC(req)
	if err != nil {
		return nil, err
	}

	return vpc, nil
}

// DeleteCloudVPC 删除VPC
// @Summary 更新VPC
// @Tags    Logs
// @Produce json
// @Success 200 {array} k8sclient.Container
// @Router  /cloudvpc [delete]
func DeleteCloudVPC(c context.Context, req *DeleteCloudVPCReq) (*bcs.CloudVPC, error) {
	result, err := bcs.DeleteCloudVPC(req.CloudID, req.VpcID)
	if err != nil {
		return nil, err
	}

	return result, nil
}
