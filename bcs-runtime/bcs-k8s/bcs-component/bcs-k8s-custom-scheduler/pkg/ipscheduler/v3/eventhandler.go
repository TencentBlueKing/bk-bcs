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

package v3

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

type ipHandler struct {
	cache *PoolCache
}

func newIPHander(cache *PoolCache) *ipHandler {
	return &ipHandler{
		cache: cache,
	}
}

// OnAdd implements event handler
func (ih *ipHandler) OnAdd(obj interface{}) {
	addUnstruct, ok := obj.(*unstructured.Unstructured)
	if !ok {
		blog.Warnf("added object %v is not unstructured", obj)
		return
	}
	ip := &BCSNetIP{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		addUnstruct.UnstructuredContent(), ip); err != nil {
		blog.Warnf("failed to convert unstructured ip %v", addUnstruct)
		return
	}
	syncCachedPoolByIP(ip)
}

// OnUpdate implements event handler
func (ih *ipHandler) OnUpdate(oldObj, obj interface{}) {
	newUnstruct, ok := obj.(*unstructured.Unstructured)
	if !ok {
		blog.Warnf("new object %v is not unstructured", obj)
		return
	}
	ip := &BCSNetIP{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		newUnstruct.UnstructuredContent(), ip); err != nil {
		blog.Warnf("failed to convert unstructured ip %v", newUnstruct)
		return
	}
	syncCachedPoolByIP(ip)
}

// OnDelete implements event handler
func (ih *ipHandler) OnDelete(obj interface{}) {
	var ok bool
	var ipStruct *unstructured.Unstructured
	var tombstone cache.DeletedFinalStateUnknown
	ipStruct, ok = obj.(*unstructured.Unstructured)
	if !ok {
		tombstone, ok = obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			blog.Warnf("error decoding object, invalid type")
			return
		}
		ipStruct, ok = tombstone.Obj.(*unstructured.Unstructured)
		if !ok {
			blog.Warnf("error decoding object tombstone, invalid type")
			return
		}
		blog.Infof("recovered deleted ip '%s/%s' from tombstone", ipStruct.GetName())
	}
	ip := &BCSNetIP{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		ipStruct.UnstructuredContent(), ip); err != nil {
		blog.Warnf("failed to convert unstructured ip %v", ipStruct)
		return
	}
	syncCachedPoolByIP(ip)
}
