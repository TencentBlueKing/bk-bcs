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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

// Clean the data out of date
func cleanEventOutDate(maxDays int64) {
	if maxDays <= 0 {
		blog.Infof("event maxDays is %d, event day cleaner will not be launched")
		return
	}

	for {
		deadTime := time.Now().Add(time.Duration(-24*maxDays) * time.Hour)
		condition := operator.BaseCondition.AddOp(operator.Lt, createTimeTag, deadTime)

		tank := getNewTank().From(tableName).Filter(condition).RemoveAll()
		if err := tank.GetError(); err == nil {
			blog.Infof(dateCleanOutTitle("Clean the events data before %s, total: %d"), deadTime.String(), tank.GetChangeInfo().Removed)
		} else {
			blog.Errorf(dateCleanOutTitle("Clean the events data failed. err: %v"), err)
		}
		tank.Close()
		time.Sleep(1 * time.Hour)
	}
}

// Clean the over flow data for each cluster
func cleanEventOutCap(maxCaps int) {
	if maxCaps <= 0 {
		blog.Infof("event maxCaps is %d, event cap cleaner will not be launched")
		return
	}

	left := maxCaps / 3 * 2
	tank := getNewTank().From(tableName)
	event, cancel := tank.Watch(&operator.WatchOptions{})
	defer func() {
		cancel()
		blog.Warnf("cleanEventOutCap exist")
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
			blog.Infof("event cap status\n%s", status)
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
					blog.Errorf(capCleanOutTitle("%s | Count event failed. err: %v"), clusterId, err)
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
					blog.Errorf(capCleanOutTitle("%s | Query event failed. err: %v"), clusterId, err)
					continue
				}
				r := t.GetValue()
				if len(r) < left {
					continue
				}
				deadTimeStr, ok := r[len(r)-1].(map[string]interface{})[createTimeTag].(string)
				if !ok {
					blog.Errorf(capCleanOutTitle("%s | Convert event time failed: %v"), clusterId, r[len(r)-1])
					continue
				}

				deadTime, err := time.Parse(time.RFC3339, deadTimeStr)
				deadTime = deadTime.UTC()
				if err != nil {
					blog.Errorf(capCleanOutTitle("%s | Parse event time failed: %v"), clusterId, r[len(r)-1])
					continue
				}
				blog.Infof(capCleanOutTitle("%s | deadTime: %v"), clusterId, deadTime)

				condition = condition.AddOp(operator.Lt, createTimeTag, deadTime)
				t = tank.Filter(condition).RemoveAll()
				if err := t.GetError(); err != nil {
					blog.Errorf(capCleanOutTitle("%s | Remove event failed. err: %v"), clusterId, err)
					continue
				}
				clusterPool[clusterId] = count - t.GetChangeInfo().Removed
				blog.Infof(capCleanOutTitle("%s | Remove event: %d of %d, %d left"), clusterId, t.GetChangeInfo().Removed, count, clusterPool[clusterId])

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
	return fmt.Sprintf("Clean out %s of events | %s", t, s)
}

