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

package watchk8smesos

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	restful "github.com/emicklei/go-restful/v3"

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

func get(req *restful.Request, env string) (interface{}, error) {
	// 表名
	resType := types.ObjectType(req.PathParameter(tableTag))
	// 参数
	key := types.ObjectKey{
		ClusterID: req.PathParameter(clusterIDTag),
		Namespace: req.PathParameter(namespaceTag),
		Name:      req.PathParameter(nameTag),
	}

	// option
	opt := &sto.GetOptions{Env: env}

	// 查询数据
	data, err := GetData(req.Request.Context(), resType, key, opt)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func list(req *restful.Request, env string) ([]string, error) {
	// 表名
	resType := types.ObjectType(req.PathParameter(tableTag))
	// 参数
	clusterID := req.PathParameter(clusterIDTag)
	ns := req.PathParameter(namespaceTag)

	// option
	opt := &sto.ListOptions{
		Env:       env,
		Cluster:   clusterID,
		Namespace: ns,
	}

	return GetList(req.Request.Context(), resType, opt)
}

func put(req *restful.Request, env string) error {
	// 表名
	resType := types.ObjectType(req.PathParameter(tableTag))
	// 参数
	clusterID := req.PathParameter(clusterIDTag)
	ns := req.PathParameter(namespaceTag)
	name := req.PathParameter(nameTag)

	dataRaw, err := getReqData(req)
	if err != nil {
		return err
	}
	data, ok := dataRaw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid data format")
	}
	data[updateTimeTag] = time.Now()

	newObj := &types.RawObject{
		Meta: types.Meta{
			Type:      resType,
			ClusterID: clusterID,
			Namespace: ns,
			Name:      name,
		},
		Data: data,
	}

	// option
	opt := &sto.UpdateOptions{
		Env:             env,
		CreateNotExists: true,
	}

	return PutData(req.Request.Context(), newObj, opt)
}

func remove(req *restful.Request, env string) error {
	// 表名
	resType := req.PathParameter(tableTag)
	// 参数
	clusterID := req.PathParameter(clusterIDTag)
	ns := req.PathParameter(namespaceTag)
	name := req.PathParameter(nameTag)

	newObj := &types.RawObject{
		Meta: types.Meta{
			Type:      types.ObjectType(resType),
			ClusterID: clusterID,
			Namespace: ns,
			Name:      name,
		},
	}

	// option
	opt := &sto.DeleteOptions{
		Env:            env,
		IgnoreNotFound: true,
	}

	return RemoveDta(req.Request.Context(), newObj, opt)
}

func urlK8SPath(oldURL string) string {
	return urlK8SPrefix + oldURL
}

func urlMesosPath(oldURL string) string {
	return urlMesosPrefix + oldURL
}
