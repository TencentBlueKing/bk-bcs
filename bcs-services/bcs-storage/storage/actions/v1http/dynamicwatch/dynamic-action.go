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
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
)

const (
	resourceTypeTag = "resourceType"
	tableTag        = resourceTypeTag
)

// Use Mongodb for storage.
const dbConfig = "mongodb/dynamic"

// WatchDynamic watch dynamic data
func WatchDynamic(req *restful.Request, resp *restful.Response) {
	request := newReqDynamic(req, resp)
	request.watch()
}

// WatchContainer watch container data
func WatchContainer(req *restful.Request, resp *restful.Response) {
	request := newReqDynamic(req, resp)
	request.watchContainer()
}

func init() {
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: "/dynamic/watch/{clusterId}/{resourceType}",
		Params: nil, Handler: WatchDynamic})
	actions.RegisterV1Action(actions.Action{
		Verb: "POST", Path: "/dynamic/watch/containers/{clusterId}",
		Params: nil, Handler: WatchContainer})
}
