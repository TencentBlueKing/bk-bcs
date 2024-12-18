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

package events

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/emicklei/go-restful"
	"go-micro.dev/v4/broker"
	mopt "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
)

// getExtra get extra
func getExtra(req *restful.Request) operator.M {
	raw := req.QueryParameter(extraTag)
	if raw == "" {
		return nil
	}

	extra := make(operator.M)
	err := lib.NewExtra(raw).Unmarshal(&extra)
	if err != nil {
		blog.Errorf("decode extra %s failed, err %s", raw, err)
	}
	return extra
}

// getExtraContain get extracontain
func getExtraContain(req *restful.Request) operator.M {
	raw := req.QueryParameter(extraConTag)
	if raw == "" {
		return nil
	}

	extraContain := make(operator.M)
	err := lib.NewExtra(raw).Unmarshal(&extraContain)
	if err != nil {
		blog.Errorf("decode extraContain %s failed, err %s", raw, err)
	}
	return extraContain
}

// getCondition get condition
func getCondition(req *restful.Request) *operator.Condition {
	timeConds := getTimeConds(req)
	commonConds := getCommonConds(req)
	commonConds = append(commonConds, timeConds...)
	var condition *operator.Condition
	if len(commonConds) != 0 {
		condition = operator.NewBranchCondition(operator.And, commonConds...)
	} else {
		condition = operator.EmptyCondition
	}

	// handle the extra field
	var extraConds []*operator.Condition
	extra := getExtra(req)
	features := make(operator.M)
	for k, v := range extra {
		if _, ok := v.([]interface{}); !ok {
			features[k] = []interface{}{v}
			continue
		}
		features[k] = v
	}

	if len(features) > 0 {
		extraConds = append(extraConds, operator.NewLeafCondition(operator.In, features))
	}

	// handle the extra contain field
	extraCon := getExtraContain(req)
	featuresCon := make(operator.M)
	for k, v := range extraCon {
		if _, ok := v.(string); !ok {
			continue
		}
		featuresCon[k] = v.(string)
	}

	if len(featuresCon) > 0 {
		extraConds = append(extraConds, operator.NewLeafCondition(operator.Con, featuresCon))
	}
	if len(extraConds) != 0 {
		condition = operator.NewBranchCondition(operator.And, extraConds...)
	}
	return condition
}

// getCommonConds get common conds
func getCommonConds(req *restful.Request) []*operator.Condition {
	var condList []*operator.Condition
	for _, k := range conditionTagList {
		if v := req.QueryParameter(k); v != "" {
			condList = append(condList, operator.NewLeafCondition(operator.In,
				operator.M{k: strings.Split(v, ",")}))
		}
	}
	return condList
}

// getTimeConds get time conds
func getTimeConds(req *restful.Request) []*operator.Condition {
	var condList []*operator.Condition
	if tmp, _ := strconv.ParseInt(req.QueryParameter(timeBeginTag), 10, 64); tmp > 0 {
		condList = append(condList, operator.NewLeafCondition(operator.Gt, operator.M{
			eventTimeTag: time.Unix(tmp, 0)}))
	}

	if tmp, _ := strconv.ParseInt(req.QueryParameter(timeEndTag), 10, 64); tmp > 0 {
		condList = append(condList, operator.NewLeafCondition(operator.Lt, operator.M{
			eventTimeTag: time.Unix(tmp, 0)}))
	}

	return condList
}

// listEvent list event
func listEvent(req *restful.Request) ([]operator.M, int64, error) {
	clusterIDs := lib.GetQueryParamStringArray(req, clusterIDTag, ",")
	if clusterIDs == nil {
		return nil, 0, fmt.Errorf("clusterID is empty")
	}

	blog.Infof("clusterIDs: %s", clusterIDs)
	fields := lib.GetQueryParamStringArray(req, fieldTag, ",")

	limit, err := lib.GetQueryParamInt64(req, limitTag, 0)
	if err != nil {
		return nil, 0, err
	}

	offset, err := lib.GetQueryParamInt64(req, offsetTag, 0)
	if err != nil {
		return nil, 0, err
	}

	condition := getCondition(req)

	// set read preference, read from secondary node
	secondary := readpref.Secondary()
	dbOpts := mopt.Database().SetReadPreference(secondary)

	// option
	opt := &lib.StoreGetOption{
		Fields: fields,
		Sort: map[string]int{
			eventTimeTag: -1,
		},
		Cond:            condition,
		Offset:          offset,
		Limit:           limit,
		DatabaseOptions: dbOpts,
	}

	return GetEventList(req.Request.Context(), clusterIDs, opt)
}

