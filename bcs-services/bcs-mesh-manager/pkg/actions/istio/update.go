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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/store/entity"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

// UpdateIstioAction handles istio update request
type UpdateIstioAction struct {
	model store.MeshManagerModel
	req   *meshmanager.UpdateIstioRequest
	resp  *meshmanager.UpdateIstioResponse
}

// NewUpdateIstioAction create update istio action
func NewUpdateIstioAction(model store.MeshManagerModel) *UpdateIstioAction {
	return &UpdateIstioAction{
		model: model,
	}
}

// Handle processes the istio update request
func (u *UpdateIstioAction) Handle(
	ctx context.Context,
	req *meshmanager.UpdateIstioRequest,
	resp *meshmanager.UpdateIstioResponse,
) error {
	u.req = req
	u.resp = resp

	if err := u.req.Validate(); err != nil {
		blog.Errorf("update mesh failed, invalid request, %s, param: %v", err.Error(), u.req)
		u.setResp(common.ParamErrorCode, err.Error())
		return nil
	}

	if err := u.update(ctx); err != nil {
		blog.Errorf("update mesh failed, %s, meshID: %s", err.Error(), u.req.MeshID)
		u.setResp(common.DBErrorCode, err.Error())
		return nil
	}

	u.setResp(common.SuccessCode, "")
	blog.Infof("update mesh successfully, meshID: %s", u.req.MeshID)
	return nil
}

// setResp sets the response with code and message
func (u *UpdateIstioAction) setResp(code uint32, message string) {
	u.resp.Code = code
	u.resp.Message = message
}

// ServiceDiscoveryFromProto converts proto ServiceDiscovery to entity.ServiceDiscovery
func ServiceDiscoveryFromProto(proto *meshmanager.ServiceDiscovery) *entity.ServiceDiscovery {
	if proto == nil {
		return nil
	}
	sd := &entity.ServiceDiscovery{
		Clusters:           proto.Clusters,
		AutoInjectNS:       make(map[string][]string),
		DisabledInjectPods: make(map[string]map[string][]string),
	}
	for clusterID, namespaceList := range proto.AutoInjectNS {
		sd.AutoInjectNS[clusterID] = namespaceList.Namespaces
	}
	for clusterID, namespacePods := range proto.DisabledInjectPods {
		clusterPods := make(map[string][]string)
		for namespace, podList := range namespacePods.NamespacePods {
			clusterPods[namespace] = podList.Pods
		}
		sd.DisabledInjectPods[clusterID] = clusterPods
	}
	return sd
}

// update implements the business logic for updating mesh
func (u *UpdateIstioAction) update(ctx context.Context) error {
	mesh := &entity.MeshIstio{}
	updateFields := mesh.UpdateFromProto(u.req)

	// 迁移 ServiceDiscovery 字段的转换逻辑到 action 层
	if u.req.ServiceDiscovery != nil {
		updateFields["serviceDiscovery"] = ServiceDiscoveryFromProto(u.req.ServiceDiscovery)
	}

	return u.model.Update(ctx, u.req.MeshID, updateFields)
}
