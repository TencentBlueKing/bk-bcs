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
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

func cleanAlarmOutDate(maxDays int64) {
	if maxDays <= 0 {
		blog.Infof("alarm maxDays is %d, alarm day cleaner will not be launched")
		return
	}

	for {
		deadTime := time.Now().Add(time.Duration(-24*maxDays) * time.Hour)
		condition := operator.BaseCondition.AddOp(operator.Lt, createTimeTag, deadTime)

		tank := getNewTank().From(tableName).Filter(condition).RemoveAll()
		if err := tank.GetError(); err == nil {
			blog.Infof(dateCleanOutTitle("Clean the alarms data before %s, total: %d"), deadTime.String(), tank.GetChangeInfo().Removed)
		} else {
			blog.Errorf(dateCleanOutTitle("Clean the alarms data failed. err: %v"), err)
		}
		tank.Close()
		time.Sleep(1 * time.Hour)
	}
}

func cleanAlarmOutCap(maxCaps int) {
	if maxCaps <= 0 {
		blog.Infof("alarm maxCaps is %d, alarm cap cleaner will not be launched")
		return
	}

	left := maxCaps / 3 * 2
	tank := getNewTank().From(tableName)
	event, cancel := tank.Watch(&operator.WatchOptions{})
	defer func() {
		cancel()
		blog.Warnf("cleanAlarmOutCap exist")
	}()

	clusterPool := make(map[string]int)
	reportTick := time.NewTicker(30 * time.Minute)
	defer reportTick.Stop()
	refreshTick := time.NewTicker(24 * time.Hour)
	defer refreshTick.Stop()
	for {
		select {
		case <-reportTick.C:
			status := ""
			i := 0
			for clusterId, count := range clusterPool {
				status += fmt.Sprintf(" %s: %d", clusterId, count)
				if i++; i%5 == 0 {
					status += "\n"
				}
			}
			blog.Infof("alarm cap status\n%s", status)
		case <-refreshTick.C:
			clusterPool = make(map[string]int)
		case e := <-event:
			if e.Type != operator.Add {
				continue
			}
			clusterId, ok := e.Value[clusterIdTag].(string)
			if !ok {
				continue
			}

			var count int
			if count, ok = clusterPool[clusterId]; !ok {
				condition := operator.BaseCondition.AddOp(operator.Eq, clusterIdTag, clusterId)
				t := tank.Filter(condition).Count()
				if err := t.GetError(); err != nil {
					blog.Errorf(capCleanOutTitle("%s | Count alarm failed. err: %v"), clusterId, err)
					continue
				}
				count = t.GetLen()
			} else {
				count++
			}
			clusterPool[clusterId] = count

			if count > maxCaps {
				condition := operator.BaseCondition.AddOp(operator.Eq, clusterIdTag, clusterId)
				t := tank.Filter(condition).OrderBy("-" + createTimeTag).Select(createTimeTag).Limit(left).Query()
				if err := t.GetError(); err != nil {
					blog.Errorf(capCleanOutTitle("%s | Query alarm failed. err: %v"), clusterId, err)
					continue
				}
				r := t.GetValue()
				if len(r) < left {
					continue
				}
				deadTimeStr, ok := r[len(r)-1].(map[string]interface{})[createTimeTag].(string)
				if !ok {
					blog.Errorf(capCleanOutTitle("%s | Convert alarm time failed: %v"), clusterId, r[len(r)-1])
					continue
				}

				deadTime, err := time.Parse(time.RFC3339, deadTimeStr)
				deadTime = deadTime.UTC()
				if err != nil {
					blog.Errorf(capCleanOutTitle("%s | Parse alarm time failed: %v"), clusterId, r[len(r)-1])
					continue
				}
				blog.Infof(capCleanOutTitle("%s | deadTime: %v"), clusterId, deadTime)

				condition = condition.AddOp(operator.Lt, createTimeTag, deadTime)
				t = tank.Filter(condition).RemoveAll()
				if err := t.GetError(); err != nil {
					blog.Errorf(capCleanOutTitle("%s | Remove alarm failed. err: %v"), clusterId, err)
					continue
				}
				clusterPool[clusterId] = count - t.GetChangeInfo().Removed
				blog.Infof(capCleanOutTitle("%s | Remove alarm: %d of %d, %d left"), clusterId, t.GetChangeInfo().Removed, count, clusterPool[clusterId])

				if t.GetChangeInfo().Removed == 0 {
					delete(clusterPool, clusterId)
				}
			}
		}
	}
}

func dateCleanOutTitle(s string) string {
	return cleanOutTitle("date", s)
}

func capCleanOutTitle(s string) string {
	return cleanOutTitle("cap", s)
}

func cleanOutTitle(t, s string) string {
	return fmt.Sprintf("Clean out %s of alarms | %s", t, s)
}

// reqAlarm define a unit for operating alarm data based
// on restful.Request.
type reqAlarm struct {
	req  *restful.Request
	tank operator.Tank

	offset    int
	limit     int
	table     string
	selector  []string
	condition *operator.Condition
	features  operator.M
	data      operator.M
}

// get a new instance of reqAlarm, getNewTank() will be called and
// return a init Tank which is ready for operating
func newReqAlarm(req *restful.Request) *reqAlarm {
	return &reqAlarm{
		req:    req,
		tank:   getNewTank(),
		table:  tableName,
		offset: 0,
		limit:  -1,
	}
}

