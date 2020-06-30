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

package hostConfig

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	storageErr "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/errors"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	restful "github.com/emicklei/go-restful"
)

// reqHost define a unit for operating host data based
// on restful.Request.
// host data manages cluster info for cluster-keeper since BCS deployed
// in a non-PaaS env such as AWS.
type reqHost struct {
	req  *restful.Request
	tank operator.Tank

	condition *operator.Condition
	features  operator.M
	data      operator.M
	table     string
	ips       []string
	clusterId string
}

// get a new instance of reqHost, getNewTank() will be called and
// return a init Tank which is ready for operating
func newReqHost(req *restful.Request) *reqHost {
	return &reqHost{
		req:   req,
		tank:  getNewTank(),
		table: tableName,
	}
}

// reset clean the condition, data etc. so that the reqDynamic can be ready for
// next op.
func (rh *reqHost) reset() {
	rh.condition = nil
	rh.features = nil
	rh.data = nil
	rh.ips = nil
}

// host data table is "host"
func (rh *reqHost) getTable() string {
	return rh.table
}

func (rh *reqHost) getHostFeat() *operator.Condition {
	return rh.getFeat(hostFeatTags)
}

func (rh *reqHost) getQueryHostFeat() *operator.Condition {
	return rh.getFeat(hostQueryFeatTags)
}

func (rh *reqHost) getFeat(resourceFeatList []string) *operator.Condition {
	if rh.condition == nil {
		features := make(operator.M, len(resourceFeatList))
		for _, key := range resourceFeatList {
			features[key] = rh.req.PathParameter(key)
		}
		rh.features = features
		rh.condition = operator.NewCondition(operator.Eq, features)
	}
	return rh.condition
}

func (rh *reqHost) getRelationFeat() (*operator.Condition, error) {
	if rh.condition == nil {
		var tmp *types.BcsStorageClusterRelationIf
		if err := codec.DecJsonReader(rh.req.Request.Body, &tmp); err != nil {
			return nil, err
		}

		rh.ips = tmp.Ips
		rh.condition = operator.BaseCondition.AddOp(operator.In, ipTag, tmp.Ips)
	}
	return rh.condition, nil
}

func (rh *reqHost) getReqData() (operator.M, error) {
	if rh.data == nil {
		var tmp types.BcsStorageHostIf
		if err := codec.DecJsonReader(rh.req.Request.Body, &tmp); err != nil {
			return nil, err
		}
		data := lib.CopyMap(rh.features)
		data[clusterIdTag] = tmp.ClusterId
		data[dataTag] = tmp.Data
		rh.data = data
	}
	return rh.data, nil
}

func (rh *reqHost) getRelationData() operator.M {
	if rh.data == nil {
		rh.clusterId = rh.req.PathParameter(clusterIdTag)
		rh.data = operator.M{
			clusterIdTag: rh.clusterId,
		}
	}
	return rh.data
}

func (rh *reqHost) getHost() ([]interface{}, error) {
	return rh.get(rh.getHostFeat())
}

func (rh *reqHost) queryHost() ([]interface{}, error) {
	return rh.get(rh.getQueryHostFeat())
}

func (rh *reqHost) get(condition *operator.Condition) (r []interface{}, err error) {
	tank := rh.tank.From(rh.getTable()).Filter(condition).Query()

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

func (rh *reqHost) putHost() error {
	return rh.put(rh.getHostFeat())
}

func (rh *reqHost) put(condition *operator.Condition) (err error) {
	tank := rh.tank.From(rh.getTable()).Filter(condition)

	data, err := rh.getReqData()
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

func (rh *reqHost) removeHost() error {
	return rh.remove(rh.getHostFeat())
}

func (rh *reqHost) remove(condition *operator.Condition) (err error) {
	tank := rh.tank.From(rh.getTable()).Filter(condition).RemoveAll()

	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to remove. err: %v", err)
		return
	}
	if changeInfo := tank.GetChangeInfo(); changeInfo.Removed == 0 {
		return storageErr.ResourceDoesNotExist
	}
	return
}

// List all host whose cluster equals clusterId
// and make their cluster empty string
func (rh *reqHost) cleanCluster(clusterId string) (err error) {
	now := time.Now()
	condition := operator.BaseCondition.AddOp(operator.Eq, clusterIdTag, clusterId)
	tank := rh.tank.From(rh.getTable()).Filter(condition).UpdateAll(operator.M{clusterIdTag: "", updateTimeTag: now})
	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to clean clusterId %s err: %v", rh.clusterId, err)
		return
	}
	return
}

// doRelation has 2 options:
//  - put(isPut=true):   make cluster clusterId just contains ips, for instance,
//                       put(127.0.0.1, 127.0.0.2)=bcs-10001, then cluster=bcs-10001 contains 2 ips, 127.0.0.1, 127.0.0.2,
//                       no matter what host it contains before.
//  - post(isPut=false): add ips to cluster clusterId, for instance,
//                       post(127.0.0.3)=bcs-10001, then cluster=bcs-10001 contains 3 ips, 127.0.0.1, 127.0.0.2, 127.0.0.3.
//                       it just add.
func (rh *reqHost) doRelation(isPut bool) (err error) {
	rawTank := rh.tank.From(rh.getTable())
	condition, err := rh.getRelationFeat()
	if err != nil {
		blog.Errorf("Failed to get relation features: %v", err)
		return
	}

	tank := rawTank.Filter(condition).Distinct(ipTag).Query()
	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to query. err: %v", err)
		return
	}

	// currentIpList is the current ipList in db which match the clusterId
	currentIpList := listInterface2String(tank.GetValue())
	// expectIpList is the ipList which match the clusterId we expected
	expectIpList := rh.ips

	// insert the ip with clusterId="" which is not in db yet, preparing for next ops
	insertList := make([]operator.M, 0, len(expectIpList))
	now := time.Now()
	for _, ip := range expectIpList {
		if !inList(ip, currentIpList) {
			insertList = append(insertList, operator.M{
				ipTag:         ip,
				clusterIdTag:  "",
				createTimeTag: now,
				updateTimeTag: now,
			})
		}
	}

	if len(insertList) > 0 {
		tank = rawTank.Insert(insertList...)
		if err = tank.GetError(); err != nil {
			blog.Errorf("Failed to insert new host in expected list. err: %v", err)
			return
		}
	}

	data := rh.getRelationData()
	// put will clean the all cluster first, if not then just update
	if isPut {
		if err = rh.cleanCluster(rh.clusterId); err != nil {
			return
		}
	}

	tank = rawTank.Filter(condition).UpdateAll(data.Update(updateTimeTag, now))
	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to update relation err: %v", err)
		return
	}
	return
}

// exit() should be called after all ops in reqDynamic to close the connection
// to database.
func (rh *reqHost) exit() {
	if rh.tank != nil {
		rh.tank.Close()
	}
}

func urlPath(oldUrl string) string {
	return oldUrl
}

func inList(key string, list []string) bool {
	for _, d := range list {
		if key == d {
			return true
		}
	}
	return false
}

func listInterface2String(s []interface{}) (r []string) {
	r = make([]string, len(s))
	for _, v := range s {
		if vv, ok := v.(string); ok {
			r = append(r, vv)
		}
	}
	return
}
