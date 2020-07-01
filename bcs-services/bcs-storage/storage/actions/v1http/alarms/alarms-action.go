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

package alarms

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
	dataTag         = "data"
	extraTag        = "extra"
	fieldTag        = "field"
	typeTag         = "type"
	clusterIdTag    = "clusterId"
	namespaceTag    = "namespace"
	messageTag      = "message"
	sourceTag       = "source"
	moduleTag       = "module"
	offsetTag       = "offset"
	limitTag        = "length"
	tableName       = "alarm"
	createTimeTag   = "createTime"
	receivedTimeTag = "receivedTime"
	timeBeginTag    = "timeBegin"
	timeEndTag      = "timeEnd"
	timeLayout      = "2006-01-02 15:04:05"
)

var needTimeFormatList = [...]string{createTimeTag, receivedTimeTag}
var conditionTagList = [...]string{clusterIdTag, namespaceTag, sourceTag, moduleTag}

// Use Mongodb for storage.
const dbConfig = "alarm"

var getNewTank operator.GetNewTank = lib.GetMongodbTank(dbConfig)

func PostAlarm(req *restful.Request, resp *restful.Response) {
	request := newReqAlarm(req)
	defer request.exit()
	if err := request.insert(); err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, ErrCode: common.BcsErrStoragePutResourceFail, Message: common.BcsErrStoragePutResourceFailStr})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

func ListAlarm(req *restful.Request, resp *restful.Response) {
	request := newReqAlarm(req)
	defer request.exit()
	r, total, err := request.listAlarm()
	extra := map[string]interface{}{"total": total}
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: []string{}, ErrCode: common.BcsErrStorageListResourceFail, Message: common.BcsErrStorageListResourceFailStr, Extra: extra})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r, Extra: extra})
}

func CleanAlarmsOutDate() {
	cleanAlarmOutDate(apiserver.GetAPIResource().Conf.AlarmMaxTime)
}

func CleanAlarmsOutCap() {
	cleanAlarmOutCap(apiserver.GetAPIResource().Conf.AlarmMaxCap)
}

func init() {
	alarmPath := urlPath("/alarms")
	actions.RegisterV1Action(actions.Action{Verb: "POST", Path: alarmPath, Params: nil, Handler: lib.MarkProcess(PostAlarm)})
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: alarmPath, Params: nil, Handler: lib.MarkProcess(ListAlarm)})

	actions.RegisterDaemonFunc(CleanAlarmsOutDate)
	actions.RegisterDaemonFunc(CleanAlarmsOutCap)
}
