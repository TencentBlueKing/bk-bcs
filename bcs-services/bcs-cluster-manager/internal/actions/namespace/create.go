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
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"

	k8scorev1 "k8s.io/api/core/v1"
)

// CreateAction action for create namespace
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CreateNamespaceReq
	resp  *cmproto.CreateNamespaceResp
}

// NewCreateAction create namespace action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

func (ca *CreateAction) queryFederationCluster(clusterID string) error {
	_, err := ca.model.GetCluster(ca.ctx, clusterID)
	return err
}

func (ca *CreateAction) createNamespace() error {
	if len(ca.req.MaxQuota) != 0 {
		maxQuota := &k8scorev1.ResourceQuota{}
		if err := json.Unmarshal([]byte(ca.req.MaxQuota), maxQuota); err != nil {
			blog.Warnf("decode max quota %s to k8s ResourceQuota failed, err %s", ca.req.MaxQuota, err.Error())
			return fmt.Errorf("decode max quota %s to k8s ResourceQuota failed, err %s", ca.req.MaxQuota, err.Error())
		}
	}
	createTime := time.Now().Format(time.RFC3339)
	newNs := &cmproto.Namespace{
		Name:                ca.req.Name,
		FederationClusterID: ca.req.FederationClusterID,
		ProjectID:           ca.req.ProjectID,
		BusinessID:          ca.req.BusinessID,
		Labels:              ca.req.Labels,
		MaxQuota:            ca.req.MaxQuota,
		CreateTime:          createTime,
		UpdateTime:          createTime,
	}
	return ca.model.CreateNamespace(ca.ctx, newNs)
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create namespace request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateNamespaceReq, resp *cmproto.CreateNamespaceResp) {
	if req == nil || resp == nil {
		blog.Errorf("create namespace failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := req.Validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ca.queryFederationCluster(ca.req.FederationClusterID); err != nil {
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	if err := ca.createNamespace(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return
		}
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
