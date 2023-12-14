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

// Package view provides a simple way to create HTTP handlers that implement the http.Handler interface.
package view

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
)

// GenericFunc View函数类型, 使用view.GenericFunc(customHandler)
type GenericFunc func(r *http.Request) (interface{}, error)

// ServeHTTP handler 函数实现
func (h GenericFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := h(r)
	// handle returned error here.
	if err != nil {
		ww, ok := w.(*GenericResponseWriter)
		if ok {
			ww.SetError(err)
		}

		if errors.Is(err, errf.ErrPermissionDenied) {
			render.Render(w, r, rest.PermissionDenied(err, nil))
		} else {
			render.Render(w, r, rest.BadRequest(err))
		}
		return
	}

	//  返回的 data 可能为空, 不能序列化
	if data == nil {
		return
	}

	switch v := data.(type) {
	case render.Renderer:
		render.Render(w, r, v)
	default:
		render.JSON(w, r, v)
	}
}
