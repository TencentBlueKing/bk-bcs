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

package metricwatch

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
)

func watch(req *restful.Request, resp *restful.Response) {
	opt := &lib.WatchServerOption{
		Store:     GetStore(),
		TableName: req.PathParameter(clusterIDTag),
		Req:       req,
		Resp:      resp,
	}
	ws, err := lib.NewWatchServer(opt)
	ws.Writer = func(resp *restful.Response, event *lib.Event) bool {
		if event.Type != lib.Del && event.Value[resourceTypeTag] != req.PathParameter(resourceTypeTag) {
			return false
		}

		if err = codec.EncJsonWriter(event, resp.ResponseWriter); err != nil {
			blog.Errorf("defaultWriter error: %v", err)
			return false
		}
		return true
	}

	if err != nil {
		_, _ = resp.Write(lib.EventWatchBreakBytes)
		return
	}

	ws.Go(context.Background())
}