// get json extra
func getJsonExtra(params map[string]string) operator.M {
	raw := params[extraTag]

	extra := make(operator.M)
	err := lib.NewExtra(raw).Unmarshal(&extra)
	if err != nil {
		blog.Errorf("decode extra %s failed, err %s", raw, err)
	}
	return extra
}

// getJsonExtraContain get json extra contain
func getJsonExtraContain(params map[string]string) operator.M {
	raw := params[extraConTag]

	extraContain := make(operator.M)
	err := lib.NewExtra(raw).Unmarshal(&extraContain)
	if err != nil {
		blog.Errorf("decode extraContain %s failed, err %s", raw, err)
	}
	return extraContain
}

// getJsonCondition get json condition
func getJsonCondition(params map[string]string) *operator.Condition {
	timeConds := getJsonTimeConds(params)
	commonConds := getJsonCommonConds(params)
	commonConds = append(commonConds, timeConds...)
	var condition *operator.Condition
	if len(commonConds) != 0 {
		condition = operator.NewBranchCondition(operator.And, commonConds...)
	} else {
		condition = operator.EmptyCondition
	}

	// handle the extra field
	var extraConds []*operator.Condition
	extra := getJsonExtra(params)
	features := make(operator.M)
	for k, v := range extra {
		if _, ok := v.([]interface{}); !ok {
			features[k] = []interface{}{v}
			continue
		}
		features[k] = v
	}

	if len(features) > 0 {
		extraConds = append(extraConds, operator.NewLeafCondition(operator.In, features))
	}

	// handle the extra contain field
	extraCon := getJsonExtraContain(params)
	featuresCon := make(operator.M)
	for k, v := range extraCon {
		if _, ok := v.(string); !ok {
			continue
		}
		featuresCon[k] = v.(string)
	}

	if len(featuresCon) > 0 {
		extraConds = append(extraConds, operator.NewLeafCondition(operator.Con, featuresCon))
	}
	if len(extraConds) != 0 {
		condition = operator.NewBranchCondition(operator.And, extraConds...)
	}
	return condition
}

// getJsonCommonConds get json common conds
func getJsonCommonConds(params map[string]string) []*operator.Condition {
	var condList []*operator.Condition
	for _, k := range conditionTagList {
		if v := params[k]; v != "" {
			condList = append(condList, operator.NewLeafCondition(operator.In,
				operator.M{k: strings.Split(v, ",")}))
		}
	}
	return condList
}

// getJsonTimeConds get json time conds
func getJsonTimeConds(params map[string]string) []*operator.Condition {
	var condList []*operator.Condition
	if tmp, _ := strconv.ParseInt(params[timeBeginTag], 10, 64); tmp > 0 {
		condList = append(condList, operator.NewLeafCondition(operator.Gt, operator.M{
			eventTimeTag: time.Unix(tmp, 0)}))
	}

	if tmp, _ := strconv.ParseInt(params[timeEndTag], 10, 64); tmp > 0 {
		condList = append(condList, operator.NewLeafCondition(operator.Lt, operator.M{
			eventTimeTag: time.Unix(tmp, 0)}))
	}

	return condList
}

// post event
func postEvent(req *restful.Request) ([]operator.M, int64, error) {
	eventParams := map[string]string{}
	if err := codec.DecJsonReader(req.Request.Body, eventParams); err != nil {
		return nil, 0, err
	}

	clusterIDs := lib.GetJsonParamStringArray(eventParams, clusterIDTag, ",")
	if clusterIDs == nil {
		return nil, 0, fmt.Errorf("clusterID is empty")
	}

	blog.Infof("clusterIDs: %s", clusterIDs)
	fields := lib.GetJsonParamStringArray(eventParams, fieldTag, ",")

	limit, err := lib.GetJsonParamInt64(eventParams, limitTag, 0)
	if err != nil {
		return nil, 0, err
	}

	offset, err := lib.GetJsonParamInt64(eventParams, offsetTag, 0)
	if err != nil {
		return nil, 0, err
	}

	condition := getJsonCondition(eventParams)

	// set read preference, read from secondary node
	secondary := readpref.Secondary()
	dbOpts := mopt.Database().SetReadPreference(secondary)

	// option
	opt := &lib.StoreGetOption{
		Fields: fields,
		Sort: map[string]int{
			eventTimeTag: -1,
		},
		Cond:            condition,
		Offset:          offset,
		Limit:           limit,
		DatabaseOptions: dbOpts,
	}

	return GetEventList(req.Request.Context(), clusterIDs, opt)
}

