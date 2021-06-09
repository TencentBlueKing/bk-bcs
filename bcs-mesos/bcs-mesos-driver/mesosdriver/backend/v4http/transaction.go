/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v4http

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"

	restful "github.com/emicklei/go-restful"
)

func (s *Scheduler) listTransactionHandler(req *restful.Request, resp *restful.Response) {
	objKind := req.QueryParameter("objKind")
	objName := req.QueryParameter("objName")
	namespace := req.PathParameter("ns")
	url := fmt.Sprintf("%s/v1/transactions/%s?objKind=%s&objName=%s",
		s.GetHost(), namespace, objKind, objName)
	reply, err := s.client.GET(url, nil, nil)
	if err != nil {
		blog.Errorf("list transaction to url (%s) failed, err (%s)", url, err.Error())
		err = bhttp.InternalError(common.BcsErrCommRequestDataErr, err.Error())
		resp.Write([]byte(err.Error()))
		return
	}
	resp.Write([]byte(reply))
}

func (s *Scheduler) deleteTransactionHandler(req *restful.Request, resp *restful.Response) {
	// get namespace
	namespace := req.PathParameter("ns")
	// get name
	name := req.PathParameter("name")
	url := fmt.Sprintf("%s/v1/transactions/%s/%s", s.GetHost(), namespace, name)
	reply, err := s.client.DELETE(url, nil, nil)
	if err != nil {
		blog.Errorf("failed to delete transaction namespace %s name %s", namespace, name)
		err = bhttp.InternalError(common.BcsErrCommHttpDo, common.BcsErrCommHttpDoStr+err.Error())
		resp.Write([]byte(err.Error()))
		return
	}
	resp.Write([]byte(reply))
}
