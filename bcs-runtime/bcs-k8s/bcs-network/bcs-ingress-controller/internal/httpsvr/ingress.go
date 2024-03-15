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

package httpsvr

import (
	"context"

	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func (h *HttpServerClient) listIngress(request *restful.Request, response *restful.Response) {
	ingressList := &networkextensionv1.IngressList{}
	if err := h.Mgr.GetClient().List(context.Background(), ingressList); err != nil {
		blog.Errorf("list ext ingresses failed when collect metrics, err %s", err.Error())
		return
	}
	data := CreateResponseData(nil, "success", ingressList)
	_, _ = response.Write(data)
}
