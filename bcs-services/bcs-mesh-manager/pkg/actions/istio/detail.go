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

package istio

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/helmmanager"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/cmd/mesh-manager/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/clients/helm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/utils"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// GetIstioDetailAction action for get istio detail
type GetIstioDetailAction struct {
	istioConfig *options.IstioConfig
	model       store.MeshManagerModel
	req         *meshmanager.GetIstioDetailRequest
	resp        *meshmanager.GetIstioDetailResponse
}

// NewGetIstioDetailAction create get istio detail action
func NewGetIstioDetailAction(istioConfig *options.IstioConfig, model store.MeshManagerModel) *GetIstioDetailAction {
	return &GetIstioDetailAction{
		istioConfig: istioConfig,
		model:       model,
	}
}

// Handle processes the mesh list request
func (l *GetIstioDetailAction) Handle(
	ctx context.Context,
	req *meshmanager.GetIstioDetailRequest,
	resp *meshmanager.GetIstioDetailResponse,
) error {
	l.req = req
	l.resp = resp

	if err := l.req.Validate(); err != nil {
		blog.Errorf("get istio detail failed, invalid request, %s, param: %v", err.Error(), l.req)
		l.setResp(common.ParamErrorCode, err.Error(), nil)
		return nil
	}

	result, err := l.getDetail(ctx)
	if err != nil {
		blog.Errorf("get istio detail failed, %s, meshID: %s", err.Error(), l.req.MeshID)
		l.setResp(common.DBErrorCode, err.Error(), nil)
		return nil
	}

	// 设置成功响应
	l.setResp(common.SuccessCode, "", result)
	blog.Infof("get istio detail successfully, meshID: %s", l.req.MeshID)
	return nil
}

// setResp sets the response with code, message and data
func (l *GetIstioDetailAction) setResp(code uint32, message string, data *meshmanager.IstioDetailInfo) {
	l.resp.Code = code
	l.resp.Message = message
	l.resp.Data = data
}

func (l *GetIstioDetailAction) getDetail(ctx context.Context) (*meshmanager.IstioDetailInfo, error) {
	// 构建查询条件
	cond := l.buildQueryConditions()

	meshIstio, err := l.model.Get(ctx, cond)
	if err != nil {
		return nil, err
	}
	if meshIstio == nil {
		blog.Errorf("get mesh istio failed, meshID: %s", l.req.MeshID)
		return nil, fmt.Errorf("get mesh istio failed, meshID: %s", l.req.MeshID)
	}

	// 检查 PrimaryClusters 是否为空
	if len(meshIstio.PrimaryClusters) == 0 {
		blog.Errorf("mesh PrimaryClusters is empty, meshID: %s", meshIstio.MeshID)
		return nil, fmt.Errorf("mesh PrimaryClusters is empty, meshID: %s", meshIstio.MeshID)
	}

	// 安装中或安装失败的istio不获取release的values信息
	if meshIstio.Status == common.IstioStatusInstalling || meshIstio.Status == common.IstioStatusInstallFailed {
		return meshIstio.Transfer2ProtoForDetail(), nil
	}

	// istio 状态不在安装中才需要查询 release 的 values 信息
	clusterID := meshIstio.PrimaryClusters[0]
	namespace := common.IstioNamespace
	istiodName := common.IstioInstallIstiodName

	// 调用 RPC 接口获取 release 详情
	release, err := helm.GetReleaseDetail(
		ctx,
		&helmmanager.GetReleaseDetailV1Req{
			ProjectCode: &l.req.ProjectCode,
			ClusterID:   &clusterID,
			Namespace:   &namespace,
			Name:        &istiodName,
		},
	)
	if err != nil {
		blog.Errorf("get release detail failed, clusterID: %s, err: %s", clusterID, err.Error())
		return nil, fmt.Errorf("get release detail failed, clusterID: %s, err: %s", clusterID, err.Error())
	}
	if release == nil || release.Data == nil {
		blog.Errorf("get release detail failed, clusterID: %s, release is nil", clusterID)
		return nil, fmt.Errorf("get release detail failed, clusterID: %s, release is nil", clusterID)
	}

	if len(release.Data.Values) == 0 {
		blog.Errorf("release values is empty, clusterID: %s", clusterID)
		return nil, fmt.Errorf("release values is empty, clusterID: %s", clusterID)
	}
	value := release.Data.Values[0]

	// 解析 values 为 IstiodInstallValues 结构
	istiodValues := &common.IstiodInstallValues{}
	if err = yaml.Unmarshal([]byte(value), istiodValues); err != nil {
		blog.Errorf("unmarshal istiod values failed, clusterID: %s, err: %s", clusterID, err.Error())
		return nil, fmt.Errorf("unmarshal istiod values failed, clusterID: %s", clusterID)
	}

	// 基于实际的部署配置构建返回的 IstioListItem
	result, err := utils.ConvertValuesToIstioDetailInfo(meshIstio, istiodValues)
	if err != nil {
		blog.Errorf("build istio list item failed, clusterID: %s, err: %s", clusterID, err.Error())
		return nil, fmt.Errorf("build istio list item failed, clusterID: %s", clusterID)
	}

	// 转换资源配置单位
	if err := utils.NormalizeResourcesConfig(result); err != nil {
		blog.Errorf("normalize resources failed, clusterID: %s, err: %s",
			clusterID, err.Error())
		return nil, fmt.Errorf("normalize istio detail resources failed, clusterID: %s, err: %s",
			clusterID, err.Error())
	}

	return result, nil
}

// buildQueryConditions 构建查询条件
func (l *GetIstioDetailAction) buildQueryConditions() *operator.Condition {
	conditions := make([]*operator.Condition, 0)

	if l.req.ProjectCode != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			"projectCode": l.req.ProjectCode,
		}))
	}
	if l.req.MeshID != "" {
		conditions = append(conditions, operator.NewLeafCondition(operator.Eq, operator.M{
			"meshID": l.req.MeshID,
		}))
	}

	if len(conditions) > 0 {
		return operator.NewBranchCondition(operator.And, conditions...)
	}
	return operator.NewBranchCondition(operator.And)
}
