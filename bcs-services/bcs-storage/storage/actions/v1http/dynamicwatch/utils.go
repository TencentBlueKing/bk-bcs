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

package dynamicwatch

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
)

type reqDynamic struct {
	req   *restful.Request
	resp  *restful.Response
	store *lib.Store
}

func newReqDynamic(req *restful.Request, resp *restful.Response) *reqDynamic {
	return &reqDynamic{
		req:   req,
		resp:  resp,
		store: GetStore(),
	}
}

func (rd *reqDynamic) getTable() string {
	return rd.req.PathParameter(tableTag)
}

func (rd *reqDynamic) watch() {
	watchOption := &lib.WatchServerOption{
		Store:     rd.store,
		TableName: rd.getTable(),
		Cond:      operator.M{constants.ClusterIDTag: rd.req.PathParameter(constants.ClusterIDTag)},
		Req:       rd.req,
		Resp:      rd.resp,
	}

	ws, err := lib.NewWatchServer(watchOption)
	if err != nil {
		blog.Error("dynamic get watch server failed, err %s", err.Error())
		_, _ = rd.resp.Write(lib.EventWatchBreakBytes)
		return
	}

	ws.Go(context.Background())
}

func (rd *reqDynamic) watchContainer() {
	watchOption := &lib.WatchServerOption{
		Store:     rd.store,
		TableName: constants.ContainerInfo,
		Cond:      operator.M{constants.ClusterIDTag: rd.req.PathParameter(constants.ClusterIDTag)},
		Req:       rd.req,
		Resp:      rd.resp,
	}

	ws, err := lib.NewWatchServer(watchOption)
	if err != nil {
		blog.Error("dynamic get watch server failed, err %s", err.Error())
		_, _ = rd.resp.Write(lib.EventWatchBreakBytes)
		return
	}

	ws.Go(context.Background())
}
