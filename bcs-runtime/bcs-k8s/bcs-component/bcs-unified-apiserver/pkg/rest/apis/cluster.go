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

package apis

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/rest"
)

// ClusterInterface 集群 API 元信息 需要实现的方法
type ClusterInterface interface {
	GetServerGroups() (*metav1.APIGroupList, error) // GetServerGroups /apis 返回
}

// ClusterHandler
type ClusterHandler struct {
	handler ClusterInterface
}

// NewClusterHandler
func NewClusterHandler(handler ClusterInterface) *ClusterHandler {
	return &ClusterHandler{handler: handler}
}

// Service Resource Verb Handler
func (h *ClusterHandler) Serve(c *rest.RequestContext) error {
	var (
		obj runtime.Object
		err error
	)
	switch c.Options.Verb {
	case rest.GetVerb:
		if c.Path == "/apis" {
			obj, err = h.handler.GetServerGroups()
		}

	default:
		// 未实现的功能
		return rest.ErrNotImplemented
	}

	if err != nil {
		return err
	}
	rest.AddTypeInformationToObject(obj)
	c.Write(obj)
	return nil
}
