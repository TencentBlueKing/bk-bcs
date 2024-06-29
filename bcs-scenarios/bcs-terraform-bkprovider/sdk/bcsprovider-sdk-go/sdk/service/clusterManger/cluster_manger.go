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
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/header"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

/*
	集群管理
*/

// ImportCluster 集群创建(云凭证方式).
// ImportClusterReq参数说明, 如下:
//
// clusterName不能为空;
// provider不能为空;
// region不能为空;
// projectID不能为空;
// businessID不能为空;
// environment不能为空(stag,debug,prod);
// cloudAccountID不能为空(账号id, 如: BCS-tencentCloud-xxxxx);
// cloudMode.cloudID不能为空(云集群ID, 如: cls-xxxxxx);
//
// cloudMode.inter可以为空, 默认false, 表示为外网方式导入, 为true时, 表示以内网方式导入;
// manageType可以为空, 集群管理类型, 默认是 INDEPENDENT_CLUSTER(独立集群, 自行维护), MANAGED_CLUSTER(云上托管集群), 仅公有云时生效;
//
// engineType推荐为空, 默认为k8s;
// isExclusive推荐为空, 默认为true;
// clusterType推荐为空, 默认为single;
// creator推荐为空, 默认为当前用户;
// clusterCategory推荐为空, 默认为importer;
// is_shared推荐为空, 默认为false;
func (h *handler) ImportCluster(ctx context.Context, req *pb.ImportClusterReq) (*pb.ImportClusterResp, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	// 参数校验
	if err := h.fillImportClusterReq(req); err != nil {
		return nil, err
	}

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.ImportClusterReq marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Post(ctx, requestID, h.importClusterApi(), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "import cluster failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.ImportClusterResp)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// fillImportClusterReq 参数校验
func (h *handler) fillImportClusterReq(req *pb.ImportClusterReq) error {
	if len(req.ClusterName) == 0 {
		return errors.Errorf("clusterName cannot be empty.")
	}
	if len(req.Provider) == 0 {
		return errors.Errorf("provider cannot be empty.")
	}
	if len(req.Region) == 0 {
		return errors.Errorf("region cannot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return errors.Errorf("projectID cannot be empty.")
	}
	if len(req.BusinessID) == 0 {
		return errors.Errorf("businessID cannot be empty.")
	}
	if len(req.Environment) == 0 {
		return errors.Errorf("environment cannot be empty.")
	}
	if len(req.AccountID) == 0 {
		return errors.Errorf("accountID cannot be empty.")
	}
	if req.CloudMode == nil {
		return errors.Errorf("cloudMode cannot be empty.")
	}
	if len(req.CloudMode.CloudID) == 0 {
		return errors.Errorf("cloudMode.CloudID cannot be empty.")
	}

	switch req.ManageType {
	case IndependentCluster, ManagedCluster:
	default:
		req.ManageType = IndependentCluster
	}

	req.EngineType = "k8s"
	req.IsExclusive = wrapperspb.Bool(true)
	req.ClusterType = "single"
	// 操作人默认为当前用户
	req.Creator = h.config.Username
	req.IsShared = false

	return nil
}

// importClusterApi 集群创建(云凭证方式) post
func (h *handler) importClusterApi() string {
	return h.backendApi[importClusterApi]
}

// CreateCluster 集群创建(直接创建).
// CreateClusterReq参数说明, 如下:
//
// region不能为空;
// cloudAccountID不能为空(账号id, 如: BCS-tencentCloud-xxxxx);
// clusterName不能为空;
// environment不能为空(stag,debug,prod);
// manageType不能为空, 集群管理类型: INDEPENDENT_CLUSTER(独立集群, 自行维护), MANAGED_CLUSTER(云上托管集群), 仅公有云时生效;
// projectID不能为空;
// vpcID不能为空;
// provider不能为空;
// businessID不能为空;
// networkSettings字段不能为空;
// clusterBasicSettings.clusterLevel表示集群规格,若为托管集群时, 则本字段必填;
// clusterBasicSettings.version表示集群版本, 不能为空;
// clusterBasicSettings.module.workerModuleID表示node节点cc模块, 不能为空;
// clusterAdvanceSettings.containerRuntime表示运行时, 不能为空;
// clusterAdvanceSettings.runtimeVersion表示运行时版本, 不能为空;
// networkSettings表示集群网络基础设置, 不能为空;
// networkSettings.clusterIPv4CIDR表示IPv4地址池, 不能为空;
// networkSettings.maxNodePodNum表示节点上最大Pod数量, 不能为空;
// networkSettings.maxServiceNum表示集群最大的Service数量, 不能为空;
// nodes表示worker节点，不能为空;
// nodeSettings.workerLogin.initLoginPassword表示worker登录密码，不能为空;
//
// networkSettings.clusterIpType表示IP类型, 可以为空, 默认为ipv4;
// networkSettings.isStaticIpMode表示是否为非固定IP, 可以为空, 默认为false;
// clusterBasicSettings.OS表示镜像名称, 可以为空, 默认使用tlinux3.2x86_64;
// clusterBasicSettings.area.bkCloudID表示云区域id, 可以为空, 默认为0;
// clusterAdvanceSettings.networkType表示网络插件, 可以为空, 默认使用GR(Global Router，推荐使用)插件, 也可以使用VPC-CNI;
// clusterAdvanceSettings.clusterConnectSetting.isExtranet集群访问方式, 可以为空, 默认为false, 表示内网访问;若需求外网访问, 则设置为true;
// extraInfo.cloudProjectId表示云项目, 默认为0;
// networkType表示集群网络类型, 可以为空, 默认为“overlay”;
// description可以为空;
//
// engineType推荐为空, 默认为k8s;
// creator推荐为空, 默认为当前用户;
// isExclusive推荐为空, 默认为true;
// clusterType推荐为空, 默认为single;
// is_shared推荐为空, 默认为false;
// clusterCategory推荐为空, 默认为builder;
// clusterBasicSettings.isAutoUpgradeClusterLevel表示是否自动升级或降级, 默认为true;
func (h *handler) CreateCluster(ctx context.Context, req *pb.CreateClusterReq) (*pb.CreateClusterResp, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	switch req.ManageType {
	case IndependentCluster, ManagedCluster:
	default:
		return nil, errors.Errorf("manageType cannot be empty.")
	}
	if req.NetworkSettings == nil {
		return nil, errors.Errorf("networkSettings cannot be empty.")
	}
	if req.ClusterBasicSettings == nil {
		return nil, errors.Errorf("clusterBasicSettings cannot be empty.")
	}
	if req.ClusterBasicSettings.Module == nil {
		return nil, errors.Errorf("clusterBasicSettings.module cannot be empty.")
	}
	if req.ClusterAdvanceSettings == nil {
		return nil, errors.Errorf("clusterAdvanceSettings cannot be empty.")
	}
	if req.NodeSettings == nil {
		return nil, errors.Errorf("nodeSettings cannot be empty.")
	}
	if req.NodeSettings.WorkerLogin == nil {
		return nil, errors.Errorf("nodeSettings.workerLogin cannot be empty.")
	}
	// 参数校验
	if err := h.checkCreateClusterReq(req); err != nil {
		return nil, err
	}
	// 填充默认值
	h.fillCreateClusterReq(req)

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.CreateClusterReq marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Post(ctx, requestID, h.createClusterApi(), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "create cluster failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.CreateClusterResp)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// checkCreateClusterReq 空值检查
func (h *handler) checkCreateClusterReq(req *pb.CreateClusterReq) error {
	if len(req.Region) == 0 {
		return errors.Errorf("region cannot be empty.")
	}
	if len(req.CloudAccountID) == 0 {
		return errors.Errorf("cloudAccountID cannot be empty.")
	}
	if len(req.ClusterName) == 0 {
		return errors.Errorf("clusterName cannot be empty.")
	}
	if len(req.Environment) == 0 {
		return errors.Errorf("environment cannot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return errors.Errorf("projectID cannot be empty.")
	}
	if len(req.VpcID) == 0 {
		return errors.Errorf("vpcID cannot be empty.")
	}
	if len(req.Provider) == 0 {
		return errors.Errorf("provider cannot be empty.")
	}
	if len(req.BusinessID) == 0 {
		return errors.Errorf("businessID cannot be empty.")
	}
	if len(req.ClusterBasicSettings.Module.WorkerModuleID) == 0 {
		return errors.Errorf("clusterBasicSettings.module.workerModuleID cannot be empty.")
	}
	if len(req.ClusterBasicSettings.Version) == 0 {
		return errors.Errorf("clusterBasicSettings.version cannot be empty.")
	}
	if req.ManageType == ManagedCluster && len(req.ClusterBasicSettings.ClusterLevel) == 0 {
		return errors.Errorf("clusterBasicSettings.clusterLevel cannot be empty.")
	}
	if len(req.ClusterAdvanceSettings.ContainerRuntime) == 0 {
		return errors.Errorf("clusterAdvanceSettings.containerRuntime cannot be empty.")
	}
	if len(req.ClusterAdvanceSettings.RuntimeVersion) == 0 {
		return errors.Errorf("clusterAdvanceSettings.runtimeVersion cannot be empty.")
	}
	if len(req.NetworkSettings.ClusterIPv4CIDR) == 0 {
		return errors.Errorf("networkSettings.clusterIPv4CIDR cannot be empty.")
	}
	if req.NetworkSettings.MaxNodePodNum == 0 {
		return errors.Errorf("networkSettings.maxNodePodNum cannot be empty.")
	}
	if req.NetworkSettings.MaxServiceNum == 0 {
		return errors.Errorf("networkSettings.maxServiceNum cannot be empty.")
	}
	if len(req.Nodes) == 0 {
		return errors.Errorf("nodes cannot be empty.")
	}
	if len(req.NodeSettings.WorkerLogin.InitLoginPassword) == 0 {
		return errors.Errorf("nodeSettings.workerLogin.initLoginPassword cannot be empty.")
	}
	return nil
}

// fillCreateClusterReq 参数校验
func (h *handler) fillCreateClusterReq(req *pb.CreateClusterReq) {

	if len(req.NetworkSettings.ClusterIpType) == 0 {
		req.NetworkSettings.ClusterIpType = "ipv4"
	}
	if len(req.ClusterBasicSettings.OS) == 0 {
		req.ClusterBasicSettings.OS = "tlinux3.2x86_64"
	}
	if req.ClusterBasicSettings.Area == nil {
		req.ClusterBasicSettings.Area = &pb.CloudArea{
			BkCloudID: 0,
		}
	}
	if len(req.ClusterAdvanceSettings.NetworkType) == 0 {
		req.ClusterAdvanceSettings.NetworkType = "GR"
	}
	if req.ClusterAdvanceSettings.ClusterConnectSetting == nil {
		req.ClusterAdvanceSettings.ClusterConnectSetting = &pb.ClusterConnectSetting{
			IsExtranet: false,
		}
	}
	if len(req.ExtraInfo) == 0 {
		req.ExtraInfo = map[string]string{
			"cloudProjectId": "0",
		}
	}
	if len(req.NetworkType) == 0 {
		req.NetworkType = "overlay"
	}

	req.EngineType = "k8s"
	// 操作人默认为当前用户
	req.Creator = h.config.Username
	req.IsExclusive = true
	req.ClusterType = "single"
	req.IsShared = false
	req.ClusterCategory = "builder"
	req.ClusterBasicSettings.IsAutoUpgradeClusterLevel = true
}

// createClusterApi 集群创建(直接创建) post
func (h *handler) createClusterApi() string {
	return h.backendApi[createClusterApi]
}

// DeleteCluster 删除集群.
// DeleteClusterReq参数说明, 如下:
//
// ClusterID不能为空;
func (h *handler) DeleteCluster(ctx context.Context, req *pb.DeleteClusterReq) (*pb.DeleteClusterResp, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ClusterID) == 0 {
		return nil, errors.Errorf("req.clusterID cannot be empty.")
	}
	// 操作人默认为当前用户
	req.Operator = h.config.Username

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.DeleteClusterReq marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Delete(ctx, requestID, h.deleteClusterApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "delete cluster failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.DeleteClusterResp)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// deleteClusterApi 删除集群 delete ( clusterID )
func (h *handler) deleteClusterApi(req *pb.DeleteClusterReq) string {
	base := fmt.Sprintf(h.backendApi[deleteClusterApi], req.ClusterID)
	return fmt.Sprintf("%s?isForced=%+v&instanceDeleteMode=%s&onlyDeleteInfo=%+v&operator=%s&deleteClusterRecord=%+v",
		base, req.IsForced, req.InstanceDeleteMode, req.OnlyDeleteInfo, req.Operator, req.DeleteClusterRecord)
}

// UpdateCluster 更新集群.
// UpdateClusterReq参数说明, 如下:
//
// ClusterID不能为空;
//
// 更多可以修改参数如下:
// xxxx
func (h *handler) UpdateCluster(ctx context.Context, req *pb.UpdateClusterReq) (*pb.UpdateClusterResp, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ClusterID) == 0 {
		return nil, errors.Errorf("req.clusterID cannot be empty.")
	}
	// 操作人默认为当前用户
	req.Updater = h.config.Username

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.UpdateClusterReq marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Put(ctx, requestID, h.updateClusterApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "update cluster failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.UpdateClusterResp)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// updateClusterApi  更新集群 put ( clusterID )
func (h *handler) updateClusterApi(req *pb.UpdateClusterReq) string {
	return fmt.Sprintf(h.backendApi[updateClusterApi], req.ClusterID)
}

// GetCluster 查询集群详细信息.
// GetClusterReq参数说明, 如下:
//
// ClusterID不能为空;
func (h *handler) GetCluster(ctx context.Context, req *pb.GetClusterReq) (*pb.GetClusterResp, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ClusterID) == 0 {
		return nil, errors.Errorf("req.clusterID cannot be empty.")
	}
	req.CloudInfo = true

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.GetClusterReq marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.getClusterApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "get cluster failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.GetClusterResp)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// getClusterApi  查询集群 get ( clusterID )
func (h *handler) getClusterApi(req *pb.GetClusterReq) string {
	base := fmt.Sprintf(h.backendApi[getClusterApi], req.ClusterID)
	return fmt.Sprintf("%s?cloudInfo=%+v", base, req.CloudInfo)
}

// ListProjectCluster 查询某个项目下的Cluster列表.
// ListProjectClusterReq参数说明, 如下:
//
// ProjectID不能为空;
func (h *handler) ListProjectCluster(ctx context.Context, req *pb.ListProjectClusterReq) (*pb.ListProjectClusterResp,
	error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return nil, errors.Errorf("req.projectID cannot be empty.")
	}
	// 操作人默认为当前用户
	req.Operator = h.config.Username

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, req); err != nil {
		return nil, errors.Wrapf(err, "pb.ListProjectClusterReq marhsal failed.")
	}

	requestID := header.GenUUID()
	// 请求 (提示：1、切换http方法;  2、切换api方法;  3、设置http body)
	resp, err := h.roundtrip.Get(ctx, requestID, h.listProjectClusterApi(req), body.Bytes())
	if err != nil {
		return nil, errors.Wrapf(err, "list cluster in project failed, traceId: %s, body: %s", requestID,
			cast.ToString(body))
	}

	result := new(pb.ListProjectClusterResp)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// listProjectClusterApi  查询某个项目下的Cluster列表 get ( projectID )
func (h *handler) listProjectClusterApi(req *pb.ListProjectClusterReq) string {
	base := fmt.Sprintf(h.backendApi[listProjectClusterApi], req.ProjectID)
	return fmt.Sprintf("%s?region=%s&provider=%s&operator=%s", base, req.Region, req.Provider, req.Operator)
}
