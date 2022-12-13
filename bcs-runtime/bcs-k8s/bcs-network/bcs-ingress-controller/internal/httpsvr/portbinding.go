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

package httpsvr

import (
	"context"
	"fmt"

	"github.com/emicklei/go-restful"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8sapitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func (h *HttpServerClient) listPortbinding(request *restful.Request, response *restful.Response) {
	var data string
	portPool := &networkextensionv1.PortPool{}
	if err := h.Mgr.GetClient().Get(context.Background(), k8sapitypes.NamespacedName{
		Name:      request.PathParameter("name"),
		Namespace: request.PathParameter("namespace"),
	}, portPool); err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Infof("portpool %s/%s not found", request.PathParameter("namespace"), request.PathParameter("name"))
			data = CreateResponseData(err, "failed", nil)
		}
	}

	portBindingList := &networkextensionv1.PortBindingList{}
	labelKey := fmt.Sprintf(networkextensionv1.PortPoolBindingLabelKeyFromat, portPool.GetName(), portPool.GetNamespace())
	if err := h.Mgr.GetClient().List(context.Background(), portBindingList,
		client.MatchingLabels{labelKey: portPool.Name}); err != nil {
		blog.Error("list portBinding with label['%s'='%s'] failed",
			labelKey, portPool.Name)
	}

	data = CreateResponseData(nil, "success", portBindingList)
	_, _ = response.Write([]byte(data))
}
