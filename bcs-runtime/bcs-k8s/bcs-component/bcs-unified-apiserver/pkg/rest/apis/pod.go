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

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/rest"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type PodInterface interface {
	List(ctx context.Context, namespace string, opts metav1.ListOptions) (*v1.PodList, error)
	ListAsTable(ctx context.Context, namespace string, acceptHeader string, opts metav1.ListOptions) (*metav1.Table, error)
	Get(ctx context.Context, namespace string, name string, opts metav1.GetOptions) (*v1.Pod, error)
	GetAsTable(ctx context.Context, namespace string, name string, acceptHeader string, opts metav1.GetOptions) (*metav1.Table, error)
	Delete(ctx context.Context, namespace string, name string, opts metav1.DeleteOptions) (*v1.Pod, error)
	Watch(ctx context.Context, namespace string, opts metav1.ListOptions) (watch.Interface, error)
}

type PodHandler struct {
	stor PodInterface
}

func NewPodHandler(handler PodInterface) *PodHandler {
	return &PodHandler{stor: handler}
}

func (h *PodHandler) Serve(c *rest.RequestInfo) error {
	var (
		result runtime.Object
		err    error
	)
	switch c.Options.Verb {
	case rest.ListVerb:
		result, err = h.stor.List(c.Request.Context(), c.Namespace, *c.Options.ListOptions)
	case rest.ListAsTableVerb:
		result, err = h.stor.ListAsTable(c.Request.Context(), c.Namespace, c.Options.AcceptHeader, *c.Options.ListOptions)
	case rest.GetVerb:
		result, err = h.stor.Get(c.Request.Context(), c.Namespace, c.Name, *c.Options.GetOptions)
	case rest.GetAsTableVerb:
		result, err = h.stor.GetAsTable(c.Request.Context(), c.Namespace, c.Name, c.Options.AcceptHeader, *c.Options.GetOptions)
	case rest.DeleteVerb:
		result, err = h.stor.Delete(c.Request.Context(), c.Namespace, c.Name, *c.Options.DeleteOptions)
	case rest.WatchVerb:
		watch, err := h.stor.Watch(c.Request.Context(), c.Namespace, *c.Options.ListOptions)
		if err != nil {
			return err
		}
		firstChunk := true
		for event := range watch.ResultChan() {
			err = rest.AddTypeInformationToObject(event.Object)
			if err != nil {
				return err
			}
			c.WriteChunk(event, firstChunk)
			firstChunk = false
		}
		return nil
	default:
		return rest.ErrNotImplemented
	}
	if err != nil {
		return err
	}
	c.Write(result)
	return nil
}
