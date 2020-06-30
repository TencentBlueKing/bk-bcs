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

package metric

import (
	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// JSONStatus json status
type JSONStatus struct {
	hFunc restful.RouteFunction
}

// NewJSONStatus new json status object
func NewJSONStatus(hf restful.RouteFunction) Resource {
	return &JSONStatus{
		hFunc: hf,
	}
}

// Register implements metric.Resource interface
func (js *JSONStatus) Register(container *restful.Container) {
	blog.Infof("register json status resource to metric")
	ws := new(restful.WebService)
	ws.Path("/status").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)
	ws.Route(ws.GET("/").To(js.hFunc))
	container.Add(ws)
}
