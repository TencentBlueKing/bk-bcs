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

package lib

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"

	restful "github.com/emicklei/go-restful"
)

// MarkProcess does the following things:
// 1. print log when a request comes in and returns.
// 2. print request body to log.
// 3. flow control.
func MarkProcess(f restful.RouteFunction) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		logTracer := blog.WithID("markProcess", blog.GetTraceFromRequest(req.Request).ID())
		apiConf := apiserver.GetAPIResource().Conf
		entranceTime := time.Now()
		// print request body to log
		var stringBody = "Not Parsed"
		if apiConf.PrintBody {
			body, _ := ioutil.ReadAll(req.Request.Body)
			stringBody = string(body)
			req.Request.Body = ioutil.NopCloser(strings.NewReader(stringBody))
		}
		// print log when a request comes in and returns
		logTracer.Infof("Receive %s %s?%s, body: %s",
			req.Request.Method, req.Request.URL.Path, req.Request.URL.RawQuery, stringBody)
		f(req, resp)
		logTracer.Infof("Return [%d] %s %s", resp.StatusCode(), req.Request.Method, req.Request.URL.Path)
		if apiConf.PrintManager {
			// Count request time
			if req.Request.Method == "GET" {
				go managerGet.Add(time.Since(entranceTime))
			} else {
				go managerSet.Add(time.Since(entranceTime))
			}
		}
	}
}

var (
	managerGet *Manager
	managerSet *Manager
)

func bringUpManagerGet() {
	apiConf := apiserver.GetAPIResource().Conf
	if apiConf.PrintManager {
		managerGet = NewManager(apiConf.WatchTimeSep, "Manager Get")
		managerGet.Start()
	}
}

func bringUpManagerSet() {
	apiConf := apiserver.GetAPIResource().Conf
	if apiConf.PrintManager {
		managerSet = NewManager(apiConf.WatchTimeSep, "Manager Set")
		managerSet.Start()
	}
}

func init() {
	actions.RegisterDaemonFunc(bringUpManagerGet)
	actions.RegisterDaemonFunc(bringUpManagerSet)
}
