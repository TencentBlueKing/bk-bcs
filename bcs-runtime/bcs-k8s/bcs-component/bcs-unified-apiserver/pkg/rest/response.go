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

package rest

import (
	"fmt"
	"net/http"

	v1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes/scheme"
)

func (c *RequestInfo) AbortWithError(err error) {
	if c.TableReq.IsTable {
		c.abortWithErrorTable(err)
		return
	}

	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store")
	result := apierrors.NewNotFound(v1.Resource("secrets"), c.Request.URL.Path)
	c.Writer.WriteHeader(int(result.ErrStatus.Code))
	json.NewEncoder(c.Writer).Encode(result)
}

func AbortWithError(rw http.ResponseWriter, err error) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Cache-Control", "no-cache, no-store")
	result := apierrors.NewNotFound(v1.Resource("secrets"), "")
	rw.WriteHeader(int(result.ErrStatus.Code))
	json.NewEncoder(rw).Encode(result)
}

func AbortWithErrorTable(rw http.ResponseWriter, err error) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Cache-Control", "no-cache, no-store")
	result := apierrors.NewNotFound(v1.Resource("secrets"), "")
	rw.WriteHeader(int(result.ErrStatus.Code))
	json.NewEncoder(rw).Encode(result)
}

func (c *RequestInfo) abortWithErrorTable(err error) {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store")
	result := apierrors.NewNotFound(v1.Resource("secrets"), c.Request.URL.Path)
	c.Writer.WriteHeader(int(result.ErrStatus.Code))
	json.NewEncoder(c.Writer).Encode(result)
}

func (c *RequestInfo) Write(obj runtime.Object) {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store")
	c.Writer.WriteHeader(http.StatusOK)
	json.NewEncoder(c.Writer).Encode(obj)
}

func (c *RequestInfo) Serve() {

}

func AddTypeInformationToObject(obj runtime.Object) error {
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		return fmt.Errorf("missing apiVersion or kind and cannot assign it; %w", err)
	}

	for _, gvk := range gvks {
		if len(gvk.Kind) == 0 {
			continue
		}
		if len(gvk.Version) == 0 || gvk.Version == runtime.APIVersionInternal {
			continue
		}
		obj.GetObjectKind().SetGroupVersionKind(gvk)
		break
	}

	return nil
}
