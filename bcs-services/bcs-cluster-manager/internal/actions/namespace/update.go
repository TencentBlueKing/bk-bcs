/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package namespace

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"

	k8scorev1 "k8s.io/api/core/v1"
)

// UpdateAction action for update namespace
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateNamespaceReq
	resp  *cmproto.UpdateNamespaceResp
	ns    *cmproto.Namespace
}

// NewUpdateAction create action for udpate
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) validate() error {
	if err := ua.req.Validate(); err != nil {
		return err
	}
	return nil
}

func (ua *UpdateAction) getNamespace() error {
	ns, err := ua.model.GetNamespace(ua.ctx, ua.req.Name, ua.req.FederationClusterID)
	if err != nil {
		return err
	}
	ua.ns = ns
	return nil
}

func (ua *UpdateAction) updateNamespace() error {
	if len(ua.req.MaxQuota) != 0 {
		maxQuota := &k8scorev1.ResourceQuota{}
		if err := json.Unmarshal([]byte(ua.req.MaxQuota), maxQuota); err != nil {
			blog.Warnf("decode max quota %s to k8s ResourceQuota failed, err %s", ua.req.MaxQuota, err.Error())
			return fmt.Errorf("decode max quota %s to k8s ResourceQuota failed, err %s", ua.req.MaxQuota, err.Error())
		}
	}
	newNs := &cmproto.Namespace{
		Name:                ua.ns.Name,
		FederationClusterID: ua.ns.FederationClusterID,
		MaxQuota:            ua.req.MaxQuota,
		Labels:              ua.req.Labels,
		CreateTime:          ua.ns.CreateTime,
		UpdateTime:          time.Now().Format(time.RFC3339),
	}
	return ua.model.UpdateNamespace(ua.ctx, newNs)
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle update namespace request
func (ua *UpdateAction) Handle(ctx context.Context,
	req *cmproto.UpdateNamespaceReq, resp *cmproto.UpdateNamespaceResp) {
	if req == nil || resp == nil {
		blog.Errorf("update cluster failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ua.updateNamespace(); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