// reqEvent define a unit for operating event data based
// on restful.Request.
type reqEvent struct {
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

// get a new instance of reqEvent, getNewTank() will be called and
// return a init Tank which is ready for operating
func newReqEvent(req *restful.Request) *reqEvent {
	return &reqEvent{
		req:    req,
		tank:   getNewTank(),
		table:  tableName,
		offset: 0,
		limit:  -1,
	}
}

// reset clean the condition, data etc. so that the reqEvent can be ready for
// next op.
func (re *reqEvent) reset() {
	re.condition = nil
	re.selector = nil
	re.data = nil
}

// event data table is "event", one big table handle all these.
func (re *reqEvent) getTable() string {
	return re.table
}

// getSelector return a slice of string contains select key for db query.
// re.selector will be save since first call, so reset() should be called if doing another op.
func (re *reqEvent) getSelector() []string {
	if re.selector == nil {
		s := re.req.QueryParameter(fieldTag)
		re.selector = strings.Split(s, ",")
	}
	return re.selector
}

func (re *reqEvent) getOffset() int {
	s := re.req.QueryParameter(offsetTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		re.offset = r
	}
	return re.offset
}

func (re *reqEvent) getLimit() int {
	s := re.req.QueryParameter(limitTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		if r <= 0 {
			r = 1000
		}
		re.limit = r
	}
	return re.limit
}

func (re *reqEvent) getExtra() (extra operator.M) {
	raw := re.req.QueryParameter(extraTag)
	if raw == "" {
		return
	}

	err := lib.NewExtra(raw).Unmarshal(&extra)
	blog.Errorf("err: %v", err)
	return
}

func (re *reqEvent) getExtraContain() (extraContain operator.M) {
	raw := re.req.QueryParameter(extraConTag)
	if raw == "" {
		return
	}

	_ = lib.NewExtra(raw).Unmarshal(&extraContain)
	return
}

func (re *reqEvent) getFeat() *operator.Condition {
	if re.condition == nil {
		r := re.getCommonFeat().And(re.getTimeFeat())

		// handle the extra field
		extra := re.getExtra()
		features := make(operator.M)
		for k, v := range extra {
			if _, ok := v.([]interface{}); !ok {
				features[k] = []interface{}{v}
				continue
			}
			features[k] = v
		}

		if len(features) > 0 {
			r = r.And(operator.NewCondition(operator.In, features))
		}

		// handle the extra contain field
		extraCon := re.getExtraContain()
		featuresCon := make(operator.M)
		for k, v := range extraCon {
			if _, ok := v.(string); !ok {
				continue
			}
			featuresCon[k] = v.(string)
		}

		if len(featuresCon) > 0 {
			r = r.And(operator.NewCondition(operator.Con, featuresCon))
		}

		re.condition = r
	}
	return re.condition
}

func (re *reqEvent) getCommonFeat() *operator.Condition {
	r := operator.BaseCondition
	for _, k := range conditionTagList {
		if v := re.req.QueryParameter(k); v != "" {
			r = r.AddOp(operator.In, k, strings.Split(v, ","))
		}
	}
	return r
}

func (re *reqEvent) getTimeFeat() *operator.Condition {
	r := operator.BaseCondition

	if tmp, _ := strconv.ParseInt(re.req.QueryParameter(timeBeginTag), 10, 64); tmp > 0 {
		r = r.AddOp(operator.Gt, eventTimeTag, time.Unix(tmp, 0))
	}

	if tmp, _ := strconv.ParseInt(re.req.QueryParameter(timeEndTag), 10, 64); tmp > 0 {
		r = r.AddOp(operator.Lt, eventTimeTag, time.Unix(tmp, 0))
	}

	return r
}

func (re *reqEvent) listEvent() ([]interface{}, int, error) {
	return re.get(re.getFeat())
}

func (re *reqEvent) get(condition *operator.Condition) (r []interface{}, total int, err error) {
	tank := re.tank.From(re.getTable()).Filter(condition)

	if err = tank.GetError(); err != nil {
		blog.Error("Failed to query. err: %v", err)
		return
	}
	// Get total length
	total = tank.Count().GetLen()

	tank = tank.Offset(re.getOffset()).Limit(re.getLimit()).
		Select(re.getSelector()...).OrderBy("-" + eventTimeTag).Query()

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

func (re *reqEvent) getReqData() (operator.M, error) {
	if re.data == nil {
		var tmp types.BcsStorageEventIf
		if err := codec.DecJsonReader(re.req.Request.Body, &tmp); err != nil {
			return nil, err
		}
		re.data = operator.M{
			dataTag:      tmp.Data,
			idTag:        tmp.ID,
			envTag:       tmp.Env,
			kindTag:      tmp.Kind,
			levelTag:     tmp.Level,
			componentTag: tmp.Component,
			typeTag:      tmp.Type,
			describeTag:  tmp.Describe,
			clusterIdTag: tmp.ClusterId,
			extraInfoTag: tmp.ExtraInfo,
			eventTimeTag: time.Unix(tmp.EventTime, 0),
		}
	}
	return re.data, nil
}

func (re *reqEvent) insert() (err error) {
	tank := re.tank.From(re.getTable())

	data, err := re.getReqData()
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
func (re *reqEvent) exit() {
	if re.tank != nil {
		re.tank.Close()
	}
}

func urlPath(oldUrl string) string {
	return oldUrl
}
