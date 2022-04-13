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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	"github.com/emicklei/go-restful"
	"github.com/micro/go-micro/v2/broker"
)

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

func listEvent(req *restful.Request) ([]operator.M, int, error) {
	clusterID := req.QueryParameter(clusterIDTag)
	if clusterID == "" {
		blog.Errorf("request clusterID is empty")
		return nil, 0, fmt.Errorf("request clusterID is empty")
	}
	blog.Infof("clusterID: %s", clusterID)
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

	getOption := &lib.StoreGetOption{
		Fields: fields,
		Sort: map[string]int{
			eventTimeTag: -1,
		},
		Cond:   condition,
		Offset: offset,
		Limit:  limit,
	}

	eventDBClient := apiserver.GetAPIResource().GetDBClient(dbConfig)

	store := lib.NewStore(
		eventDBClient,
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	var mList []operator.M

	mList, err = store.Get(req.Request.Context(), tablePrefix+clusterID, getOption)
	if err != nil {
		return nil, 0, err
	}
	if int64(len(mList)) < limit {
		getOption.Limit = limit - int64(len(mList))
		tmpList, err := store.Get(req.Request.Context(), tableName, getOption)
		if err != nil {
			return nil, 0, err
		}
		mList = append(mList, tmpList...)
	}

	lib.FormatTime(mList, []string{eventTimeTag})
	return mList, len(mList), nil
}

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

func insert(req *restful.Request) error {
	data, err := getReqData(req)
	if err != nil {
		return err
	}

	putOption := &lib.StorePutOption{}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	data[createTimeTag] = time.Now()
	err = store.Put(req.Request.Context(), tablePrefix+data[clusterIDTag].(string), data, putOption)
	if err != nil {
		return fmt.Errorf("failed to insert, err %s", err.Error())
	}

	queueData := lib.CopyMap(data)
	queueData[resourceTypeTag] = EventResource
	env := typeofToString(queueData[envTag])

	if extra, ok := queueData[extraInfoTag]; ok {
		if d, ok := extra.(types.EventExtraInfo); ok {
			queueData[nameSpaceTag] = d.Namespace
			queueData[resourceNameTag] = d.Name
			queueData[resourceKindTag] =
				func(env string) interface{} {
					switch env {
					case string(types.Event_Env_K8s):
						return data[kindTag]
					case string(types.Event_Env_Mesos):
						return d.Kind
					}

					return ""
				}(env)
		}
	}

	// queueFlag true
	if apiserver.GetAPIResource().GetMsgQueue().QueueFlag {
		err = publishEventResourceToQueue(queueData, eventFeatTags, msgqueue.EventTypeUpdate)
		if err != nil {
			blog.Errorf("publishEventResourceToQueue failed, err %s", err.Error())
		}
	}

	return nil
}

func watch(req *restful.Request, resp *restful.Response) {
	clusterID := req.QueryParameter(clusterIDTag)
	if clusterID == "" {
		blog.Errorf("request clusterID is empty")
		resp.WriteError(http.StatusBadRequest, fmt.Errorf("request clusterID is empty"))
		return
	}
	newWatchOption := &lib.WatchServerOption{
		Store: lib.NewStore(
			apiserver.GetAPIResource().GetDBClient(dbConfig),
			apiserver.GetAPIResource().GetEventBus(dbConfig)),
		TableName: tablePrefix + clusterID,
		Req:       req,
		Resp:      resp,
	}
	ws, err := lib.NewWatchServer(newWatchOption)
	if err != nil {
		blog.Errorf("event get watch server failed, err %s", err.Error())
		resp.Write(lib.EventWatchBreakBytes)
		return
	}

	ws.Go(context.Background())
}

func urlPath(oldURL string) string {
	return oldURL
}

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
