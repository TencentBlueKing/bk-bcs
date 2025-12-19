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

package hostconfig

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
)

func getHostFeat(req *restful.Request) *operator.Condition {
	return getFeat(req, hostFeatTags)
}

func listHostFeat(req *restful.Request) *operator.Condition {
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
	return QueryHost(req.Request.Context(), getHostFeat(req))
}

func listHost(req *restful.Request) ([]operator.M, error) {
	return QueryHost(req.Request.Context(), listHostFeat(req))
}

func putHost(req *restful.Request) error {
	return put(req, getFeatM(req, hostFeatTags))
}

func put(req *restful.Request, features operator.M) error {
	// 参数
	condition := operator.NewLeafCondition(operator.Eq, features)
	data, err := getReqData(req, features)
	if err != nil {
		return err
	}
	return PutHostToDB(req.Request.Context(), data, condition)
}

func removeHost(req *restful.Request) error {
	return remove(req, getHostFeat(req))
}

func remove(req *restful.Request, condition *operator.Condition) error {
	return RemoveHost(req.Request.Context(), condition)
}

// doRelation has 2 options:
//   - put(isPut=true): make cluster clusterId just contains ips, for instance,
//     put(127.0.0.1, 127.0.0.2)=bcs-10001, then cluster=bcs-10001 contains 2 ips, 127.0.0.1, 127.0.0.2,
//     no matter what host it contains before.
//   - post(isPut=false): add ips to cluster clusterId, for instance,
//     post(127.0.0.3)=bcs-10001, then cluster=bcs-10001 contains 3 ips, 127.0.0.1, 127.0.0.2, 127.0.0.3.
//     it just add.
func doRelation(req *restful.Request, isPut bool) error {
	// 参数
	relation, err := getRelationFeat(req)
	if err != nil {
		return fmt.Errorf("failed to QueryHost relation features, err %s", err.Error())
	}
	data := getRelationData(req)
	condition := operator.NewLeafCondition(operator.In, operator.M{ipTag: relation.Ips})

	// option
	opt := &lib.StoreGetOption{
		Cond:   condition,
		Fields: []string{ipTag},
	}

	return DoRelation(req.Request.Context(), opt, data, isPut, relation)
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
