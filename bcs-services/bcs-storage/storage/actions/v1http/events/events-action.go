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

package events

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/operator"

	"github.com/emicklei/go-restful"
)

const (
	tableName     = "event"
	dataTag       = "data"
	extraTag      = "extra"
	extraConTag   = "extra_contain"
	fieldTag      = "field"
	idTag         = "id"
	envTag        = "env"
	kindTag       = "kind"
	levelTag      = "level"
	componentTag  = "component"
	typeTag       = "type"
	describeTag   = "describe"
	clusterIdTag  = "clusterId"
	extraInfoTag  = "extraInfo"
	offsetTag     = "offset"
	limitTag      = "length"
	timeBeginTag  = "timeBegin"
	timeEndTag    = "timeEnd"
	createTimeTag = "createTime"
	eventTimeTag  = "eventTime"
	timeLayout    = "2006-01-02 15:04:05"
)

var needTimeFormatList = [...]string{createTimeTag, eventTimeTag}
var conditionTagList = [...]string{idTag, envTag, kindTag, levelTag, componentTag, typeTag, clusterIdTag, "extraInfo.name", "extraInfo.namespace", "extraInfo.kind"}

// Use Mongodb for storage.
const dbConfig = "event"

var getNewTank operator.GetNewTank = lib.GetMongodbTank(dbConfig)

func PutEvent(req *restful.Request, resp *restful.Response) {
	request := newReqEvent(req)
	defer request.exit()
	if err := request.insert(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStoragePutResourceFail, Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func ListEvent(req *restful.Request, resp *restful.Response) {
	request := newReqEvent(req)
	defer request.exit()
	r, total, err := request.listEvent()
	extra := map[string]interface{}{"total": total}
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr, Extra: extra})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r, Extra: extra})
}

func CleanEventsOutDate() {
	cleanEventOutDate(apiserver.GetAPIResource().Conf.EventMaxTime)
}

func CleanEventsOutCap() {
	cleanEventOutCap(apiserver.GetAPIResource().Conf.EventMaxCap)
}

func init() {
	eventPath := urlPath("/events")
	actions.RegisterV1Action(actions.Action{Verb: "PUT", Path: eventPath, Params: nil, Handler: lib.MarkProcess(PutEvent)})
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: eventPath, Params: nil, Handler: lib.MarkProcess(ListEvent)})

	actions.RegisterDaemonFunc(CleanEventsOutDate)
	actions.RegisterDaemonFunc(CleanEventsOutCap)
}
