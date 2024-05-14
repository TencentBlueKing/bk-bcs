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

package middleware

import (
	"fmt"

	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/constant"
	blog "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
)

// ProjectFilter parse project code from path params
func ProjectFilter(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
	projectParam := request.PathParameter("project_code")
	if projectParam == "" {
		chain.ProcessFilter(request, response)
		return
	}
	project, err := component.GetProjectWithCache(request.Request.Context(), projectParam)
	if err != nil {
		blog.Log(request.Request.Context()).Errorf("get project %s failed, err %s", projectParam, err.Error())
		utils.ResponseParamsError(response, fmt.Errorf("project %s not found", projectParam))
		return
	}
	request.SetAttribute(constant.ProjectAttr, project)
	chain.ProcessFilter(request, response)
}
