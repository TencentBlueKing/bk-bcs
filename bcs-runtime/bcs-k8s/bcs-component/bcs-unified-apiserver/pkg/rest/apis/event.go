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
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/rest"
)

// EventInterface Event Handler 需要实现的方法
type EventInterface interface {
	List(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.EventList, error)
	ListAsTable(ctx context.Context, namespace string, acceptHeader string, opts metav1.ListOptions) (*metav1.Table, error)
	Get(ctx context.Context, namespace string, name string, opts metav1.GetOptions) (*v1.Event, error)
	GetAsTable(ctx context.Context, namespace string, name string, acceptHeader string, opts metav1.GetOptions) (*metav1.Table, error)
}

// EventHandler
type EventHandler struct {
	handler EventInterface
}

// NewEventHandler
func NewEventHandler(handler EventInterface) *EventHandler {
	return &EventHandler{handler: handler}
}

// Event Resource Verb Handler
func (h *EventHandler) Serve(c *rest.RequestContext) error {
	var (
		obj runtime.Object
		err error
	)
	ctx := c.Request.Context()

	switch c.Options.Verb {
	case rest.ListVerb:
		obj, err = h.handler.List(ctx, c.Namespace, *c.Options.ListOptions)
	case rest.ListAsTableVerb:
		obj, err = h.handler.ListAsTable(ctx, c.Namespace, c.Options.AcceptHeader, *c.Options.ListOptions)
	case rest.GetVerb:
		obj, err = h.handler.Get(ctx, c.Namespace, c.Name, *c.Options.GetOptions)
	case rest.GetAsTableVerb:
		obj, err = h.handler.GetAsTable(ctx, c.Namespace, c.Name, c.Options.AcceptHeader, *c.Options.GetOptions)
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
