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
	"context"
	"strings"
	"time"

	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
	v1http "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/v1http/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/clean"
)

const (
	tablePrefix   = "event_"
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
	clusterIDTag  = "clusterId"
	extraInfoTag  = "extraInfo"
	offsetTag     = "offset"
	limitTag      = "length"
	timeBeginTag  = "timeBegin"
	timeEndTag    = "timeEnd"
	createTimeTag = "createTime"
	eventTimeTag  = "eventTime"
	timeLayout    = "2006-01-02 15:04:05"

	nameSpaceTag    = "namespace"
	resourceTypeTag = "resourceType"
	resourceKindTag = "resourceKind"
	resourceNameTag = "resourceName"

	EventResource = "Event"
)

var needTimeFormatList = [...]string{createTimeTag, eventTimeTag}
var conditionTagList = [...]string{
	idTag, envTag, kindTag, levelTag, componentTag, typeTag, clusterIDTag,
	"extraInfo.name", "extraInfo.namespace", "extraInfo.kind"}
var eventFeatTags = []string{idTag, envTag, kindTag, levelTag, componentTag, typeTag,
	clusterIDTag, nameSpaceTag, resourceTypeTag, resourceKindTag, resourceNameTag}

var eventIndexKeys = []string{"data.metadata.name", "data.metadata.resourceVersion"}

// Use Mongodb for storage.
const dbConfig = "mongodb/event"

// PutEvent put event
func PutEvent(req *restful.Request, resp *restful.Response) {
	const (
		handler = "PutEvent"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	if err := insert(req); err != nil {
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

// ListEvent list event
func ListEvent(req *restful.Request, resp *restful.Response) {
	const (
		handler = "ListEvent"
	)
	span := v1http.SetHTTPSpanContextInfo(req, handler)
	defer span.Finish()

	r, total, err := listEvent(req)
	extra := map[string]interface{}{"total": total}
	if err != nil {
		utils.SetSpanLogTagError(span, err)
		blog.Errorf("%s | err: %v", common.BcsErrStorageListResourceFailStr, err)
		lib.ReturnRest(&lib.RestResponse{
			Resp: resp, Data: []string{},
			ErrCode: common.BcsErrStorageListResourceFail,
			Message: common.BcsErrStorageListResourceFailStr, Extra: extra})
		return
	}
	lib.ReturnRest(&lib.RestResponse{Resp: resp, Data: r, Extra: extra})
}

// WatchEvent watch event
func WatchEvent(req *restful.Request, resp *restful.Response) {
	watch(req, resp)
}

// CleanEvents clean event
func CleanEvents() {
	maxCap := apiserver.GetAPIResource().Conf.EventMaxCap
	maxTime := apiserver.GetAPIResource().Conf.EventMaxTime
	eventDBClient := apiserver.GetAPIResource().GetDBClient(dbConfig)
	tables, err := eventDBClient.ListTableNames(context.TODO())
	if err != nil {
		blog.Errorf("list table name failed, err: %v", err)
		return
	}
	for _, table := range tables {
		if strings.HasPrefix(table, tableName) {
			cleaner := clean.NewDBCleaner(eventDBClient, table, time.Hour)
			cleaner.WithMaxEntryNum(maxCap)
			cleaner.WithMaxDuration(time.Duration(maxTime*24)*time.Hour, createTimeTag)
			go cleaner.Run(context.TODO())
		}
	}
}

func init() {
	eventPath := urlPath("/events")
	actions.RegisterV1Action(actions.Action{
		Verb: "PUT", Path: eventPath, Params: nil, Handler: lib.MarkProcess(PutEvent)})
	actions.RegisterV1Action(actions.Action{
		Verb: "GET", Path: eventPath, Params: nil, Handler: lib.MarkProcess(ListEvent)})

	eventWatchPath := urlPath("/events/watch")
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: eventWatchPath, Params: nil, Handler: lib.MarkProcess(WatchEvent)})

	actions.RegisterDaemonFunc(CleanEvents)
}
