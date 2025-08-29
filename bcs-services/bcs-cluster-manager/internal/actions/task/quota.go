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

package task

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	itypes "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/types"
)

// ListClusterTaskQuotaAction action for list online cluster credential
type ListClusterTaskQuotaAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.ListClusterTaskQuotaRequest
	resp  *cmproto.ListClusterTaskQuotaResponse
}

// NewListClusterTaskQuotaAction create list action for cluster task quota
func NewListClusterTaskQuotaAction(model store.ClusterManagerModel) *ListClusterTaskQuotaAction {
	return &ListClusterTaskQuotaAction{
		model: model,
	}
}

func (la *ListClusterTaskQuotaAction) listTaskQuota() error {

	tasks, err := la.model.ListTaskQuota(la.ctx, la.req.ClusterID, &storeopt.ListOption{})
	if err != nil {
		return err
	}
	quotaData := convertQuotaData(tasks)
	la.resp.Data = quotaData
	return nil
}

// convertQuotaData convert quota data to proto data
func convertQuotaData(tasks []*itypes.ClusterTaskQuota) []*cmproto.ClusterTask {
	quotaData := []*cmproto.ClusterTask{}
	quotaDataUnique := map[string][]*cmproto.ClusterTaskQuota{}
	for i := range tasks {
		quotaDataUnique[tasks[i].ClusterId] = append(quotaDataUnique[tasks[i].ClusterId], &cmproto.ClusterTaskQuota{
			SuccessRate:      tasks[i].SuccessRate,
			ExecutionTime:    tasks[i].AvgExecutionTime,
			FailureTopReason: tasks[i].TopFailReason,
			TaskType:         tasks[i].TaskType,
		})
	}

	for clusterID, quota := range quotaDataUnique {
		quotaData = append(quotaData, &cmproto.ClusterTask{
			ClusterId:  clusterID,
			TaskQuotas: quota,
		})
	}
	return quotaData
}

func (la *ListClusterTaskQuotaAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle list cluster task quota
func (la *ListClusterTaskQuotaAction) Handle(
	ctx context.Context, req *cmproto.ListClusterTaskQuotaRequest, resp *cmproto.ListClusterTaskQuotaResponse) {
	if req == nil || resp == nil {
		blog.Errorf("list Task Quota failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listTaskQuota(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
