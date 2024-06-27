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
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

/*
	节点池
*/

// CreateNodeGroup 创建节点池.
// req参数说明, 如下:
// Name不能为空;
// ClusterID不能为空;
// LaunchTemplate.InstanceType不能为空;
// LaunchTemplate.KeyPair和LaunchTemplate.InitLoginPassword二选一, 不能同时为空;
// LaunchTemplate.SecurityGroupIDs为安全组配置, 需要用户自行确定安全组; sdk仅做为空判断;
// EnableAutoscale需要用户自行确认是否开启节点池, 默认为不开启;
//
// Creator推荐为空, 默认为当前用户;
// Provider推荐为空, 默认与集群保持一致;
// NodeGroupType推荐为空, 默认为普通节点池(normal);
// Region推荐为空, 强制与集群同地域; sdk会自动对齐地域值;
// VpcID推荐为空, 强制与集群同vpcID; sdk会自动对齐vpcID;
// NodeTemplate推荐为空, 如无特殊需求, 请保持为空即可; sdk会自动补齐默认值;
// LaunchTemplate.Cpu推荐为空, sdk自动补齐该字段;
// LaunchTemplate.Mem推荐为空, sdk自动补齐该字段;
//
// AutoScaling.Zones可以为空, 默认为该地域下所有可用区; 如对可用区有要求, 在本字段指定可用区即可;
// AutoScaling.SubnetIDs可以为空, sdk会自动补齐子网信息;
// AutoScaling.ScalingMode可以为空, 默认为"ClassicScaling"( 扩容时创建新实例,缩容时销毁实例 );
// AutoScaling.MultiZoneSubnetPolicy可以为空, 默认为"Priority"(在高优先级的子网与可用区创建实例);
// AutoScaling.RetryPolicy可以为空, 默认为"ImmediateRetry"(立即重试, 在较短时间内快速重试, 连续失败超过一定次数(5次)后不再重试);
//
// LaunchTemplate.ImageInfo可以为空, 默认为公共linux镜像(tlinux3.2x86_64);
// LaunchTemplate.SystemDisk可以为空, 默认为SSD+50GB;
// LaunchTemplate.InternetAccess可以为空, 默认无外网IP;
// LaunchTemplate.IsSecurityService可以为空, 默认开启;
// LaunchTemplate.IsMonitorService可以为空, 默认开启;
// LaunchTemplate.InstanceChargeType可以为空, 默认为POSTPAID_BY_HOUR(后付费, 即按量计费);
// LaunchTemplate.DataDisks可以为空, 默认没有数据盘(注意如果需要数据盘, 除了本字段需要填写, 另外NodeTemplate.DataDisks字段也需要填写, 并且是完整填写所有disk参数);
//
// 当前接口为异步接口, 请注意判断response的status, 通常有: CREATING、DELETING、UPDATING、RUNNING、
// DELETE-FAILURE、CREATE-FAILURE、UPDATE-FAILURE状态
func (h *handler) CreateNodeGroup(ctx context.Context, req *pb.CreateNodeGroupRequest) (*pb.CreateNodeGroupResponse,
	error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if err := h.fillCreateNodeGroupRequest(ctx, req); err != nil {
		return nil, err
	}

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.CreateNodeGroupRequest marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Post(ctx, requestID, h.createNodeGroupApi(), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "create node group failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.CreateNodeGroupResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// fillAutoScaling 补齐 AutoScalingGroup 参数
func (h *handler) fillAutoScaling(as *pb.AutoScalingGroup, subs []*pb.Subnet) {
	if as.MaxSize == 0 {
		as.MaxSize = 1 // 不能为0
	}
	// 扩容模式
	switch as.ScalingMode {
	case ClassicScaling, WakeUpStoppedScaling:
	default:
		as.ScalingMode = ClassicScaling
	}
	// 可用区子网模式
	switch as.MultiZoneSubnetPolicy {
	case Priority, Equality:
	default:
		as.MultiZoneSubnetPolicy = Priority
	}
	// 重试策略
	switch as.RetryPolicy {
	case ImmediateRetry, IncrementalIntervals, NoRetry:
	default:
		as.RetryPolicy = ImmediateRetry
	}
	// 已设置子网
	if len(as.SubnetIDs) != 0 {
		return
	}

	// 未设置子网
	currentZone := make(map[string]bool)
	for _, zone := range as.Zones {
		currentZone[zone] = true
	}

	subnetIDs := make([]string, 0)
	for _, subnet := range subs {
		if subnet.AvailableIPAddressCount == 0 {
			continue
		}

		if len(as.Zones) == 0 {
			//如果zones为空, 则表示当前节点池为任意可用区, 这种情况建议把当前vpc下所有子网都选择上。
			subnetIDs = append(subnetIDs, subnet.SubnetID)
			continue
		}

		//如果zones不为空, 有设置可用区或设置多个可用区, 这种情况建议把vpc可用区下所有子网也都选择上。
		if _, ok := currentZone[subnet.Zone]; ok {
			subnetIDs = append(subnetIDs, subnet.SubnetID)
		}
	}

	// 补齐子网信息
	as.SubnetIDs = subnetIDs
}

// fillLaunchTemplate 填充 LaunchConfiguration 参数
func (h *handler) fillLaunchTemplate(lc *pb.LaunchConfiguration, insTypes []*pb.InstanceType) {
	if lc.ImageInfo == nil {
		// note: 推荐动态查询镜像(tlinux3.2x86_64)
		lc.ImageInfo = &pb.ImageInfo{
			ImageID:   "img-9qrfy1xt",
			ImageType: PublicImage,
		}
	}
	// 填充cpu、mem信息
	for _, ins := range insTypes {
		if lc.InstanceType == ins.NodeType {
			lc.CPU = ins.Cpu
			lc.Mem = ins.Memory
			break
		}
	}
	// 填充系统盘信息
	if lc.SystemDisk == nil {
		lc.SystemDisk = &pb.DataDisk{
			DiskSize: "50",
			DiskType: CloudPremium,
		}
	}
	// 填充外网信息
	if lc.InternetAccess == nil {
		lc.InternetAccess = &pb.InternetAccessible{
			InternetChargeType:   BandwidthPackage,
			InternetMaxBandwidth: "0",
			PublicIPAssigned:     false,
			BandwidthPackageId:   "",
		}
	}
	// 付费信息
	switch lc.InstanceChargeType {
	case Prepaid, PostpaidByHour, Spotpaid:
	default:
		// 默认为预付费
		lc.InstanceType = Prepaid
	}
	// 其他配置
	lc.IsMonitorService = true
	lc.IsSecurityService = true
}

// fillNodeTemplate 填充 NodeTemplate 参数
// DockerGraphPath推荐为空, 默认为bcs统一设定;
func (h *handler) fillNodeTemplate(nt *pb.NodeTemplate) {
	// 默认参数
	nt.UnSchedulable = 1
	nt.AllowSkipScaleOutWhenFailed = false
	nt.AllowSkipScaleInWhenFailed = true
	// bcs规范, 强制统一
	nt.DockerGraphPath = dockerGraphPath
	//// 检查数据盘
	//if len(nt.DataDisks) != 0 {
	//	// 当数据盘不为空时, 默认视为参数填写正确, 不做二次校验
	//	return
	//}
	//// 没有数据盘时, 数据盘推荐购买100GB
	//disk := &pb.CloudDataDisk{
	//	DiskType:           CloudPremium,
	//	DiskSize:           "100",
	//	FileSystem:         "ext4",
	//	MountTarget:        "/data",
	//	AutoFormatAndMount: true,
	//}
	//nt.DataDisks = append(nt.DataDisks, disk)
}

// fillCreateNodeGroupRequest 字段填充与检查
func (h *handler) fillCreateNodeGroupRequest(ctx context.Context, req *pb.CreateNodeGroupRequest) error {
	if len(req.Name) == 0 {
		return errors.Errorf("cluster name connot be empty.")
	}
	if len(req.ClusterID) == 0 {
		return errors.Errorf("clusterID connot be empty.")
	}
	if req.AutoScaling == nil {
		req.AutoScaling = new(pb.AutoScalingGroup)
	}
	as := req.AutoScaling
	if req.LaunchTemplate == nil {
		return errors.Errorf("launchTemplate connot be empty.")
	}
	lt := req.LaunchTemplate
	if len(lt.InstanceType) == 0 {
		return errors.Errorf("launchTemplate.InstanceType connot be empty.")
	}
	if len(lt.SecurityGroupIDs) == 0 {
		return errors.Errorf("launchTemplate.SecurityGroupIDs connot be empty.")
	}
	if len(lt.InitLoginPassword) == 0 && lt.KeyPair == nil { // 密码和密钥不能同时为空
		return errors.Errorf("launchTemplate.InitLoginPassword and launchTemplate.KeyPair connot be empty.")
	}
	if req.NodeTemplate == nil {
		req.NodeTemplate = new(pb.NodeTemplate)
	}
	nt := req.NodeTemplate

	/*
		获取依赖数据 -- 后续封装为一键获取
	*/
	clusterResp, err := h.GetCluster(ctx, &pb.GetClusterReq{ClusterID: req.ClusterID})
	if err != nil || !clusterResp.Result || clusterResp.Code != 0 {
		return errors.Wrapf(err, "failed to obtain dependency data bcs-cluster, clusterResp: %s",
			utils.ObjToJson(clusterResp))
	}
	bcsCluster := clusterResp.Data

	subReq := &pb.ListCloudSubnetsRequest{
		VpcID:     bcsCluster.VpcID,
		Region:    bcsCluster.Region,
		CloudID:   bcsCluster.Provider,       // 如： TencentCloud
		AccountID: bcsCluster.CloudAccountID, // 云账号
	}
	subResp, err := h.ListCloudSubnets(ctx, subReq)
	if err != nil || !subResp.Result || subResp.Code != 0 {
		return errors.Wrapf(err, "failed to obtain dependency data cloud subnet, subResp: %s",
			utils.ObjToJson(subResp))
	}
	cloudSubs := subResp.Data

	insReq := &pb.ListCloudInstanceTypeRequest{
		Region:    bcsCluster.Region,
		AccountID: bcsCluster.CloudAccountID, // 云账号
		CloudID:   bcsCluster.Provider,       // 如： TencentCloud
	}
	insResp, err := h.ListCloudInstanceTypes(ctx, insReq)
	if err != nil || !subResp.Result || subResp.Code != 0 {
		return errors.Wrapf(err, "failed to obtain dependency data instance types, insResp: %s",
			utils.ObjToJson(insResp))
	}
	insTypes := insResp.Data

	// fill as
	h.fillAutoScaling(as, cloudSubs)
	// fill lt
	h.fillLaunchTemplate(lt, insTypes)
	// fill nt
	h.fillNodeTemplate(nt)

	// 其他字段
	as.VpcID = bcsCluster.VpcID
	req.Creator = h.config.Username
	req.Provider = bcsCluster.Provider
	req.Region = bcsCluster.Region // 必须和集群一个地域
	req.NodeGroupType = "normal"   // 默认为普通节点池

	return nil
}

// createNodeGroupApi post
func (h *handler) createNodeGroupApi() string {
	return h.backendApi[createNodeGroupApi]
}

// DeleteNodeGroup 删除节点池.
// DeleteNodeGroupRequest参数说明, 如下:
//
// NodeGroupID不能为空;
//
// Operator推荐为空, 默认为当前用户;
//
// 其他参数不做验证;
func (h *handler) DeleteNodeGroup(ctx context.Context, req *pb.DeleteNodeGroupRequest) (*pb.DeleteNodeGroupResponse,
	error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.NodeGroupID) == 0 {
		return nil, errors.Errorf("req.nodeGroupID cannot be empty.")
	}
	// 操作人默认为当前用户
	req.Operator = h.config.Username

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.DeleteNodeGroupRequest marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Delete(ctx, requestID, h.deleteNodeGroupApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "delete node group failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.DeleteNodeGroupResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// deleteNodeGroupApi delete
func (h *handler) deleteNodeGroupApi(req *pb.DeleteNodeGroupRequest) string {
	// ( nodeGroupID )
	base := fmt.Sprintf(h.backendApi[deleteNodeGroupApi], req.NodeGroupID)
	return fmt.Sprintf("%s?isForce=%+v&reserveNodesInCluster=%+v&keepNodesInstance=%+v&operator=%s&onlyDeleteInfo=%+v",
		base, req.IsForce, req.ReserveNodesInCluster, req.KeepNodesInstance, req.Operator, req.OnlyDeleteInfo)
}

// UpdateNodeGroup 修改节点池.
// 请注意：不太推荐修改节点池参数; 云上节点池众多参数一经创建是不可修改的;
// 如果创建参数不正确, 基本上是无法修改的, 推荐先删除, 再重建.
//
// UpdateNodeGroupRequest参数说明, 支持修改参数如下:
//
// Name, 名称;
// Labels, 标签;
// Taints, 污点;
// Tags, 注解;
// EnableAutoscale, 是否开启节点池;
// AutoScalingGroup.MinSize, 下限;
// AutoScalingGroup.MaxSize, 上限;
// AutoScalingGroup.DesiredSize, 期望节点数;
// AutoScalingGroup.MultiZoneSubnetPolicy, 可用区子网模式;
// AutoScalingGroup.RetryPolicy, 重试策略;
// AutoScalingGroup.ScalingMode, 扩容模式;
func (h *handler) UpdateNodeGroup(ctx context.Context, req *pb.UpdateNodeGroupRequest) (*pb.UpdateNodeGroupResponse,
	error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.NodeGroupID) == 0 {
		return nil, errors.Errorf("req.nodeGroupID cannot be empty.")
	}
	// 操作人默认为当前用户
	req.Updater = h.config.Username

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.UpdateNodeGroupRequest marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Put(ctx, requestID, h.updateNodeGroupApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "update node group info failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.UpdateNodeGroupResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// updateNodeGroupApi put
func (h *handler) updateNodeGroupApi(req *pb.UpdateNodeGroupRequest) string {
	// ( nodeGroupID )
	return fmt.Sprintf(h.backendApi[updateNodeGroupApi], req.NodeGroupID)
}

// UpdateGroupDesiredNode 更新期望节点数量.
// UpdateGroupDesiredNodeRequest参数说明, 如下:
//
// NodeGroupID不能为空;
// DesiredNode表示期望节点数, 可以为0;当为0时, 表示把该节点池中所有节点都缩容掉;
//
// Operator推荐为空, 默认为当前用户;
// Manual推荐为空, 默认为true(表示手动);
func (h *handler) UpdateGroupDesiredNode(ctx context.Context, req *pb.UpdateGroupDesiredNodeRequest,
) (*pb.UpdateGroupDesiredNodeResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.NodeGroupID) == 0 {
		return nil, errors.Errorf("req.nodeGroupID cannot be empty.")
	}
	// 默认为手动
	req.Manual = true
	// 操作人默认为当前用户
	req.Operator = h.config.Username

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.UpdateGroupDesiredNodeRequest marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Post(ctx, requestID, h.updateGroupDesiredNodeApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "update node group desired failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.UpdateGroupDesiredNodeResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// updateGroupDesiredNodeApi post
func (h *handler) updateGroupDesiredNodeApi(req *pb.UpdateGroupDesiredNodeRequest) string {
	//( nodeGroupID )
	return fmt.Sprintf(h.backendApi[updateGroupDesiredNodeApi], req.NodeGroupID)
}

// UpdateGroupMinMaxSize 更新节点池上限、下限.
// UpdateGroupMinMaxSizeRequest参数说明, 如下:
//
// NodeGroupID不能为空;
// MaxSize不能为0;
//
// Operator推荐为空, 默认为当前用户;
func (h *handler) UpdateGroupMinMaxSize(ctx context.Context, req *pb.UpdateGroupMinMaxSizeRequest,
) (*pb.UpdateGroupMinMaxSizeResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.NodeGroupID) == 0 {
		return nil, errors.Errorf("req.nodeGroupID cannot be empty.")
	}
	if req.MaxSize == 0 {
		req.MaxSize = 1
	}
	// 操作人默认为当前用户
	req.Operator = h.config.Username

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.UpdateGroupMinMaxSizeRequest marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Post(ctx, requestID, h.updateGroupMinMaxSizeApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "update node group min or max size failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.UpdateGroupMinMaxSizeResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// updateGroupMinMaxSizeApi post
func (h *handler) updateGroupMinMaxSizeApi(req *pb.UpdateGroupMinMaxSizeRequest) string {
	// ( nodeGroupID )
	return fmt.Sprintf(h.backendApi[updateGroupMinMaxSizeApi], req.NodeGroupID)
}

// GetNodeGroup 查询节点池详细信息.
// GetNodeGroupRequest参数说明, 如下:
//
// NodeGroupID不能为空;
func (h *handler) GetNodeGroup(ctx context.Context, req *pb.GetNodeGroupRequest) (*pb.GetNodeGroupResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.NodeGroupID) == 0 {
		return nil, errors.Errorf("req.nodeGroupID cannot be empty.")
	}

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.GetNodeGroupRequest marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.getNodeGroupApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "get node group detail failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.GetNodeGroupResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// getNodeGroupApi get
func (h *handler) getNodeGroupApi(req *pb.GetNodeGroupRequest) string {
	//( nodeGroupID )
	return fmt.Sprintf(h.backendApi[getNodeGroupApi], req.NodeGroupID)
}

// ListClusterNodeGroup 查询当前集群所有节点池.
// ListClusterNodeGroupRequest参数说明, 如下:
//
// ClusterID不能为空;
func (h *handler) ListClusterNodeGroup(ctx context.Context, req *pb.ListClusterNodeGroupRequest) (
	*pb.ListClusterNodeGroupResponse, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ClusterID) == 0 {
		return nil, errors.Errorf("req.clusterID cannot be empty.")
	}

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.ListClusterNodeGroupRequest marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.listClusterNodeGroupApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "list node group failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.ListClusterNodeGroupResponse)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// listClusterNodeGroupApi get
func (h *handler) listClusterNodeGroupApi(req *pb.ListClusterNodeGroupRequest) string {
	// ( clusterID )
	base := fmt.Sprintf(h.backendApi[listClusterNodeGroupApi], req.ClusterID)
	return fmt.Sprintf("%s?enableFilter=%+v", base, req.EnableFilter)
}
