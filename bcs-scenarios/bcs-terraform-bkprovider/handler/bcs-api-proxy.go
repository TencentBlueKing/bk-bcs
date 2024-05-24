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

// Package handler xxx
package handler

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tvpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/middleware/xbknodeman"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/middleware/xtencentcloud"
	pb "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/proto"
)

// BcsApiHandler api handler
type BcsApiHandler struct {
	adminUsers []string

	bkAppCode   string
	bkAppSecret string
	bkEnv       string

	bkOuterIP        string
	bkAddrTemplateID string

	tVpcCli *xtencentcloud.VpcClient
}

// NewBcsApiHandler return new instance
func NewBcsApiHandler(opt *common.Options) *BcsApiHandler {
	h := &BcsApiHandler{
		adminUsers: opt.AdminUsers,

		bkAppCode:        opt.BkSystem.BkAppCode,
		bkAppSecret:      opt.BkSystem.BkAppSecret,
		bkEnv:            opt.BkSystem.BkEnv,
		bkAddrTemplateID: opt.BkSystem.BkAddressTemplateID,
		bkOuterIP:        opt.BkSystem.BkOuterIP,
	}

	tVpcCli, err := xtencentcloud.NewClient(opt.TencentCloud.VpcDomain, opt.TencentCloud.SecretID,
		opt.TencentCloud.SecretKey, opt.TencentCloud.Region)
	if err != nil {
		panic(err)
	}
	h.tVpcCli = tVpcCli
	return h
}

// InstallJob create job
func (b *BcsApiHandler) InstallJob(ctx context.Context, request *pb.InstallJobRequest,
	response *pb.InstallJobResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	installReq := &xbknodeman.InstallJobRequest{}
	for _, host := range request.Hosts {
		installReq.Hosts = append(installReq.Hosts, &xbknodeman.InstallHost{
			BkCloudId: host.BkCloudId,
			BkBizId:   host.BkBizId,
			BkHostID:  host.BkHostId,
			OsType:    host.OsType,
			InnerIp:   host.InnerIp,
			OuterIp:   host.OuterIp,
			LoginIp:   host.LoginIp,
			Account:   host.Account,
			Port:      host.Port,
			AuthType:  host.AuthType,
			Password:  host.Password,
			ApId:      host.ApId,
			Key:       host.Key,
		})
	}
	installReq.JobType = request.JobType
	resp, err := nodeManCli.InstallJob(ctx, installReq)
	if err != nil {
		blog.Errorf("install job failed,req:%s, err: %s", common.JsonMarshal(request), err.Error())
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}
	response.Code = common.CodeSuccess
	response.Message = resp.Message
	response.Data = &pb.InstallJobData{
		JobId: resp.Data.JobId,
	}
	return nil
}

