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

package dynamic

import (
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

// reqDynamic define a unit for operating dynamic data based
// on restful.Request.
type reqDynamic struct {
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

// get a new instance of reqDynamic, getNewTank() will be called and
// return a init Tank which is ready for operating
func newReqDynamic(req *restful.Request) *reqDynamic {
	return &reqDynamic{
		req:    req,
		tank:   getNewTank(),
		offset: 0,
		limit:  -1,
	}
}

// reset clean the condition, data etc. so that the reqDynamic can be ready for
// next op.
func (rd *reqDynamic) reset() {
	rd.condition = nil
	rd.selector = nil
	rd.data = nil
	rd.table = ""
}

// dynamic data table is like: "clusterId_Type", for instance: "BCS-K8S-10001_Deployment".
// rd.table will be save since first call, so reset() should be called if doing another op.
func (rd *reqDynamic) getTable() string {
	if rd.table == "" {

		// process stored with application in application-table
		tablePart := rd.req.PathParameter(tableTag)
		if tablePart == processTypeName {
			tablePart = applicationTypeName
		}
		rd.table = rd.req.PathParameter(clusterIdTag) + "_" + tablePart
	}
	return rd.table
}

// getSelector return a slice of string contains select key for db query.
// rd.selector will be save since first call, so reset() should be called if doing another op.
func (rd *reqDynamic) getSelector() []string {
	if rd.selector == nil {
		s := rd.req.QueryParameter(fieldTag)
		rd.selector = strings.Split(s, ",")
	}
	return rd.selector
}

func (rd *reqDynamic) getOffset() int {
	s := rd.req.QueryParameter(offsetTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		rd.offset = r
	}
	return rd.offset
}

func (rd *reqDynamic) getLimit() int {
	s := rd.req.QueryParameter(limitTag)
	r, err := strconv.Atoi(s)
	if err == nil {
		rd.limit = r
	}
	return rd.limit
}

func (rd *reqDynamic) getExtra() (extra operator.M) {
	raw := rd.req.QueryParameter(extraTag)
	if raw == "" {
		return
	}

	lib.NewExtra(raw).Unmarshal(&extra)
	return
}

func (rd *reqDynamic) getNsFeat() *operator.Condition {
	return rd.getFeat(nsFeatTags)
}

func (rd *reqDynamic) getCsFeat() *operator.Condition {
	return rd.getFeat(csFeatTags)
}

func (rd *reqDynamic) getNsListFeat() *operator.Condition {
	return rd.getFeat(nsListFeatTags)
}

func (rd *reqDynamic) getCsListFeat() *operator.Condition {
	return rd.getFeat(csListFeatTags)
}

func (rd *reqDynamic) getNsRemoveFeat() *operator.Condition {
	if rd.condition == nil {
		condition := rd.getFeat(nsListFeatTags)
		if timeFeat := rd.getTimeFeat(); timeFeat != operator.BaseCondition {
			condition = condition.And(timeFeat)
		}
		rd.condition = condition
	}
	return rd.condition
}

func (rd *reqDynamic) getCsRemoveFeat() *operator.Condition {
	if rd.condition == nil {
		condition := rd.getFeat(csListFeatTags)
		if timeFeat := rd.getTimeFeat(); timeFeat != operator.BaseCondition {
			condition = condition.And(timeFeat)
		}
		rd.condition = condition
	}
	return rd.condition
}

// timeFeat provide a Condition that updateTimeTag should be
// between data.UpdateTimeBegin and data.UpdateTimeEnd
func (rd *reqDynamic) getTimeFeat() *operator.Condition {
	var data types.BcsStorageDynamicBatchDeleteIf
	if err := codec.DecJsonReader(rd.req.Request.Body, &data); err != nil {
		return operator.BaseCondition
	}

	r := operator.BaseCondition

	if data.UpdateTimeBegin > 0 {
		r = r.AddOp(operator.Gt, updateTimeTag, time.Unix(data.UpdateTimeBegin, 0))
	}
	if data.UpdateTimeEnd > 0 {
		r = r.AddOp(operator.Lt, updateTimeTag, time.Unix(data.UpdateTimeEnd, 0))
	}
	return r
}

func (rd *reqDynamic) getFeat(resourceFeatList []string) *operator.Condition {
	if rd.condition == nil {
		features := make(operator.M)
		featuresExcept := make(operator.M)
		for _, key := range resourceFeatList {
			features[key] = rd.req.PathParameter(key)

			// For historical reasons, mesos process is stored with application in one table(same clusters).
			// And process's construction is almost the same with application, except with field 'data.kind'.
			// If 'data.kind'='process', then this object is a process stored in application-table,
			// If 'data.kind'='application' or '', then this object is an application stored in application-table.
			//
			// For this case, we should:
			// 1. Change the key 'resourceType' from 'process' to 'application' when the caller ask for 'process'.
			// 2. Besides, getFeat() should add an extra condition that mentions the 'data.kind' to distinguish 'process' and 'application'.
			// 3. Make sure the table is application-table whether the type is 'application' or 'process'. (with getTable())
			if key == resourceTypeTag {
				switch features[key] {
				case applicationTypeName:
					featuresExcept[kindTag] = processTypeName
				case processTypeName:
					features[key] = applicationTypeName
					features[kindTag] = processTypeName
				}
			}
		}
		rd.features = features

		// handle the extra field
		extra := rd.getExtra()
		for k, v := range extra {
			features[k] = v
		}

		rd.condition = operator.NewCondition(operator.Eq, features)
		if len(featuresExcept) > 0 {
			rd.condition.And(operator.NewCondition(operator.Ne, featuresExcept))
		}
	}
	return rd.condition
}

func (rd *reqDynamic) getReqData() (operator.M, error) {
	if rd.data == nil {
		var tmp types.BcsStorageDynamicIf
		if err := codec.DecJsonReader(rd.req.Request.Body, &tmp); err != nil {
			return nil, err
		}
		data := lib.CopyMap(rd.features)
		data[dataTag] = tmp.Data
		rd.data = data
	}
	return rd.data, nil
}

func (rd *reqDynamic) nsGet() ([]interface{}, error) {
	return rd.get(rd.getNsFeat())
}

func (rd *reqDynamic) csGet() ([]interface{}, error) {
	return rd.get(rd.getCsFeat())
}

func (rd *reqDynamic) nsList() ([]interface{}, error) {
	return rd.get(rd.getNsListFeat())
}

func (rd *reqDynamic) csList() ([]interface{}, error) {
	return rd.get(rd.getCsListFeat())
}

func (rd *reqDynamic) get(condition *operator.Condition) (r []interface{}, err error) {
	tank := rd.tank.From(rd.getTable()).Filter(condition).
		Offset(rd.getOffset()).Limit(rd.getLimit()).Select(rd.getSelector()...).Query()

	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to query. err: %v", err)
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

func (rd *reqDynamic) nsPut() error {
	return rd.put(rd.getNsFeat())
}

func (rd *reqDynamic) csPut() error {
	return rd.put(rd.getCsFeat())
}

// put try update first, if target is no found the try insert.
func (rd *reqDynamic) put(condition *operator.Condition) (err error) {
	tank := rd.tank.From(rd.getTable()).Filter(condition)

	data, err := rd.getReqData()
	if err != nil {
		return
	}

	// Update or insert
	timeNow := time.Now()

	queryTank := tank.Query()
	if err = queryTank.GetError(); err != nil {
		blog.Errorf("Failed to check if resource exist. err: %v", err)
		return
	}
	if queryTank.GetLen() == 0 {
		data.Update(createTimeTag, timeNow)
	}

	tank = tank.Index(indexKeys...).Upsert(data.Update(updateTimeTag, timeNow))
	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to update. err: %v", err)
		return
	}
	return
}

func (rd *reqDynamic) nsRemove() error {
	return rd.remove(rd.getNsFeat(), true)
}

func (rd *reqDynamic) csRemove() error {
	return rd.remove(rd.getCsFeat(), true)
}

func (rd *reqDynamic) nsBatchRemove() error {
	return rd.remove(rd.getNsRemoveFeat(), false)
}

func (rd *reqDynamic) csBatchRemove() error {
	return rd.remove(rd.getCsRemoveFeat(), false)
}

func (rd *reqDynamic) remove(condition *operator.Condition, mustMatch bool) (err error) {
	tank := rd.tank.From(rd.getTable()).Filter(condition).RemoveAll()

	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to remove. err: %v", err)
		return
	}

	changeInfo := tank.GetChangeInfo()
	if changeInfo.Removed != changeInfo.Matched {
		return storageErr.RemoveLessThanMatch
	}

	if mustMatch && changeInfo.Matched == 0 {
		return storageErr.ResourceDoesNotExist
	}
	return
}

// exit() should be called after all ops in reqDynamic to close the connection
// to database.
func (rd *reqDynamic) exit() {
	if rd.tank != nil {
		rd.tank.Close()
	}
}

func urlPath(oldUrl string) string {
	return urlPrefix + oldUrl
}
