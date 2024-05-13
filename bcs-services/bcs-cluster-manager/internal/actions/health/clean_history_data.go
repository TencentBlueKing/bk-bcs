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

package health

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	// DataTask data task
	DataTask = "task"
	// DataOperationLog data operation log
	DataOperationLog = "operationlog"
)

// CleanDBDataAction action for delete db data
type CleanDBDataAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CleanDbHistoryDataRequest
	resp  *cmproto.CleanDbHistoryDataResponse
	start string
	end   string
}

// NewCleanDBDataAction delete db data action
func NewCleanDBDataAction(model store.ClusterManagerModel) *CleanDBDataAction {
	return &CleanDBDataAction{
		model: model,
	}
}

func (ha *CleanDBDataAction) validate() error {
	if err := ha.req.Validate(); err != nil {
		return err
	}

	if ha.req.GetEndTime() <= ha.req.GetStartTime() {
		return fmt.Errorf("endTime(%v) lt startTime(%v)", ha.req.GetEndTime(), ha.req.GetStartTime())
	}

	startTime := utils.TransTsToTime(int64(ha.req.GetStartTime()))
	endTime := utils.TransTsToTime(int64(ha.req.GetEndTime()))

	// the cleanup time must not exceed one year.
	if endTime.Sub(startTime).Hours() > 24*365*time.Hour.Hours() {
		return fmt.Errorf("clean history data time range exceed one year")
	}

	// data from within the last three months cannot be cleaned.
	if endTime.Add(time.Hour*24*90).Sub(time.Now()) > 0 { // nolint
		return fmt.Errorf("cannot clean data within the last three months")
	}

	ha.start = utils.TransTsToStr(int64(ha.req.GetStartTime()))
	ha.end = utils.TransTsToStr(int64(ha.req.GetEndTime()))

	return nil
}

func (ha *CleanDBDataAction) setResp(code uint32, msg string) {
	ha.resp.Code = code
	ha.resp.Message = msg
	ha.resp.Result = false
	if code == common.BcsErrClusterManagerSuccess {
		ha.resp.Result = true
	}
}

func (ha *CleanDBDataAction) cleanDataTypeHistory() error {
	switch ha.req.DataType {
	case DataTask:
		if err := ha.model.DeleteFinishedTaskByDate(ha.ctx, ha.start, ha.end); err != nil {
			return err
		}
	case DataOperationLog:
		if err := ha.model.DeleteOperationLogByDate(ha.ctx, ha.start, ha.end); err != nil {
			return err
		}
	default:
		return fmt.Errorf("not supported dataType[%s]", ha.req.DataType)
	}

	return nil
}

// Handle handle delete db data check
func (ha *CleanDBDataAction) Handle(
	ctx context.Context, req *cmproto.CleanDbHistoryDataRequest, resp *cmproto.CleanDbHistoryDataResponse) {
	if req == nil || resp == nil {
		blog.Errorf("delete db data check failed, req or resp is empty")
		return
	}
	ha.ctx = ctx
	ha.req = req
	ha.resp = resp

	if err := ha.validate(); err != nil {
		ha.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	err := ha.cleanDataTypeHistory()
	if err != nil {
		ha.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	ha.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
