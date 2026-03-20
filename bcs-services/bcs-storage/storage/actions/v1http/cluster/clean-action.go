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

// Package cluster xxx
package cluster

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	v1http "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/clean"
)

// CleanClusterData HTTP接口：清理集群所有数据
// DELETE /bcsstorage/v1/clusters/{clusterId}/data
func CleanClusterData(req *restful.Request, resp *restful.Response) {
	const (
		handler = "CleanClusterData"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	// 解析URL路径参数
	clusterId := req.PathParameter("clusterId")
	if clusterId == "" {
		ReturnError(handler, span, resp, common.BcsErrStorageGetResourceFail,
			"clusterId is required", fmt.Errorf("empty clusterId"))
		return
	}

	blog.Infof("%s: clusterId=%s", handler, clusterId)

	cleaner := clean.NewClusterCleaner()

	// 执行清理
	deletedCounts, err := cleaner.CleanClusterData(req.Request.Context(), clusterId)
	if err != nil {
		ReturnError(handler, span, resp, common.BcsErrStorageDeleteResourceFail,
			fmt.Sprintf("clean cluster data failed: %v", err), err)
		return
	}

	responseData := map[string]interface{}{
		"clusterId":     clusterId,
		"deletedCounts": deletedCounts,
	}

	blog.Infof("%s success: clusterId=%s, deleted=%v",
		handler, clusterId, deletedCounts)

	lib.ReturnRest(&lib.RestResponse{
		Resp:    resp,
		Data:    []interface{}{responseData},
		ErrCode: common.BcsSuccess,
		Message: common.BcsSuccessStr,
	})
}

func init() {
	// 注册集群数据清理路由
	// DELETE /bcsstorage/v1/clusters/{clusterId}/data
	clusterCleanPath := "/clusters/{clusterId}/data"
	actions.RegisterV1Action(actions.Action{
		Verb:    "DELETE",
		Path:    clusterCleanPath,
		Params:  nil,
		Handler: lib.MarkProcess(CleanClusterData),
	})
	blog.Infof("cluster clean route registered: DELETE %s", clusterCleanPath)
}
