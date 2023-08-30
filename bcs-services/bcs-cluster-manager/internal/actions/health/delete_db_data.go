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
 *
 */

package health

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

const (
	DataTask         = "task"
	DataOperationLog = "operationLog"
)

// DeleteDBDataAction action for delete db data
type DeleteDBDataAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.DeleteDBDataReq
	resp  *cmproto.DeleteDBDataResp
}

// NewDeleteDBDataAction delete db data action
func NewDeleteDBDataAction(model store.ClusterManagerModel) *DeleteDBDataAction {
	return &DeleteDBDataAction{
		model: model,
	}
}

func (ha *DeleteDBDataAction) validate() error {
	if err := ha.req.Validate(); err != nil {
		return err
	}

	layout := "2006-01-02"

	st, err := time.ParseInLocation(layout, ha.req.StartTime, time.Local)
	if err != nil {
		return fmt.Errorf("start time parse error: %s", err)
	}

	ha.req.StartTime = st.Format(time.RFC3339)

	et, err := time.ParseInLocation(layout, ha.req.EndTime, time.Local)
	if err != nil {
		return fmt.Errorf("end time parse error: %s", err)
	}

	ha.req.EndTime = et.Format(time.RFC3339)

	return nil
}

func (ha *DeleteDBDataAction) setResp(code uint32, msg string) {
	ha.resp.Code = code
	ha.resp.Message = msg
	ha.resp.Result = false
	if code == common.BcsErrClusterManagerSuccess {
		ha.resp.Result = true
	}
}

// Handle handle delete db data check
func (ha *DeleteDBDataAction) Handle(
	ctx context.Context, req *cmproto.DeleteDBDataReq, resp *cmproto.DeleteDBDataResp) {
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

	types := strings.Split(ha.req.DataType, ",")
	for _, t := range types {
		if t == DataTask {
			if err := ha.model.DeleteFinishTaskByDate(ha.ctx, ha.req.StartTime, ha.req.EndTime); err != nil {
				ha.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
				return
			}
		} else if t == DataOperationLog {
			if err := ha.model.DeleteOperationLogByDate(ha.ctx, ha.req.StartTime, ha.req.EndTime); err != nil {
				ha.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
				return
			}
		}
	}

	ha.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