// CreateCloud create cloud
func (b *BcsApiHandler) CreateCloud(ctx context.Context, request *pb.CloudCreateRequest,
	response *pb.CloudCreateResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	resp, err := nodeManCli.CreateCloud(ctx, &xbknodeman.CreateCloudRequest{
		BkCloudName: request.BkCloudName,
		Isp:         request.Isp,
		ApID:        int64(request.ApId),
	})
	if err != nil {
		blog.Errorf("CreateCloud failed,req:%s, err: %s", common.JsonMarshal(request), err.Error())
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	response.Code = common.CodeSuccess
	response.Message = resp.Message
	response.Data = int32(resp.Data.BkCloudID)
	return nil
}

// UpdateCloud update cloud
func (b *BcsApiHandler) UpdateCloud(ctx context.Context, request *pb.CloudUpdateRequest,
	response *pb.CloudUpdateResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	resp, err := nodeManCli.UpdateCloud(ctx, &xbknodeman.UpdateCloudRequest{
		BkCloudID:   int64(request.BkCloudId),
		BkCloudName: request.BkCloudName,
		Isp:         request.Isp,
		ApID:        int64(request.ApId),
	})
	if err != nil {
		blog.Errorf("UpdateCloud failed,req:%s, err: %s", common.JsonMarshal(request), err.Error())
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	response.Code = common.CodeSuccess
	response.Message = resp.Message
	return nil
}

// ListCloud list all clouds
func (b *BcsApiHandler) ListCloud(ctx context.Context, request *pb.CloudListRequest,
	response *pb.CloudListResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	resp, err := nodeManCli.ListCloud(ctx, &xbknodeman.ListCloudRequest{})
	if err != nil {
		blog.Errorf("ListCloud failed,req:%s, err: %s", common.JsonMarshal(request), err.Error())
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	response.Code = common.CodeSuccess
	response.Message = resp.Message
	for _, cloud := range resp.Data {
		response.Data = append(response.Data, &pb.Cloud{
			BkCloudId:   int32(cloud.BkCloudId),
			BkCloudName: cloud.BkCloudName,
			Isp:         cloud.Isp,
			ApId:        int32(cloud.ApId),
		})
	}
	return nil
}

// DeleteCloud delete a cloud
func (b *BcsApiHandler) DeleteCloud(ctx context.Context, request *pb.CloudDeleteRequest,
	response *pb.CloudDeleteResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	resp, err := nodeManCli.DeleteCloud(ctx, &xbknodeman.DeleteCloudRequest{
		BkCloudID: int64(request.BkCloudId),
	})
	if err != nil {
		blog.Errorf("DeleteCloud failed,req:%s, err: %s", common.JsonMarshal(request), err.Error())
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	response.Code = common.CodeSuccess
	response.Message = resp.Message
	return nil
}

// ListHost list hosts
func (b *BcsApiHandler) ListHost(ctx context.Context, request *pb.ListHostRequest,
	response *pb.ListHostResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	listReq := &xbknodeman.ListHostRequest{
		Page:     int64(request.Page),
		PageSize: int64(request.Pagesize),
	}

	for _, cond := range request.Conditions {
		listReq.Conditions = append(listReq.Conditions, xbknodeman.Condition{
			Key:   cond.Key,
			Value: cond.Value,
		})
	}
	for _, bkBizID := range request.BkBizIds {
		listReq.BkBizId = append(listReq.BkBizId, int64(bkBizID))
	}

	resp, err := nodeManCli.ListHosts(ctx, listReq)
	if err != nil {
		blog.Errorf("ListHosts failed,req:%s, err: %s", common.JsonMarshal(request), err.Error())
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	err = common.JsonConvert(resp.Data, &response.Data)
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}
	response.Code = common.CodeSuccess
	response.Message = resp.Message
	return nil
}

// ListProxyHost list proxy host
func (b *BcsApiHandler) ListProxyHost(ctx context.Context, request *pb.ListProxyHostRequest,
	response *pb.ListProxyHostResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	listReq := &xbknodeman.GetProxyHostRequest{
		BkCloudId: request.BkCloudId,
	}

	resp, err := nodeManCli.GetProxyHost(ctx, listReq)
	if err != nil {
		blog.Errorf("GetProxyHost failed,req:%s, err: %s", common.JsonMarshal(request), err.Error())
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	err = common.JsonConvert(resp.Data, &response.Data)
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}
	response.Code = common.CodeSuccess
	response.Message = resp.Message
	return nil
}

// GetJobDetail get job detail
func (b *BcsApiHandler) GetJobDetail(ctx context.Context, request *pb.GetJobDetailRequest,
	response *pb.GetJobDetailResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	nodeManCli := b.newBkNodeManCli(user.GetUsername())
	jobDetailReq := &xbknodeman.GetJobDetailRequest{
		JobID:    request.JobId,
		Page:     int64(request.Page),
		PageSize: int64(request.Pagesize),
	}

	for _, cond := range request.Conditions {
		jobDetailReq.Conditions = append(jobDetailReq.Conditions, xbknodeman.Condition{
			Key:   cond.Key,
			Value: cond.Value,
		})
	}
	jobDetail, err := nodeManCli.GetJobDetails(ctx, jobDetailReq)
	if err != nil {
		blog.Errorf("GetJobDetails failed,req:%s, err: %s", common.JsonMarshal(request), err.Error())
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	err = common.JsonConvert(jobDetail.Data, &response.Data)
	if err != nil {
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}
	response.Code = common.CodeSuccess
	response.Message = jobDetail.Message
	return nil
}

// RegisterBkWhitelist register whitelist
func (b *BcsApiHandler) RegisterBkWhitelist(ctx context.Context, request *pb.RegisterBkWhitelistRequest, response *pb.RegisterBkWhitelistResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}
	blog.Infof("registerBkWhiteList, user: %s, biz: %s, ip_list: %s", user.GetUsername(), request.GetBizName(),
		common.ToJsonString(request.GetIpList()))
	resp, err := b.tVpcCli.DescribeAddressTemplate(b.bkAddrTemplateID)
	if err != nil {
		blog.Errorf("DescribeAddressTemplate failed,req:%s, err: %s", common.JsonMarshal(request), err.Error())
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	if getUint64Value(resp.Response.TotalCount) == 0 {
		blog.Errorf("get addr template failed,template_id:%s", b.bkAddrTemplateID)
		response.Code = common.CodeInternalError
		response.Message = fmt.Sprintf("internal error, invali whitelist")
		return nil
	}

	existedAddr := make(map[string]interface{})
	bkAddrTemplate := resp.Response.AddressTemplateSet[0]
	for _, addr := range bkAddrTemplate.AddressSet {
		existedAddr[*addr] = struct{}{}
	}

	toAddAddr := make([]*tvpc.MemberInfo, 0)
	for _, addr := range request.GetIpList() {
		// 不变更已存在ip
		if _, ok := existedAddr[addr]; !ok {
			toAddAddr = append(toAddAddr, &tvpc.MemberInfo{
				Member:      tcommon.StringPtr(addr),
				Description: tcommon.StringPtr(fmt.Sprintf("operator: %s, biz: %s", user.Username, request.GetBizName())),
			})
		}
	}

	if err = b.tVpcCli.AddTemplateMember(b.bkAddrTemplateID, toAddAddr); err != nil {
		blog.Errorf("AddTemplateMember failed,req:%s, err: %s", common.JsonMarshal(request), err.Error())
		response.Code = common.CodeInternalError
		response.Message = fmt.Sprintf("add whitelist failed, err: %s", err.Error())
		return nil
	}

	response.Code = common.CodeSuccess
	response.Message = common.MsgSuccess
	return nil
}

// GetBkOuterIP return bk outer ip
func (b *BcsApiHandler) GetBkOuterIP(ctx context.Context, request *pb.GetBkOuterIPRequest, response *pb.GetBkOuterIPResponse) error {
	response.Code = common.CodeSuccess
	response.Message = common.MsgSuccess
	response.Data = []string{b.bkOuterIP}
	return nil
}

// ListBkWhitelist 获取蓝鲸网关白名单(仅测试用)
func (b *BcsApiHandler) ListBkWhitelist(ctx context.Context, request *pb.ListBkWhiteListRequest, response *pb.ListBkWhiteListResponse) error {
	user, code, msg := getUserInfo(ctx)
	if code != common.CodeSuccess {
		response.Code = code
		response.Message = msg
		return nil
	}

	if !stringInSlice(user.GetUsername(), b.adminUsers) {
		response.Code = common.CodeInternalError
		response.Message = fmt.Sprintf("invalid user")
		return nil
	}

	resp, err := b.tVpcCli.DescribeAddressTemplate(b.bkAddrTemplateID)
	if err != nil {
		blog.Errorf("DescribeAddressTemplate failed,req:%s, err: %s", b.bkAddrTemplateID, err.Error())
		response.Code = common.CodeInternalError
		response.Message = err.Error()
		return nil
	}

	if getUint64Value(resp.Response.TotalCount) == 0 {
		blog.Errorf("get addr template failed,template_id:%s", b.bkAddrTemplateID)
		response.Code = common.CodeInternalError
		response.Message = fmt.Sprintf("internal error, invali whitelist")
		return nil
	}

	whiteIpList := make([]string, 0)
	for _, addr := range resp.Response.AddressTemplateSet[0].AddressSet {
		whiteIpList = append(whiteIpList, *addr)
	}

	response.Code = common.CodeSuccess
	response.Message = common.MsgSuccess
	response.Data = whiteIpList
	return nil
}

func (b *BcsApiHandler) newBkNodeManCli(userName string) *xbknodeman.Client {
	return xbknodeman.NewClient(b.bkEnv, b.bkAppCode, b.bkAppSecret, "", userName)
}
