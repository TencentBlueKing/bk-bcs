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
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
)

func (c *RequestInfo) AbortWithError(err error) {
	AbortWithError(c.Writer, err)
}

func AbortWithError(rw http.ResponseWriter, err error) {
	var status metav1.Status

	switch v := err.(type) {
	case *apierrors.StatusError:
		status = v.ErrStatus
	default:
		status = apierrors.NewBadRequest(err.Error()).ErrStatus
	}

	status.Kind = "Status"
	status.APIVersion = "v1"

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.Header().Set("Cache-Control", "no-cache, no-store")
	rw.WriteHeader(int(status.Code))
	json.NewEncoder(rw).Encode(status)
}

func (c *RequestInfo) Write(obj runtime.Object) {
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-store")
	c.Writer.WriteHeader(http.StatusOK)
	json.NewEncoder(c.Writer).Encode(obj)
}

func (c *RequestInfo) WriteChunk(obj watch.Event, firstChunk bool) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		panic("expected http.ResponseWriter to be an http.Flusher")
	}
	if firstChunk {
		c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Writer.Header().Set("Cache-Control", "no-cache, private")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Writer.WriteHeader(http.StatusOK)
	}
	eJsons, err := json.Marshal(obj) //转换成JSON返回的是byte[]
	if err != nil {
		panic(err)
	}
	eJsonStr := string(eJsons) + "\r\n"
	// 大小写字符串替换，参考：
	// k8s.io/apimachinery@v0.21.3/pkg/apis/meta/v1/watch.go:31
	// k8s.io/apimachinery@v0.21.3/pkg/watch/watch.go:57
	eJsonStr = strings.Replace(eJsonStr, "Type", "type", 1)
	eJsonStr = strings.Replace(eJsonStr, "Object", "object", 1)
	c.Writer.Write([]byte(eJsonStr))
	flusher.Flush()
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
