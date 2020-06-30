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

package api

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	http2 "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"net/http"

	restful "github.com/emicklei/go-restful"
)

type RestResponse struct {
	Resp     *restful.Response
	HTTPCode int

	Data    interface{}
	ErrCode int
	Message string
	Extra   map[string]interface{}

	WrapFunc func([]byte) []byte
}

func ReturnRest(resp *RestResponse) {
	if resp.HTTPCode == 0 {
		resp.HTTPCode = http.StatusOK
	}
	if resp.ErrCode == 0 && resp.Message == "" {
		resp.Message = common.BcsSuccessStr
	}
	if resp.Data == nil {
		resp.Data = map[string]interface{}{}
	}
	result, err := http2.GetResponseEx(resp.ErrCode, resp.Message, resp.Data, resp.Extra)
	if err != nil {
		blog.Errorf("%s | err: %v", common.BcsErrCommJsonDecode, err)
		resp.HTTPCode = http.StatusOK
		resp.Data = nil
		resp.ErrCode = common.BcsErrCommJsonDecode
		resp.Message = common.BcsErrCommJsonDecodeStr
		ReturnRest(resp)
		return
	}
	if resp.WrapFunc != nil {
		result = resp.WrapFunc(result)
	}
	resp.Resp.WriteHeader(resp.HTTPCode)
	resp.Resp.Write(result)
}
