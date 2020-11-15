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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

// // Clean the data out of date
// func cleanEventOutDate(maxDays int64) {
// 	if maxDays <= 0 {
// 		blog.Infof("event maxDays is %d, event day cleaner will not be launched")
// 		return
// 	}

// 	for {
// 		deadTime := time.Now().Add(time.Duration(-24*maxDays) * time.Hour)
// 		condition := operator.BaseCondition.AddOp(operator.Lt, createTimeTag, deadTime)

// 		tank := getNewTank().From(tableName).Filter(condition).RemoveAll()
// 		if err := tank.GetError(); err == nil {
// 			blog.Infof(dateCleanOutTitle("Clean the events data before %s, total: %d"), deadTime.String(), tank.GetChangeInfo().Removed)
// 		} else {
// 			blog.Errorf(dateCleanOutTitle("Clean the events data failed. err: %v"), err)
// 		}
// 		tank.Close()
// 		time.Sleep(1 * time.Hour)
// 	}
// }

// // Clean the over flow data for each cluster
// func cleanEventOutCap(maxCaps int) {
// 	if maxCaps <= 0 {
// 		blog.Infof("event maxCaps is %d, event cap cleaner will not be launched")
// 		return
// 	}

// 	left := maxCaps / 3 * 2
// 	tank := getNewTank().From(tableName)
// 	event, cancel := tank.Watch(&operator.WatchOptions{})
// 	defer func() {
// 		cancel()
// 		blog.Warnf("cleanEventOutCap exist")
// 	}()

// 	clusterPool := make(map[string]int)
// 	reportTick := time.NewTicker(30 * time.Minute)
// 	refreshTick := time.NewTicker(24 * time.Hour)
// 	for {
// 		select {
// 		case <-reportTick.C:
// 			status := ""
// 			i := 0
// 			for clusterId, count := range clusterPool {
// 				status += fmt.Sprintf(" %s: %d", clusterId, count)
// 				if i++; i%5 == 0 {
// 					status += "\n"
// 				}
// 			}
// 			blog.Infof("event cap status\n%s", status)
// 		case <-refreshTick.C:
// 			clusterPool = make(map[string]int)
// 		case e := <-event:
// 			if e.Type != operator.Add {
// 				continue
// 			}
// 			clusterId, ok := e.Value[clusterIDTag].(string)
// 			if !ok {
// 				continue
// 			}

// 			var count int
// 			if count, ok = clusterPool[clusterId]; !ok {
// 				condition := operator.BaseCondition.AddOp(operator.Eq, clusterIDTag, clusterId)
// 				t := tank.Filter(condition).Count()
// 				if err := t.GetError(); err != nil {
// 					blog.Errorf(capCleanOutTitle("%s | Count event failed. err: %v"), clusterId, err)
// 					continue
// 				}
// 				count = t.GetLen()
// 			} else {
// 				count++
// 			}
// 			clusterPool[clusterId] = count

// 			if count > maxCaps {
// 				condition := operator.BaseCondition.AddOp(operator.Eq, clusterIDTag, clusterId)
// 				t := tank.Filter(condition).OrderBy("-" + createTimeTag).Select(createTimeTag).Limit(left).Query()
// 				if err := t.GetError(); err != nil {
// 					blog.Errorf(capCleanOutTitle("%s | Query event failed. err: %v"), clusterId, err)
// 					continue
// 				}
// 				r := t.GetValue()
// 				if len(r) < left {
// 					continue
// 				}
// 				deadTimeStr, ok := r[len(r)-1].(map[string]interface{})[createTimeTag].(string)
// 				if !ok {
// 					blog.Errorf(capCleanOutTitle("%s | Convert event time failed: %v"), clusterId, r[len(r)-1])
// 					continue
// 				}

// 				deadTime, err := time.Parse(time.RFC3339, deadTimeStr)
// 				deadTime = deadTime.UTC()
// 				if err != nil {
// 					blog.Errorf(capCleanOutTitle("%s | Parse event time failed: %v"), clusterId, r[len(r)-1])
// 					continue
// 				}
// 				blog.Infof(capCleanOutTitle("%s | deadTime: %v"), clusterId, deadTime)

// 				condition = condition.AddOp(operator.Lt, createTimeTag, deadTime)
// 				t = tank.Filter(condition).RemoveAll()
// 				if err := t.GetError(); err != nil {
// 					blog.Errorf(capCleanOutTitle("%s | Remove event failed. err: %v"), clusterId, err)
// 					continue
// 				}
// 				clusterPool[clusterId] = count - t.GetChangeInfo().Removed
// 				blog.Infof(capCleanOutTitle("%s | Remove event: %d of %d, %d left"), clusterId, t.GetChangeInfo().Removed, count, clusterPool[clusterId])

// 				if t.GetChangeInfo().Removed == 0 {
// 					delete(clusterPool, clusterId)
// 				}
// 			}
// 		}
// 	}
// }

// func dateCleanOutTitle(s string) string {
// 	return cleanOutTitle("date", s)
// }

// func capCleanOutTitle(s string) string {
// 	return cleanOutTitle("cap", s)
// }

// func cleanOutTitle(t, s string) string {
// 	return fmt.Sprintf("Clean out %s of events | %s", t, s)
// }

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
	condition := operator.NewBranchCondition(operator.And, commonConds...)

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

	condition = operator.NewBranchCondition(operator.And, extraConds...)
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

func listEvent(req *restful.Request) ([]interface{}, int, error) {
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

	store := lib.NewStore(apiserver.GetAPIResource().GetDBClient(dbConfig))
	mList, err := store.Get(req.Request.Context(), tableName, getOption)
	if err != nil {
		return nil, 0, err
	}

	return []interface{}{mList}, len(mList), nil
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
	store := lib.NewStore(apiserver.GetAPIResource().GetDBClient(dbConfig))
	data[createTimeTag] = time.Now()
	err = store.Put(req.Request.Context(), tableName, data, putOption)
	if err != nil {
		return fmt.Errorf("failed to insert, err %s", err.Error())
	}
	return nil
}

func urlPath(oldURL string) string {
	return oldURL
}
