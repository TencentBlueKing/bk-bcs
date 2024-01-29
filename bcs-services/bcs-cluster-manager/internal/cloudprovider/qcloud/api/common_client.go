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

package api

import (
	"context"
	"errors"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tchttp "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)

// APIVersion xxx
const APIVersion = "2018-05-25"

// Client tke common client
type Client struct {
	common.Client
}

// NewClient init client
func NewClient(credential common.CredentialIface, region string,
	clientProfile *profile.ClientProfile) (client *Client, err error) {
	client = &Client{}
	client.Init(region).
		WithCredential(credential).
		WithProfile(clientProfile)
	return
}

// NewEnableExternalNodeSupportRequest xxx
func NewEnableExternalNodeSupportRequest() (request *EnableExternalNodeSupportRequest) {
	request = &EnableExternalNodeSupportRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "EnableExternalNodeSupport")

	return
}

// NewEnableExternalNodeSupportResponse xxx
func NewEnableExternalNodeSupportResponse() (response *EnableExternalNodeSupportResponse) {
	response = &EnableExternalNodeSupportResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// EnableExternalNodeSupport 开启第三方节点池支持
func (c *Client) EnableExternalNodeSupport(request *EnableExternalNodeSupportRequest) (
	response *EnableExternalNodeSupportResponse, err error) {
	return c.EnableExternalNodeSupportWithContext(context.Background(), request)
}

// EnableExternalNodeSupportWithContext 开启第三方节点池支持
func (c *Client) EnableExternalNodeSupportWithContext(
	ctx context.Context, request *EnableExternalNodeSupportRequest) (
	response *EnableExternalNodeSupportResponse, err error) {
	if request == nil {
		request = NewEnableExternalNodeSupportRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("EnableExternalNodeSupport require credential")
	}

	request.SetContext(ctx)

	response = NewEnableExternalNodeSupportResponse()
	err = c.Send(request, response)
	return
}

// NewDescribeExternalNodeScriptRequest request
func NewDescribeExternalNodeScriptRequest() (request *DescribeExternalNodeScriptRequest) {
	request = &DescribeExternalNodeScriptRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeExternalNodeScript")

	return
}

// NewDescribeExternalNodeScriptResponse response
func NewDescribeExternalNodeScriptResponse() (response *DescribeExternalNodeScriptResponse) {
	response = &DescribeExternalNodeScriptResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeExternalNodeScript 获取第三方节点添加脚本
func (c *Client) DescribeExternalNodeScript(request *DescribeExternalNodeScriptRequest) (
	response *DescribeExternalNodeScriptResponse, err error) {
	return c.DescribeExternalNodeScriptWithContext(context.Background(), request)
}

// DescribeExternalNodeScriptWithContext 获取第三方节点添加脚本
func (c *Client) DescribeExternalNodeScriptWithContext(ctx context.Context,
	request *DescribeExternalNodeScriptRequest) (response *DescribeExternalNodeScriptResponse, err error) {
	if request == nil {
		request = NewDescribeExternalNodeScriptRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeExternalNodeScript require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeExternalNodeScriptResponse()
	err = c.Send(request, response)
	return
}

// NewDeleteExternalNodeRequest request
func NewDeleteExternalNodeRequest() (request *DeleteExternalNodeRequest) {
	request = &DeleteExternalNodeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeleteExternalNode")

	return
}

// NewDeleteExternalNodeResponse response
func NewDeleteExternalNodeResponse() (response *DeleteExternalNodeResponse) {
	response = &DeleteExternalNodeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteExternalNode 删除第三方节点
func (c *Client) DeleteExternalNode(request *DeleteExternalNodeRequest) (
	response *DeleteExternalNodeResponse, err error) {
	return c.DeleteExternalNodeWithContext(context.Background(), request)
}

// DeleteExternalNodeWithContext 删除第三方节点
func (c *Client) DeleteExternalNodeWithContext(ctx context.Context, request *DeleteExternalNodeRequest) (
	response *DeleteExternalNodeResponse, err error) {
	if request == nil {
		request = NewDeleteExternalNodeRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DeleteExternalNode require credential")
	}

	request.SetContext(ctx)

	response = NewDeleteExternalNodeResponse()
	err = c.Send(request, response)
	return
}

// NewDrainExternalNodeRequest request
func NewDrainExternalNodeRequest() (request *DrainExternalNodeRequest) {
	request = &DrainExternalNodeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DrainExternalNode")

	return
}

// NewDrainExternalNodeResponse response
func NewDrainExternalNodeResponse() (response *DrainExternalNodeResponse) {
	response = &DrainExternalNodeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DrainExternalNode 驱逐第三方节点
func (c *Client) DrainExternalNode(request *DrainExternalNodeRequest) (response *DrainExternalNodeResponse, err error) {
	return c.DrainExternalNodeWithContext(context.Background(), request)
}

// DrainExternalNodeWithContext 驱逐第三方节点
func (c *Client) DrainExternalNodeWithContext(ctx context.Context, request *DrainExternalNodeRequest) (
	response *DrainExternalNodeResponse, err error) {
	if request == nil {
		request = NewDrainExternalNodeRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DrainExternalNode require credential")
	}

	request.SetContext(ctx)

	response = NewDrainExternalNodeResponse()
	err = c.Send(request, response)
	return
}

// NewDeleteExternalNodePoolRequest request
func NewDeleteExternalNodePoolRequest() (request *DeleteExternalNodePoolRequest) {
	request = &DeleteExternalNodePoolRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DeleteExternalNodePool")

	return
}

// NewDeleteExternalNodePoolResponse response
func NewDeleteExternalNodePoolResponse() (response *DeleteExternalNodePoolResponse) {
	response = &DeleteExternalNodePoolResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DeleteExternalNodePool 删除第三方节点池
func (c *Client) DeleteExternalNodePool(request *DeleteExternalNodePoolRequest) (
	response *DeleteExternalNodePoolResponse, err error) {
	return c.DeleteExternalNodePoolWithContext(context.Background(), request)
}

// DeleteExternalNodePoolWithContext 删除第三方节点池
func (c *Client) DeleteExternalNodePoolWithContext(ctx context.Context, request *DeleteExternalNodePoolRequest) (
	response *DeleteExternalNodePoolResponse, err error) {
	if request == nil {
		request = NewDeleteExternalNodePoolRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DeleteExternalNodePool require credential")
	}

	request.SetContext(ctx)

	response = NewDeleteExternalNodePoolResponse()
	err = c.Send(request, response)
	return
}

// NewDescribeExternalNodePoolsRequest request
func NewDescribeExternalNodePoolsRequest() (request *DescribeExternalNodePoolsRequest) {
	request = &DescribeExternalNodePoolsRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeExternalNodePools")

	return
}

// NewDescribeExternalNodePoolsResponse response
func NewDescribeExternalNodePoolsResponse() (response *DescribeExternalNodePoolsResponse) {
	response = &DescribeExternalNodePoolsResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeExternalNodePools 查看第三方节点池列表
func (c *Client) DescribeExternalNodePools(request *DescribeExternalNodePoolsRequest) (
	response *DescribeExternalNodePoolsResponse, err error) {
	return c.DescribeExternalNodePoolsWithContext(context.Background(), request)
}

// DescribeExternalNodePoolsWithContext 查看第三方节点池列表
func (c *Client) DescribeExternalNodePoolsWithContext(ctx context.Context, request *DescribeExternalNodePoolsRequest) (
	response *DescribeExternalNodePoolsResponse, err error) {
	if request == nil {
		request = NewDescribeExternalNodePoolsRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeExternalNodePools require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeExternalNodePoolsResponse()
	err = c.Send(request, response)
	return
}

// NewDescribeExternalNodeSupportConfigRequest request
func NewDescribeExternalNodeSupportConfigRequest() (request *DescribeExternalNodeSupportConfigRequest) {
	request = &DescribeExternalNodeSupportConfigRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeExternalNodeSupportConfig")

	return
}

// NewDescribeExternalNodeSupportConfigResponse response
func NewDescribeExternalNodeSupportConfigResponse() (response *DescribeExternalNodeSupportConfigResponse) {
	response = &DescribeExternalNodeSupportConfigResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeExternalNodeSupportConfig 查看开启第三方节点池配置信息
func (c *Client) DescribeExternalNodeSupportConfig(request *DescribeExternalNodeSupportConfigRequest) (
	response *DescribeExternalNodeSupportConfigResponse, err error) {
	return c.DescribeExternalNodeSupportConfigWithContext(context.Background(), request)
}

// DescribeExternalNodeSupportConfigWithContext 查看开启第三方节点池配置信息
func (c *Client) DescribeExternalNodeSupportConfigWithContext(ctx context.Context,
	request *DescribeExternalNodeSupportConfigRequest) (response *DescribeExternalNodeSupportConfigResponse,
	err error) {
	if request == nil {
		request = NewDescribeExternalNodeSupportConfigRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeExternalNodeSupportConfig require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeExternalNodeSupportConfigResponse()
	err = c.Send(request, response)
	return
}

// NewCreateExternalNodePoolRequest request
func NewCreateExternalNodePoolRequest() (request *CreateExternalNodePoolRequest) {
	request = &CreateExternalNodePoolRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "CreateExternalNodePool")

	return
}

// NewCreateExternalNodePoolResponse response
func NewCreateExternalNodePoolResponse() (response *CreateExternalNodePoolResponse) {
	response = &CreateExternalNodePoolResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// CreateExternalNodePool 创建第三方节点池
func (c *Client) CreateExternalNodePool(request *CreateExternalNodePoolRequest) (
	response *CreateExternalNodePoolResponse, err error) {
	return c.CreateExternalNodePoolWithContext(context.Background(), request)
}

// CreateExternalNodePoolWithContext 创建第三方节点池
func (c *Client) CreateExternalNodePoolWithContext(ctx context.Context, request *CreateExternalNodePoolRequest) (
	response *CreateExternalNodePoolResponse, err error) {
	if request == nil {
		request = NewCreateExternalNodePoolRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("CreateExternalNodePool require credential")
	}

	request.SetContext(ctx)

	response = NewCreateExternalNodePoolResponse()
	err = c.Send(request, response)
	return
}

// NewModifyExternalNodePoolRequest request
func NewModifyExternalNodePoolRequest() (request *ModifyExternalNodePoolRequest) {
	request = &ModifyExternalNodePoolRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "ModifyExternalNodePool")

	return
}

// NewModifyExternalNodePoolResponse response
func NewModifyExternalNodePoolResponse() (response *ModifyExternalNodePoolResponse) {
	response = &ModifyExternalNodePoolResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// ModifyExternalNodePool 修改第三方节点池
func (c *Client) ModifyExternalNodePool(request *ModifyExternalNodePoolRequest) (
	response *ModifyExternalNodePoolResponse, err error) {
	return c.ModifyExternalNodePoolWithContext(context.Background(), request)
}

// ModifyExternalNodePoolWithContext 修改第三方节点池
func (c *Client) ModifyExternalNodePoolWithContext(ctx context.Context, request *ModifyExternalNodePoolRequest) (
	response *ModifyExternalNodePoolResponse, err error) {
	if request == nil {
		request = NewModifyExternalNodePoolRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("ModifyExternalNodePool require credential")
	}

	request.SetContext(ctx)

	response = NewModifyExternalNodePoolResponse()
	err = c.Send(request, response)
	return
}

// NewDescribeExternalNodeRequest request
func NewDescribeExternalNodeRequest() (request *DescribeExternalNodeRequest) {
	request = &DescribeExternalNodeRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeExternalNode")

	return
}

// NewDescribeExternalNodeResponse response
func NewDescribeExternalNodeResponse() (response *DescribeExternalNodeResponse) {
	response = &DescribeExternalNodeResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeExternalNode 查看第三方节点列表
func (c *Client) DescribeExternalNode(request *DescribeExternalNodeRequest) (
	response *DescribeExternalNodeResponse, err error) {
	return c.DescribeExternalNodeWithContext(context.Background(), request)
}

// DescribeExternalNodeWithContext 查看第三方节点列表
func (c *Client) DescribeExternalNodeWithContext(ctx context.Context, request *DescribeExternalNodeRequest) (
	response *DescribeExternalNodeResponse, err error) {
	if request == nil {
		request = NewDescribeExternalNodeRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeExternalNode require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeExternalNodeResponse()
	err = c.Send(request, response)
	return
}

// DescribeOSImages 获取OS聚合信息
func (c *Client) DescribeOSImages(request *DescribeOSImagesRequest) (response *DescribeOSImagesResponse, err error) {
	return c.DescribeOSImagesWithContext(context.Background(), request)
}

// DescribeOSImagesWithContext 获取OS聚合信息
func (c *Client) DescribeOSImagesWithContext(ctx context.Context, request *DescribeOSImagesRequest) (
	response *DescribeOSImagesResponse, err error) {
	if request == nil {
		request = NewDescribeOSImagesRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeOSImages require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeOSImagesResponse()
	err = c.Send(request, response)
	return
}

// NewDescribeInstanceCreateProgressRequest add node progress
func NewDescribeInstanceCreateProgressRequest() (request *DescribeInstanceCreateProgressRequest) {
	request = &DescribeInstanceCreateProgressRequest{
		BaseRequest: &tchttp.BaseRequest{},
	}
	request.Init().WithApiInfo("tke", APIVersion, "DescribeInstanceCreateProgress")

	return
}

// NewDescribeInstanceCreateProgressResponse add node progress
func NewDescribeInstanceCreateProgressResponse() (response *DescribeInstanceCreateProgressResponse) {
	response = &DescribeInstanceCreateProgressResponse{
		BaseResponse: &tchttp.BaseResponse{},
	}
	return
}

// DescribeInstanceCreateProgress 获取节点创建进度
func (c *Client) DescribeInstanceCreateProgress(
	request *DescribeInstanceCreateProgressRequest) (response *DescribeInstanceCreateProgressResponse, err error) {
	return c.DescribeInstanceCreateProgressWithContext(context.Background(), request)
}

// DescribeInstanceCreateProgressWithContext  获取节点创建进度
func (c *Client) DescribeInstanceCreateProgressWithContext(ctx context.Context,
	request *DescribeInstanceCreateProgressRequest) (response *DescribeInstanceCreateProgressResponse, err error) {
	if request == nil {
		request = NewDescribeInstanceCreateProgressRequest()
	}

	if c.GetCredential() == nil {
		return nil, errors.New("DescribeInstanceCreateProgress require credential")
	}

	request.SetContext(ctx)

	response = NewDescribeInstanceCreateProgressResponse()
	err = c.Send(request, response)
	return
}
