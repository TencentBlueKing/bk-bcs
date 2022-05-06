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
	"encoding/json"
	"io/ioutil"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	types "k8s.io/apimachinery/pkg/types"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/rest"
)

// StatefulSetInterface StatefulSet Handler 需要实现的方法
type StatefulSetInterface interface {
	List(ctx context.Context, namespace string, opts metav1.ListOptions) (*appsv1.StatefulSetList, error)
	ListAsTable(ctx context.Context, namespace string, acceptHeader string, opts metav1.ListOptions) (*metav1.Table, error)
	Get(ctx context.Context, namespace string, name string, opts metav1.GetOptions) (*appsv1.StatefulSet, error)
	GetAsTable(ctx context.Context, namespace string, name string, acceptHeader string, opts metav1.GetOptions) (*metav1.Table, error)
	Create(ctx context.Context, namespace string, deployment *appsv1.StatefulSet, opts metav1.CreateOptions) (*appsv1.StatefulSet, error)
	Update(ctx context.Context, namespace string, deployment *appsv1.StatefulSet, opts metav1.UpdateOptions) (*appsv1.StatefulSet, error)
	Patch(ctx context.Context, namespace string, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*appsv1.StatefulSet, error)
	Delete(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) (*metav1.Status, error)
}

// StatefulSetHandler
type StatefulSetHandler struct {
	handler StatefulSetInterface
}

// NewStatefulSetHandler
func NewStatefulSetHandler(handler StatefulSetInterface) *StatefulSetHandler {
	return &StatefulSetHandler{handler: handler}
}

// StatefulSet Resource Verb Handler
func (h *StatefulSetHandler) Serve(c *rest.RequestContext) error {
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
	case rest.CreateVerb: // kubectl create 操作
		newObj := appsv1.StatefulSet{}
		if decodeErr := json.NewDecoder(c.Request.Body).Decode(&newObj); decodeErr != nil {
			return decodeErr
		}
		obj, err = h.handler.Create(ctx, c.Namespace, &newObj, *c.Options.CreateOptions)
	case rest.UpdateVerb: // kubectl replace 操作
		newObj := appsv1.StatefulSet{}
		if decodeErr := json.NewDecoder(c.Request.Body).Decode(&newObj); decodeErr != nil {
			return decodeErr
		}
		obj, err = h.handler.Update(ctx, c.Namespace, &newObj, *c.Options.UpdateOptions)
	case rest.PatchVerb: // kubectl edit/apply 操作
		data, rErr := ioutil.ReadAll(c.Request.Body)
		if rErr != nil {
			return rErr
		}
		obj, err = h.handler.Patch(ctx, c.Namespace, c.Name, c.Options.PatchType, data, *c.Options.PatchOptions, c.Subresource)
	case rest.DeleteVerb: // kubectl delete 操作
		obj, err = h.handler.Delete(ctx, c.Namespace, c.Name, *c.Options.DeleteOptions)
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
