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

package cluster

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/utils"
	"github.com/emicklei/go-restful/v3"
	"github.com/opentracing/opentracing-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/lib"
)

// ReturnError 统一的错误处理函数
func ReturnError(handler string, span opentracing.Span, resp *restful.Response, errCode int, errMsg string, err error) {
	utils.SetSpanLogTagError(span, err)
	blog.Errorf("%s | %s: %v", handler, errMsg, err)
	lib.ReturnRest(&lib.RestResponse{
		Resp:    resp,
		Data:    nil,
		ErrCode: errCode,
		Message: errMsg,
	})
}
