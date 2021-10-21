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
	"github.com/Tencent/bk-bcs/bcs-common/common/version"

	restful "github.com/emicklei/go-restful"
)

type VersionMetricResp struct {
	Version   string `json:"version"`
	Tag       string `json:"tag"`
	BuildTime string `json:"buildtime"`
	GitHash   string `json:"githash"`
}

type VersionMetric struct {
}

func NewVersionMetric() Resource {
	return &VersionMetric{}
}

func (vm *VersionMetric) Register(container *restful.Container) {
	// 创建webservice
	ws := new(restful.WebService)
	//指定路径以及支持的媒体类型
	ws.Path("/version").
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON, restful.MIME_XML)
	ws.Route(ws.GET("/").To(vm.getVersion))
	//创建container 不创建则使用restful.Add添加到DefaultContainer
	container.Add(ws)
}

func (vm *VersionMetric) getVersion(req *restful.Request, resp *restful.Response) {
	newResp := VersionMetricResp{
		Version:   version.BcsVersion,
		Tag:       version.BcsTag,
		BuildTime: version.BcsBuildTime,
		GitHash:   version.BcsGitHash,
	}
	resp.WriteEntity(newResp)
}
