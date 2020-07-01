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

package watch

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/codec"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

type reqWatch struct {
	req  *restful.Request
	tank operator.Tank

	typeTable string
	nodeTable string
	nodeT     string
	data      interface{}
}

func newReqWatch(req *restful.Request) *reqWatch {
	return &reqWatch{
		req:  req,
		tank: getNewTank(),
	}
}

// reset clean the condition, data etc. so that the reqWatch can be ready for
// next op.
func (rw *reqWatch) reset() {
	rw.typeTable = ""
	rw.nodeTable = ""
	rw.data = nil
}

func (rw *reqWatch) getTypeTable() string {
	if rw.typeTable == "" {
		rw.typeTable = fmt.Sprintf("%s/%s/%s",
			env,
			rw.req.PathParameter(clusterIdTag),
			rw.req.PathParameter(tableTag),
		)
	}
	return rw.typeTable
}

func (rw *reqWatch) getNodeTable() string {
	if rw.nodeTable == "" {
		rw.nodeTable = fmt.Sprintf("%s/%s/%s/%s.%s",
			env,
			rw.req.PathParameter(clusterIdTag),
			rw.req.PathParameter(tableTag),
			rw.req.PathParameter(namespaceTag),
			rw.req.PathParameter(nameTag),
		)
	}
	return rw.nodeTable
}

func (rw *reqWatch) getReqData() (interface{}, error) {
	if rw.data == nil {
		var tmp types.BcsStorageWatchIf
		if err := codec.DecJsonReader(rw.req.Request.Body, &tmp); err != nil {
			return nil, err
		}
		rw.data = tmp.Data
	}
	return rw.data, nil
}

func (rw *reqWatch) get() (r []interface{}, err error) {
	tank := rw.tank.From(rw.getNodeTable()).GetTableV()

	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to query. err: %v", err)
		return
	}
	r = tank.GetValue()
	if len(r) > 0 {
		var tmp interface{}
		err = codec.DecJson([]byte(r[0].(string)), &tmp)
		r = []interface{}{tmp}
	}
	return
}

func (rw *reqWatch) list() (r []interface{}, err error) {
	tank := rw.tank.From(rw.getTypeTable()).Tables()

	if err = tank.GetError(); err != nil {
		blog.Errorf("Failed to query. err: %v", err)
		return
	}
	r = tank.GetValue()
	return
}

func (rw *reqWatch) put() (err error) {
	tank := rw.tank.From(rw.getNodeTable())

	dataRaw, err := rw.getReqData()
	if err != nil {
		return
	}
	data := dataRaw.(map[string]interface{})
	data[updateTimeTag] = time.Now()

	err = tank.SetTableV(data).GetError()
	return
}

func (rw *reqWatch) remove() error {
	return rw.tank.From(rw.getNodeTable()).Remove().GetError()
}

// exit() should be called after all ops in reqWatch to close the connection
// to database.
func (rw *reqWatch) exit() {
	if rw.tank != nil {
		rw.tank.Close()
	}
}

func urlPath(oldUrl string) string {
	return urlPrefix + oldUrl
}
