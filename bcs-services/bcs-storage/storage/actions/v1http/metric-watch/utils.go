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

package metricWatch

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

type reqMetric struct {
	req  *restful.Request
	resp *restful.Response
	tank operator.Tank

	condition *operator.Condition
	table     string
}

func newReqMetric(req *restful.Request, resp *restful.Response) *reqMetric {
	return &reqMetric{
		req:  req,
		resp: resp,
		tank: getNewTank(),
	}
}

func (rd *reqMetric) getFeat() *operator.Condition {
	return operator.BaseCondition
}

func (rd *reqMetric) getTable() string {
	if rd.table == "" {
		rd.table = rd.req.PathParameter(clusterIdTag)
	}
	return rd.table
}

func (rd *reqMetric) watch() {
	tank := rd.tank.From(rd.getTable()).Filter(rd.getFeat())
	ws, err := lib.NewWatchServer(rd.req, rd.resp, tank)

	ws.Writer = func(resp *restful.Response, event *operator.Event) bool {
		if event.Type != operator.Del && event.Value[resourceTypeTag] != rd.req.PathParameter(resourceTypeTag) {
			return false
		}

		if err = codec.EncJsonWriter(event, resp.ResponseWriter); err != nil {
			blog.Errorf("defaultWriter error: %v", err)
			return false
		}
		return true
	}

	if err != nil {
		rd.resp.Write(operator.EventWatchBreakBytes)
		return
	}

	ws.Go(context.Background())
}
