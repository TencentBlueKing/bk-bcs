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

package watchk8smesos

import (
	"fmt"
	"time"

	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	sto "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/types"
)

func getReqData(req *restful.Request) (interface{}, error) {
	var tmp bcstypes.BcsStorageWatchIf
	if err := codec.DecJsonReader(req.Request.Body, &tmp); err != nil {
		return nil, err
	}
	return tmp.Data, nil
}

func get(req *restful.Request) (interface{}, error) {
	resType := lib.GetQueryParamString(req, tableTag)
	clusterID := lib.GetQueryParamString(req, clusterIDTag)
	ns := lib.GetQueryParamString(req, namespaceTag)
	name := lib.GetQueryParamString(req, nameTag)
	store := apiserver.GetAPIResource().GetStoreClient(dbConfig)
	rawObj, err := store.Get(req.Request.Context(), types.ObjectType(resType), types.ObjectKey{
		ClusterID: clusterID,
		Namespace: ns,
		Name:      name,
	})
	if err != nil {
		return nil, err
	}
	return rawObj.GetData(), nil
}

func list(req *restful.Request) ([]string, error) {
	resType := lib.GetQueryParamString(req, tableTag)
	clusterID := lib.GetQueryParamString(req, clusterIDTag)
	ns := lib.GetQueryParamString(req, namespaceTag)

	store := apiserver.GetAPIResource().GetStoreClient(dbConfig)
	objList, err := store.List(req.Request.Context(), types.ObjectType(resType), &sto.ListOptions{
		Cluster:   clusterID,
		Namespace: ns,
	})
	if err != nil {
		return nil, err
	}
	var retList []string
	for _, obj := range objList {
		retList = append(retList, fmt.Sprintf("%s.%s", obj.GetNamespace(), obj.GetName()))
	}
	return retList, nil
}

func put(req *restful.Request) error {
	resType := lib.GetQueryParamString(req, tableTag)
	clusterID := lib.GetQueryParamString(req, clusterIDTag)
	ns := lib.GetQueryParamString(req, namespaceTag)
	name := lib.GetQueryParamString(req, nameTag)
	dataRaw, err := getReqData(req)
	if err != nil {
		return err
	}

	data := dataRaw.(map[string]interface{})
	data[updateTimeTag] = time.Now()
	newObj := &types.RawObject{
		Meta: types.Meta{
			Type:      types.ObjectType(resType),
			ClusterID: clusterID,
			Namespace: ns,
			Name:      name,
		},
		Data: data,
	}
	store := apiserver.GetAPIResource().GetStoreClient(dbConfig)
	return store.Update(req.Request.Context(), newObj, &sto.UpdateOptions{
		CreateNotExists: true,
	})
}

func remove(req *restful.Request) error {
	resType := lib.GetQueryParamString(req, tableTag)
	clusterID := lib.GetQueryParamString(req, clusterIDTag)
	ns := lib.GetQueryParamString(req, namespaceTag)
	name := lib.GetQueryParamString(req, nameTag)
	newObj := &types.RawObject{
		Meta: types.Meta{
			Type:      types.ObjectType(resType),
			ClusterID: clusterID,
			Namespace: ns,
			Name:      name,
		},
	}
	store := apiserver.GetAPIResource().GetStoreClient(dbConfig)
	return store.Delete(req.Request.Context(), newObj, &sto.DeleteOptions{
		IgnoreNotFound: true,
	})
}

func urlK8SPath(oldURL string) string {
	return urlK8SPrefix + oldURL
}

func urlMesosPath(oldURL string) string {
	return urlMesosPrefix + oldURL
}
