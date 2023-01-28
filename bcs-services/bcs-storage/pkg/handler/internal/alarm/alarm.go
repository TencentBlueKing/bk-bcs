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

package alarm

import (
	"context"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	storage "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/proto"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/alarms"
)

var (
	// alarmConditionTagList alarm 条件 tag list
	alarmConditionTagList = []string{constants.ClusterIDTag, constants.NamespaceTag, constants.SourceTag,
		constants.ModuleTag,
	}
)

// alarmToM alarm 对象转 operator.M 结构
func alarmToM(req *storage.PostAlarmRequest) operator.M {
	return operator.M{
		constants.ClusterIDTag:    req.ClusterId,
		constants.NamespaceTag:    req.Namespace,
		constants.MessageTag:      req.Message,
		constants.SourceTag:       req.Source,
		constants.ModuleTag:       req.Module,
		constants.TypeTag:         req.Type,
		constants.ReceivedTimeTag: time.Unix(req.ReceivedTime, 0),
		constants.DataTag:         req.Data,
	}
}

// getCondition 构建查询条件
func getCondition(req *storage.ListAlarmRequest) *operator.Condition {
	var condList []*operator.Condition

	condList = append(condList, getAlarmCondition(req))
	condList = append(condList, getTimeCondition(req))
	condList = append(condList, getTypeCondition(req))

	return operator.NewBranchCondition(operator.And, condList...)
}

// getAlarmCondition 获取alarm条件
func getAlarmCondition(req *storage.ListAlarmRequest) *operator.Condition {
	var condList []*operator.Condition

	params := []string{req.ClusterId, req.Namespace, req.Source, req.Module}

	for i, v := range params {
		if v != "" {
			condList = append(
				condList,
				operator.NewLeafCondition(
					operator.In,
					operator.M{
						alarmConditionTagList[i]: strings.Split(v, ","),
					},
				),
			)
		}
	}

	if len(condList) == 0 {
		return operator.EmptyCondition
	}

	return operator.NewBranchCondition(operator.And, condList...)
}

// getTimeCondition 获取time条件
func getTimeCondition(req *storage.ListAlarmRequest) *operator.Condition {
	var condList []*operator.Condition

	if req.TimeBegin > 0 {
		condList = append(condList, operator.NewLeafCondition(
			operator.Gt, operator.M{constants.ReceivedTimeTag: time.Unix(req.TimeBegin, 0)}))
	}

	if req.TimeEnd > 0 {
		condList = append(condList, operator.NewLeafCondition(
			operator.Lt, operator.M{constants.ReceivedTimeTag: time.Unix(req.TimeEnd, 0)}))
	}

	if len(condList) == 0 {
		return operator.EmptyCondition
	}

	return operator.NewBranchCondition(operator.And, condList...)

}

// getTypeCondition 获取type条件
func getTypeCondition(req *storage.ListAlarmRequest) *operator.Condition {
	if req.Type == "" {
		return operator.EmptyCondition
	}

	var condList []*operator.Condition
	for _, v := range strings.Split(req.Type, ",") {
		condList = append(
			condList,
			operator.NewLeafCondition(
				operator.Con,
				operator.M{
					constants.TypeTag: v,
				},
			),
		)
	}
	return operator.NewBranchCondition(operator.Or, condList...)
}

// HandlerPostAlarm PostAlarm业务方法
func HandlerPostAlarm(ctx context.Context, req *storage.PostAlarmRequest) error {
	data := alarmToM(req)
	opt := &lib.StorePutOption{
		CreateTimeKey: constants.CreateTimeTag,
	}
	return alarms.PutData(ctx, data, opt)
}

// HandlerListAlarm  ListAlarm业务方法
func HandlerListAlarm(ctx context.Context, req *storage.ListAlarmRequest) ([]operator.M, error) {
	opt := &lib.StoreGetOption{
		Offset: int64(req.Offset),
		Limit:  int64(req.Limit),
		Cond:   getCondition(req),
	}
	if len(req.Fields) != 0 {
		opt.Fields = req.Fields
	}

	return alarms.GetData(ctx, opt)
}
