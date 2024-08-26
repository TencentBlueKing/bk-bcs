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

// Package helmManger helm-service
package helmManger

import (
	"bytes"
	"context"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/pkg/errors"
	"github.com/spf13/cast"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/header"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/roundtrip"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-bkprovider/sdk/bcsprovider-sdk-go/sdk/common/utils"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// Service 对接 bcs-helm-manager
type Service interface {
	// InstallAddons 安装集群组件
	InstallAddons(ctx context.Context, req *InstallAddonsRequest) (*pb.UpgradeAddonsResp, error)

	// UninstallAddons 卸载集群组件
	UninstallAddons(ctx context.Context, req *UninstallAddonsRequest) (*pb.UninstallAddonsResp, error)

	// UpgradeAddons 升级或更新集群组件
	UpgradeAddons(ctx context.Context, req *UpgradeAddonsRequest) (*pb.UpgradeAddonsResp, error)

	// GetAddonsDetail 查询集群组件详情
	GetAddonsDetail(ctx context.Context, req *GetAddonsDetailRequest) (*pb.GetAddonsDetailResp, error)

	// ListAddons 查询集群组件列表
	ListAddons(ctx context.Context, req *ListAddonsRequest) (*pb.ListAddonsResp, error)
}

// NewService return Service
func NewService(config *options.Config, roundtrip roundtrip.Client) (Service, error) {
	if config == nil {
		return nil, errors.Errorf("config cannot be empty.")
	}
	if roundtrip == nil {
		return nil, errors.Errorf("roundtrip cannot be empty.")
	}

	h := &handler{
		config: config,
		// roundtrip
		roundtrip: roundtrip,
		// api
		backendApi: map[string]string{},
	}
	if err := h.init(); err != nil {
		return nil, errors.Wrapf(err, "init handler failed")
	}

	return h, nil
}

// handler impl Service
type handler struct {
	// config 配置
	config *options.Config

	// roundtrip http client
	roundtrip roundtrip.Client

	// backendApi 后端完整路径
	backendApi map[string]string
}

func (h *handler) init() error {
	apis := []string{
		installAddonsApi,
		uninstallAddonsApi,
		upgradeAddonsApi,
		getAddonsDetailApi,
		listAddonsApi,
	}
	addr := h.config.BcsGatewayAddr

	for _, api := range apis {
		h.backendApi[api] = utils.PathJoin(addr, api)
	}

	return nil
}

// InstallAddons 安装集群组件
func (h *handler) InstallAddons(ctx context.Context, req *InstallAddonsRequest) (*pb.UpgradeAddonsResp, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return nil, errors.Errorf("project_code connot be empty.")
	}
	if len(req.ClusterID) == 0 {
		return nil, errors.Errorf("project_clusterID connot be empty.")
	}
	if len(req.Name) == 0 {
		return nil, errors.Errorf("project_name connot be empty.")
	}
	if len(req.Version) == 0 {
		return nil, errors.Errorf("project_version connot be empty.")
	}
	if len(req.Values) == 0 {
		return nil, errors.Errorf("project_values connot be empty.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	// 转换(提示：1、切换方法)
	body, err := h.installAddonsReqToBytes(req)
	if err != nil {
		return nil, err
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Put(ctx, req.RequestID, h.installAddonsApi(req), body)
	if err != nil {
		return nil, errors.Wrapf(err, "install addons failed, traceId: %s, projectCode: %s",
			req.RequestID, req.ProjectID)
	}

	result := new(pb.UpgradeAddonsResp)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// installAddonsReqToBytes to pb struct data
func (h *handler) installAddonsReqToBytes(req *InstallAddonsRequest) ([]byte, error) {
	obj := &pb.UpgradeAddonsReq{
		ProjectCode: utils.ToPtr(req.ProjectID),
		ClusterID:   utils.ToPtr(req.ClusterID),
		Name:        utils.ToPtr(req.Name),
		Version:     utils.ToPtr(req.Version),
		Values:      utils.ToPtr(req.Values),
	}

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, obj); err != nil {
		return nil, errors.Wrapf(err, "pb.UpgradeAddonsReq marhsal failed.")
	}

	return body.Bytes(), nil
}

// installAddonsApi api
func (h *handler) installAddonsApi(req *InstallAddonsRequest) string {
	// install or upgrade 共用一个接口 : projectCode + clusterID + name
	return fmt.Sprintf(h.backendApi[upgradeAddonsApi], req.ProjectID, req.ClusterID, req.Name)
}

// UninstallAddons 卸载集群组件
func (h *handler) UninstallAddons(ctx context.Context, req *UninstallAddonsRequest) (*pb.UninstallAddonsResp, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return nil, errors.Errorf("project_id connot be empty.")
	}
	if len(req.ClusterID) == 0 {
		return nil, errors.Errorf("project_clusterID connot be empty.")
	}
	if len(req.Name) == 0 {
		return nil, errors.Errorf("project_name connot be empty.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Delete(ctx, req.RequestID, h.uninstallAddonsApi(req), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "uninstall addons failed, traceId: %s, projectId: %s",
			req.RequestID, req.ProjectID)
	}

	result := new(pb.UninstallAddonsResp)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// uninstallAddonsApi api
func (h *handler) uninstallAddonsApi(req *UninstallAddonsRequest) string {
	// projectCode + clusterID + name
	return fmt.Sprintf(h.backendApi[uninstallAddonsApi], req.ProjectID, req.ClusterID, req.Name)
}

// UpgradeAddons 升级或更新集群组件
func (h *handler) UpgradeAddons(ctx context.Context, req *UpgradeAddonsRequest) (*pb.UpgradeAddonsResp, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return nil, errors.Errorf("project_code connot be empty.")
	}
	if len(req.ClusterID) == 0 {
		return nil, errors.Errorf("project_clusterID connot be empty.")
	}
	if len(req.Name) == 0 {
		return nil, errors.Errorf("project_name connot be empty.")
	}
	if len(req.Version) == 0 {
		return nil, errors.Errorf("project_version connot be empty.")
	}
	if len(req.Values) == 0 {
		return nil, errors.Errorf("project_values connot be empty.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	// 转换(提示：1、切换方法)
	body, err := h.updateProjectReqToBytes(req)
	if err != nil {
		return nil, err
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Put(ctx, req.RequestID, h.updateProjectApi(req), body)
	if err != nil {
		return nil, errors.Wrapf(err, "update addons failed, traceId: %s, projectCode: %s",
			req.RequestID, req.ProjectID)
	}

	result := new(pb.UpgradeAddonsResp)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// updateProjectReqToBytes to pb struct data
func (h *handler) updateProjectReqToBytes(req *UpgradeAddonsRequest) ([]byte, error) {
	obj := &pb.UpgradeAddonsReq{
		ProjectCode: utils.ToPtr(req.ProjectID),
		ClusterID:   utils.ToPtr(req.ClusterID),
		Name:        utils.ToPtr(req.Name),
		Version:     utils.ToPtr(req.Version),
		Values:      utils.ToPtr(req.Values),
	}

	body := bytes.NewBuffer(nil)
	// note: 两种json方式 json.marshal or proto.Marshal
	if err := pbMarshaller.Marshal(body, obj); err != nil {
		return nil, errors.Wrapf(err, "pb.UpgradeAddonsReq marhsal failed.")
	}

	return body.Bytes(), nil
}

// upgradeAddonsApi api
func (h *handler) updateProjectApi(req *UpgradeAddonsRequest) string {
	// install or upgrade 共用一个接口 : projectCode + clusterID + name
	return fmt.Sprintf(h.backendApi[upgradeAddonsApi], req.ProjectID, req.ClusterID, req.Name)
}

// GetAddonsDetail 查询集群组件详情
func (h *handler) GetAddonsDetail(ctx context.Context, req *GetAddonsDetailRequest) (*pb.GetAddonsDetailResp, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return nil, errors.Errorf("project_code connot be empty.")
	}
	if len(req.ClusterID) == 0 {
		return nil, errors.Errorf("project_clusterID connot be empty.")
	}
	if len(req.Name) == 0 {
		return nil, errors.Errorf("project_name connot be empty.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, req.RequestID, h.getAddonsDetailApi(req), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "get addons detail failed, traceId: %s, projectId: %s",
			req.RequestID, req.ProjectID)
	}

	result := new(pb.GetAddonsDetailResp)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// getAddonsDetailApi api
func (h *handler) getAddonsDetailApi(req *GetAddonsDetailRequest) string {
	// projectCode + clusterID + name
	return fmt.Sprintf(h.backendApi[getAddonsDetailApi], req.ProjectID, req.ClusterID, req.Name)
}

// ListAddons 查询集群组件列表
func (h *handler) ListAddons(ctx context.Context, req *ListAddonsRequest) (*pb.ListAddonsResp, error) {
	if req == nil {
		return nil, errors.Errorf("req cannot be empty.")
	}
	if len(req.ProjectID) == 0 {
		return nil, errors.Errorf("project_code connot be empty.")
	}
	if len(req.ClusterID) == 0 {
		return nil, errors.Errorf("project_clusterID connot be empty.")
	}
	if len(req.RequestID) == 0 {
		req.RequestID = header.GenUUID()
	}

	// 请求 (提示：1、切换http方法；2、切换api方法；3、设置http body)
	resp, err := h.roundtrip.Get(ctx, req.RequestID, h.listProjectsApi(req), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "list addons failed, traceId: %s", req.RequestID)
	}

	result := new(pb.ListAddonsResp)
	if err = jsonpb.Unmarshal(bytes.NewBuffer(resp), result); err != nil {
		return nil, errors.Wrapf(err, "resp unmarshal failed, resp: %s", cast.ToString(resp))
	}

	return result, nil
}

// listAddonsApi api
func (h *handler) listProjectsApi(req *ListAddonsRequest) string {
	// projectCode + clusterID
	return fmt.Sprintf(h.backendApi[listAddonsApi], req.ProjectID, req.ClusterID)
}
