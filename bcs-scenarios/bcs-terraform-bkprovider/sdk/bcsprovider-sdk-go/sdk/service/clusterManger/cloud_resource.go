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

// Package clusterManger cluster-service
package clusterManger

import (
	"bytes"
	"context"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	"github.com/spf13/cast"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/header"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

/*
	云上资源查询(创建集群辅助接口)
*/

// ListCloudOsImage 查询Node操作系统镜像列表.
// CloudID 云区域ID
// ***不能为空***
// 推荐结合使用：cloudID + region + accountID + provider(如果没有特殊需求,一般默认为共镜像PublicImage)
func (h *handler) ListCloudOsImage(ctx context.Context, req *pb.ListCloudOsImageRequest) (*pb.ListCloudOsImageResponse,
	error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	// 云区域
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	// 提供者
	switch req.Provider {
	case PublicImage, PrivateImage, SharedImage, MarketImage, All:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}

	requestID := header.GenUUID()
	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.ListCloudOsImageRequest marhsal failed.") // set pb.*
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.listCloudOsImageApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "list cloud os image failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.ListCloudOsImageResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// listCloudOsImageApi get
func (h *handler) listCloudOsImageApi(req *pb.ListCloudOsImageRequest) string {
	//cloudID
	base := fmt.Sprintf(h.backendApi[listCloudOsImageApi], req.CloudID)
	return fmt.Sprintf("%s?region=%s&accountID=%s&provider=%s&projectID=%s", base, req.Region, req.AccountID,
		req.Provider, req.ProjectID)
}

// ListCloudInstanceTypes 查询Node机型.
// CloudID 云区域ID
// ***不能为空***
// 推荐结合使用：cloudID + region + accountID
func (h *handler) ListCloudInstanceTypes(ctx context.Context, req *pb.ListCloudInstanceTypeRequest,
) (*pb.ListCloudInstanceTypeResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	requestID := header.GenUUID()

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.ListCloudInstanceTypeRequest marhsal failed.")
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.listCloudInstanceTypeApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "lisst cloud instance type failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.ListCloudInstanceTypeResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// listCloudInstanceTypeApi get
func (h *handler) listCloudInstanceTypeApi(req *pb.ListCloudInstanceTypeRequest) string {
	// cloudID
	base := fmt.Sprintf(h.backendApi[listCloudInstanceTypeApi], req.CloudID)
	return fmt.Sprintf("%s?region=%s&accountID=%s&zone=%s&nodeFamily=%s&cpu=%d&memory=%d&bizID=%s&provider=%s&"+
		"resourceType=%s", base, req.Region, req.AccountID, req.Zone, req.NodeFamily, req.Cpu, req.Memory, req.BizID,
		req.Provider, req.ResourceType)
}

// ListCloudSecurityGroups 查询安全组列表.
// CloudID 云区域ID
// ***不能为空***
// 推荐结合使用：cloudID + region + accountID
func (h *handler) ListCloudSecurityGroups(ctx context.Context, req *pb.ListCloudSecurityGroupsRequest,
) (*pb.ListCloudSecurityGroupsResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	requestID := header.GenUUID()

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.ListCloudSecurityGroupsRequest marhsal failed.") // set pb.*
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.listCloudSecurityGroupsApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "list cloud security groups failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.ListCloudSecurityGroupsResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// listCloudSecurityGroupsApi get
func (h *handler) listCloudSecurityGroupsApi(req *pb.ListCloudSecurityGroupsRequest) string {
	//cloudID
	base := fmt.Sprintf(h.backendApi[listCloudSecurityGroupsApi], req.CloudID)
	return fmt.Sprintf("%s?region=%s&accountID=%s&resourceGroupName=%s", base, req.Region, req.AccountID,
		req.ResourceGroupName)
}

// GetCloudRegions 查询cloud地域列表.
// CloudID 云区域ID
// ***不能为空***
// 推荐结合使用：cloudID + accountID
func (h *handler) GetCloudRegions(ctx context.Context, req *pb.GetCloudRegionsRequest,
) (*pb.GetCloudRegionsResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	requestID := header.GenUUID()

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.GetCloudRegionsRequest marhsal failed.") // set pb.*
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.getCloudRegionsApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "get cloud region failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.GetCloudRegionsResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// getCloudRegionsApi get
func (h *handler) getCloudRegionsApi(req *pb.GetCloudRegionsRequest) string {
	//cloudID
	base := fmt.Sprintf(h.backendApi[getCloudRegionsApi], req.CloudID)
	return fmt.Sprintf("%s?accountID=%s", base, req.AccountID)
}

// GetCloudRegionZones 查询cloud地域可用区列表.
// CloudID 云区域ID
// ***不能为空***
// 推荐结合使用：cloudID + region + accountID
func (h *handler) GetCloudRegionZones(ctx context.Context, req *pb.GetCloudRegionZonesRequest,
) (*pb.GetCloudRegionZonesResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	requestID := header.GenUUID()

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.GetCloudRegionZonesRequest marhsal failed.") // set pb.*
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.getCloudRegionZonesApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "get zones in region failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.GetCloudRegionZonesResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// getCloudRegionZonesApi get
func (h *handler) getCloudRegionZonesApi(req *pb.GetCloudRegionZonesRequest) string {
	//cloudID
	base := fmt.Sprintf(h.backendApi[getCloudRegionZonesApi], req.CloudID)
	return fmt.Sprintf("%s?region=%s&accountID=%s&vpcId=%s&state=%s", base, req.Region, req.AccountID, req.VpcId,
		req.State)
}

// ListCloudVpcs 获取云所属地域vpc列表.
// CloudID 云区域ID
// ***不能为空***
// 推荐结合使用：cloudID + region + accountID
func (h *handler) ListCloudVpcs(ctx context.Context, req *pb.ListCloudVpcsRequest) (*pb.ListCloudVpcsResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	requestID := header.GenUUID()

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.ListCloudVpcsRequest marhsal failed.") // set pb.*
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.listCloudVpcsApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "list vpcs failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.ListCloudVpcsResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// listCloudVpcsApi get
func (h *handler) listCloudVpcsApi(req *pb.ListCloudVpcsRequest) string {
	//cloudID
	base := fmt.Sprintf(h.backendApi[listCloudVpcsApi], req.CloudID)
	return fmt.Sprintf("%s?region=%s&accountID=%s&vpcID=%s&resourceGroupName=%s", base, req.Region, req.AccountID,
		req.VpcID, req.ResourceGroupName)
}

// ListCloudSubnets 查询vpc子网列表.
// CloudID 云区域ID
// ***不能为空***
// 依赖可用区，如任意可用区则设置
// 推荐结合使用：cloudID + region + accountID + vpcID
func (h *handler) ListCloudSubnets(ctx context.Context, req *pb.ListCloudSubnetsRequest) (*pb.ListCloudSubnetsResponse,
	error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	requestID := header.GenUUID()

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.ListCloudSubnetsRequest marhsal failed.") // set pb.*
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.listCloudSubnetsApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "list subnets failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.ListCloudSubnetsResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// listCloudSubnetsApi get
func (h *handler) listCloudSubnetsApi(req *pb.ListCloudSubnetsRequest) string {
	// cloudID
	base := fmt.Sprintf(h.backendApi[listCloudSubnetsApi], req.CloudID)
	return fmt.Sprintf("%s?region=%s&accountID=%s&vpcID=%s&zone=%s&subnetID=%s&injectCluster=%+v&resourceGroupName=%s",
		base, req.Region, req.AccountID, req.VpcID, req.Zone, req.SubnetID, req.InjectCluster, req.ResourceGroupName)
}

// ListKeypairs 查询密钥对列表.
// CloudID 云区域ID
// ***不能为空***
// 依赖可用区，如任意可用区则设置
// 推荐结合使用：cloudID + region + accountID
func (h *handler) ListKeypairs(ctx context.Context, req *pb.ListKeyPairsRequest) (*pb.ListKeyPairsResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	requestID := header.GenUUID()

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.ListKeyPairsRequest marhsal failed.") // set pb.*
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.listKeypairsApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "list key pairs failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.ListKeyPairsResponse) // set resp
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// listKeypairsApi get
func (h *handler) listKeypairsApi(req *pb.ListKeyPairsRequest) string {
	// cloudID
	base := fmt.Sprintf(h.backendApi[listKeypairsApi], req.CloudID)

	return fmt.Sprintf("%s?region=%s&accountID=%s&resourceGroupName=%s", base, req.Region, req.AccountID,
		req.ResourceGroupName)
}

// GetCloudAccountType 查询云账号类型.
// CloudID 云区域ID
// ***不能为空***
// 依赖可用区，如任意可用区则设置
// 推荐结合使用：cloudID + region + accountID
func (h *handler) GetCloudAccountType(ctx context.Context, req *pb.GetCloudAccountTypeRequest,
) (*pb.GetCloudAccountTypeResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	requestID := header.GenUUID()

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.GetCloudAccountTypeRequest marhsal failed.") // set pb.*
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.getCloudAccountTypeApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "get cloud account type failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.GetCloudAccountTypeResponse) // set resp
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// getCloudAccountTypeApi get
func (h *handler) getCloudAccountTypeApi(req *pb.GetCloudAccountTypeRequest) string {
	// cloudID
	base := fmt.Sprintf(h.backendApi[getCloudAccountTypeApi], req.CloudID)

	return fmt.Sprintf("%s?region=%s&accountID=%s", base, req.Region, req.AccountID)
}

// GetCloudBandwidthPackages 查询云共享带宽包.
// CloudID 云区域ID
// ***不能为空***
// 依赖可用区，如任意可用区则设置
// 推荐结合使用：cloudID + region + accountID
func (h *handler) GetCloudBandwidthPackages(ctx context.Context, req *pb.GetCloudBandwidthPackagesRequest) (
	*pb.GetCloudBandwidthPackagesResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	requestID := header.GenUUID()

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.GetCloudBandwidthPackagesRequest marhsal failed.") // set pb.*
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.getCloudBandwidthPackagesApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "get bgp packages failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.GetCloudBandwidthPackagesResponse) // set resp
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// getCloudBandwidthPackagesApi get
func (h *handler) getCloudBandwidthPackagesApi(req *pb.GetCloudBandwidthPackagesRequest) string {
	// cloudID
	base := fmt.Sprintf(h.backendApi[getCloudBandwidthPackagesApi], req.CloudID)
	return fmt.Sprintf("%s?region=%s&accountID=%s", base, req.Region, req.AccountID)
}

// ListCloudProjects 获取云项目列表.
// CloudID 云区域ID
// ***不能为空***
// 依赖可用区，如任意可用区则设置
// 推荐结合使用：cloudID + region + accountID
func (h *handler) ListCloudProjects(ctx context.Context, req *pb.ListCloudProjectsRequest,
) (*pb.ListCloudProjectsResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	switch req.CloudID {
	case GcpCloud, AzureCloud, TencentCloud:
	default:
		return nil, errors.Errorf("cloudID connot be empty or incorrect type.")
	}
	requestID := header.GenUUID()

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.ListCloudProjectsRequest marhsal failed.") // set pb.*
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.listCloudProjectsApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "list cloud projects failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.ListCloudProjectsResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// listCloudProjectsApi get
func (h *handler) listCloudProjectsApi(req *pb.ListCloudProjectsRequest) string {
	//cloudID
	base := fmt.Sprintf(h.backendApi[listCloudProjectsApi], req.CloudID)
	return fmt.Sprintf("%s?region=%s&accountID=%s", base, req.Region, req.AccountID)
}