// reset clean the condition, data etc. so that the reqAlarm can be ready for
// next op.
func (ra *reqAlarm) reset() {
	ra.condition = nil
	ra.selector = nil
	ra.data = nil
}

// alarm data table is "alarm", one big table handle all these.
func (ra *reqAlarm) getTable() string {
	return ra.table
}

// getSelector return a slice of string contains select key for db query.
// ra.selector will be save since first call, so reset() should be called if doing another op.
func (ra *reqAlarm) getSelector() []string {
	if ra.selector == nil {
		s := ra.req.QueryParameter(fieldTag)
		ra.selector = strings.Split(s, ",")
	}
	return ra.selector
}

func (ra *reqAlarm) getOffset() int {
	s := ra.req.QueryParameter(offsetTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		ra.offset = r
	}
	return ra.offset
}

func (ra *reqAlarm) getLimit() int {
	s := ra.req.QueryParameter(limitTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		if r <= 0 {
			r = 1000
		}
		ra.limit = r
	}
	return ra.limit
}

func (ra *reqAlarm) getExtra() (extra operator.M) {
	raw := ra.req.QueryParameter(extraTag)
	if raw == "" {
		return
	}

	lib.NewExtra(raw).Unmarshal(&extra)
	return
}

func (ra *reqAlarm) getFeat() *operator.Condition {
	if ra.condition == nil {
		ra.condition = ra.getCommonFeat().And(ra.getTimeFeat()).And(ra.getTypeFeat())
	}
	return ra.condition
}

func (ra *reqAlarm) getCommonFeat() *operator.Condition {
	r := operator.BaseCondition
	for _, k := range conditionTagList {
		if v := ra.req.QueryParameter(k); v != "" {
			r = r.AddOp(operator.In, k, strings.Split(v, ","))
		}
	}
	return r
}

func (ra *reqAlarm) getTimeFeat() *operator.Condition {
	r := operator.BaseCondition

	if tmp, _ := strconv.ParseInt(ra.req.QueryParameter(timeBeginTag), 10, 64); tmp > 0 {
		r = r.AddOp(operator.Gt, receivedTimeTag, time.Unix(tmp, 0))
	}

	if tmp, _ := strconv.ParseInt(ra.req.QueryParameter(timeEndTag), 10, 64); tmp > 0 {
		r = r.AddOp(operator.Lt, receivedTimeTag, time.Unix(tmp, 0))
	}

	return r
}

func (ra *reqAlarm) getTypeFeat() *operator.Condition {
	r := operator.BaseCondition
	typeParam := ra.req.QueryParameter(typeTag)
	if typeParam == "" {
		return r
	}

	typeList := strings.Split(typeParam, ",")
	for _, t := range typeList {
		cond := operator.BaseCondition.AddOp(operator.Con, typeTag, t)
		r = r.Or(cond)
	}
	return r
}

func (ra *reqAlarm) listAlarm() ([]interface{}, int, error) {
	return ra.get(ra.getFeat())
}

func (ra *reqAlarm) get(condition *operator.Condition) (r []interface{}, total int, err error) {
	tank := ra.tank.From(ra.getTable()).Filter(condition)

	if err = tank.GetError(); err != nil {
		blog.Error("Failed to query. err: %v", err)
		return
	}
	// Get total length
	total = tank.Count().GetLen()

	tank = tank.Offset(ra.getOffset()).Limit(ra.getLimit()).
		Select(ra.getSelector()...).OrderBy("-" + receivedTimeTag).Query()

	if err = tank.GetError(); err != nil {
		blog.Error("Failed to query. err: %v", err)
		return
	}
	r = tank.GetValue()

	// Some time-field need to be format before return
	for i := range r {
		for _, t := range needTimeFormatList {
			tmp, ok := r[i].(map[string]interface{})[t].(time.Time)
			if !ok {
				continue
			}
			r[i].(map[string]interface{})[t] = tmp.Format(timeLayout)
		}
	}
	return
}

func (ra *reqAlarm) getReqData() (operator.M, error) {
	if ra.data == nil {
		var tmp types.BcsStorageAlarmIf
		if err := codec.DecJsonReader(ra.req.Request.Body, &tmp); err != nil {
			return nil, err
		}
		ra.data = operator.M{
			clusterIdTag:    tmp.ClusterId,
			namespaceTag:    tmp.Namespace,
			messageTag:      tmp.Message,
			sourceTag:       tmp.Source,
			moduleTag:       tmp.Module,
			typeTag:         tmp.Type,
			receivedTimeTag: time.Unix(tmp.ReceivedTime, 0),
			dataTag:         tmp.Data,
		}
	}
	return ra.data, nil
}

func (ra *reqAlarm) insert() (err error) {
	tank := ra.tank.From(ra.getTable())

	data, err := ra.getReqData()
	if err != nil {
		return
	}

	data[createTimeTag] = time.Now()
	if err = tank.Insert(data).GetError(); err != nil {
		blog.Errorf("Failed to insert. err: %v", err)
	}
	return
}

// exit() should be called after all ops in reqDynamic to close the connection
// to database.
func (ra *reqAlarm) exit() {
	if ra.tank != nil {
		ra.tank.Close()
	}
}

func urlPath(oldUrl string) string {
	return oldUrl
}
