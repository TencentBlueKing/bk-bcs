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

package dynamicWatch

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/emicklei/go-restful"
)

type reqDynamic struct {
	req  *restful.Request
	resp *restful.Response
	tank operator.Tank

	condition *operator.Condition
	table     string
}

func newReqDynamic(req *restful.Request, resp *restful.Response) *reqDynamic {
	return &reqDynamic{
		req:  req,
		resp: resp,
		tank: getNewTank(),
	}
}

func (rd *reqDynamic) getFeat() *operator.Condition {
	return operator.BaseCondition
}

func (rd *reqDynamic) getTable() string {
	if rd.table == "" {
		rd.table = rd.req.PathParameter(clusterIdTag) + "_" + rd.req.PathParameter(tableTag)
	}
	return rd.table
}

func (rd *reqDynamic) watch() {
	tank := rd.tank.From(rd.getTable()).Filter(rd.getFeat())
	ws, err := lib.NewWatchServer(rd.req, rd.resp, tank)
	if err != nil {
		blog.Error("dynamic get watch server failed: %v", err)
		rd.resp.Write(operator.EventWatchBreakBytes)
		return
	}

	ws.Go(context.Background())
}

func (rd *reqDynamic) watchContainer() {
	tableTank := rd.tank.Tables()
	if err := tableTank.GetError(); err != nil {
		blog.Errorf("dynamic container watch failed: %v", err)
		rd.resp.Write(operator.EventWatchBreakBytes)
		return
	}

	var table string
	canWatch := false
	clusterId := rd.req.PathParameter(clusterIdTag)
	r := tableTank.GetValue()
	for _, t := range containerTypeList {
		table = fmt.Sprintf("%s_%s", clusterId, t)
		if inList(table, r) {
			canWatch = true
			break
		}
	}

	if !canWatch {
		blog.Errorf("dynamic container watch failed, there is no tables in (%s) like: %v", clusterId, containerTypeList)
		rd.resp.Write(operator.EventWatchBreakBytes)
		return
	}

	blog.Infof(table)

	tank := rd.tank.From(table).Filter(rd.getFeat())
	ws, err := lib.NewWatchServer(rd.req, rd.resp, tank)
	if err != nil {
		blog.Error("dynamic get container watch server failed: %v", err)
		rd.resp.Write(operator.EventWatchBreakBytes)
		return
	}

	ws.Go(context.Background())
}

func inList(s string, l []interface{}) bool {
	for _, ll := range l {
		if ls, ok := ll.(string); ok {
			if ls == s {
				return true
			}
		}
	}
	return false
}
