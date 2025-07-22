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

package alarms

import (
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
)

func getStoreGetOption(req *restful.Request) (*lib.StoreGetOption, error) {
	fields := lib.GetQueryParamStringArray(req, fieldTag, ",")
	offset, err := lib.GetQueryParamInt64(req, offsetTag, 0)
	if err != nil {
		return nil, err
	}
	limit, err := lib.GetQueryParamInt64(req, limitTag, 0)
	if err != nil {
		return nil, err
	}
	cond, err := getCondition(req)
	if err != nil {
		return nil, err
	}
	newOpt := &lib.StoreGetOption{
		Offset: offset,
		Limit:  limit,
		Cond:   cond,
	}
	if len(fields) != 0 {
		newOpt.Fields = fields
	}

	return newOpt, nil
}

func getCondition(req *restful.Request) (*operator.Condition, error) {
	var condList []*operator.Condition
	condList = append(condList, getCommonFeat(req))
	timeCond, err := getTimeFeat(req)
	if err != nil {
		return nil, err
	}
	condList = append(condList, timeCond)
	condList = append(condList, getTypeFeat(req))
	return operator.NewBranchCondition(operator.And, condList...), nil
}

func getCommonFeat(req *restful.Request) *operator.Condition {
	var condList []*operator.Condition
	for _, k := range conditionTagList {
		if v := req.QueryParameter(k); v != "" {
			condList = append(condList,
				operator.NewLeafCondition(operator.In, operator.M{k: strings.Split(v, ",")}))
		}
	}
	if len(condList) == 0 {
		return operator.EmptyCondition
	}
	return operator.NewBranchCondition(operator.And, condList...)
}

func getTimeFeat(req *restful.Request) (*operator.Condition, error) {
	var condList []*operator.Condition
	timeBegin, err := lib.GetQueryParamInt64(req, timeBeginTag, 0)
	if err != nil {
		return nil, err
	}
	if timeBegin > 0 {
		condList = append(condList, operator.NewLeafCondition(
			operator.Gt, operator.M{receivedTimeTag: time.Unix(timeBegin, 0)}))
	}

	timeEnd, err := lib.GetQueryParamInt64(req, timeEndTag, 0)
	if err != nil {
		return nil, err
	}
	if timeEnd > 0 {
		condList = append(condList, operator.NewLeafCondition(
			operator.Lt, operator.M{receivedTimeTag: time.Unix(timeEnd, 0)}))
	}
	if len(condList) == 0 {
		return operator.EmptyCondition, nil
	}
	return operator.NewBranchCondition(operator.And, condList...), nil
}

func getTypeFeat(req *restful.Request) *operator.Condition {
	typeParam := req.QueryParameter(typeTag)
	if typeParam == "" {
		return operator.EmptyCondition
	}

	var condList []*operator.Condition
	typeList := strings.Split(typeParam, ",")
	for _, t := range typeList {
		condList = append(condList, operator.NewLeafCondition(operator.Con, operator.M{typeTag: t}))
	}
	cond := operator.NewBranchCondition(operator.Or, condList...)
	return cond
}

func getReqData(req *restful.Request) (operator.M, error) {
	var tmp types.BcsStorageAlarmIf
	if err := codec.DecJsonReader(req.Request.Body, &tmp); err != nil {
		return nil, err
	}
	return operator.M{
		clusterIDTag:    tmp.ClusterId,
		namespaceTag:    tmp.Namespace,
		messageTag:      tmp.Message,
		sourceTag:       tmp.Source,
		moduleTag:       tmp.Module,
		typeTag:         tmp.Type,
		receivedTimeTag: time.Unix(tmp.ReceivedTime, 0),
		dataTag:         tmp.Data,
	}, nil
}
