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

// Package quota xxx
package quota

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/convert"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// CreateQuotaAction action for create project quota
type CreateQuotaAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.CreateProjectQuotaRequest
	resp  *proto.ProjectQuotaResponse

	project *pm.Project
	cluster *clustermanager.Cluster
	sQuota  *quota.ProjectQuota
	task    *types.Task
	pQuota  *proto.ProjectQuota
}

// NewCreateQuotaAction new create project quota action
func NewCreateQuotaAction(model store.ProjectModel) *CreateQuotaAction {
	return &CreateQuotaAction{
		model: model,
	}
}

// validate check project quota request
func (ca *CreateQuotaAction) validate() error {
	err := ca.req.Validate()
	if err != nil {
		return err
	}

	// project validate
	p, err := checkProjectValidate(ca.model, ca.req.ProjectID, ca.req.ProjectCode, "")
	if err != nil {
		return err
	}
	ca.project = p

	// cluster validate
	cls, err := checkClusterValidate(ca.req.ClusterId)
	if err != nil {
		return err
	}
	ca.cluster = cls
	// provider validate: provider && quotaType

	return nil
}

// setDefaultReqValues set req default values
func (ca *CreateQuotaAction) setDefaultReqValues() {
	if ca.project != nil {
		if len(ca.req.GetProjectID()) == 0 {
			ca.req.ProjectID = ca.project.ProjectID
		}
		if len(ca.req.GetProjectCode()) == 0 {
			ca.req.ProjectCode = ca.project.ProjectCode
		}
		if len(ca.req.GetBusinessID()) == 0 {
			ca.req.BusinessID = ca.project.BusinessID
		}
	}

	if ca.cluster != nil {
		if len(ca.req.ClusterName) == 0 {
			ca.req.ClusterName = ca.cluster.ClusterName
		}
	}
}

// createProjectQuota create project quota && associate with provider
func (ca *CreateQuotaAction) createProjectQuota() error {
	pQuota := &quota.ProjectQuota{
		CreateTime:  time.Now().Unix(),
		UpdateTime:  time.Now().Unix(),
		QuotaId:     stringx.GenerateRandomID("quota"),
		QuotaName:   ca.req.QuotaName,
		Description: ca.req.Description,
		Provider:    ca.req.Provider,
		QuotaType:   quota.ProjectQuotaType(ca.req.QuotaType),
		Quota:       &quota.QuotaResource{},
		ProjectId:   ca.req.ProjectID,
		ProjectCode: ca.req.ProjectCode,
		ClusterId:   ca.req.ClusterId,
		Namespace:   ca.req.GetNameSpace(),
		BusinessId:  ca.req.GetBusinessID(),
		IsDeleted:   false,
		Status:      quota.Creating,
		Labels:      ca.req.GetLabels(),
	}
	// 从 context 中获取 username
	if authUser, err := middleware.GetUserFromContext(ca.ctx); err == nil {
		pQuota.Creator = authUser.GetUsername()
	}
	// trans proto quota to store quota
	pQuota.Quota = quota.TransPorto2StoreQuota(ca.req.Quota)

	ca.sQuota = pQuota
	ca.pQuota = quota.TransStore2ProtoQuota(pQuota)

	return ca.model.CreateProjectQuota(ca.ctx, pQuota)
}

// dispatchTask dispatch quota task
func (ca *CreateQuotaAction) dispatchTask() error {
	quotaMgr, err := manager.GetQuotaManager(ca.req.Provider)
	if err != nil {
		return err
	}

	task, err := quotaMgr.CreateProjectQuota(ca.pQuota, &provider.CreateProjectQuotaOptions{})
	if err != nil {
		return err
	}
	ca.task = task

	err = manager.GetTaskServer().Dispatch(task)
	if err != nil {
		return err
	}

	return nil
}

// Do create project request
func (ca *CreateQuotaAction) Do(ctx context.Context,
	req *proto.CreateProjectQuotaRequest, resp *proto.ProjectQuotaResponse) error {
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		return errorx.NewReadableErr(errorx.ParamErr, err.Error())
	}

	ca.setDefaultReqValues()
	if err := ca.createProjectQuota(); err != nil {
		return errorx.NewDBErr(err.Error())
	}
	if err := ca.dispatchTask(); err != nil {
		return errorx.NewBuildTaskErr(err.Error())
	}

	// set resp data
	task, err := convert.MarshalInterfaceToValue(ca.task)
	if err != nil {
		return err
	}
	resp.Task = task
	resp.Data = ca.pQuota

	return nil
}
