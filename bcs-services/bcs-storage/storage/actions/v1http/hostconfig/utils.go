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

package hostconfig

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"

	restful "github.com/emicklei/go-restful"
)

func getHostFeat(req *restful.Request) *operator.Condition {
	return getFeat(req, hostFeatTags)
}

func getQueryHostFeat(req *restful.Request) *operator.Condition {
	return getFeat(req, hostQueryFeatTags)
}

func getFeatM(req *restful.Request, resourceFeatList []string) operator.M {
	features := make(operator.M, len(resourceFeatList))
	for _, key := range resourceFeatList {
		features[key] = req.PathParameter(key)
	}
	return features
}

func getFeat(req *restful.Request, resourceFeatList []string) *operator.Condition {
	features := make(operator.M, len(resourceFeatList))
	for _, key := range resourceFeatList {
		features[key] = req.PathParameter(key)
	}
	return operator.NewLeafCondition(operator.Eq, features)
}

func getRelationFeat(req *restful.Request) (*types.BcsStorageClusterRelationIf, error) {
	var tmp *types.BcsStorageClusterRelationIf
	if err := codec.DecJsonReader(req.Request.Body, &tmp); err != nil {
		return nil, err
	}
	return tmp, nil
}

func getReqData(req *restful.Request, features operator.M) (operator.M, error) {
	var tmp types.BcsStorageHostIf
	if err := codec.DecJsonReader(req.Request.Body, &tmp); err != nil {
		return nil, err
	}
	data := lib.CopyMap(features)
	data[clusterIDTag] = tmp.ClusterId
	data[dataTag] = tmp.Data
	return data, nil
}

func getRelationData(req *restful.Request) operator.M {
	clusterID := req.PathParameter(clusterIDTag)
	return operator.M{
		clusterIDTag: clusterID,
	}
}

func getHost(req *restful.Request) ([]operator.M, error) {
	return get(req, getHostFeat(req))
}

func queryHost(req *restful.Request) ([]operator.M, error) {
	return get(req, getQueryHostFeat(req))
}

func get(req *restful.Request, condition *operator.Condition) ([]operator.M, error) {
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	getOption := &lib.StoreGetOption{
		Cond: condition,
	}
	mList, err := store.Get(req.Request.Context(), tableName, getOption)
	return mList, err
}

func putHost(req *restful.Request) error {
	return put(req, getFeatM(req, hostFeatTags))
}

func put(req *restful.Request, features operator.M) error {
	condition := operator.NewLeafCondition(operator.Eq, features)
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	putOption := &lib.StorePutOption{
		UniqueKey:     indexKeys,
		Cond:          condition,
		CreateTimeKey: createTimeTag,
		UpdateTimeKey: updateTimeTag,
	}

	data, err := getReqData(req, features)
	if err != nil {
		return err
	}

	return store.Put(req.Request.Context(), tableName, data, putOption)
}

func removeHost(req *restful.Request) error {
	return remove(req, getHostFeat(req))
}

func remove(req *restful.Request, condition *operator.Condition) error {
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	rmOption := &lib.StoreRemoveOption{
		Cond: condition,
	}
	return store.Remove(req.Request.Context(), tableName, rmOption)
}

// List all host whose cluster equals clusterID
// and make their cluster empty string
func cleanCluster(req *restful.Request, clusterID string) error {
	now := time.Now()
	condition := operator.NewLeafCondition(operator.Eq, operator.M{clusterIDTag: clusterID})
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	_, err := store.GetDB().Table(tableName).UpdateMany(req.Request.Context(), condition,
		operator.M{clusterIDTag: "", updateTimeTag: now})
	return err
}

// doRelation has 2 options:
// - put(isPut=true): make cluster clusterId just contains ips, for instance,
//                    put(127.0.0.1, 127.0.0.2)=bcs-10001, then cluster=bcs-10001 contains 2 ips, 127.0.0.1, 127.0.0.2,
//                       no matter what host it contains before.
// - post(isPut=false): add ips to cluster clusterId, for instance,
//                    post(127.0.0.3)=bcs-10001, then cluster=bcs-10001 contains 3 ips, 127.0.0.1, 127.0.0.2, 127.0.0.3.
//                    it just add.
func doRelation(req *restful.Request, isPut bool) error {
	tmp, err := getRelationFeat(req)
	if err != nil {
		return fmt.Errorf("Failed to get relation features, err %s", err.Error())
	}
	condition := operator.NewLeafCondition(operator.In, operator.M{ipTag: tmp.Ips})
	getOption := &lib.StoreGetOption{
		Cond:   condition,
		Fields: []string{ipTag},
	}
	store := lib.NewStore(
		apiserver.GetAPIResource().GetDBClient(dbConfig),
		apiserver.GetAPIResource().GetEventBus(dbConfig))
	mList, err := store.Get(req.Request.Context(), tableName, getOption)
	if err != nil {
		return fmt.Errorf("failed to query, err %s", err.Error())
	}
	var ipList []string
	for _, doc := range mList {
		ip, ok := doc[ipTag]
		if !ok {
			return fmt.Errorf("failed to get ip from %+v", doc)
		}
		ipStr, aok := ip.(string)
		if !aok {
			return fmt.Errorf("failed to parse ip from %+v", doc)
		}
		ipList = append(ipList, ipStr)
	}
	currentIPList := deduplicateStringSlice(ipList)
	// expectIpList is the ipList which match the clusterId we expected
	expectIPList := tmp.Ips

	// insert the ip with clusterId="" which is not in db yet, preparing for next ops
	insertList := make([]operator.M, 0, len(expectIPList))
	now := time.Now()
	for _, ip := range expectIPList {
		if !inList(ip, currentIPList) {
			insertList = append(insertList, operator.M{
				ipTag:         ip,
				clusterIDTag:  "",
				createTimeTag: now,
				updateTimeTag: now,
			})
		}
	}

	if len(insertList) > 0 {
		if store.GetDB().Table(tableName).Insert(req.Request.Context(), []interface{}{insertList}); err != nil {
			return err
		}
	}

	data := getRelationData(req)
	// put will clean the all cluster first, if not then just update
	if isPut {
		clusterID, ok := data[clusterIDTag].(string)
		if !ok {
			return fmt.Errorf("cannot parse clusterID from %+v", data)
		}
		if err = cleanCluster(req, clusterID); err != nil {
			return err
		}
	}
	data.Update(updateTimeTag, now)
	_, err = store.GetDB().Table(tableName).UpdateMany(req.Request.Context(), condition, data)
	return err
}

func urlPath(oldURL string) string {
	return oldURL
}

func deduplicateStringSlice(mList []string) []string {
	tmpMap := make(map[string]bool)
	var retList []string
	for _, value := range mList {
		if _, ok := tmpMap[value]; !ok {
			retList = append(retList, value)
		}
	}
	return retList
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
