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

package cluster

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/utils"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	v1http "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
)

// PutClusterInfo put cluster info
func PutClusterInfo(req *restful.Request, resp *restful.Response) {
	const (
		handler = "PutClusterInfo"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := putClusterInfo(req); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStoragePutResourceFail,
			Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// DeleteClusterInfo delete cluster info
func DeleteClusterInfo(req *restful.Request, resp *restful.Response) {
	const (
		handler = "DeleteClusterInfo"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := deleteClusterInfo(req); err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageDeleteResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStorageDeleteResourceFail,
			Message: common.BcsErrStorageDeleteResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func init() {
	clusterPath := urlPath("/clusters/{clusterID}")
	actions.RegisterV1Action(actions.Action{
		Verb: "PUT", Path: clusterPath, Params: nil, Handler: lib.MarkProcess(PutClusterInfo)})
	actions.RegisterV1Action(actions.Action{
		Verb: "DELETE", Path: clusterPath, Params: nil, Handler: lib.MarkProcess(DeleteClusterInfo)})
}
