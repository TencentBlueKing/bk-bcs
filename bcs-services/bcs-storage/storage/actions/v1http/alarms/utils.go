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

package alarms

import (
	"strings"
	"time"

	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"
)

// func cleanAlarmOutDate(maxDays int64) {
// 	if maxDays <= 0 {
// 		blog.Infof("alarm maxDays is %d, alarm day cleaner will not be launched")
// 		return
// 	}

// 	for {
// 		deadTime := time.Now().Add(time.Duration(-24*maxDays) * time.Hour)
// 		store := lib.NewStore(apiserver.GetAPIResource().GetDBClient(dbConfig))
// 		condition := operator.NewLeafCondition(operator.Lt, operator.M{createTimeTag: deadTime})
// 		if err := store.Remove(context.TODO(), tableName, &lib.StoreRemoveOption{
// 			IgnoreNotFound: true,
// 			Cond: condition,
// 		}); err == nil {
// 			blog.Infof(dateCleanOutTitle(
// 				"Clean the alarms data before %s, total: %d"),
// 				deadTime.String(), tank.GetChangeInfo().Removed)
// 		} else {
// 			blog.Errorf(dateCleanOutTitle("Clean the alarms data failed. err: %v"), err)
// 		}
// 		tank.Close()
// 		time.Sleep(1 * time.Hour)
// 	}
// }

// func cleanAlarmOutCap(maxCaps int) {
// 	if maxCaps <= 0 {
// 		blog.Infof("alarm maxCaps is %d, alarm cap cleaner will not be launched")
// 		return
// 	}

// 	left := maxCaps / 3 * 2
// 	tank := getNewTank().From(tableName)
// 	event, cancel := tank.Watch(&operator.WatchOptions{})
// 	defer func() {
// 		cancel()
// 		blog.Warnf("cleanAlarmOutCap exist")
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
// 			blog.Infof("alarm cap status\n%s", status)
// 		case <-refreshTick.C:
// 			clusterPool = make(map[string]int)
// 		case e := <-event:
// 			if e.Type != operator.Add {
// 				continue
// 			}
// 			clusterId, ok := e.Value[clusterIdTag].(string)
// 			if !ok {
// 				continue
// 			}

// 			var count int
// 			if count, ok = clusterPool[clusterId]; !ok {
// 				condition := operator.BaseCondition.AddOp(operator.Eq, clusterIdTag, clusterId)
// 				t := tank.Filter(condition).Count()
// 				if err := t.GetError(); err != nil {
// 					blog.Errorf(capCleanOutTitle("%s | Count alarm failed. err: %v"), clusterId, err)
// 					continue
// 				}
// 				count = t.GetLen()
// 			} else {
// 				count++
// 			}
// 			clusterPool[clusterId] = count

// 			if count > maxCaps {
// 				condition := operator.BaseCondition.AddOp(operator.Eq, clusterIdTag, clusterId)
// 				t := tank.Filter(condition).OrderBy("-" + createTimeTag).Select(createTimeTag).Limit(left).Query()
// 				if err := t.GetError(); err != nil {
// 					blog.Errorf(capCleanOutTitle("%s | Query alarm failed. err: %v"), clusterId, err)
// 					continue
// 				}
// 				r := t.GetValue()
// 				if len(r) < left {
// 					continue
// 				}
// 				deadTimeStr, ok := r[len(r)-1].(map[string]interface{})[createTimeTag].(string)
// 				if !ok {
// 					blog.Errorf(capCleanOutTitle("%s | Convert alarm time failed: %v"), clusterId, r[len(r)-1])
// 					continue
// 				}

// 				deadTime, err := time.Parse(time.RFC3339, deadTimeStr)
// 				deadTime = deadTime.UTC()
// 				if err != nil {
// 					blog.Errorf(capCleanOutTitle("%s | Parse alarm time failed: %v"), clusterId, r[len(r)-1])
// 					continue
// 				}
// 				blog.Infof(capCleanOutTitle("%s | deadTime: %v"), clusterId, deadTime)

// 				condition = condition.AddOp(operator.Lt, createTimeTag, deadTime)
// 				t = tank.Filter(condition).RemoveAll()
// 				if err := t.GetError(); err != nil {
// 					blog.Errorf(capCleanOutTitle("%s | Remove alarm failed. err: %v"), clusterId, err)
// 					continue
// 				}
// 				clusterPool[clusterId] = count - t.GetChangeInfo().Removed
// 				blog.Infof(capCleanOutTitle("%s | Remove alarm: %d of %d, %d left"), clusterId, t.GetChangeInfo().Removed, count, clusterPool[clusterId])

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
// 	return fmt.Sprintf("Clean out %s of alarms | %s", t, s)
// }

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
