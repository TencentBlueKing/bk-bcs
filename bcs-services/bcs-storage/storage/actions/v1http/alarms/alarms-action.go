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

// Package alarms xxx
package alarms

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/utils"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	v1http "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/clean"
)

const (
	dataTag         = "data"
	extraTag        = "extra"
	fieldTag        = "field"
	typeTag         = "type"
	clusterIDTag    = "clusterId"
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

var conditionTagList = [...]string{clusterIDTag, namespaceTag, sourceTag, moduleTag}

// Use Mongodb for storage.
const dbConfig = "mongodb/alarm"

// PostAlarm post alarms
func PostAlarm(req *restful.Request, resp *restful.Response) {
	const (
		handler = "PostAlarm"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	errFunc := func(err error) {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStoragePutResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			ErrCode: common.BcsErrStoragePutResourceFail,
			Message: common.BcsErrStoragePutResourceFailStr})
	}

	// data
	data, err := getReqData(req)
	if err != nil {
		errFunc(err)
		return
	}
	// option
	opt := &lib.StorePutOption{
		CreateTimeKey: createTimeTag,
	}

	if err = PutData(req.Request.Context(), data, opt); err != nil {
		errFunc(err)
		return
	}

	lib.ReturnRest(&lib.RestResponse{Resp: resp})
}

// ListAlarm list alarm
func ListAlarm(req *restful.Request, resp *restful.Response) {
	const (
		handler = "ListAlarm"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	errFunc := func(err error) {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp:    resp,
			Data:    []string{},
			ErrCode: common.BcsErrStorageListResourceFail,
			Message: common.BcsErrStorageListResourceFailStr})
	}
	// option
	opt, err := getStoreGetOption(req)
	if err != nil {
		errFunc(err)
		return
	}

	// get data list
	r, err := GetData(req.Request.Context(), opt)
	if err != nil {
		errFunc(err)
		return
	}

	extra := map[string]interface{}{"total": len(r)}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r, Extra: extra})
}

// CleanAlarm clean alarm
func CleanAlarm() {
	maxCap := apiserver.GetAPIResource().Conf.AlarmMaxCap
	maxTime := apiserver.GetAPIResource().Conf.AlarmMaxTime
	cleaner := clean.NewDBCleaner(apiserver.GetAPIResource().GetDBClient(dbConfig), tableName, time.Hour)
	cleaner.WithMaxEntryNum(maxCap)
	cleaner.WithMaxDuration(time.Duration(maxTime*24)*time.Hour, time.Duration(0), createTimeTag)
	cleaner.Run(context.TODO())
}

func init() {
	alarmPath := "/alarms"
	actions.RegisterV1Action(actions.Action{Verb: "POST", Path: alarmPath, Params: nil,
		Handler: lib.MarkProcess(PostAlarm)})
	actions.RegisterV1Action(actions.Action{Verb: "GET", Path: alarmPath, Params: nil,
		Handler: lib.MarkProcess(ListAlarm)})

	actions.RegisterDaemonFunc(CleanAlarm)
}
