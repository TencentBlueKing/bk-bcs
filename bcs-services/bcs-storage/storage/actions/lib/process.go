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

package lib

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	restful "github.com/emicklei/go-restful/v3"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
)

// MarkProcess does the following things:
// 1. print log when a request comes in and returns.
// 2. print request body to log.
// 3. flow control.
func MarkProcess(f restful.RouteFunction) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		entranceTime := time.Now()
		apiConf := apiserver.GetAPIResource().Conf
		// print request body to log
		var stringBody = "Not Parsed"
		if apiConf.PrintBody {
			body, _ := ioutil.ReadAll(req.Request.Body)
			stringBody = string(body)
			req.Request.Body = ioutil.NopCloser(strings.NewReader(stringBody))
		}
		// print log when a request comes in and returns
		method := req.Request.Method
		path := req.Request.URL.Path
		blog.Infof("Receive %s %s?%s, body: %s", method, path, req.Request.URL.RawQuery, stringBody)
		f(req, resp)
		blog.Infof("Return [%d] %s %s", resp.StatusCode(), method, path)
		if apiConf.PrintManager {
			// Count request time
			if method == "GET" {
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