// getReqData get req data
func getReqData(req *restful.Request) (operator.M, error) {
	var tmp types.BcsStorageEventIf
	if err := codec.DecJsonReader(req.Request.Body, &tmp); err != nil {
		return nil, err
	}
	data := operator.M{
		dataTag:      tmp.Data,
		idTag:        tmp.ID,
		envTag:       tmp.Env,
		kindTag:      tmp.Kind,
		levelTag:     tmp.Level,
		componentTag: tmp.Component,
		typeTag:      tmp.Type,
		describeTag:  tmp.Describe,
		clusterIDTag: tmp.ClusterId,
		extraInfoTag: tmp.ExtraInfo,
		eventTimeTag: time.Unix(tmp.EventTime, 0),
	}
	return data, nil
}

// insert 插入
func insert(req *restful.Request) error {
	// 参数
	data, err := getReqData(req)
	if err != nil {
		return err
	}
	// 表名
	resourceType := TablePrefix + data[clusterIDTag].(string)
	// option
	opt := &lib.StorePutOption{
		UniqueKey: EventIndexKeys,
	}

	return AddEvent(req.Request.Context(), resourceType, data, opt)
}

// watch
func watch(req *restful.Request, resp *restful.Response) {
	clusterID := req.QueryParameter(clusterIDTag)
	if clusterID == "" {
		blog.Errorf("request clusterID is empty")
		_ = resp.WriteError(http.StatusBadRequest, fmt.Errorf("request clusterID is empty"))
		return
	}
	newWatchOption := &lib.WatchServerOption{
		Store:     GetStore(),
		TableName: TablePrefix + clusterID,
		Req:       req,
		Resp:      resp,
	}
	ws, err := lib.NewWatchServer(newWatchOption)
	if err != nil {
		blog.Errorf("event get watch server failed, err %s", err.Error())
		_, _ = resp.Write(lib.EventWatchBreakBytes)
		return
	}

	ws.Go(context.Background())
}

func urlPath(oldURL string) string {
	return oldURL
}

// isExistResourceQueue is exist resource queue
func isExistResourceQueue(features map[string]string) bool {
	if len(features) == 0 {
		return false
	}

	resourceType, ok := features[resourceTypeTag]
	if !ok {
		return false
	}

	if _, ok := apiserver.GetAPIResource().GetMsgQueue().ResourceToQueue[resourceType]; !ok {
		return false
	}

	return true
}

// publishEventResourceToQueue publish event resource to queue
func publishEventResourceToQueue(data operator.M, featTags []string, event msgqueue.EventKind) error {
	var (
		err     error
		message = &broker.Message{
			Header: map[string]string{},
		}
	)

	startTime := time.Now()
	for _, feat := range featTags {
		if v, ok := data[feat]; ok {
			message.Header[feat] = typeofToString(v)
		}
	}
	message.Header[string(msgqueue.EventType)] = string(event)

	exist := isExistResourceQueue(message.Header)
	if !exist {
		return nil
	}

	// NOCC:revive/early-return(设计如此:)
	// nolint
	if v, ok := data[dataTag]; ok {
		codec.EncJson(v, &message.Body)
	} else {
		blog.Infof("object[%v] not exist data", data[dataTag])
		return nil
	}

	err = apiserver.GetAPIResource().GetMsgQueue().MsgQueue.Publish(message)
	if err != nil {
		return err
	}

	if queueName, ok := message.Header[resourceTypeTag]; ok {
		metrics.ReportQueuePushMetrics(queueName, err, startTime)
	}

	return nil
}

// typeofToString convert type of to string
func typeofToString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return v.(string)
	case types.EventEnv:
		return string(v.(types.EventEnv))
	case types.EventKind:
		return string(v.(types.EventKind))
	case types.EventComponent:
		return string(v.(types.EventComponent))
	case types.EventLevel:
		return string(v.(types.EventLevel))
	case types.ExtraKind:
		return string(v.(types.ExtraKind))
	default:
		_ = t
		return ""
	}
}
